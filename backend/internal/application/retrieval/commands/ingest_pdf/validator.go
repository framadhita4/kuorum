package ingest_pdf

import (
	"errors"
	"path/filepath"
	"strings"

	"kuorum/backend/internal/domain/retrieval"
)

func (c IngestPdfCommand) Validate() error {
	if c.FilePath == "" && len(c.FileBytes) == 0 && c.DownloadURL == "" && c.Metadata == nil && c.PaperID == "" {
		return errors.New("either filePath, fileBytes, downloadUrl, metadata, or paperId must be provided")
	}
	if c.FileName != "" {
		ext := strings.ToLower(filepath.Ext(c.FileName))
		if ext != ".pdf" {
			return errors.New("file must have a .pdf extension")
		}
	}
	if c.PaperID != "" {
		if _, err := retrieval.ParsePaperID(c.PaperID); err != nil {
			return err
		}
	}
	if len(c.FileBytes) > 50*1024*1024 { // 50 MB
		return errors.New("file size exceeds maximum limit of 50MB")
	}
	return nil
}
