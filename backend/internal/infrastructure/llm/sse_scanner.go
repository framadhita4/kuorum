package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
)

type SSEScanner struct {
	reader *bufio.Reader
	closer io.Closer
}

func NewSSEScanner(body io.ReadCloser) *SSEScanner {
	return &SSEScanner{
		reader: bufio.NewReader(body),
		closer: body,
	}
}

// StreamTokens reads SSE lines until EOF or [DONE], piping tokens to tokenChan and errors to errChan.
func (s *SSEScanner) StreamTokens(ctx context.Context, tokenChan chan<- string, errChan chan<- error) {
	defer func() {
		_ = s.closer.Close()
		close(tokenChan)
		close(errChan)
	}()

	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}

		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				errChan <- err
			}
			return
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		lineStr := string(line)
		if !strings.HasPrefix(lineStr, "data:") {
			continue
		}

		dataPayload := strings.TrimSpace(strings.TrimPrefix(lineStr, "data:"))
		if dataPayload == "[DONE]" {
			return
		}

		var chunk ChatCompletionResponse
		if err := json.Unmarshal([]byte(dataPayload), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case tokenChan <- choice.Delta.Content:
				}
			}
		}
	}
}
