package pdf

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type PopplerPdfParser struct {
	pdftotextBin string
	pdftohtmlBin string
	pdfinfoBin   string
}

// NewPopplerPdfParser constructs a high-fidelity PDF parser using Poppler utilities.
func NewPopplerPdfParser() interfaces.IPdfParser {
	pdftotext := "pdftotext"
	if custom := os.Getenv("PDFTOTEXT_BIN"); custom != "" {
		pdftotext = custom
	}

	pdftohtml := "pdftohtml"
	if custom := os.Getenv("PDFTOHTML_BIN"); custom != "" {
		pdftohtml = custom
	}

	pdfinfo := "pdfinfo"
	if custom := os.Getenv("PDFINFO_BIN"); custom != "" {
		pdfinfo = custom
	}

	return &PopplerPdfParser{
		pdftotextBin: pdftotext,
		pdftohtmlBin: pdftohtml,
		pdfinfoBin:   pdfinfo,
	}
}

func (p *PopplerPdfParser) Parse(ctx context.Context, reader io.Reader, filename string) (*retrieval.PdfDocument, error) {
	if reader == nil {
		return nil, retrieval.ErrInvalidPdf
	}

	tmpFile, err := os.CreateTemp("", "kuorum_upload_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("creating temp file failed: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("writing PDF to temp file failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("closing temp file failed: %w", err)
	}

	doc, err := p.ParseFile(ctx, tmpPath)
	if err != nil {
		return nil, err
	}

	if filename != "" {
		doc.FileName = filename
		if parsedID, err := retrieval.ParsePaperID(filename); err == nil && parsedID != "" {
			doc.ID = parsedID
		}
	}
	return doc, nil
}

func (p *PopplerPdfParser) ParseFile(ctx context.Context, filePath string) (*retrieval.PdfDocument, error) {
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("%w: %v", retrieval.ErrInvalidPdf, err)
	}

	// 1. Run pdfinfo to inspect encryption, page count, and metadata
	meta, err := p.extractPdfInfo(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("pdfinfo extraction failed: %w", err)
	}

	if meta.encrypted {
		return nil, retrieval.ErrEncryptedPdf
	}

	// 2. Extract pages using flow-aware layout extraction (pdftohtml XML or pdftotext fallback)
	pages, totalRawLength, err := p.extractPages(ctx, filePath)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, retrieval.ErrEmptyPdf
	}

	// 3. Sectionize pages into academic sections
	sections := SectionizePages(pages)

	// 4. Extract bibliography references from REFERENCES section
	var references []retrieval.ExtractedReference
	var abstractText string

	for i := range sections {
		if sections[i].Type == retrieval.SectionReferences {
			references = ParseReferences(sections[i].Content)
		}
		if sections[i].Type == retrieval.SectionAbstract && abstractText == "" {
			abstractText = sections[i].Content
		}
	}

	// 5. Build document metadata
	title := meta.title
	authors := meta.authors
	if title == "" && len(sections) > 0 {
		// Heuristic: Extract first non-empty line from preamble as title
		firstSec := sections[0]
		if len(firstSec.Paragraphs) > 0 {
			lines := strings.Split(firstSec.Paragraphs[0], "\n")
			if len(lines) > 0 {
				title = strings.TrimSpace(lines[0])
			}
		}
	}
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	totalPages := meta.pages
	if totalPages == 0 {
		totalPages = len(pages)
	}

	// 6. Look for canonical DOI in preamble or article info sections
	detectedDOI := ""
	for _, sec := range sections {
		if sec.Type == retrieval.SectionOther || sec.Type == retrieval.SectionAbstract {
			if matches := doiRegex.FindStringSubmatch(sec.Content); len(matches) > 1 {
				detectedDOI = strings.TrimRight(matches[1], ".,;)")
				break
			}
		}
	}

	doc := &retrieval.PdfDocument{
		ID:            generateDocID(title),
		FileName:      filepath.Base(filePath),
		Title:         title,
		Authors:       authors,
		Abstract:      abstractText,
		DOI:           detectedDOI,
		TotalPages:    totalPages,
		Sections:      sections,
		References:    references,
		RawTextLength: totalRawLength,
		ParsedAt:      time.Now().UTC(),
	}

	return doc, nil
}

// extractPages attempts flow-based layout extraction with pdftohtml first, then falls back to pdftotext.
func (p *PopplerPdfParser) extractPages(ctx context.Context, filePath string) ([]rawPage, int, error) {
	// 1. Try flow-aware XML extraction via pdftohtml
	pages, rawLen, err := p.extractPagesWithPdftohtml(ctx, filePath)
	if err == nil && len(pages) > 0 {
		return pages, rawLen, nil
	}

	// 2. Fall back to pdftotext
	return p.extractPagesWithPdftotext(ctx, filePath)
}

