package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/pipeline"
)

type SqlitePipelineRepository struct {
	db *sql.DB
}

func NewSqlitePipelineRepository(database *sql.DB) interfaces.IPipelineRepository {
	return &SqlitePipelineRepository{db: database}
}

func (r *SqlitePipelineRepository) SavePipeline(ctx context.Context, p *pipeline.ResearchPipeline) error {
	if p == nil {
		return errors.New("cannot save nil research pipeline")
	}

	stageDataJSON, err := json.Marshal(p.StageData)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline stage data: %w", err)
	}

	query := `
		INSERT INTO chat_histories (
			id, title, initial_prompt, model, current_stage, status, mode, approach, checkpoint,
			selected_angle_id, anchor_paper_id, stage_data_json, pinned, is_archived,
			error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			model = excluded.model,
			current_stage = excluded.current_stage,
			status = excluded.status,
			mode = excluded.mode,
			approach = excluded.approach,
			checkpoint = excluded.checkpoint,
			selected_angle_id = excluded.selected_angle_id,
			anchor_paper_id = excluded.anchor_paper_id,
			stage_data_json = excluded.stage_data_json,
			pinned = excluded.pinned,
			is_archived = excluded.is_archived,
			error_message = excluded.error_message,
			updated_at = excluded.updated_at;
	`

	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	appr := string(p.Approach)
	if appr == "" {
		appr = string(pipeline.ApproachNormal)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for pipeline %q: %w", p.ID, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, query,
		p.ID,
		p.Title,
		p.InitialPrompt,
		p.Model,
		string(p.CurrentStage),
		string(p.Status),
		string(p.Mode),
		appr,
		string(p.Checkpoint),
		p.SelectedAngleID,
		p.AnchorPaperID,
		string(stageDataJSON),
		p.Pinned,
		p.IsArchived,
		p.ErrorMessage,
		p.CreatedAt,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving chat history %q failed: %w", p.ID, err)
	}

	// Persist associated chat messages
	for i := range p.Messages {
		msg := &p.Messages[i]
		if msg.ChatID == "" {
			msg.ChatID = p.ID
		}
		if err := r.saveMessageTx(ctx, tx, msg); err != nil {
			return fmt.Errorf("failed to save chat message %q for chat %q: %w", msg.ID, p.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing pipeline %q failed: %w", p.ID, err)
	}

	return nil
}

func (r *SqlitePipelineRepository) GetPipeline(ctx context.Context, id string) (*pipeline.ResearchPipeline, error) {
	query := `
		SELECT id, title, initial_prompt, model, current_stage, status, mode, approach, checkpoint,
		       selected_angle_id, anchor_paper_id, stage_data_json, pinned, is_archived,
		       error_message, created_at, updated_at
		FROM chat_histories
		WHERE id = ?;
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var p pipeline.ResearchPipeline
	var stage, status, mode, approach, checkpoint, stageDataStr string

	err := row.Scan(
		&p.ID,
		&p.Title,
		&p.InitialPrompt,
		&p.Model,
		&stage,
		&status,
		&mode,
		&approach,
		&checkpoint,
		&p.SelectedAngleID,
		&p.AnchorPaperID,
		&stageDataStr,
		&p.Pinned,
		&p.IsArchived,
		&p.ErrorMessage,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pipeline.ErrPipelineNotFound
		}
		return nil, fmt.Errorf("querying chat history %q failed: %w", id, err)
	}

	p.CurrentStage = pipeline.Stage(stage)
	p.Status = pipeline.Status(status)
	p.Mode = pipeline.ExecutionMode(mode)
	if approach == "" {
		approach = string(pipeline.ApproachNormal)
	}
	p.Approach = pipeline.ResearchApproach(approach)
	p.Checkpoint = pipeline.CheckpointType(checkpoint)

	if stageDataStr != "" {
		_ = json.Unmarshal([]byte(stageDataStr), &p.StageData)
	}

	// Fetch associated chat messages
	msgs, err := r.GetMessages(ctx, p.ID)
	if err == nil {
		p.Messages = msgs
	}

	return &p, nil
}

func (r *SqlitePipelineRepository) ListPipelines(ctx context.Context, limit int) ([]pipeline.ResearchPipeline, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, title, initial_prompt, model, current_stage, status, mode, approach, checkpoint,
		       selected_angle_id, anchor_paper_id, stage_data_json, pinned, is_archived,
		       error_message, created_at, updated_at
		FROM chat_histories
		ORDER BY pinned DESC, created_at DESC
		LIMIT ?;
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("listing chat histories failed: %w", err)
	}
	defer rows.Close()

	var results []pipeline.ResearchPipeline
	for rows.Next() {
		var p pipeline.ResearchPipeline
		var stage, status, mode, approach, checkpoint, stageDataStr string

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.InitialPrompt,
			&p.Model,
			&stage,
			&status,
			&mode,
			&approach,
			&checkpoint,
			&p.SelectedAngleID,
			&p.AnchorPaperID,
			&stageDataStr,
			&p.Pinned,
			&p.IsArchived,
			&p.ErrorMessage,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning chat history row failed: %w", err)
		}

		p.CurrentStage = pipeline.Stage(stage)
		p.Status = pipeline.Status(status)
		p.Mode = pipeline.ExecutionMode(mode)
		if approach == "" {
			approach = string(pipeline.ApproachNormal)
		}
		p.Approach = pipeline.ResearchApproach(approach)
		p.Checkpoint = pipeline.CheckpointType(checkpoint)

		if stageDataStr != "" {
			_ = json.Unmarshal([]byte(stageDataStr), &p.StageData)
		}

		results = append(results, p)
	}

	return results, nil
}

func (r *SqlitePipelineRepository) DeletePipeline(ctx context.Context, id string) error {
	query := `DELETE FROM chat_histories WHERE id = ?;`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting chat history %q failed: %w", id, err)
	}
	return nil
}

func (r *SqlitePipelineRepository) SaveMessage(ctx context.Context, msg *pipeline.ChatMessage) error {
	if msg == nil {
		return errors.New("cannot save nil chat message")
	}
	return r.saveMessageTx(ctx, r.db, msg)
}

func (r *SqlitePipelineRepository) saveMessageTx(ctx context.Context, execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, msg *pipeline.ChatMessage) error {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now

	stepsJSON, _ := json.Marshal(msg.Steps)
	thinkingStepsJSON, _ := json.Marshal(msg.ThinkingSteps)
	statsJSON, _ := json.Marshal(msg.StatsSummary)
	citationsJSON, _ := json.Marshal(msg.Citations)
	quoteJSON, _ := json.Marshal(msg.Quote)
	checkpointJSON, _ := json.Marshal(msg.Checkpoint)
	followUpsJSON, _ := json.Marshal(msg.SuggestedFollowUps)

	query := `
		INSERT INTO chat_messages (
			id, chat_id, role, content, status, stage, steps_json,
			thinking_steps_json, stats_summary_json,
			citations_json, quote_json, checkpoint_json, suggested_follow_ups_json,
			error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			status = excluded.status,
			stage = excluded.stage,
			steps_json = excluded.steps_json,
			thinking_steps_json = excluded.thinking_steps_json,
			stats_summary_json = excluded.stats_summary_json,
			citations_json = excluded.citations_json,
			quote_json = excluded.quote_json,
			checkpoint_json = excluded.checkpoint_json,
			suggested_follow_ups_json = excluded.suggested_follow_ups_json,
			error_message = excluded.error_message,
			updated_at = excluded.updated_at;
	`

	_, err := execer.ExecContext(ctx, query,
		msg.ID,
		msg.ChatID,
		string(msg.Role),
		msg.Content,
		string(msg.Status),
		string(msg.Stage),
		string(stepsJSON),
		string(thinkingStepsJSON),
		string(statsJSON),
		string(citationsJSON),
		string(quoteJSON),
		string(checkpointJSON),
		string(followUpsJSON),
		msg.ErrorMessage,
		msg.CreatedAt,
		msg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("persisting chat message %q failed: %w", msg.ID, err)
	}

	return nil
}

func (r *SqlitePipelineRepository) GetMessages(ctx context.Context, chatID string) ([]pipeline.ChatMessage, error) {
	query := `
		SELECT id, chat_id, role, content, status, stage, steps_json,
		       thinking_steps_json, stats_summary_json,
		       citations_json, quote_json, checkpoint_json, suggested_follow_ups_json,
		       error_message, created_at, updated_at
		FROM chat_messages
		WHERE chat_id = ?
		ORDER BY created_at ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("querying chat messages for chat %q failed: %w", chatID, err)
	}
	defer rows.Close()

	var msgs []pipeline.ChatMessage
	for rows.Next() {
		var msg pipeline.ChatMessage
		var roleStr, statusStr, stageStr string
		var stepsStr, thinkingStepsStr, statsStr, citationsStr, quoteStr, checkpointStr, followUpsStr string

		if err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&roleStr,
			&msg.Content,
			&statusStr,
			&stageStr,
			&stepsStr,
			&thinkingStepsStr,
			&statsStr,
			&citationsStr,
			&quoteStr,
			&checkpointStr,
			&followUpsStr,
			&msg.ErrorMessage,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning chat message failed: %w", err)
		}

		msg.Role = pipeline.ChatMessageRole(roleStr)
		msg.Status = pipeline.ChatMessageStatus(statusStr)
		msg.Stage = pipeline.Stage(stageStr)

		if stepsStr != "" && stepsStr != "null" {
			_ = json.Unmarshal([]byte(stepsStr), &msg.Steps)
		}
		if thinkingStepsStr != "" && thinkingStepsStr != "null" && thinkingStepsStr != "[]" {
			_ = json.Unmarshal([]byte(thinkingStepsStr), &msg.ThinkingSteps)
		}
		if statsStr != "" && statsStr != "null" && statsStr != "{}" {
			_ = json.Unmarshal([]byte(statsStr), &msg.StatsSummary)
		}
		if citationsStr != "" && citationsStr != "null" {
			_ = json.Unmarshal([]byte(citationsStr), &msg.Citations)
		}
		if quoteStr != "" && quoteStr != "null" && quoteStr != "{}" {
			_ = json.Unmarshal([]byte(quoteStr), &msg.Quote)
		}
		if checkpointStr != "" && checkpointStr != "null" && checkpointStr != "{}" {
			_ = json.Unmarshal([]byte(checkpointStr), &msg.Checkpoint)
		}
		if followUpsStr != "" && followUpsStr != "null" {
			_ = json.Unmarshal([]byte(followUpsStr), &msg.SuggestedFollowUps)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (r *SqlitePipelineRepository) PinChat(ctx context.Context, id string, pinned bool) error {
	query := `UPDATE chat_histories SET pinned = ?, updated_at = ? WHERE id = ?;`
	_, err := r.db.ExecContext(ctx, query, pinned, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("updating pin status for chat %q failed: %w", id, err)
	}
	return nil
}
