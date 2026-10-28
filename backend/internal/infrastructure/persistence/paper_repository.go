package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
	"kuorum/backend/internal/infrastructure/persistence/db"
)

type SqlitePaperRepository struct {
	queries *db.Queries
}

func NewSqlitePaperRepository(database *sql.DB) interfaces.IPaperRepository {
	return &SqlitePaperRepository{
		queries: db.New(database),
	}
}

func (r *SqlitePaperRepository) SavePaper(ctx context.Context, paper *retrieval.PaperMetadata) error {
	if paper == nil {
		return errors.New("cannot save nil paper metadata")
	}

	authorsJSON, err := json.Marshal(paper.Authors)
	if err != nil {
		return fmt.Errorf("failed to marshal authors: %w", err)
	}

	fieldsJSON, err := json.Marshal(paper.FieldsOfStudy)
	if err != nil {
		return fmt.Errorf("failed to marshal fields of study: %w", err)
	}

	fetchedAt := paper.FetchedAt
	if fetchedAt.IsZero() {
		fetchedAt = time.Now().UTC()
	}

	err = r.queries.UpsertPaper(ctx, db.UpsertPaperParams{
		ID:                       paper.ID,
		Title:                    paper.Title,
		Abstract:                 paper.Abstract,
		AuthorsJson:              string(authorsJSON),
		Year:                     int64(paper.Year),
		CitationCount:            int64(paper.CitationCount),
		InfluentialCitationCount: int64(paper.InfluentialCitationCount),
		Tldr:                     paper.TLDR,
		FieldsOfStudyJson:        string(fieldsJSON),
		OpenAccessPdfUrl:         paper.OpenAccessPdfURL,
		Doi:                      paper.Doi,
		Venue:                    paper.Venue,
		FetchedAt:                fetchedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert paper %q: %w", paper.ID, err)
	}

	return nil
}

func (r *SqlitePaperRepository) GetPaper(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	row, err := r.queries.GetPaper(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, retrieval.ErrPaperNotFound
		}
		return nil, fmt.Errorf("failed to get paper %q: %w", id, err)
	}

	return toDomainPaper(row)
}

func (r *SqlitePaperRepository) SearchLocalPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	if limit <= 0 {
		limit = 10
	}

	param := sql.NullString{String: query, Valid: true}
	rows, err := r.queries.SearchPapers(ctx, db.SearchPapersParams{
		Column1: param,
		Column2: param,
		Limit:   int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search local papers: %w", err)
	}

	results := make([]retrieval.PaperMetadata, 0, len(rows))
	for _, row := range rows {
		p, err := toDomainPaper(row)
		if err != nil {
			return nil, err
		}
		results = append(results, *p)
	}

	return results, nil
}

func toDomainPaper(row db.Paper) (*retrieval.PaperMetadata, error) {
	var authors []retrieval.Author
	if row.AuthorsJson != "" {
		if err := json.Unmarshal([]byte(row.AuthorsJson), &authors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal authors for paper %s: %w", row.ID, err)
		}
	}

	var fields []string
	if row.FieldsOfStudyJson != "" {
		if err := json.Unmarshal([]byte(row.FieldsOfStudyJson), &fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields of study for paper %s: %w", row.ID, err)
		}
	}

	return &retrieval.PaperMetadata{
		ID:                       row.ID,
		Title:                    row.Title,
		Abstract:                 row.Abstract,
		Authors:                  authors,
		Year:                     int(row.Year),
		CitationCount:            int(row.CitationCount),
		InfluentialCitationCount: int(row.InfluentialCitationCount),
		TLDR:                     row.Tldr,
		FieldsOfStudy:            fields,
		OpenAccessPdfURL:         row.OpenAccessPdfUrl,
		Doi:                      row.Doi,
		Venue:                    row.Venue,
		FetchedAt:                row.FetchedAt,
	}, nil
}
