package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported by client writer")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher.Flush()

	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}, nil
}

func (s *SSEWriter) WriteEvent(event string, data any) error {
	var payloadBytes []byte
	switch v := data.(type) {
	case string:
		payloadBytes = []byte(v)
	case []byte:
		payloadBytes = v
	default:
		var err error
		payloadBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal sse payload failed: %w", err)
		}
	}

	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, string(payloadBytes)); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *SSEWriter) WriteStatus(status string) error {
	return s.WriteEvent("status", map[string]string{"status": status})
}

func (s *SSEWriter) WriteToken(token string) error {
	return s.WriteEvent("token", map[string]string{"token": token})
}

func (s *SSEWriter) WriteDone(data any) error {
	return s.WriteEvent("done", data)
}

func (s *SSEWriter) WriteError(errMessage string) error {
	return s.WriteEvent("error", map[string]string{"error": errMessage})
}
