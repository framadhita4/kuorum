package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/infrastructure/persistence/db"
)

type SqliteGraphRepository struct {
	queries *db.Queries
}

func NewSqliteGraphRepository(database *sql.DB) interfaces.IGraphRepository {
	return &SqliteGraphRepository{
		queries: db.New(database),
	}
}

func (r *SqliteGraphRepository) SaveGraph(ctx context.Context, g *graph.CitationGraph) error {
	if g == nil {
		return errors.New("cannot save nil citation graph")
	}

	nodesJSON, err := json.Marshal(g.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal graph nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(g.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal graph edges: %w", err)
	}

	traversedAt := g.TraversedAt
	if traversedAt.IsZero() {
		traversedAt = time.Now().UTC()
	}

	err = r.queries.UpsertCitationGraph(ctx, db.UpsertCitationGraphParams{
		AnchorPaperID: g.AnchorPaperID,
		NodesJson:     string(nodesJSON),
		EdgesJson:     string(edgesJSON),
		TraversedAt:   traversedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert citation graph for anchor %q: %w", g.AnchorPaperID, err)
	}

	return nil
}

func (r *SqliteGraphRepository) GetGraph(ctx context.Context, anchorPaperID string) (*graph.CitationGraph, error) {
	row, err := r.queries.GetCitationGraph(ctx, anchorPaperID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, graph.ErrGraphNotFound
		}
		return nil, fmt.Errorf("failed to query citation graph for anchor %q: %w", anchorPaperID, err)
	}

	var nodes []graph.GraphNode
	if row.NodesJson != "" {
		if err := json.Unmarshal([]byte(row.NodesJson), &nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal graph nodes: %w", err)
		}
	}

	var edges []graph.GraphEdge
	if row.EdgesJson != "" {
		if err := json.Unmarshal([]byte(row.EdgesJson), &edges); err != nil {
			return nil, fmt.Errorf("failed to unmarshal graph edges: %w", err)
		}
	}

	return &graph.CitationGraph{
		AnchorPaperID: row.AnchorPaperID,
		Nodes:         nodes,
		Edges:         edges,
		TraversedAt:   row.TraversedAt,
	}, nil
}

func (r *SqliteGraphRepository) DeleteGraph(ctx context.Context, anchorPaperID string) error {
	if err := r.queries.DeleteCitationGraph(ctx, anchorPaperID); err != nil {
		return fmt.Errorf("failed to delete citation graph for anchor %q: %w", anchorPaperID, err)
	}
	return nil
}
