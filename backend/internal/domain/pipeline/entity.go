package pipeline

import (
	"errors"
	"time"

	"kuorum/backend/internal/domain/exploration"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/domain/retrieval"
)

var (
	ErrInvalidStateTransition  = errors.New("invalid pipeline state transition")
	ErrPipelineAlreadyFinished = errors.New("pipeline is already in a terminal state")
	ErrMissingAnchorPaper      = errors.New("cannot proceed to ingestion without an anchor paper")
	ErrPipelineNotFound        = errors.New("research pipeline not found")
	ErrEmptyPrompt             = errors.New("initial prompt cannot be empty")
)

type Stage string

const (
	StageExploration Stage = "EXPLORATION"
	StageRetrieval   Stage = "RETRIEVAL"
	StageIngestion   Stage = "INGESTION"
	StageGraph       Stage = "GRAPH"
	StageExtraction  Stage = "EXTRACTION"
	StageCompleted   Stage = "COMPLETED"
)

type Status string

const (
	StatusDraft        Status = "DRAFT"
	StatusRunning      Status = "RUNNING"
	StatusWaitingInput Status = "WAITING_INPUT"
	StatusCompleted    Status = "COMPLETED"
	StatusFailed       Status = "FAILED"
	StatusCancelled    Status = "CANCELLED"
)

type ExecutionMode string

const (
	ModeAutonomous  ExecutionMode = "AUTONOMOUS"
	ModeInteractive ExecutionMode = "INTERACTIVE"
)

type ResearchApproach string

const (
	ApproachNormal ResearchApproach = "NORMAL"
	ApproachDeep   ResearchApproach = "DEEP_RESEARCH"
)

type CheckpointType string

const (
	CheckpointNone            CheckpointType = "NONE"
	CheckpointAngleSelection  CheckpointType = "ANGLE_SELECTION"
	CheckpointAnchorSelection CheckpointType = "ANCHOR_SELECTION"
)

type ChatMessageRole string

const (
	RoleUser      ChatMessageRole = "user"
	RoleAssistant ChatMessageRole = "assistant"
	RoleSystem    ChatMessageRole = "system"
)

type ChatMessageStatus string

const (
	MessageStatusSending   ChatMessageStatus = "sending"
	MessageStatusStreaming ChatMessageStatus = "streaming"
	MessageStatusDone      ChatMessageStatus = "done"
	MessageStatusError     ChatMessageStatus = "error"
)

// ChatCitationRef represents verified academic citation metadata attached to a message.
type ChatCitationRef struct {
	ID      string `json:"id"`
	Num     int    `json:"num"`
	Title   string `json:"title"`
	Authors string `json:"authors"`
	Year    int    `json:"year"`
	Journal string `json:"journal"`
	DOI     string `json:"doi,omitempty"`
}

// ChatQuote represents empirical verbatim snippet evidence attached to a message.
type ChatQuote struct {
	Text    string `json:"text"`
	Source  string `json:"source"`
	PaperID string `json:"paperId,omitempty"`
}

// ChatCheckpointData captures interactive checkpoint options (angles, candidate papers).
type ChatCheckpointData struct {
	Type            CheckpointType              `json:"type"`
	Angles          []exploration.ResearchAngle `json:"angles,omitempty"`
	CandidatePapers []retrieval.PaperMetadata   `json:"candidatePapers,omitempty"`
}

// PaperScoreInfo captures point calculation and rapid classification metrics for a paper.
type PaperScoreInfo struct {
	PaperID            string  `json:"paperId"`
	Title              string  `json:"title"`
	TotalPoints        float64 `json:"totalPoints"` // 0 - 100 points
	RelevanceScore     float64 `json:"relevanceScore"`
	AuthorityScore     float64 `json:"authorityScore"`
	RecencyScore       float64 `json:"recencyScore"`
	OpenAccessScore    float64 `json:"openAccessScore"`
	ClefRelevanceScore float64 `json:"clefRelevanceScore,omitempty"` // 0.0 - 1.0 from Cloudflare Clef
	ClefLabel          string  `json:"clefLabel,omitempty"`          // e.g. "directly_relevant"
	Rank               int     `json:"rank"`
	IsSelected         bool    `json:"isSelected"`
	Reason             string  `json:"reason,omitempty"`
}

// ChatPaperSelectionData captures paper evaluation and automated selection metadata.
type ChatPaperSelectionData struct {
	Approach       ResearchApproach          `json:"approach"`
	TotalRetrieved int                       `json:"totalRetrieved"`
	SelectedCount  int                       `json:"selectedCount"`
	SelectedPapers []retrieval.PaperMetadata `json:"selectedPapers"`
	Scores         []PaperScoreInfo          `json:"scores,omitempty"`
}

// ThinkingStep captures fine-grained, Consensus-style reasoning and literature milestones.
type ThinkingStep struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "thought", "search", "read_abstract", "read_fulltext", "references", "citations", "deconstruct", "audit"
	Label       string `json:"label,omitempty"`
	Description string `json:"description"`
	Target      string `json:"target,omitempty"`
	Count       any    `json:"count,omitempty"`
	Status      string `json:"status"` // "pending", "in_progress", "completed", "error"
}

// StatsSummary captures top-level quantitative literature metrics (Retrieved, Eligible, Included).
type StatsSummary struct {
	Retrieved any `json:"retrieved"`
	Eligible  any `json:"eligible"`
	Included  any `json:"included"`
}

