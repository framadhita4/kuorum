package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"

	"golang.org/x/sync/errgroup"
)

// ClefClassifier implements interfaces.IRapidClassifier utilizing Cloudflare Workers AI Clef
// decision models with an ultra-fast in-memory semantic decision fallback.
type ClefClassifier struct {
	accountID  string
	apiToken   string
	model      string
	httpClient *http.Client
}

// NewClefClassifier instantiates a new Clef rapid classification client.
func NewClefClassifier(accountID, apiToken, model string, httpClient ...*http.Client) *ClefClassifier {
	if model == "" {
		model = "@cf/cloudflare/clef-flash"
	}
	var client *http.Client
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	} else {
		client = &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	return &ClefClassifier{
		accountID:  strings.TrimSpace(accountID),
		apiToken:   strings.TrimSpace(apiToken),
		model:      model,
		httpClient: client,
	}
}

// Cloudflare Workers AI Clef Request Schema
type clefRequest struct {
	State     string                  `json:"state"`
	Questions map[string]clefQuestion `json:"questions"`
}

type clefQuestion struct {
	Type         string   `json:"type"`
	Instructions string   `json:"instructions"`
	Options      []string `json:"options,omitempty"`
}

// Cloudflare Workers AI Clef Response Schema
type clefResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
	Result struct {
		Questions struct {
			IsRelevant struct {
				True  float64 `json:"true"`
				False float64 `json:"false"`
			} `json:"is_relevant"`
			RelevanceLevel map[string]float64 `json:"relevance_level"`
		} `json:"questions"`
	} `json:"result"`
}

// ClassifyAbstract evaluates a single paper abstract against the research topic.
func (c *ClefClassifier) ClassifyAbstract(ctx context.Context, topic string, paper retrieval.PaperMetadata) (*interfaces.ClassificationResult, error) {
	// If Cloudflare credentials are configured, attempt real Cloudflare Workers AI call
	if c.accountID != "" && c.apiToken != "" {
		res, err := c.classifyWithCloudflare(ctx, topic, paper)
		if err == nil && res != nil {
			return res, nil
		}
		// Log notice and gracefully fallback to rapid local decision evaluator
		log.Printf("[ClefClassifier] Notice: Cloudflare Workers AI request failed (%v), using rapid decision fallback", err)
	}

	// High-throughput rapid semantic decision evaluator fallback
	return c.rapidDecisionFallback(topic, paper), nil
}

