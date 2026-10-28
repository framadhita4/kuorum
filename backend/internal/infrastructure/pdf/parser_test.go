package pdf

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"kuorum/backend/internal/domain/retrieval"
)

func TestCleanPageText(t *testing.T) {
	raw := "Ragel State Machine Compiler User Guide\n" +
		"1 / 50\n\n" +
		"In today's computing landscape, regular expressions are used heavily for the\n" +
		"purpose of specifying parsers. They are normally treated as black boxes, lin-\n" +
		"ked together with program logic.\n\n" +
		"Page 1 of 50"

	cleaned := CleanPageText(raw, 1)

	if strings.Contains(cleaned, "1 / 50") {
		t.Errorf("expected running header '1 / 50' to be removed, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "Page 1 of 50") {
		t.Errorf("expected running footer 'Page 1 of 50' to be removed, got: %s", cleaned)
	}
	if !strings.Contains(cleaned, "linked together with program logic") {
		t.Errorf("expected de-hyphenated word 'linked', got: %s", cleaned)
	}
}

func TestSectionizePages(t *testing.T) {
	pages := []rawPage{
		{
			pageNumber: 1,
			text: "Attention Is All You Need\n\n" +
				"Ashish Vaswani, Noam Shazeer, Niki Parmar\n\n" +
				"Abstract\n\n" +
				"The dominant sequence transduction models are based on complex recurrent or convolutional neural networks.\n\n" +
				"1. Introduction\n\n" +
				"Recurrent neural networks, long short-term memory and gated recurrent neural networks in particular, have been firmly established as state of the art approaches in sequence modeling.",
		},
		{
			pageNumber: 2,
			text: "2. Related Work\n\n" +
				"The goal of reducing sequential computation also forms the foundation of the Extended Neural GPU, ByteNet and ConvS2S.\n\n" +
				"3. Model Architecture\n\n" +
				"Most competitive neural sequence transduction models have an encoder-decoder structure. Here, the encoder maps an input sequence of symbol representations.",
		},
		{
			pageNumber: 3,
			text: "4. Results\n\n" +
				"On the WMT 2014 English-to-German translation task, the big transformer model outperforms the best previously reported models.\n\n" +
				"5. Conclusion\n\n" +
				"In this work, we presented the Transformer, the first sequence transduction model based entirely on attention.\n\n" +
				"References\n\n" +
				"[1] Jimmy Lei Ba, Jamie Ryan Kiros, and Geoffrey E Hinton. Layer normalization. arXiv preprint arXiv:1607.06450, 2016.\n\n" +
				"[2] Dzmitry Bahdanau, Kyunghyun Cho, and Yoshua Bengio. Neural machine translation by jointly learning to align and translate. ICLR, 2015. doi:10.1000/182",
		},
	}

	sections := SectionizePages(pages)

	expectedOrder := []retrieval.SectionType{
		retrieval.SectionOther, // Title & preamble
		retrieval.SectionAbstract,
		retrieval.SectionIntroduction,
		retrieval.SectionRelatedWork,
		retrieval.SectionMethodology,
		retrieval.SectionResults,
		retrieval.SectionConclusion,
		retrieval.SectionReferences,
	}

	if len(sections) != len(expectedOrder) {
		t.Fatalf("expected %d sections, got %d", len(expectedOrder), len(sections))
	}

	for i, expectedType := range expectedOrder {
		if sections[i].Type != expectedType {
			t.Errorf("section %d: expected type %s, got %s (title: %q)", i, expectedType, sections[i].Type, sections[i].Title)
		}
	}

	// Verify StartPage and EndPage
	refsSection := sections[7]
	if refsSection.StartPage != 3 || refsSection.EndPage != 3 {
		t.Errorf("expected references on page 3, got start=%d end=%d", refsSection.StartPage, refsSection.EndPage)
	}
}

func TestParseReferences(t *testing.T) {
	rawRefs := `[1] Ashish Vaswani, Noam Shazeer, and Niki Parmar. "Attention is all you need." In NeurIPS, 2017. doi:10.1145/3065386

[2] Jacob Devlin, Ming-Wei Chang, Kenton Lee, and Kristina Toutanova. "BERT: Pre-training of deep bidirectional transformers." arXiv:1810.04805, 2018.`

	refs := ParseReferences(rawRefs)

	if len(refs) != 2 {
		t.Fatalf("expected 2 parsed references, got %d", len(refs))
	}

	// Check ref 1
	if refs[0].Year != 2017 {
		t.Errorf("expected year 2017, got %d", refs[0].Year)
	}
	if refs[0].DOI != "10.1145/3065386" {
		t.Errorf("expected DOI '10.1145/3065386', got %q", refs[0].DOI)
	}
	if refs[0].Title != "Attention is all you need." {
		t.Errorf("expected title 'Attention is all you need.', got %q", refs[0].Title)
	}

	// Check ref 2
	if refs[1].Year != 2018 {
		t.Errorf("expected year 2018, got %d", refs[1].Year)
	}
	if refs[1].ArXivID != "1810.04805" {
		t.Errorf("expected arXiv ID '1810.04805', got %q", refs[1].ArXivID)
	}
}

func TestGetTargetSectionsForIntent(t *testing.T) {
	doc := &retrieval.PdfDocument{
		Sections: []retrieval.ParsedSection{
			{Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Abstract summary"},
			{Type: retrieval.SectionIntroduction, Title: "1. Introduction", Content: "Intro content"},
			{Type: retrieval.SectionMethodology, Title: "2. Architecture", Content: "Method content"},
			{Type: retrieval.SectionResults, Title: "3. Experiments", Content: "Empirical results table"},
			{Type: retrieval.SectionDiscussion, Title: "4. Discussion", Content: "Discussion analysis"},
			{Type: retrieval.SectionReferences, Title: "References", Content: "[1] Ref"},
		},
	}

	// Test "results" intent
	anchor, targets := doc.GetTargetSectionsForIntent("what are the empirical results and performance?")
	if anchor == nil || anchor.Type != retrieval.SectionAbstract {
		t.Errorf("expected Abstract as anchor")
	}
	if len(targets) != 2 || targets[0].Type != retrieval.SectionResults || targets[1].Type != retrieval.SectionDiscussion {
		t.Errorf("expected Results and Discussion sections, got %d targets", len(targets))
	}

	// Test "theorems" / "definitions" intent
	_, defTargets := doc.GetTargetSectionsForIntent("what is the definition and theorem proof?")
	if len(defTargets) != 2 || defTargets[0].Type != retrieval.SectionIntroduction || defTargets[1].Type != retrieval.SectionMethodology {
		t.Errorf("expected Introduction and Methodology for definitions intent, got %d targets", len(defTargets))
	}

	// Test "mechanism" intent
	_, mechTargets := doc.GetTargetSectionsForIntent("how does the algorithm and system architecture work?")
	if len(mechTargets) != 1 || mechTargets[0].Type != retrieval.SectionMethodology {
		t.Errorf("expected Methodology for mechanism intent, got %d targets", len(mechTargets))
	}
}

func TestPopplerPdfParser_RealSystemPdf(t *testing.T) {
	samplePdf := "/usr/share/doc/ragel/ragel-guide.pdf"
	if _, err := os.Stat(samplePdf); os.IsNotExist(err) {
		t.Skip("sample system PDF not found; skipping integration test")
	}

	parser := NewPopplerPdfParser()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := parser.ParseFile(ctx, samplePdf)
	if err != nil {
		t.Fatalf("failed to parse sample PDF: %v", err)
	}

	if doc.TotalPages == 0 {
		t.Errorf("expected total pages > 0, got %d", doc.TotalPages)
	}
	if len(doc.Sections) == 0 {
		t.Errorf("expected parsed sections > 0, got %d", len(doc.Sections))
	}
	if doc.RawTextLength == 0 {
		t.Errorf("expected non-empty raw text length")
	}

	t.Logf("Successfully parsed %q: TotalPages=%d, Sections=%d, RawLength=%d chars, Title=%q",
		samplePdf, doc.TotalPages, len(doc.Sections), doc.RawTextLength, doc.Title)
}

func TestUnspaceKerning(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"A B S T R A C T", "ABSTRACT"},
		{"A B S T R A K", "ABSTRAK"},
		{"A R T I C L E   I N F O", "ARTICLE INFO"},
		{"1. I N T R O D U C T I O N", "1. INTRODUCTION"},
		{"IV. E X P E R I M E N T S", "IV. EXPERIMENTS"},
		{"1. Introduction", "1. Introduction"},
		{"3. Model Architecture", "3. Model Architecture"},
	}

	for _, c := range cases {
		actual := UnspaceKerning(c.input)
		if actual != c.expected {
			t.Errorf("UnspaceKerning(%q) = %q, expected %q", c.input, actual, c.expected)
		}
	}
}

