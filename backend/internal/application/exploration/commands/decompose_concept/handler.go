package decompose_concept

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/exploration"
)

type DecomposeConceptCommandHandler struct {
	llmClient interfaces.ILLMClient
	repo      interfaces.IExplorationRepository
}

func NewDecomposeConceptCommandHandler(llmClient interfaces.ILLMClient, repo interfaces.IExplorationRepository) *DecomposeConceptCommandHandler {
	return &DecomposeConceptCommandHandler{
		llmClient: llmClient,
		repo:      repo,
	}
}

type llmDecompositionOutput struct {
	NormalizedTerm string `json:"normalizedTerm"`
	Angles         []struct {
		Title            string   `json:"title"`
		Description      string   `json:"description"`
		SuggestedQueries []string `json:"suggestedQueries"`
	} `json:"angles"`
}

func (h *DecomposeConceptCommandHandler) Handle(ctx context.Context, cmd DecomposeConceptCommand) (*DecomposeConceptResult, error) {
	prompt := strings.TrimSpace(cmd.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("decompose concept failed: %w", exploration.ErrEmptyConceptPrompt)
	}

	if cmd.Model != "" {
		ctx = interfaces.WithModel(ctx, cmd.Model)
	}

	systemPrompt := `You are an expert academic research assistant.
Given a scientific concept or user query, normalize the terminology to standard international academic English.
Decompose the concept into 3 to 4 distinct research dimensions/angles (e.g., Theoretical Foundation, Algorithmic Mechanism, Industrial/Clinical Applications, Evaluation Benchmarks).
For each angle, provide a clear description and 2-3 specific suggested academic search queries.

You must respond strictly with a valid JSON object adhering to this schema:
{
  "normalizedTerm": "Standardized academic concept name",
  "angles": [
    {
      "title": "Dimension Title",
      "description": "Explanation of this angle",
      "suggestedQueries": ["academic search query 1", "academic search query 2"]
    }
  ]
}`

	var llmOutput llmDecompositionOutput
	var err error

	if h.llmClient != nil {
		err = h.llmClient.CompleteStructured(ctx, systemPrompt, prompt, &llmOutput)
	}

	// Fallback/Low-implementation generator if LLM client is nil or fails (ensures app works offline/out-of-the-box)
	if h.llmClient == nil || err != nil || len(llmOutput.Angles) == 0 {
		caser := cases.Title(language.English)
		normalized := caser.String(strings.ToLower(prompt))
		llmOutput = llmDecompositionOutput{
			NormalizedTerm: normalized,
			Angles: []struct {
				Title            string   `json:"title"`
				Description      string   `json:"description"`
				SuggestedQueries []string `json:"suggestedQueries"`
			}{
				{
					Title:       "Theoretical Foundations & Mathematical Formulation",
					Description: fmt.Sprintf("Underlying mathematical models, core theorems, and foundational principles of %s.", normalized),
					SuggestedQueries: []string{
						fmt.Sprintf("%s foundational principles", normalized),
						fmt.Sprintf("%s mathematical formulation baseline", normalized),
					},
				},
				{
					Title:       "Algorithmic Architecture & System Mechanics",
					Description: fmt.Sprintf("Detailed pipeline mechanics, state transitions, and implementation architectures for %s.", normalized),
					SuggestedQueries: []string{
						fmt.Sprintf("%s system architecture algorithm", normalized),
						fmt.Sprintf("%s computational optimization methodology", normalized),
					},
				},
				{
					Title:       "Empirical Applications & Real-World Use Cases",
					Description: fmt.Sprintf("Practical deployments, industry implementations, and case studies utilizing %s.", normalized),
					SuggestedQueries: []string{
						fmt.Sprintf("%s practical application case study", normalized),
						fmt.Sprintf("%s industry deployment benchmark", normalized),
					},
				},
				{
					Title:       "Evaluation Benchmarks & Comparative Analysis",
					Description: fmt.Sprintf("State-of-the-art benchmarks, performance metrics, and comparative trade-offs for %s.", normalized),
					SuggestedQueries: []string{
						fmt.Sprintf("%s comparative benchmark analysis", normalized),
						fmt.Sprintf("%s evaluation metrics ablation", normalized),
					},
				},
			},
		}
	}

	domainAngles := make([]exploration.ResearchAngle, len(llmOutput.Angles))
	resultAngles := make([]AngleDTO, len(llmOutput.Angles))
	for i, a := range llmOutput.Angles {
		angleID := fmt.Sprintf("angle-%d", i+1)
		domainAngles[i] = exploration.ResearchAngle{
			ID:               angleID,
			Title:            a.Title,
			Description:      a.Description,
			SuggestedQueries: a.SuggestedQueries,
		}
		resultAngles[i] = AngleDTO{
			AngleID:          angleID,
			Title:            a.Title,
			Description:      a.Description,
			SuggestedQueries: a.SuggestedQueries,
		}
	}

	// Persist to repository if registered
	if h.repo != nil {
		if decomp, err := exploration.NewConceptDecomposition(prompt, llmOutput.NormalizedTerm, domainAngles); err == nil {
			_ = h.repo.SaveDecomposition(ctx, decomp)
		}
	}

	return &DecomposeConceptResult{
		Prompt:         prompt,
		NormalizedTerm: llmOutput.NormalizedTerm,
		Angles:         resultAngles,
	}, nil
}
