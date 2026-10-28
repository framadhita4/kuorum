package build_graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/domain/retrieval"
)

type BuildCitationGraphCommandHandler struct {
	paperApi  interfaces.IPaperApi
	graphRepo interfaces.IGraphRepository
	paperRepo interfaces.IPaperRepository
	cache     interfaces.ICacheRepository
}

func NewBuildCitationGraphCommandHandler(
	paperApi interfaces.IPaperApi,
	graphRepo interfaces.IGraphRepository,
	paperRepo interfaces.IPaperRepository,
	cache ...interfaces.ICacheRepository,
) *BuildCitationGraphCommandHandler {
	h := &BuildCitationGraphCommandHandler{
		paperApi:  paperApi,
		graphRepo: graphRepo,
		paperRepo: paperRepo,
	}
	if len(cache) > 0 {
		h.cache = cache[0]
	}
	return h
}

func (h *BuildCitationGraphCommandHandler) Handle(ctx context.Context, cmd BuildCitationGraphCommand) (*BuildCitationGraphResult, error) {
	maxRefs := cmd.MaxReferences
	if maxRefs <= 0 {
		maxRefs = 10
	}
	maxCites := cmd.MaxCitations
	if maxCites <= 0 {
		maxCites = 10
	}

	// 1. Check existing graph cache unless ForceRefresh is requested
	if !cmd.ForceRefresh && h.graphRepo != nil {
		if existing, err := h.graphRepo.GetGraph(ctx, cmd.AnchorPaperID); err == nil && existing != nil && len(existing.Nodes) > 0 {
			return mapDomainToResult(existing, true), nil
		}
	}

	// 2. Fetch Anchor Paper metadata
	var anchorPaper *retrieval.PaperMetadata
	if h.paperRepo != nil {
		anchorPaper, _ = h.paperRepo.GetPaper(ctx, cmd.AnchorPaperID)
	}
	if anchorPaper == nil && h.paperApi != nil {
		var err error
		anchorPaper, err = h.paperApi.GetPaperDetails(ctx, cmd.AnchorPaperID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch anchor paper details: %w", err)
		}
	}
	if anchorPaper == nil {
		anchorPaper = &retrieval.PaperMetadata{
			ID:    cmd.AnchorPaperID,
			Title: "Anchor Paper (" + cmd.AnchorPaperID + ")",
			Year:  time.Now().Year(),
		}
	}

	// 3. Fetch Backward References (Foundational Theories)
	var rawRefs []retrieval.PaperMetadata
	if h.paperApi != nil {
		rawRefs, _ = h.paperApi.GetReferences(ctx, cmd.AnchorPaperID)
	}

	// Sort references by citation count descending to capture most influential baselines
	sort.SliceStable(rawRefs, func(i, j int) bool {
		return rawRefs[i].CitationCount > rawRefs[j].CitationCount
	})
	if len(rawRefs) > maxRefs {
		rawRefs = rawRefs[:maxRefs]
	}

	// 4. Fetch Forward Citations (Downstream Applications)
	var rawCites []retrieval.PaperMetadata
	if h.paperApi != nil {
		rawCites, _ = h.paperApi.GetCitations(ctx, cmd.AnchorPaperID)
	}

	// Sort citations by recency then citation count
	sort.SliceStable(rawCites, func(i, j int) bool {
		if rawCites[i].Year == rawCites[j].Year {
			return rawCites[i].CitationCount > rawCites[j].CitationCount
		}
		return rawCites[i].Year > rawCites[j].Year
	})
	if len(rawCites) > maxCites {
		rawCites = rawCites[:maxCites]
	}

	// 5. Construct Graph Topology (Nodes & Directed Edges)
	nodeMap := make(map[string]*graph.GraphNode)
	var edges []graph.GraphEdge

	// Anchor Node
	anchorNode := &graph.GraphNode{
		PaperID:       anchorPaper.ID,
		Title:         anchorPaper.Title,
		Year:          anchorPaper.Year,
		CitationCount: anchorPaper.CitationCount,
		Role:          graph.NodeAnchor,
	}
	nodeMap[anchorPaper.ID] = anchorNode

	// Foundational Nodes & Edges (Anchor -> Reference)
	for _, ref := range rawRefs {
		if ref.ID == "" || ref.ID == anchorPaper.ID {
			continue
		}
		nodeMap[ref.ID] = &graph.GraphNode{
			PaperID:       ref.ID,
			Title:         ref.Title,
			Year:          ref.Year,
			CitationCount: ref.CitationCount,
			Role:          graph.NodeFoundationalReference,
		}
		edges = append(edges, graph.GraphEdge{
			SourcePaperID: anchorPaper.ID,
			TargetPaperID: ref.ID,
			Intent:        "foundational_reference",
		})
	}

	// Application Nodes & Edges (CitingPaper -> Anchor)
	for _, cite := range rawCites {
		if cite.ID == "" || cite.ID == anchorPaper.ID {
			continue
		}
		if _, exists := nodeMap[cite.ID]; !exists {
			nodeMap[cite.ID] = &graph.GraphNode{
				PaperID:       cite.ID,
				Title:         cite.Title,
				Year:          cite.Year,
				CitationCount: cite.CitationCount,
				Role:          graph.NodePracticalApplication,
			}
		}
		edges = append(edges, graph.GraphEdge{
			SourcePaperID: cite.ID,
			TargetPaperID: anchorPaper.ID,
			Intent:        "practical_application",
		})
	}

	// Selected Cohort Nodes & Exploration Relationships
	for _, cohortID := range cmd.CohortPaperIDs {
		cleanCohortID := strings.TrimSpace(cohortID)
		if cleanCohortID == "" || cleanCohortID == anchorPaper.ID {
			continue
		}

		var cohortMeta *retrieval.PaperMetadata
		if h.paperRepo != nil {
			cohortMeta, _ = h.paperRepo.GetPaper(ctx, cleanCohortID)
		}
		if cohortMeta == nil && h.paperApi != nil {
			cohortMeta, _ = h.paperApi.GetPaperDetails(ctx, cleanCohortID)
		}
		if cohortMeta == nil {
			cohortMeta = &retrieval.PaperMetadata{
				ID:    cleanCohortID,
				Title: "Selected Study (" + cleanCohortID + ")",
				Year:  time.Now().Year(),
			}
		}

		if existingNode, exists := nodeMap[cleanCohortID]; !exists {
			nodeMap[cleanCohortID] = &graph.GraphNode{
				PaperID:       cohortMeta.ID,
				Title:         cohortMeta.Title,
				Year:          cohortMeta.Year,
				CitationCount: cohortMeta.CitationCount,
				Role:          graph.NodeSelectedStudy,
			}
		} else if existingNode.Role != graph.NodeAnchor {
			existingNode.Role = graph.NodeSelectedStudy
		}

		// Ensure citation connectivity to the Anchor or Foundational Baselines
		hasDirectEdge := false
		for _, e := range edges {
			if (e.SourcePaperID == anchorPaper.ID && e.TargetPaperID == cleanCohortID) ||
				(e.SourcePaperID == cleanCohortID && e.TargetPaperID == anchorPaper.ID) {
				hasDirectEdge = true
				break
			}
		}

		if !hasDirectEdge {
			if cohortMeta.Year > 0 && anchorPaper.Year > 0 && cohortMeta.Year < anchorPaper.Year {
				edges = append(edges, graph.GraphEdge{
					SourcePaperID: anchorPaper.ID,
					TargetPaperID: cleanCohortID,
					Intent:        "foundational_reference",
				})
			} else {
				edges = append(edges, graph.GraphEdge{
					SourcePaperID: cleanCohortID,
					TargetPaperID: anchorPaper.ID,
					Intent:        "selected_cohort",
				})
			}
		}
	}

	// Compute In-Degree & Out-Degree
	for _, edge := range edges {
		if src, ok := nodeMap[edge.SourcePaperID]; ok {
			src.OutDegree++
		}
		if tgt, ok := nodeMap[edge.TargetPaperID]; ok {
			tgt.InDegree++
		}
	}

	// Flatten nodes slice
	var nodes []graph.GraphNode
	for _, n := range nodeMap {
		nodes = append(nodes, *n)
	}

	// Sort nodes: Anchor first, then foundational, then applications
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Role == graph.NodeAnchor {
			return true
		}
		if nodes[j].Role == graph.NodeAnchor {
			return false
		}
		return nodes[i].CitationCount > nodes[j].CitationCount
	})

	citationGraph := &graph.CitationGraph{
		AnchorPaperID: cmd.AnchorPaperID,
		Nodes:         nodes,
		Edges:         edges,
		TraversedAt:   time.Now().UTC(),
	}

	// 6. Persist to Graph Repository
	if h.graphRepo != nil {
		_ = h.graphRepo.SaveGraph(ctx, citationGraph)
	}

	return mapDomainToResult(citationGraph, false), nil
}

func mapDomainToResult(g *graph.CitationGraph, fromCache bool) *BuildCitationGraphResult {
	foundational := 0
	apps := 0
	cohort := 0
	for _, n := range g.Nodes {
		switch n.Role {
		case graph.NodeFoundationalReference:
			foundational++
		case graph.NodePracticalApplication:
			apps++
		case graph.NodeSelectedStudy:
			cohort++
		}
	}

	return &BuildCitationGraphResult{
		AnchorPaperID:     g.AnchorPaperID,
		Nodes:             g.Nodes,
		Edges:             g.Edges,
		TotalNodes:        len(g.Nodes),
		TotalEdges:        len(g.Edges),
		FoundationalCount: foundational,
		ApplicationCount:  apps,
		CohortCount:       cohort,
		FromCache:         fromCache,
	}
}
