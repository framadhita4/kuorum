package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/infrastructure/persistence/db"
)

type SqliteExtractionRepository struct {
	queries *db.Queries
}

func NewSqliteExtractionRepository(database *sql.DB) interfaces.IExtractionRepository {
	return &SqliteExtractionRepository{
		queries: db.New(database),
	}
}

func (r *SqliteExtractionRepository) SaveSynthesis(ctx context.Context, s *extraction.GroundedConceptSynthesis) error {
	if s == nil {
		return errors.New("cannot save nil grounded concept synthesis")
	}

	defsJSON, err := json.Marshal(s.Definitions)
	if err != nil {
		return fmt.Errorf("failed to marshal definitions: %w", err)
	}

	mechanismsJSON, err := json.Marshal(s.Mechanisms)
	if err != nil {
		return fmt.Errorf("failed to marshal mechanisms: %w", err)
	}

	casesJSON, err := json.Marshal(s.CaseStudies)
	if err != nil {
		return fmt.Errorf("failed to marshal case studies: %w", err)
	}

	generatedAt := s.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	err = r.queries.UpsertGroundedSynthesis(ctx, db.UpsertGroundedSynthesisParams{
		AnchorPaperID:   s.AnchorPaperID,
		Concept:         s.Concept,
		DefinitionsJson: string(defsJSON),
		MechanismsJson:  string(mechanismsJSON),
		CaseStudiesJson: string(casesJSON),
		GeneratedAt:     generatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert grounded synthesis: %w", err)
	}

	return nil
}

func (r *SqliteExtractionRepository) GetSynthesis(ctx context.Context, anchorPaperID string, concept string) (*extraction.GroundedConceptSynthesis, error) {
	row, err := r.queries.GetGroundedSynthesis(ctx, db.GetGroundedSynthesisParams{
		AnchorPaperID: anchorPaperID,
		Concept:       concept,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, extraction.ErrUngroundedClaim
		}
		return nil, fmt.Errorf("failed to query grounded synthesis: %w", err)
	}

	var defs []extraction.GroundedDefinition
	if row.DefinitionsJson != "" {
		if err := json.Unmarshal([]byte(row.DefinitionsJson), &defs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal definitions: %w", err)
		}
	}

	var mechanisms []extraction.GroundedMechanism
	if row.MechanismsJson != "" {
		if err := json.Unmarshal([]byte(row.MechanismsJson), &mechanisms); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mechanisms: %w", err)
		}
	}

	var cases []extraction.GroundedCaseStudy
	if row.CaseStudiesJson != "" {
		if err := json.Unmarshal([]byte(row.CaseStudiesJson), &cases); err != nil {
			return nil, fmt.Errorf("failed to unmarshal case studies: %w", err)
		}
	}

	return &extraction.GroundedConceptSynthesis{
		AnchorPaperID: row.AnchorPaperID,
		Concept:       row.Concept,
		Definitions:   defs,
		Mechanisms:    mechanisms,
		CaseStudies:   cases,
		GeneratedAt:   row.GeneratedAt,
	}, nil
}
