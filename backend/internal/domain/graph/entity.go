package graph

import (
	"errors"
	"time"
)

var (
	ErrGraphNotFound = errors.New("citation graph not found")
	ErrInvalidNode   = errors.New("invalid graph node")
)

type NodeType string

const (
	NodeAnchor                NodeType = "ANCHOR"
	NodeSelectedStudy         NodeType = "SELECTED_STUDY"
	NodeFoundationalReference NodeType = "FOUNDATIONAL_REFERENCE"
	NodePracticalApplication  NodeType = "PRACTICAL_APPLICATION"
)

type GraphNode struct {
	PaperID       string   `json:"paperId"`
	Title         string   `json:"title"`
	Year          int      `json:"year"`
	CitationCount int      `json:"citationCount"`
	Role          NodeType `json:"role"`
	InDegree      int      `json:"inDegree"`
	OutDegree     int      `json:"outDegree"`
}

type GraphEdge struct {
	SourcePaperID string `json:"sourcePaperId"`
	TargetPaperID string `json:"targetPaperId"`
	Intent        string `json:"intent,omitempty"`
}

type CitationGraph struct {
	AnchorPaperID string      `json:"anchorPaperId"`
	Nodes         []GraphNode `json:"nodes"`
	Edges         []GraphEdge `json:"edges"`
	TraversedAt   time.Time   `json:"traversedAt"`
}
