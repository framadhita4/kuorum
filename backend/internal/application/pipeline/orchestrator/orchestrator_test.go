package orchestrator_test

import (
	"context"
	"testing"
	"time"

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

type mockMediator struct {
	decomposeResult  *decompose_concept.DecomposeConceptResult
	searchResult     *search_papers.SearchPapersResult
	ingestResult     *ingest_pdf.IngestPdfResult
	graphResult      *build_graph.BuildCitationGraphResult
	synthesizeResult *synthesize_concept.SynthesizeConceptResult
}

func (m *mockMediator) RegisterCommandHandler(cmdType any, handler any) {}
func (m *mockMediator) RegisterQueryHandler(queryType any, handler any) {}
func (m *mockMediator) AddBehavior(behavior mediator.PipelineBehavior)  {}

func (m *mockMediator) Send(ctx context.Context, cmd any) (any, error) {
	switch cmd.(type) {
	case decompose_concept.DecomposeConceptCommand:
		return m.decomposeResult, nil
	case ingest_pdf.IngestPdfCommand:
		return m.ingestResult, nil
	case build_graph.BuildCitationGraphCommand:
		return m.graphResult, nil
	case synthesize_concept.SynthesizeConceptCommand:
		return m.synthesizeResult, nil
	}
	return nil, nil
}

func (m *mockMediator) Query(ctx context.Context, query any) (any, error) {
	switch query.(type) {
	case search_papers.SearchPapersQuery:
		return m.searchResult, nil
	}
	return nil, nil
}

type mockPipelineRepo struct {
	pipelines map[string]*pipeline.ResearchPipeline
}

func (m *mockPipelineRepo) SavePipeline(ctx context.Context, p *pipeline.ResearchPipeline) error {
	if m.pipelines == nil {
		m.pipelines = make(map[string]*pipeline.ResearchPipeline)
	}
	m.pipelines[p.ID] = p
	return nil
}

func (m *mockPipelineRepo) GetPipeline(ctx context.Context, id string) (*pipeline.ResearchPipeline, error) {
	if m.pipelines == nil {
		return nil, pipeline.ErrPipelineNotFound
	}
	p, ok := m.pipelines[id]
	if !ok {
		return nil, pipeline.ErrPipelineNotFound
	}
	return p, nil
}

func (m *mockPipelineRepo) ListPipelines(ctx context.Context, limit int) ([]pipeline.ResearchPipeline, error) {
	var list []pipeline.ResearchPipeline
	for _, p := range m.pipelines {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockPipelineRepo) DeletePipeline(ctx context.Context, id string) error {
	delete(m.pipelines, id)
	return nil
}

func (m *mockPipelineRepo) SaveMessage(ctx context.Context, msg *pipeline.ChatMessage) error {
	if m.pipelines != nil {
		if p, ok := m.pipelines[msg.ChatID]; ok {
			p.AddMessage(*msg)
		}
	}
	return nil
}

func (m *mockPipelineRepo) GetMessages(ctx context.Context, chatID string) ([]pipeline.ChatMessage, error) {
	if m.pipelines != nil {
		if p, ok := m.pipelines[chatID]; ok {
			return p.Messages, nil
		}
	}
	return nil, nil
}

func (m *mockPipelineRepo) PinChat(ctx context.Context, id string, pinned bool) error {
	if m.pipelines != nil {
		if p, ok := m.pipelines[id]; ok {
			p.Pinned = pinned
		}
	}
	return nil
}

func setupMockMediator() *mockMediator {
	return &mockMediator{
		decomposeResult: &decompose_concept.DecomposeConceptResult{
			NormalizedTerm: "Solid State Battery",
			Angles: []decompose_concept.AngleDTO{
				{
					AngleID:          "angle-1",
					Title:            "Dendrite Suppression",
					SuggestedQueries: []string{"solid state battery lithium dendrite"},
				},
			},
		},
		searchResult: &search_papers.SearchPapersResult{
			Papers: []retrieval.PaperMetadata{
				{
					ID:               "paper-anchor-1",
					Title:            "Solid State Battery Lithium Dendrite Suppression",
					Abstract:         "Detailed abstract describing electrolyte interphase engineering.",
					CitationCount:    250,
					Year:             2023,
					OpenAccessPdfURL: "https://arxiv.org/pdf/2301.0001.pdf",
				},
			},
			Total: 1,
		},
		ingestResult: &ingest_pdf.IngestPdfResult{
			Success:      true,
			PaperID:      "paper-anchor-1",
			SectionCount: 5,
			RefCount:     12,
		},
		graphResult: &build_graph.BuildCitationGraphResult{
			AnchorPaperID:     "paper-anchor-1",
			TotalNodes:        8,
			TotalEdges:        7,
			FoundationalCount: 5,
			ApplicationCount:  2,
		},
		synthesizeResult: &synthesize_concept.SynthesizeConceptResult{
			Synthesis: &extraction.GroundedConceptSynthesis{
				AnchorPaperID: "paper-anchor-1",
				Concept:       "Solid State Battery",
				Definitions: []extraction.GroundedDefinition{
					{
						Term:       "Artificial SEI",
						Definition: "Solid electrolyte protective interphase",
						Citation: extraction.CitationProof{
							PaperID:      "paper-anchor-1",
							VerbatimText: "Solid electrolyte protective interphase",
						},
					},
				},
				GeneratedAt: time.Now(),
			},
			Audit: synthesize_concept.AuditSummary{
				TotalClaims:      1,
				VerifiedClaims:   1,
				VerificationRate: 1.0,
				AuditPassed:      true,
			},
		},
	}
}

func TestPipelineOrchestrator_AutonomousMode(t *testing.T) {
	med := setupMockMediator()
	repo := &mockPipelineRepo{}
	ranker := services.NewAnchorPaperRanker()
	orchestrator := orchestrator.NewPipelineOrchestrator(med, repo, ranker)

	ctx := context.Background()
	p, err := orchestrator.StartSync(ctx, "pipe-auto-1", "Dendrite suppression in solid state battery", pipeline.ModeAutonomous)
	if err != nil {
		t.Fatalf("autonomous pipeline start failed: %v", err)
	}

	if p.Status != pipeline.StatusCompleted {
		t.Errorf("expected pipeline status COMPLETED, got %s", p.Status)
	}
	if p.CurrentStage != pipeline.StageCompleted {
		t.Errorf("expected stage COMPLETED, got %s", p.CurrentStage)
	}
	if p.AnchorPaperID != "paper-anchor-1" {
		t.Errorf("expected anchor paper 'paper-anchor-1', got %s", p.AnchorPaperID)
	}
	if p.StageData.Synthesis == nil {
		t.Errorf("expected non-nil synthesis in final stage data")
	}
}

func TestPipelineOrchestrator_AsyncStartAndEventStream(t *testing.T) {
	med := setupMockMediator()
	repo := &mockPipelineRepo{}
	ranker := services.NewAnchorPaperRanker()
	orchestrator := orchestrator.NewPipelineOrchestrator(med, repo, ranker)

	ctx := context.Background()
	pipeID := "pipe-async-1"

	// Subscribe to event stream before or right after start
	events, unsubscribe := orchestrator.Subscribe(pipeID)
	defer unsubscribe()

	p, err := orchestrator.Start(ctx, pipeID, "Solid state batteries", pipeline.ModeAutonomous)
	if err != nil {
		t.Fatalf("async start failed: %v", err)
	}

	// Verify immediate non-blocking return with status RUNNING
	if p.Status != pipeline.StatusRunning {
		t.Errorf("expected immediate status RUNNING, got %s", p.Status)
	}
	if p.ID != pipeID {
		t.Errorf("expected ID %s, got %s", pipeID, p.ID)
	}

	// Verify at least one progress event received on the channel
	select {
	case ev := <-events:
		if ev.PipelineID != pipeID {
			t.Errorf("expected event for %s, got %s", pipeID, ev.PipelineID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for broadcast event")
	}
}

func TestPipelineOrchestrator_InteractiveMode(t *testing.T) {
	med := setupMockMediator()
	repo := &mockPipelineRepo{}
	ranker := services.NewAnchorPaperRanker()
	orchestrator := orchestrator.NewPipelineOrchestrator(med, repo, ranker)

	ctx := context.Background()

	// 1. Start pipeline in Interactive mode -> Should halt at ANGLE_SELECTION
	p, err := orchestrator.StartSync(ctx, "pipe-interactive-1", "Solid state batteries", pipeline.ModeInteractive)
	if err != nil {
		t.Fatalf("interactive start failed: %v", err)
	}

	if p.Status != pipeline.StatusWaitingInput {
		t.Fatalf("expected status WAITING_INPUT, got %s", p.Status)
	}
	if p.Checkpoint != pipeline.CheckpointAngleSelection {
		t.Fatalf("expected checkpoint ANGLE_SELECTION, got %s", p.Checkpoint)
	}

	// 2. Select Angle -> Automated selection without requiring manual paper selection, completing pipeline
	p, err = orchestrator.SelectAngleSync(ctx, p.ID, "angle-1")
	if err != nil {
		t.Fatalf("select angle failed: %v", err)
	}

	if p.Status != pipeline.StatusCompleted {
		t.Errorf("expected final status COMPLETED (automated paper selection), got %s", p.Status)
	}
	if p.StageData.Synthesis == nil {
		t.Errorf("expected non-nil final synthesis")
	}
	if len(p.StageData.SelectedPapers) == 0 {
		t.Errorf("expected automated selected papers for deep dive")
	}
}

func TestPipelineOrchestrator_DeepResearchMode(t *testing.T) {
	med := setupMockMediator()
	repo := &mockPipelineRepo{}
	ranker := services.NewAnchorPaperRanker()
	orchestrator := orchestrator.NewPipelineOrchestrator(med, repo, ranker)

	ctx := context.Background()

	p, err := orchestrator.StartSync(ctx, "pipe-deep-1", "Transformer architectures in deep learning", pipeline.ModeAutonomous, pipeline.ApproachDeep)
	if err != nil {
		t.Fatalf("deep research start failed: %v", err)
	}

	if p.Approach != pipeline.ApproachDeep {
		t.Errorf("expected ApproachDeep, got %s", p.Approach)
	}
	if p.Status != pipeline.StatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", p.Status)
	}
	if len(p.StageData.SelectedPapers) == 0 {
		t.Errorf("expected selected papers from Clef classification")
	}
	if len(p.StageData.PaperScores) == 0 {
		t.Errorf("expected Clef paper scores recorded")
	}
}

func TestPipelineOrchestrator_Cancel(t *testing.T) {
	med := setupMockMediator()
	repo := &mockPipelineRepo{}
	ranker := services.NewAnchorPaperRanker()
	orchestrator := orchestrator.NewPipelineOrchestrator(med, repo, ranker)

	ctx := context.Background()
	p, err := orchestrator.StartSync(ctx, "pipe-cancel-1", "Quantum computing", pipeline.ModeInteractive)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if err := orchestrator.Cancel(ctx, p.ID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	cancelled, err := repo.GetPipeline(ctx, p.ID)
	if err != nil {
		t.Fatalf("get cancelled pipeline failed: %v", err)
	}

	if cancelled.Status != pipeline.StatusCancelled {
		t.Errorf("expected status CANCELLED, got %s", cancelled.Status)
	}
}
