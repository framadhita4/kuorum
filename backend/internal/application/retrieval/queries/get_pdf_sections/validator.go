package get_pdf_sections

import (
	"errors"
	"strings"
)

func (q GetPdfSectionsQuery) Validate() error {
	if strings.TrimSpace(q.PaperID) == "" {
		return errors.New("paperId cannot be empty")
	}
	return nil
}
