package persistence_test

import (
	"context"
	"path/filepath"
	"testing"

	"kuorum/backend/internal/domain/pipeline"
	"kuorum/backend/internal/infrastructure/persistence"
)

func TestSqlitePipelineRepository_CRUD(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_pipeline.db")
	db, err := persistence.NewSqliteDB(tempDB)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	repo := persistence.NewSqlitePipelineRepository(db)
	ctx := context.Background()

	p, err := pipeline.NewResearchPipeline("pipe-101", "Transformer Attention Scaling Laws", pipeline.ModeInteractive)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// 1. Save
	if err := repo.SavePipeline(ctx, p); err != nil {
		t.Fatalf("SavePipeline failed: %v", err)
	}

	// 2. Get
	fetched, err := repo.GetPipeline(ctx, "pipe-101")
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}
	if fetched.ID != "pipe-101" {
		t.Errorf("expected ID pipe-101, got %s", fetched.ID)
	}
	if fetched.CurrentStage != pipeline.StageExploration {
		t.Errorf("expected stage EXPLORATION, got %s", fetched.CurrentStage)
	}
	if fetched.Status != pipeline.StatusDraft {
		t.Errorf("expected status DRAFT, got %s", fetched.Status)
	}

	// 3. Update State
	_ = fetched.TransitionTo(pipeline.StageRetrieval, pipeline.StatusRunning, pipeline.CheckpointNone)
	fetched.SelectedAngleID = "angle-opt"
	if err := repo.SavePipeline(ctx, fetched); err != nil {
		t.Fatalf("updating pipeline failed: %v", err)
	}

	updated, err := repo.GetPipeline(ctx, "pipe-101")
	if err != nil {
		t.Fatalf("fetching updated pipeline failed: %v", err)
	}
	if updated.CurrentStage != pipeline.StageRetrieval {
		t.Errorf("expected stage RETRIEVAL, got %s", updated.CurrentStage)
	}
	if updated.SelectedAngleID != "angle-opt" {
		t.Errorf("expected selected angle 'angle-opt', got %s", updated.SelectedAngleID)
	}

	// 4. List
	p2, _ := pipeline.NewResearchPipeline("pipe-102", "Graph Neural Networks", pipeline.ModeAutonomous)
	_ = repo.SavePipeline(ctx, p2)

	list, err := repo.ListPipelines(ctx, 10)
	if err != nil {
		t.Fatalf("ListPipelines failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 pipelines in list, got %d", len(list))
	}

	// 5. Delete
	if err := repo.DeletePipeline(ctx, "pipe-101"); err != nil {
		t.Fatalf("DeletePipeline failed: %v", err)
	}
	_, err = repo.GetPipeline(ctx, "pipe-101")
	if err != pipeline.ErrPipelineNotFound {
		t.Errorf("expected ErrPipelineNotFound after delete, got %v", err)
	}
}

func TestSqlitePipelineRepository_MessagesAndPin(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_messages.db")
	db, err := persistence.NewSqliteDB(tempDB)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	repo := persistence.NewSqlitePipelineRepository(db)
	ctx := context.Background()

	p, err := pipeline.NewResearchPipeline("chat-999", "Autonomous Research Helper Architecture", pipeline.ModeInteractive)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}
	p.Model = "gemini-2.0-flash"

	if err := repo.SavePipeline(ctx, p); err != nil {
		t.Fatalf("SavePipeline failed: %v", err)
	}

	// 1. Verify initial user message was saved
	fetched, err := repo.GetPipeline(ctx, "chat-999")
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}
	if len(fetched.Messages) != 1 {
		t.Fatalf("expected 1 initial message, got %d", len(fetched.Messages))
	}
	if fetched.Messages[0].Content != "Autonomous Research Helper Architecture" {
		t.Errorf("unexpected message content: %s", fetched.Messages[0].Content)
	}

	// 2. Add an assistant response message
	astMsg := pipeline.ChatMessage{
		ID:      "msg-ast-1",
		ChatID:  "chat-999",
		Role:    pipeline.RoleAssistant,
		Content: "Found 4 academic angles for research.",
		Status:  pipeline.MessageStatusDone,
		Steps:   []string{"Decomposition", "Clustering"},
		SuggestedFollowUps: []string{
			"Show mathematical foundations",
			"Compare benchmarks",
		},
	}
	if err := repo.SaveMessage(ctx, &astMsg); err != nil {
		t.Fatalf("SaveMessage failed: %v", err)
	}

	msgs, err := repo.GetMessages(ctx, "chat-999")
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[1].Role != pipeline.RoleAssistant {
		t.Errorf("expected assistant role, got %s", msgs[1].Role)
	}
	if len(msgs[1].SuggestedFollowUps) != 2 {
		t.Errorf("expected 2 suggested followups, got %d", len(msgs[1].SuggestedFollowUps))
	}

	// 3. Test Pinning
	if err := repo.PinChat(ctx, "chat-999", true); err != nil {
		t.Fatalf("PinChat failed: %v", err)
	}
	pinnedP, err := repo.GetPipeline(ctx, "chat-999")
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}
	if !pinnedP.Pinned {
		t.Errorf("expected pipeline to be pinned")
	}
}
