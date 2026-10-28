package pdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type LocalPdfStorage struct {
	baseDir string
}

// NewLocalPdfStorage creates a disk storage provider for original PDF documents.
func NewLocalPdfStorage(baseDir string) (interfaces.IPdfStorage, error) {
	if baseDir == "" {
		baseDir = "data/papers"
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory %q: %w", baseDir, err)
	}

	return &LocalPdfStorage{baseDir: baseDir}, nil
}

func (s *LocalPdfStorage) GetFilePath(paperID string) string {
	safeID := retrieval.SanitizePaperID(paperID)
	if safeID == "" {
		safeID = filepath.Base(paperID)
	}
	return filepath.Join(s.baseDir, fmt.Sprintf("%s.pdf", safeID))
}

func (s *LocalPdfStorage) Save(ctx context.Context, paperID string, reader io.Reader) (string, error) {
	targetPath := s.GetFilePath(paperID)

	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("creating target file failed: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", fmt.Errorf("writing PDF bytes failed: %w", err)
	}

	return targetPath, nil
}

func (s *LocalPdfStorage) Open(ctx context.Context, paperID string) (io.ReadCloser, error) {
	targetPath := s.GetFilePath(paperID)
	file, err := os.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("opening PDF file %q failed: %w", targetPath, err)
	}
	return file, nil
}

func (s *LocalPdfStorage) Exists(ctx context.Context, paperID string) bool {
	targetPath := s.GetFilePath(paperID)
	info, err := os.Stat(targetPath)
	return err == nil && !info.IsDir()
}
