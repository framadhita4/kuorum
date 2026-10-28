package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"kuorum/backend/internal/api/response"
	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/application/pipeline/orchestrator"
	"kuorum/backend/internal/domain/pipeline"
)

type PipelineController struct {
	orchestrator *orchestrator.PipelineOrchestrator
	repo         interfaces.IPipelineRepository
}

func NewPipelineController(
	orchestrator *orchestrator.PipelineOrchestrator,
	repo interfaces.IPipelineRepository,
) *PipelineController {
	return &PipelineController{
		orchestrator: orchestrator,
		repo:         repo,
	}
}

type StartPipelineRequest struct {
	Prompt   string                    `json:"prompt"`
	Mode     pipeline.ExecutionMode    `json:"mode,omitempty"`
	Approach pipeline.ResearchApproach `json:"approach,omitempty"`
}

type SelectAngleRequest struct {
	AngleID string `json:"angleId"`
}

type SelectAnchorRequest struct {
	AnchorPaperID string `json:"anchorPaperId"`
}

// Start godoc
// @Summary Start a new research pipeline session
// @Description Initiates the autonomous research saga from prompt to synthesis.
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param body body StartPipelineRequest true "Pipeline startup parameters"
// @Success 201 {object} response.Envelope{data=pipeline.ResearchPipeline}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/pipelines [post]
func (c *PipelineController) Start(w http.ResponseWriter, r *http.Request) {
	var req StartPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		response.Error(w, http.StatusBadRequest, "Prompt cannot be empty", "prompt is required")
		return
	}

	pipeID := fmt.Sprintf("pipe_%d", time.Now().UnixNano())
	mode := pipeline.ModeAutonomous

	appr := pipeline.ApproachNormal
	if strings.EqualFold(string(req.Approach), string(pipeline.ApproachDeep)) || strings.EqualFold(string(req.Approach), "deep") {
		appr = pipeline.ApproachDeep
	}

	p, err := c.orchestrator.Start(r.Context(), pipeID, req.Prompt, mode, appr)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "Research pipeline initiated successfully", p)
}

// List godoc
// @Summary List research pipeline sessions
// @Description Retrieves historical research sessions ordered by creation date descending.
// @Tags Pipeline
// @Produce json
// @Param limit query int false "Max sessions to return (default: 20)"
// @Success 200 {object} response.Envelope{data=[]pipeline.ResearchPipeline}
// @Router /api/v1/pipelines [get]
func (c *PipelineController) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	list, err := c.repo.ListPipelines(r.Context(), limit)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Pipelines retrieved successfully", list)
}

// Get godoc
// @Summary Get pipeline state and stage artifacts
// @Description Fetches the aggregate state, current stage, and artifacts for a pipeline.
// @Tags Pipeline
// @Produce json
// @Param id path string true "Pipeline ID"
// @Success 200 {object} response.Envelope{data=pipeline.ResearchPipeline}
// @Failure 404 {object} response.Envelope
// @Router /api/v1/pipelines/{id} [get]
func (c *PipelineController) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	p, err := c.repo.GetPipeline(r.Context(), id)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Pipeline retrieved successfully", p)
}

// SelectAngle godoc
// @Summary Confirm research angle checkpoint
// @Description Submits chosen research angle at ANGLE_SELECTION checkpoint and continues execution.
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param body body SelectAngleRequest true "Selected angle ID"
// @Success 200 {object} response.Envelope{data=pipeline.ResearchPipeline}
// @Router /api/v1/pipelines/{id}/select-angle [post]
func (c *PipelineController) SelectAngle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req SelectAngleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	p, err := c.orchestrator.SelectAngle(r.Context(), id, req.AngleID)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Research angle confirmed", p)
}

// SelectAnchor godoc
// @Summary Confirm anchor paper checkpoint
// @Description Submits chosen anchor paper at ANCHOR_SELECTION checkpoint and continues execution.
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param body body SelectAnchorRequest true "Selected anchor paper ID"
// @Success 200 {object} response.Envelope{data=pipeline.ResearchPipeline}
// @Router /api/v1/pipelines/{id}/select-anchor [post]
func (c *PipelineController) SelectAnchor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req SelectAnchorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	p, err := c.orchestrator.SelectAnchor(r.Context(), id, req.AnchorPaperID)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Anchor paper confirmed", p)
}

// Resume godoc
// @Summary Resume research pipeline execution
// @Description Continues execution for a paused or failed pipeline session.
// @Tags Pipeline
// @Produce json
// @Param id path string true "Pipeline ID"
// @Success 200 {object} response.Envelope{data=pipeline.ResearchPipeline}
// @Router /api/v1/pipelines/{id}/resume [post]
func (c *PipelineController) Resume(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := c.orchestrator.Resume(r.Context(), id)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Pipeline resumed successfully", p)
}

