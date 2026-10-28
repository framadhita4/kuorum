package retrieval

import (
	"errors"
	"time"
)

var (
	ErrPaperNotFound        = errors.New("paper not found")
	ErrEmptyQuery           = errors.New("search query cannot be empty")
	ErrInvalidCitationCount = errors.New("citation count cannot be negative")
	ErrRateLimited          = errors.New("academic provider rate limit exceeded")
	ErrProviderNotAvailable = errors.New("academic provider not available")
)

type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PaperMetadata struct {
	ID                       string    `json:"id"`
	Title                    string    `json:"title"`
	Abstract                 string    `json:"abstract"`
	Authors                  []Author  `json:"authors"`
	Year                     int       `json:"year"`
	CitationCount            int       `json:"citationCount"`
	InfluentialCitationCount int       `json:"influentialCitationCount"`
	TLDR                     string    `json:"tldr,omitempty"`
	FieldsOfStudy            []string  `json:"fieldsOfStudy,omitempty"`
	OpenAccessPdfURL         string    `json:"openAccessPdfUrl,omitempty"`
	Doi                      string    `json:"doi,omitempty"`
	Venue                    string    `json:"venue,omitempty"`
	FetchedAt                time.Time `json:"fetchedAt"`
	Sources                  []string  `json:"sources,omitempty"`
}