// ChatMessage represents a discrete, persistent turn in a research chat history thread.
type ChatMessage struct {
	ID                 string                  `json:"id"`
	ChatID             string                  `json:"chatId"`
	Role               ChatMessageRole         `json:"role"`
	Content            string                  `json:"content"`
	Status             ChatMessageStatus       `json:"status"`
	Stage              Stage                   `json:"stage,omitempty"`
	Steps              []string                `json:"steps,omitempty"`
	ThinkingSteps      []ThinkingStep          `json:"thinkingSteps,omitempty"`
	StatsSummary       *StatsSummary           `json:"statsSummary,omitempty"`
	Citations          []ChatCitationRef       `json:"citations,omitempty"`
	Quote              *ChatQuote              `json:"quote,omitempty"`
	Checkpoint         *ChatCheckpointData     `json:"checkpoint,omitempty"`
	PaperSelection     *ChatPaperSelectionData `json:"paperSelection,omitempty"`
	SuggestedFollowUps []string                `json:"suggestedFollowUps,omitempty"`
	ErrorMessage       string                  `json:"errorMessage,omitempty"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
}

// PipelineStageData encapsulates state and results across all 5 research milestones.
type PipelineStageData struct {
	Decomposition   *exploration.ConceptDecomposition    `json:"decomposition,omitempty"`
	CandidatePapers []retrieval.PaperMetadata            `json:"candidatePapers,omitempty"`
	SelectedPapers  []retrieval.PaperMetadata            `json:"selectedPapers,omitempty"`
	PaperScores     []PaperScoreInfo                     `json:"paperScores,omitempty"`
	AnchorPaper     *retrieval.PaperMetadata             `json:"anchorPaper,omitempty"`
	IngestedDoc     *retrieval.PdfDocument               `json:"ingestedDoc,omitempty"`
	Graph           *graph.CitationGraph                 `json:"graph,omitempty"`
	Synthesis       *extraction.GroundedConceptSynthesis `json:"synthesis,omitempty"`
}

// ResearchPipeline represents the central stateful orchestrator aggregate and chat history session.
type ResearchPipeline struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	InitialPrompt    string            `json:"initialPrompt"`
	Model            string            `json:"model,omitempty"`
	CurrentStage     Stage             `json:"currentStage"`
	Status           Status            `json:"status"`
	Mode             ExecutionMode     `json:"mode"`
	Approach         ResearchApproach  `json:"approach"`
	Checkpoint       CheckpointType    `json:"checkpoint"`
	SelectedAngleID  string            `json:"selectedAngleId,omitempty"`
	AnchorPaperID    string            `json:"anchorPaperId,omitempty"`
	SelectedPaperIDs []string          `json:"selectedPaperIds,omitempty"`
	StageData        PipelineStageData `json:"stageData"`
	Messages         []ChatMessage     `json:"messages"`
	Pinned           bool              `json:"pinned"`
	IsArchived       bool              `json:"isArchived"`
	ErrorMessage     string            `json:"errorMessage,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

// NewResearchPipeline creates a new research pipeline session aggregate.
func NewResearchPipeline(id, prompt string, mode ExecutionMode, approach ...ResearchApproach) (*ResearchPipeline, error) {
	if prompt == "" {
		return nil, ErrEmptyPrompt
	}
	if mode == "" {
		mode = ModeAutonomous
	}
	appr := ApproachNormal
	if len(approach) > 0 && approach[0] != "" {
		appr = approach[0]
	}

	title := prompt
	if len(title) > 60 {
		title = title[:60] + "..."
	}

	now := time.Now().UTC()
	initMsg := ChatMessage{
		ID:        id + "-msg-user-init",
		ChatID:    id,
		Role:      RoleUser,
		Content:   prompt,
		Status:    MessageStatusDone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return &ResearchPipeline{
		ID:            id,
		Title:         title,
		InitialPrompt: prompt,
		CurrentStage:  StageExploration,
		Status:        StatusDraft,
		Mode:          mode,
		Approach:      appr,
		Checkpoint:    CheckpointNone,
		Messages:      []ChatMessage{initMsg},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (p *ResearchPipeline) AddMessage(msg ChatMessage) {
	if msg.ChatID == "" {
		msg.ChatID = p.ID
	}
	now := time.Now().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = now
	}
	p.Messages = append(p.Messages, msg)
	p.UpdatedAt = now
}

func (p *ResearchPipeline) TransitionTo(stage Stage, status Status, checkpoint CheckpointType) error {
	if p.Status == StatusCompleted || p.Status == StatusCancelled {
		return ErrPipelineAlreadyFinished
	}

	p.CurrentStage = stage
	p.Status = status
	p.Checkpoint = checkpoint
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *ResearchPipeline) Fail(err error) {
	p.Status = StatusFailed
	if err != nil {
		p.ErrorMessage = err.Error()
	}
	p.UpdatedAt = time.Now().UTC()
}

func (p *ResearchPipeline) Cancel() {
	p.Status = StatusCancelled
	p.UpdatedAt = time.Now().UTC()
}

func (p *ResearchPipeline) Complete(synthesis *extraction.GroundedConceptSynthesis) {
	p.CurrentStage = StageCompleted
	p.Status = StatusCompleted
	p.Checkpoint = CheckpointNone
	p.StageData.Synthesis = synthesis
	p.UpdatedAt = time.Now().UTC()
}

// Clone returns an independent shallow-to-medium clone of the pipeline aggregate to prevent data races.
func (p *ResearchPipeline) Clone() *ResearchPipeline {
	if p == nil {
		return nil
	}
	cp := *p
	if p.SelectedPaperIDs != nil {
		cp.SelectedPaperIDs = make([]string, len(p.SelectedPaperIDs))
		copy(cp.SelectedPaperIDs, p.SelectedPaperIDs)
	}
	if p.Messages != nil {
		cp.Messages = make([]ChatMessage, len(p.Messages))
		copy(cp.Messages, p.Messages)
	}
	return &cp
}
