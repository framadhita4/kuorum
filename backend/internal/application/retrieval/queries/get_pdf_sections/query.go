package get_pdf_sections

type GetPdfSectionsQuery struct {
	PaperID string `json:"paperId"`
	Intent  string `json:"intent,omitempty"`
}

func (q GetPdfSectionsQuery) IsQuery() {}