func TestJebavLayout_AsymmetricColumnsAndDualAbstract(t *testing.T) {
	testPdf := "/tmp/test_jebav.pdf"
	if _, err := os.Stat(testPdf); os.IsNotExist(err) {
		t.Skip("synthetic JEBAV test PDF not found; skipping integration test")
	}

	parser := NewPopplerPdfParser()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := parser.ParseFile(ctx, testPdf)
	if err != nil {
		t.Fatalf("failed to parse JEBAV layout PDF: %v", err)
	}

	// 1. Verify DOI extracted from Article Info
	if doc.DOI != "10.14414/jebav.v20i2.1126" {
		t.Errorf("expected DOI '10.14414/jebav.v20i2.1126', got %q", doc.DOI)
	}

	// 2. Verify Abstract section contains both English and Indonesian texts
	absSec := doc.GetSection(retrieval.SectionAbstract)
	if absSec == nil {
		t.Fatalf("expected Abstract section to be found")
	}
	if !strings.Contains(absSec.Content, "The study tried to analyze") {
		t.Errorf("expected English abstract text in Abstract section, got: %s", absSec.Content)
	}
	if !strings.Contains(absSec.Content, "Penelitian ini bertujuan") {
		t.Errorf("expected Indonesian abstract text in Abstract section, got: %s", absSec.Content)
	}

	// 3. Verify Introduction section is NOT horizontally scrambled
	introSec := doc.GetSection(retrieval.SectionIntroduction)
	if introSec == nil {
		t.Fatalf("expected Introduction section to be found")
	}
	if strings.Contains(introSec.Content, "veEntering the 21st century") ||
		strings.Contains(introSec.Content, "hicle burning remains, and the opening of green") && strings.Contains(introSec.Content, "used to be easily controlled, it turns into an uncon-") && strings.Index(introSec.Content, "used to be") > strings.Index(introSec.Content, "hicle burning") {
		t.Errorf("Introduction text suffered from column interleaving corruption:\n%s", introSec.Content)
	}

	if !strings.Contains(introSec.Content, "Entering the 21st century") {
		t.Errorf("expected Introduction to contain opening paragraph, got: %s", introSec.Content)
	}
	if !strings.Contains(introSec.Content, "energy-saving movements") {
		t.Errorf("expected Introduction to contain second column text, got: %s", introSec.Content)
	}
}
