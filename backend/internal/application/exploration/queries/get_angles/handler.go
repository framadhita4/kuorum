package get_angles

import (
	"context"
	"strings"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetAnglesQueryHandler struct {
	cache interfaces.ICacheRepository
	repo  interfaces.IExplorationRepository
}

func NewGetAnglesQueryHandler(cache interfaces.ICacheRepository, repo interfaces.IExplorationRepository) *GetAnglesQueryHandler {
	return &GetAnglesQueryHandler{
		cache: cache,
		repo:  repo,
	}
}

func (h *GetAnglesQueryHandler) Handle(ctx context.Context, q GetAnglesQuery) (*GetAnglesResult, error) {
	concept := strings.TrimSpace(q.Concept)
	if concept == "" {
		concept = "General Scientific Discovery"
	}

	if h.repo != nil {
		if decomp, err := h.repo.GetDecomposition(ctx, concept); err == nil && decomp != nil && len(decomp.Angles) > 0 {
			angles := make([]AngleSummaryDTO, len(decomp.Angles))
			for i, a := range decomp.Angles {
				angles[i] = AngleSummaryDTO{
					ID:          a.ID,
					Title:       a.Title,
					Description: a.Description,
				}
			}
			return &GetAnglesResult{
				Concept: decomp.NormalizedTerm,
				Angles:  angles,
			}, nil
		}
	}

	return &GetAnglesResult{
		Concept: concept,
		Angles: []AngleSummaryDTO{
			{
				ID:          "angle-1",
				Title:       "Theoretical Foundations & Principles",
				Description: "Fundamental theorems and formal baselines.",
			},
			{
				ID:          "angle-2",
				Title:       "Algorithmic Architecture & System Mechanics",
				Description: "Core components and technical workflows.",
			},
			{
				ID:          "angle-3",
				Title:       "Empirical Applications & Real-World Use Cases",
				Description: "Industry deployment, clinical use, and case studies.",
			},
			{
				ID:          "angle-4",
				Title:       "Evaluation Benchmarks & Comparative SOTA",
				Description: "Performance metrics and comparative trade-offs.",
			},
		},
	}, nil
}
