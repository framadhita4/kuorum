package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/exploration/commands/decompose_concept"
	"kuorum/backend/internal/application/extraction/commands/synthesize_concept"
	"kuorum/backend/internal/application/graph/commands/build_graph"
	"kuorum/backend/internal/application/pipeline/orchestrator"
	"kuorum/backend/internal/application/retrieval/commands/ingest_pdf"
	"kuorum/backend/internal/application/retrieval/queries/search_papers"
	"kuorum/backend/internal/application/retrieval/services"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/pipeline"
	"kuorum/backend/internal/domain/retrieval"
)

type mockPipelineMediator struct{}

func (m *mockPipelineMediator) RegisterCommandHandler(cmdType any, handler any) {}
func (m *mockPipelineMediator) RegisterQueryHandler(queryType any, handler any) {}
func (m *mockPipelineMediator) AddBehavior(behavior mediator.PipelineBehavior)  {}

func (m *mockPipelineMediator) Send(ctx context.Context, cmd any) (any, error) {
	switch cmd.(type) {
	case decompose_concept.DecomposeConceptCommand:
		return &decompose_concept.DecomposeConceptResult{
			NormalizedTerm: "Solid State Battery",
			Angles: []decompose_concept.AngleDTO{
				{AngleID: "ang-1", Title: "Electrolyte Conductivity", Description: "Ionic transport mechanisms"},
			},
		}, nil
	case ingest_pdf.IngestPdfCommand:
		return &ingest_pdf.IngestPdfResult{PaperID: "p-1", SectionCount: 5, RefCount: 10}, nil
	case build_graph.BuildCitationGraphCommand:
		return &build_graph.BuildCitationGraphResult{AnchorPaperID: "p-1", TotalNodes: 3, TotalEdges: 2}, nil
	case synthesize_concept.SynthesizeConceptCommand:
		return &synthesize_concept.SynthesizeConceptResult{
			Synthesis: &extraction.GroundedConceptSynthesis{},
			Audit:     synthesize_concept.AuditSummary{VerifiedClaims: 2, AuditPassed: true},
		}, nil
	}
	return nil, nil
}

func (m *mockPipelineMediator) Query(ctx context.Context, query any) (any, error) {
	switch query.(type) {
	case search_papers.SearchPapersQuery:
		return &search_papers.SearchPapersResult{
			Query: "Solid State Battery",
			Total: 2,
			Papers: []retrieval.PaperMetadata{
				{
					ID:            "paper-top-1",
					Title:         "High ionic conductivity in solid electrolyte",
					Abstract:      "We achieve unprecedented ionic conductivity in garnet solid electrolytes.",
					Year:          2024,
					CitationCount: 150,
				},
				{
					ID:            "paper-top-2",
					Title:         "Interfacial resistance in all-solid-state batteries",
					Abstract:      "Analysis of lithium metal solid-state electrolyte interface stability.",
					Year:          2023,
					CitationCount: 95,
				},
			},
		}, nil
	}
	return nil, nil
}

type mockPipelineStorage struct {
	mu    sync.RWMutex
	saved map[string]*pipeline.ResearchPipeline
}

func (s *mockPipelineStorage) SavePipeline(ctx context.Context, p *pipeline.ResearchPipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.saved == nil {
		s.saved = make(map[string]*pipeline.ResearchPipeline)
	}
	s.saved[p.ID] = p.Clone()
	return nil
}

func (s *mockPipelineStorage) GetPipeline(ctx context.Context, id string) (*pipeline.ResearchPipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.saved == nil {
		return nil, pipeline.ErrPipelineNotFound
	}
	p, ok := s.saved[id]
	if !ok {
		return nil, pipeline.ErrPipelineNotFound
	}
	return p.Clone(), nil
}

func (s *mockPipelineStorage) ListPipelines(ctx context.Context, limit int) ([]pipeline.ResearchPipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []pipeline.ResearchPipeline
	for _, p := range s.saved {
		list = append(list, *p.Clone())
	}
	return list, nil
}

func (s *mockPipelineStorage) DeletePipeline(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.saved, id)
	return nil
}

func (s *mockPipelineStorage) SaveMessage(ctx context.Context, msg *pipeline.ChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.saved != nil {
		if p, ok := s.saved[msg.ChatID]; ok {
			p.AddMessage(*msg)
		}
	}
	return nil
}

func (s *mockPipelineStorage) GetMessages(ctx context.Context, chatID string) ([]pipeline.ChatMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.saved != nil {
		if p, ok := s.saved[chatID]; ok {
			return p.Clone().Messages, nil
		}
	}
	return nil, nil
}

func (s *mockPipelineStorage) PinChat(ctx context.Context, id string, pinned bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.saved != nil {
		if p, ok := s.saved[id]; ok {
			p.Pinned = pinned
		}
	}
	return nil
}

func TestPipelineController_StartApproaches(t *testing.T) {
	med := &mockPipelineMediator{}
	repo := &mockPipelineStorage{}
	ranker := services.NewAnchorPaperRanker()
	orchestratorEngine := orchestrator.NewPipelineOrchestrator(med, repo, ranker)
	ctrl := NewPipelineController(orchestratorEngine, repo)

	t.Run("Start_NormalMode", func(t *testing.T) {
		payload := StartPipelineRequest{
			Prompt:   "Solid state battery electrolytes",
			Mode:     pipeline.ModeAutonomous,
			Approach: pipeline.ApproachNormal,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ctrl.Start(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool                      `json:"success"`
			Data    pipeline.ResearchPipeline `json:"data"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Data.Approach != pipeline.ApproachNormal {
			t.Errorf("expected ApproachNormal, got %s", resp.Data.Approach)
		}
	})

	t.Run("Start_DeepResearchMode", func(t *testing.T) {
		payload := StartPipelineRequest{
			Prompt:   "Lithium metal dendrite prevention",
			Mode:     pipeline.ModeAutonomous,
			Approach: pipeline.ApproachDeep,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ctrl.Start(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool                      `json:"success"`
			Data    pipeline.ResearchPipeline `json:"data"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Data.Approach != pipeline.ApproachDeep {
			t.Errorf("expected ApproachDeep, got %s", resp.Data.Approach)
		}
	})
}
