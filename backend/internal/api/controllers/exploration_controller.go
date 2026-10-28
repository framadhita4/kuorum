package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"kuorum/backend/internal/api/response"
	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/exploration/commands/decompose_concept"
	"kuorum/backend/internal/application/exploration/commands/reformulate_query"
	"kuorum/backend/internal/application/exploration/queries/get_angles"
	"kuorum/backend/internal/application/exploration/queries/get_models"
)

type ExplorationController struct {
	mediator mediator.Mediator
}

func NewExplorationController(mediator mediator.Mediator) *ExplorationController {
	return &ExplorationController{
		mediator: mediator,
	}
}

// Decompose godoc
// @Summary Decompose concept into research angles
// @Description Decomposes a research concept or user prompt into structured academic angles and query suggestions.
// @Tags Exploration
// @Accept json
// @Produce json
// @Param body body decompose_concept.DecomposeConceptCommand true "Concept decomposition payload"
// @Success 200 {object} response.Envelope{data=decompose_concept.DecomposeConceptResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/exploration/decompose [post]
func (c *ExplorationController) Decompose(w http.ResponseWriter, r *http.Request) {
	var cmd decompose_concept.DecomposeConceptCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	result, err := c.mediator.Send(r.Context(), cmd)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Concept decomposed successfully", result)
}

// Reformulate godoc
// @Summary Reformulate query for academic search
// @Description Translates informal queries into standardized English academic search keywords.
// @Tags Exploration
// @Accept json
// @Produce json
// @Param body body reformulate_query.ReformulateQueryCommand true "Query reformulation payload"
// @Success 200 {object} response.Envelope{data=reformulate_query.ReformulateQueryResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/exploration/reformulate [post]
func (c *ExplorationController) Reformulate(w http.ResponseWriter, r *http.Request) {
	var cmd reformulate_query.ReformulateQueryCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	result, err := c.mediator.Send(r.Context(), cmd)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Query reformulated successfully", result)
}

// GetAngles godoc
// @Summary Get exploration angles
// @Description Retrieves cached or predefined exploration angles for a given concept.
// @Tags Exploration
// @Produce json
// @Param concept query string false "Research concept name"
// @Success 200 {object} response.Envelope{data=get_angles.GetAnglesResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/exploration/angles [get]
func (c *ExplorationController) GetAngles(w http.ResponseWriter, r *http.Request) {
	concept := r.URL.Query().Get("concept")
	q := get_angles.GetAnglesQuery{Concept: concept}

	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Exploration angles retrieved successfully", result)
}

// GetModels godoc
// @Summary List available LLM models and providers
// @Description Returns the catalog of registered LLM providers, model specifications, context limits, and reasoning capabilities.
// @Tags Exploration
// @Produce json
// @Success 200 {object} response.Envelope{data=get_models.GetModelsResult}
// @Router /api/v1/models [get]
// @Router /api/v1/exploration/models [get]
func (c *ExplorationController) GetModels(w http.ResponseWriter, r *http.Request) {
	q := get_models.GetModelsQuery{}

	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Available models retrieved successfully", result)
}

// Stream godoc
// @Summary Stream concept decomposition via Server-Sent Events (SSE)
// @Description Decomposes a research concept in real-time, streaming status progression events and finalized structured JSON.
// @Tags Exploration
// @Accept json
// @Produce text/event-stream
// @Param body body decompose_concept.DecomposeConceptCommand true "Concept decomposition payload"
// @Success 200 {string} string "Server-Sent Events stream"
// @Failure 400 {object} response.Envelope
// @Router /api/v1/exploration/stream [post]
func (c *ExplorationController) Stream(w http.ResponseWriter, r *http.Request) {
	var cmd decompose_concept.DecomposeConceptCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	sse, err := response.NewSSEWriter(w)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Streaming initialization failed", err.Error())
		return
	}

	_ = sse.WriteStatus("initiating_concept_decomposition")

	// Execute command via mediator
	result, err := c.mediator.Send(r.Context(), cmd)
	if err != nil {
		_ = sse.WriteError(err.Error())
		return
	}

	// Stream status progression
	_ = sse.WriteStatus("synthesizing_facets")
	time.Sleep(50 * time.Millisecond)
	_ = sse.WriteStatus("finalizing_academic_queries")
	time.Sleep(50 * time.Millisecond)

	_ = sse.WriteDone(result)
}
