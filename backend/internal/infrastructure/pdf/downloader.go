package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

var (
	ErrNoDownloadURL     = errors.New("paper has no downloadable open access PDF URL")
	ErrInvalidPDFContent = errors.New("downloaded payload is not a valid PDF document (magic bytes missing)")
	ErrDownloadFailed    = errors.New("failed to download PDF from academic provider")
	ErrDownloadTimeout   = errors.New("PDF download request timed out")
)

// DownloaderConfig provides tuneable options for paper PDF fetching.
type DownloaderConfig struct {
	Timeout       time.Duration
	UserAgent     string
	MaxFileSizeMB int64
}

// DefaultDownloaderConfig defines polite and safe academic retrieval defaults.
func DefaultDownloaderConfig() DownloaderConfig {
	return DownloaderConfig{
		Timeout:       30 * time.Second,
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 (Academic Research; mailto:contact@research-helper.local)",
		MaxFileSizeMB: 50, // 50MB max file size to prevent memory exhaustion
	}
}

// PaperDownloader implements interfaces.IPdfDownloader with polite HTTP streaming and magic-byte checks.
type PaperDownloader struct {
	storage    interfaces.IPdfStorage
	httpClient *http.Client
	config     DownloaderConfig
}

// NewPaperDownloader constructs an instance of PaperDownloader.
func NewPaperDownloader(storage interfaces.IPdfStorage, client *http.Client, config ...DownloaderConfig) *PaperDownloader {
	cfg := DefaultDownloaderConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if client == nil {
		client = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &PaperDownloader{
		storage:    storage,
		httpClient: client,
		config:     cfg,
	}
}

// ResolveDownloadURL extracts or synthesizes an open access PDF URL from paper metadata.
func (d *PaperDownloader) ResolveDownloadURL(paper retrieval.PaperMetadata) string {
	if strings.TrimSpace(paper.OpenAccessPdfURL) != "" {
		return strings.TrimSpace(paper.OpenAccessPdfURL)
	}

	// arXiv DOI resolution: 10.48550/arXiv.2301.0001 -> https://arxiv.org/pdf/2301.0001.pdf
	doiLower := strings.ToLower(paper.Doi)
	if strings.Contains(doiLower, "arxiv.") {
		parts := strings.Split(doiLower, "arxiv.")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			return fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", strings.TrimSpace(parts[1]))
		}
	}

	// Direct arXiv URL in ID
	idLower := strings.ToLower(paper.ID)
	if strings.Contains(idLower, "arxiv.org/abs/") {
		parts := strings.Split(idLower, "arxiv.org/abs/")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			return fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", strings.TrimSpace(parts[1]))
		}
	} else if strings.Contains(idLower, "arxiv.org/pdf/") {
		return strings.TrimSpace(paper.ID)
	}

	// arXiv ID from Sources if available
	for _, src := range paper.Sources {
		if strings.HasPrefix(strings.ToLower(src), "arxiv:") {
			id := strings.TrimPrefix(src, "arxiv:")
			id = strings.TrimPrefix(id, "arXiv:")
			return fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", strings.TrimSpace(id))
		}
	}

	return ""
}

// DownloadPdf streams a PDF from the given URL, validates its PDF header, and saves it to storage.
func (d *PaperDownloader) DownloadPdf(ctx context.Context, paperID string, downloadURL string) (string, error) {
	if strings.TrimSpace(downloadURL) == "" {
		return "", ErrNoDownloadURL
	}

	// If already in storage, skip downloading
	if d.storage.Exists(ctx, paperID) {
		return d.storage.GetFilePath(paperID), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating download request failed: %w", err)
	}

	req.Header.Set("User-Agent", d.config.UserAgent)
	req.Header.Set("Accept", "application/pdf,application/octet-stream;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", ErrDownloadTimeout
		}
		return "", fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: provider returned HTTP %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Validate magic bytes (%PDF-) by reading the first 1024 bytes
	headerBuf := make([]byte, 1024)
	n, readErr := io.ReadFull(resp.Body, headerBuf)
	if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("reading PDF header failed: %w", readErr)
	}

	headerBytes := headerBuf[:n]
	if !bytes.Contains(headerBytes, []byte("%PDF-")) {
		return "", fmt.Errorf("%w: response did not start with PDF magic bytes (possibly HTML paywall)", ErrInvalidPDFContent)
	}

	// Combine header buffer and the rest of the body stream
	combinedReader := io.MultiReader(bytes.NewReader(headerBytes), resp.Body)

	// Enforce max file size
	maxBytes := d.config.MaxFileSizeMB * 1024 * 1024
	limitedReader := io.LimitReader(combinedReader, maxBytes)

	savedPath, err := d.storage.Save(ctx, paperID, limitedReader)
	if err != nil {
		return "", fmt.Errorf("saving downloaded PDF to storage failed: %w", err)
	}

	return savedPath, nil
}

// DownloadBatch concurrently fetches PDFs for all papers with available URLs.
func (d *PaperDownloader) DownloadBatch(ctx context.Context, papers []retrieval.PaperMetadata, maxConcurrency int) ([]interfaces.DownloadResult, error) {
	if len(papers) == 0 {
		return nil, nil
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 3
	}

	results := make([]interfaces.DownloadResult, len(papers))
	for i, p := range papers {
		results[i].PaperID = p.ID
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)

	var mu sync.Mutex

	for idx, paper := range papers {
		i := idx
		p := paper

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()

			url := d.ResolveDownloadURL(p)
			if url == "" {
				mu.Lock()
				results[i].Error = ErrNoDownloadURL
				mu.Unlock()
				return nil
			}

			path, err := d.DownloadPdf(gCtx, p.ID, url)
			mu.Lock()
			if err != nil {
				results[i].Error = err
			} else {
				results[i].FilePath = path
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, fmt.Errorf("batch download encountered error: %w", err)
	}

	return results, nil
}
