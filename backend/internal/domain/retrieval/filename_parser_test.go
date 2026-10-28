package retrieval

import (
	"errors"
	"strings"
	"testing"
)

func TestFilenameParser_Parse(t *testing.T) {
	parser := NewFilenameParser()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		targetErr   error
	}{
		{
			name:     "clean simple pdf",
			input:    "attention.pdf",
			expected: "attention",
		},
		{
			name:     "spaces in filename",
			input:    "Attention Is All You Need.pdf",
			expected: "attention-is-all-you-need",
		},
		{
			name:     "complex title with punctuation and symbols",
			input:    "BERT: Pre-training of Deep Bidirectional Transformers (2018).pdf",
			expected: "bert-pre-training-of-deep-bidirectional-transformers-2018",
		},
		{
			name:     "punctuation, colons, question marks, and parentheses",
			input:    "My Paper (2024): Part 1 - Final?.pdf",
			expected: "my-paper-2024-part-1-final",
		},
		{
			name:     "dots and underscores",
			input:    "deep_residual_learning_v1.0.pdf",
			expected: "deep_residual_learning_v1.0",
		},
		{
			name:     "arxiv identifier filename",
			input:    "1706.03762.pdf",
			expected: "1706.03762",
		},
		{
			name:     "unix path traversal stripped",
			input:    "../../secret/path/sample_paper.pdf",
			expected: "sample_paper",
		},
		{
			name:     "windows path separators stripped",
			input:    "C:\\Users\\admin\\Documents\\Research\\gnn_survey.pdf",
			expected: "gnn_survey",
		},
		{
			name:     "windows reserved device name con",
			input:    "con.pdf",
			expected: "paper-con",
		},
		{
			name:     "windows reserved device name aux",
			input:    "AUX.PDF",
			expected: "paper-aux",
		},
		{
			name:     "unicode accents normalized",
			input:    "Schrödinger_Equation_1926.pdf",
			expected: "schrodinger_equation_1926",
		},
		{
			name:     "excessive whitespace and symbols",
			input:    "   spaced   filename   ---   with   symbols   $$$   .pdf",
			expected: "spaced-filename-with-symbols",
		},
		{
			name:     "uppercase extension stripped",
			input:    "TRANSFORMERS.PDF",
			expected: "transformers",
		},
		{
			name:     "no extension provided",
			input:    "standalone-paper",
			expected: "standalone-paper",
		},
		{
			name:        "empty filename returns error",
			input:       "",
			expectError: true,
			targetErr:   ErrEmptyFilename,
		},
		{
			name:        "whitespace only filename returns error",
			input:       "    ",
			expectError: true,
			targetErr:   ErrEmptyFilename,
		},
		{
			name:        "only dot extension returns error",
			input:       ".pdf",
			expectError: true,
			targetErr:   ErrInvalidPaperID,
		},
		{
			name:        "only unsupported symbols returns error",
			input:       "???###!@#.pdf",
			expectError: true,
			targetErr:   ErrInvalidPaperID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil (result: %q)", tt.input, got)
				}
				if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
					t.Errorf("expected error %v, got %v", tt.targetErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for input %q: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("input %q => got %q, expected %q", tt.input, got, tt.expected)
				}
			}
		})
	}
}

func TestFilenameParser_MaxLength(t *testing.T) {
	parser := NewFilenameParser()

	// 200 character repetitive title
	longInput := strings.Repeat("neural-networks-", 15) + ".pdf"
	got, err := parser.Parse(longInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > 120 {
		t.Errorf("expected max length <= 120, got %d (%q)", len(got), got)
	}
	if strings.HasSuffix(got, "-") || strings.HasSuffix(got, "_") || strings.HasSuffix(got, ".") {
		t.Errorf("expected clean trailing characters, got %q", got)
	}
}

func TestPackageLevelConvenience(t *testing.T) {
	id, err := ParsePaperID("Attention Is All You Need.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "attention-is-all-you-need" {
		t.Errorf("got %q, expected %q", id, "attention-is-all-you-need")
	}

	sanitized := SanitizePaperID("Deep_Residual_Learning (2015).pdf")
	if sanitized != "deep_residual_learning-2015" {
		t.Errorf("got %q, expected %q", sanitized, "deep_residual_learning-2015")
	}
}
