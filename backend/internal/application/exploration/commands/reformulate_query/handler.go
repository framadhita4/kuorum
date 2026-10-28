package reformulate_query

import (
	"context"
	"fmt"
	"strings"

	"kuorum/backend/internal/application/common/interfaces"
)

type ReformulateQueryCommandHandler struct {
	llmClient interfaces.ILLMClient
}

func NewReformulateQueryCommandHandler(llmClient interfaces.ILLMClient) *ReformulateQueryCommandHandler {
	return &ReformulateQueryCommandHandler{
		llmClient: llmClient,
	}
}

type llmReformulationOutput struct {
	Keywords     []string `json:"keywords"`
	Synonyms     []string `json:"synonyms"`
	BooleanQuery string   `json:"booleanQuery"`
	DenseQuery   string   `json:"denseQuery"`
}

func (h *ReformulateQueryCommandHandler) Handle(ctx context.Context, cmd ReformulateQueryCommand) (*ReformulateQueryResult, error) {
	prompt := strings.TrimSpace(cmd.Prompt)
	words := strings.Fields(prompt)

	systemPrompt := `You are an academic retrieval specialist.
Convert the given informal or colloquial query into standardized English academic search components:
1. keywords: 3-5 core academic terms.
2. synonyms: 3-5 academic synonyms.
3. booleanQuery: Boolean search string with AND/OR operators.
4. denseQuery: A rich contextual paragraph for dense vector embedding search.

You must respond strictly with a valid JSON object adhering to this schema:
{
  "keywords": ["core academic term 1", "core academic term 2"],
  "synonyms": ["academic synonym 1", "academic synonym 2"],
  "booleanQuery": "keyword1 AND (keyword2 OR synonym1)",
  "denseQuery": "Contextual paragraph describing the academic topic for dense vector search"
}`

	var out llmReformulationOutput
	var err error
	if h.llmClient != nil {
		err = h.llmClient.CompleteStructured(ctx, systemPrompt, prompt, &out)
	}

	if h.llmClient == nil || err != nil || len(out.Keywords) == 0 {
		out = llmReformulationOutput{
			Keywords:     words,
			Synonyms:     []string{prompt + " paradigm", prompt + " methodology", prompt + " architecture"},
			BooleanQuery: fmt.Sprintf("(%s) AND (methodology OR architecture OR evaluation)", strings.Join(words, " AND ")),
			DenseQuery:   fmt.Sprintf("Academic research investigation focusing on %s, analyzing theoretical baselines and empirical applications.", prompt),
		}
	}

	return &ReformulateQueryResult{
		OriginalPrompt: prompt,
		Keywords:       out.Keywords,
		Synonyms:       out.Synonyms,
		BooleanQuery:   out.BooleanQuery,
		DenseQuery:     out.DenseQuery,
	}, nil
}