// Cancel godoc
// @Summary Cancel research pipeline session
// @Tags Pipeline
// @Produce json
// @Param id path string true "Pipeline ID"
// @Success 200 {object} response.Envelope
// @Router /api/v1/pipelines/{id}/cancel [post]
func (c *PipelineController) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := c.orchestrator.Cancel(r.Context(), id); err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Pipeline cancelled successfully", nil)
}

// Stream godoc
// @Summary Stream pipeline real-time events via SSE
// @Description Server-Sent Events stream delivering live stage progress, logs, and checkpoints.
// @Tags Pipeline
// @Produce text/event-stream
// @Param id path string true "Pipeline ID"
// @Router /api/v1/pipelines/{id}/stream [get]
func (c *PipelineController) Stream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	sse, err := response.NewSSEWriter(w)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "SSE initialization failed", err.Error())
		return
	}

	events, unsubscribe := c.orchestrator.Subscribe(id)
	defer unsubscribe()

	// Initial heartbeat event
	_ = sse.WriteEvent("connected", map[string]string{
		"pipelineId": id,
		"status":     "listening",
	})

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			if err := sse.WriteEvent(string(ev.Type), ev); err != nil {
				return
			}
			if ev.Type == orchestrator.EventCompleted || ev.Type == orchestrator.EventFailed {
				return
			}
		}
	}
}

type AddMessageRequest struct {
	Role               string   `json:"role"`
	Content            string   `json:"content"`
	Status             string   `json:"status,omitempty"`
	Steps              []string `json:"steps,omitempty"`
	SuggestedFollowUps []string `json:"suggestedFollowUps,omitempty"`
}

// AddMessage godoc
// @Summary Append a message to chat history
// @Description Appends a new user or assistant message to the chat history thread.
// @Tags Pipeline, Chat
// @Accept json
// @Produce json
// @Param id path string true "Chat / Pipeline ID"
// @Param body body AddMessageRequest true "Message payload"
// @Success 201 {object} response.Envelope{data=pipeline.ChatMessage}
// @Router /api/v1/pipelines/{id}/messages [post]
func (c *PipelineController) AddMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	var req AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		response.Error(w, http.StatusBadRequest, "Content cannot be empty", "content is required")
		return
	}

	role := pipeline.ChatMessageRole(req.Role)
	if role == "" {
		role = pipeline.RoleUser
	}

	status := pipeline.ChatMessageStatus(req.Status)
	if status == "" {
		status = pipeline.MessageStatusDone
	}

	now := time.Now().UTC()
	msg := pipeline.ChatMessage{
		ID:                 fmt.Sprintf("%s-msg-%d", id, now.UnixNano()),
		ChatID:             id,
		Role:               role,
		Content:            req.Content,
		Status:             status,
		Steps:              req.Steps,
		SuggestedFollowUps: req.SuggestedFollowUps,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := c.repo.SaveMessage(r.Context(), &msg); err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "Message added successfully", msg)
}

// GetMessages godoc
// @Summary Get messages for a chat history session
// @Tags Pipeline, Chat
// @Produce json
// @Param id path string true "Chat / Pipeline ID"
// @Success 200 {object} response.Envelope{data=[]pipeline.ChatMessage}
// @Router /api/v1/pipelines/{id}/messages [get]
func (c *PipelineController) GetMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	msgs, err := c.repo.GetMessages(r.Context(), id)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Messages retrieved successfully", msgs)
}

type TogglePinRequest struct {
	Pinned bool `json:"pinned"`
}

// TogglePin godoc
// @Summary Toggle pinned status for a chat history
// @Tags Pipeline, Chat
// @Accept json
// @Produce json
// @Param id path string true "Chat / Pipeline ID"
// @Param body body TogglePinRequest true "Pin status"
// @Success 200 {object} response.Envelope
// @Router /api/v1/pipelines/{id}/pin [patch]
func (c *PipelineController) TogglePin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	var req TogglePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if err := c.repo.PinChat(r.Context(), id, req.Pinned); err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Pin status updated successfully", map[string]any{
		"id":     id,
		"pinned": req.Pinned,
	})
}

// Delete godoc
// @Summary Delete a chat history session
// @Tags Pipeline, Chat
// @Produce json
// @Param id path string true "Chat / Pipeline ID"
// @Success 200 {object} response.Envelope
// @Router /api/v1/pipelines/{id} [delete]
func (c *PipelineController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing pipeline ID", "id parameter is required")
		return
	}

	if err := c.repo.DeletePipeline(r.Context(), id); err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Chat history deleted successfully", nil)
}