func (p *PopplerPdfParser) extractPagesWithPdftohtml(ctx context.Context, filePath string) ([]rawPage, int, error) {
	cmd := exec.CommandContext(ctx, p.pdftohtmlBin, "-xml", "-i", "-stdout", filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("pdftohtml failed: %w", err)
	}

	var doc pdf2Xml
	if err := xml.Unmarshal(stdout.Bytes(), &doc); err != nil {
		return nil, 0, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	if len(doc.Pages) == 0 {
		return nil, 0, fmt.Errorf("no pages found in pdftohtml output")
	}

	var pages []rawPage
	totalLen := 0

	for _, page := range doc.Pages {
		assembled := assemblePageFromXmlTexts(page.Texts)
		cleaned := CleanPageText(assembled, page.Number)
		if strings.TrimSpace(cleaned) != "" {
			pages = append(pages, rawPage{
				pageNumber: page.Number,
				text:       cleaned,
			})
			totalLen += len(cleaned)
		}
	}

	return pages, totalLen, nil
}

func (p *PopplerPdfParser) extractPagesWithPdftotext(ctx context.Context, filePath string) ([]rawPage, int, error) {
	cmd := exec.CommandContext(ctx, p.pdftotextBin, "-remove-hyphens", "all", filePath, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("%w: pdftotext failed: %s (%v)", retrieval.ErrParsingFailed, stderr.String(), err)
	}

	rawOutput := stdout.String()
	if strings.TrimSpace(rawOutput) == "" {
		return nil, 0, retrieval.ErrEmptyPdf
	}

	rawPageChunks := strings.Split(rawOutput, "\x0c")
	pages := make([]rawPage, 0, len(rawPageChunks))
	totalLen := len(rawOutput)

	for i, chunk := range rawPageChunks {
		cleaned := CleanPageText(chunk, i+1)
		if strings.TrimSpace(cleaned) != "" {
			pages = append(pages, rawPage{
				pageNumber: i + 1,
				text:       cleaned,
			})
		}
	}

	return pages, totalLen, nil
}

type pdf2Xml struct {
	Pages []xmlPage `xml:"page"`
}

type xmlPage struct {
	Number int       `xml:"number,attr"`
	Texts  []xmlText `xml:"text"`
}

type xmlText struct {
	Top     int    `xml:"top,attr"`
	Left    int    `xml:"left,attr"`
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Content string `xml:",innerxml"`
}

var htmlTagRegex = regexp.MustCompile(`<[^>]+>`)

func stripHtmlTags(s string) string {
	clean := htmlTagRegex.ReplaceAllString(s, "")
	clean = strings.ReplaceAll(clean, "&#160;", " ")
	clean = strings.ReplaceAll(clean, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = strings.ReplaceAll(clean, "&quot;", "\"")
	return strings.TrimSpace(clean)
}

func assemblePageFromXmlTexts(texts []xmlText) string {
	if len(texts) == 0 {
		return ""
	}

	var paragraphs []string
	var currentPara strings.Builder
	var lastText *xmlText

	for i := 0; i < len(texts); i++ {
		t := texts[i]
		clean := stripHtmlTags(t.Content)
		if clean == "" {
			continue
		}

		if lastText == nil {
			currentPara.WriteString(clean)
			lastText = &t
			continue
		}

		// Check if on same line (vertical delta <= 4)
		isSameLine := math.Abs(float64(t.Top-lastText.Top)) <= 4
		lineGap := t.Top - (lastText.Top + lastText.Height)

		// Check if new paragraph / block boundary
		isNewBlock := !isSameLine && (lineGap > 8 || math.Abs(float64(t.Left-lastText.Left)) > 50)

		if isNewBlock {
			if currentPara.Len() > 0 {
				paragraphs = append(paragraphs, strings.TrimSpace(currentPara.String()))
				currentPara.Reset()
			}
			currentPara.WriteString(clean)
		} else if isSameLine {
			// Append on same line with space
			currentPara.WriteString(" ")
			currentPara.WriteString(clean)
		} else {
			// Wrapped line in same paragraph
			currStr := currentPara.String()
			if strings.HasSuffix(currStr, "-") {
				// Dehyphenate: remove hyphen and merge directly
				trimmed := strings.TrimSuffix(currStr, "-")
				currentPara.Reset()
				currentPara.WriteString(trimmed)
				currentPara.WriteString(clean)
			} else {
				currentPara.WriteString(" ")
				currentPara.WriteString(clean)
			}
		}

		lastText = &t
	}

	if currentPara.Len() > 0 {
		paragraphs = append(paragraphs, strings.TrimSpace(currentPara.String()))
	}

	return strings.Join(paragraphs, "\n\n")
}

type pdfMetadata struct {
	title     string
	authors   []string
	pages     int
	encrypted bool
}

func (p *PopplerPdfParser) extractPdfInfo(ctx context.Context, filePath string) (*pdfMetadata, error) {
	cmd := exec.CommandContext(ctx, p.pdfinfoBin, filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &pdfMetadata{}, nil
	}

	meta := &pdfMetadata{}
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Title":
			meta.title = val
		case "Author":
			if val != "" {
				meta.authors = splitAuthors(val)
			}
		case "Pages":
			if p, err := strconv.Atoi(val); err == nil {
				meta.pages = p
			}
		case "Encrypted":
			if strings.HasPrefix(strings.ToLower(val), "yes") {
				meta.encrypted = true
			}
		}
	}

	return meta, nil
}

func generateDocID(title string) string {
	clean := strings.ToLower(title)
	clean = strings.ReplaceAll(clean, " ", "-")
	var b strings.Builder
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	res := b.String()
	if len(res) > 50 {
		res = res[:50]
	}
	if res == "" {
		res = fmt.Sprintf("doc-%d", time.Now().Unix())
	}
	return res
}
