package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/exploration"
	"kuorum/backend/internal/infrastructure/persistence/db"
)

type SqliteExplorationRepository struct {
	queries *db.Queries
}

func NewSqliteExplorationRepository(database *sql.DB) interfaces.IExplorationRepository {
	return &SqliteExplorationRepository{
		queries: db.New(database),
	}
}

func (r *SqliteExplorationRepository) SaveDecomposition(ctx context.Context, decomp *exploration.ConceptDecomposition) error {
	if decomp == nil {
		return errors.New("cannot save nil concept decomposition")
	}

	anglesJSON, err := json.Marshal(decomp.Angles)
	if err != nil {
		return fmt.Errorf("failed to marshal exploration angles: %w", err)
	}

	err = r.queries.UpsertExploration(ctx, db.UpsertExplorationParams{
		Prompt:         decomp.OriginalPrompt,
		NormalizedTerm: decomp.NormalizedTerm,
		AnglesJson:     string(anglesJSON),
		CreatedAt:      decomp.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to persist exploration: %w", err)
	}

	return nil
}

func (r *SqliteExplorationRepository) GetDecomposition(ctx context.Context, prompt string) (*exploration.ConceptDecomposition, error) {
	row, err := r.queries.GetExploration(ctx, prompt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exploration.ErrAngleNotFound
		}
		return nil, fmt.Errorf("failed to query exploration: %w", err)
	}

	var angles []exploration.ResearchAngle
	if err := json.Unmarshal([]byte(row.AnglesJson), &angles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exploration angles: %w", err)
	}

	return &exploration.ConceptDecomposition{
		OriginalPrompt: row.Prompt,
		NormalizedTerm: row.NormalizedTerm,
		Angles:         angles,
		CreatedAt:      row.CreatedAt,
	}, nil
}