// classifyWithCloudflare queries the Cloudflare Workers AI REST API for the Clef model.
func (c *ClefClassifier) classifyWithCloudflare(ctx context.Context, topic string, paper retrieval.PaperMetadata) (*interfaces.ClassificationResult, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", c.accountID, c.model)

	stateText := fmt.Sprintf("Research Topic: %s\nTitle: %s\nAbstract: %s", topic, paper.Title, paper.Abstract)
	if len(stateText) > 4000 {
		stateText = stateText[:4000]
	}

	payload := clefRequest{
		State: stateText,
		Questions: map[string]clefQuestion{
			"is_relevant": {
				Type:         "boolean",
				Instructions: "Is this academic paper directly relevant to the specified research topic?",
			},
			"relevance_level": {
				Type:         "choice",
				Instructions: "Classify the depth of relevance of this paper to the research topic",
				Options:      []string{"directly_relevant", "tangentially_related", "irrelevant"},
			},
		},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal clef request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create clef http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("clef request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("clef API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var cfResp clefResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, fmt.Errorf("failed to decode clef response: %w", err)
	}

	if !cfResp.Success {
		errMsg := "unknown cloudflare error"
		if len(cfResp.Errors) > 0 {
			errMsg = cfResp.Errors[0].Message
		}
		return nil, fmt.Errorf("cloudflare clef error: %s", errMsg)
	}

	// Extract calibrated probability
	probRelevant := cfResp.Result.Questions.IsRelevant.True
	levelDirect := cfResp.Result.Questions.RelevanceLevel["directly_relevant"]
	levelTangential := cfResp.Result.Questions.RelevanceLevel["tangentially_related"]

	// Composite relevance score prioritizing direct relevance probability
	score := probRelevant
	if levelDirect > 0 {
		score = (probRelevant * 0.6) + (levelDirect * 0.4)
	}
	score = math.Round(score*10000) / 10000

	label := "directly_relevant"
	if score < 0.40 || (levelTangential > levelDirect && score < 0.65) {
		if score < 0.30 {
			label = "irrelevant"
		} else {
			label = "tangentially_related"
		}
	}

	return &interfaces.ClassificationResult{
		PaperID:        paper.ID,
		Title:          paper.Title,
		RelevanceScore: score,
		IsRelevant:     score >= 0.50,
		Confidence:     score,
		Label:          label,
		ModelUsed:      c.model,
		Reason:         fmt.Sprintf("Cloudflare Clef probability: %.2f (Level: %s)", score, label),
	}, nil
}

// BatchClassify evaluates a collection of candidate papers concurrently using an errgroup.
func (c *ClefClassifier) BatchClassify(ctx context.Context, topic string, papers []retrieval.PaperMetadata, concurrency int) ([]interfaces.ClassificationResult, error) {
	if len(papers) == 0 {
		return nil, nil
	}
	if concurrency <= 0 {
		concurrency = 10
	}

	results := make([]interfaces.ClassificationResult, len(papers))
	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, concurrency)

	var mu sync.Mutex
	for i, p := range papers {
		idx := i
		paper := p
		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			res, err := c.ClassifyAbstract(gCtx, topic, paper)
			if err != nil {
				// Fallback to local rapid decision evaluator on individual failure
				res = c.rapidDecisionFallback(topic, paper)
			}

			mu.Lock()
			results[idx] = *res
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// rapidDecisionFallback evaluates semantic alignment between topic keywords and the paper abstract.
func (c *ClefClassifier) rapidDecisionFallback(topic string, paper retrieval.PaperMetadata) *interfaces.ClassificationResult {
	topicTokens := tokenizeClassifierText(topic)
	if len(topicTokens) == 0 {
		return &interfaces.ClassificationResult{
			PaperID:        paper.ID,
			Title:          paper.Title,
			RelevanceScore: 0.50,
			IsRelevant:     true,
			Confidence:     0.50,
			Label:          "directly_relevant",
			ModelUsed:      "clef-decision-evaluator",
			Reason:         "Default baseline score for broad query",
		}
	}

	titleTokens := tokenizeClassifierText(paper.Title)
	abstractTokens := tokenizeClassifierText(paper.Abstract + " " + paper.TLDR)

	var rawScore float64
	matchCount := 0
	titleMatchCount := 0

	for _, token := range topicTokens {
		tokenWeight := 1.0
		if len(token) > 6 {
			tokenWeight = 1.3 // Longer technical/domain words carry higher semantic specificity
		}

		inTitle := containsClassifierToken(titleTokens, token)
		inAbstract := containsClassifierToken(abstractTokens, token)

		if inTitle {
			rawScore += 2.5 * tokenWeight
			matchCount++
			titleMatchCount++
		}
		if inAbstract {
			rawScore += 1.0 * tokenWeight
			matchCount++
		}
	}

	// Keyphrase & Multi-token Bonus
	if len(topicTokens) >= 2 {
		lowerTitle := strings.ToLower(paper.Title)
		lowerAbstract := strings.ToLower(paper.Abstract + " " + paper.TLDR)
		for i := 0; i < len(topicTokens)-1; i++ {
			phrase := topicTokens[i] + " " + topicTokens[i+1]
			if strings.Contains(lowerTitle, phrase) {
				rawScore += 4.0
				matchCount += 2
				titleMatchCount++
			} else if strings.Contains(lowerAbstract, phrase) {
				rawScore += 2.0
				matchCount++
			}
		}
	}

	// Normalize score with non-linear sigmoid-like saturation
	maxPossible := float64(len(topicTokens))*3.5 + 4.0
	ratio := rawScore / maxPossible
	if ratio > 1.0 {
		ratio = 1.0
	}

	// Calibrated probability curve: 0.1 baseline + sigmoid ratio
	prob := 1.0 / (1.0 + math.Exp(-7.0*(ratio-0.32)))
	prob = math.Round(prob*10000) / 10000

	label := "irrelevant"
	isRel := false
	if prob >= 0.65 || (titleMatchCount >= 2 && prob >= 0.50) {
		label = "directly_relevant"
		isRel = true
	} else if prob >= 0.38 || titleMatchCount >= 1 {
		label = "tangentially_related"
		isRel = prob >= 0.48
	}

	reason := fmt.Sprintf("Matched %d/%d key concept facets (%d in title)", matchCount, len(topicTokens)*2, titleMatchCount)

	return &interfaces.ClassificationResult{
		PaperID:        paper.ID,
		Title:          paper.Title,
		RelevanceScore: prob,
		IsRelevant:     isRel,
		Confidence:     prob,
		Label:          label,
		ModelUsed:      "clef-decision-evaluator",
		Reason:         reason,
	}
}

var classifierStopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
	"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
	"that": true, "the": true, "to": true, "was": true, "were": true, "will": true,
	"with": true, "dan": true, "di": true, "ke": true, "dari": true, "yang": true,
	"ini": true, "itu": true, "untuk": true, "pada": true, "adalah": true,
	"dengan": true, "riset": true, "paper": true, "research": true, "study": true,
}

func tokenizeClassifierText(s string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else if current.Len() > 0 {
			token := current.String()
			current.Reset()
			if len(token) > 2 && !classifierStopWords[token] {
				tokens = append(tokens, token)
			}
		}
	}
	if current.Len() > 0 {
		token := current.String()
		if len(token) > 2 && !classifierStopWords[token] {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func containsClassifierToken(tokens []string, target string) bool {
	for _, t := range tokens {
		if t == target || strings.Contains(t, target) || strings.Contains(target, t) {
			return true
		}
	}
	return false
}
