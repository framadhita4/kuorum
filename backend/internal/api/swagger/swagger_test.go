package swagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSwaggerHandler_ServeSpec(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/swagger.json", nil)
	rr := httptest.NewRecorder()

	h.ServeSpec(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected application/json content type, got %s", contentType)
	}

	var parsed map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("openapi spec is not valid JSON: %v", err)
	}

	if parsed["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %v", parsed["openapi"])
	}

	paths, ok := parsed["paths"].(map[string]any)
	if !ok {
		t.Fatalf("missing paths in openapi spec")
	}

	requiredPaths := []string{
		"/api/v1/health",
		"/api/v1/papers/search",
		"/api/v1/papers/providers",
		"/api/v1/papers/details",
		"/api/v1/papers/{id}",
		"/api/v1/papers/references",
		"/api/v1/papers/citations",
		"/api/v1/papers/upload-pdf",
		"/api/v1/papers/{id}/sections",
		"/api/v1/papers/{id}/pdf",
		"/api/v1/exploration/decompose",
		"/api/v1/exploration/stream",
		"/api/v1/exploration/reformulate",
		"/api/v1/exploration/angles",
		"/api/v1/models",
		"/api/v1/exploration/models",
	}

	for _, p := range requiredPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("expected path %s in OpenAPI spec", p)
		}
	}
}

func TestSwaggerHandler_ServeUI(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rr := httptest.NewRecorder()

	h.ServeUI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected SwaggerUIBundle in HTML body")
	}
	if !strings.Contains(body, "/api/v1/swagger.json") {
		t.Errorf("expected spec url /api/v1/swagger.json in HTML body")
	}
}
