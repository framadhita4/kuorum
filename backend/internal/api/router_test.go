package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kuorum/backend/internal/api/controllers"
	"kuorum/backend/internal/api/swagger"
	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/exploration/queries/get_models"
	"kuorum/backend/internal/application/retrieval/queries/get_providers"
	"kuorum/backend/internal/domain/retrieval"
)

type mockRouterPaperApi struct{}

func (m *mockRouterPaperApi) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockRouterPaperApi) SearchPapersWithProviders(ctx context.Context, query string, limit int, providers []string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockRouterPaperApi) GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockRouterPaperApi) GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockRouterPaperApi) GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockRouterPaperApi) AvailableProviders() []string {
	return []string{"openalex", "semanticscholar", "arxiv", "crossref"}
}

func TestRouter_SwaggerAndApiEndpoints(t *testing.T) {
	med := mediator.NewInMemoryMediator()
	mock := &mockRouterPaperApi{}
	med.RegisterQueryHandler(get_providers.GetProvidersQuery{}, get_providers.NewGetProvidersQueryHandler(mock))
	med.RegisterQueryHandler(get_models.GetModelsQuery{}, get_models.NewGetModelsQueryHandler(nil))

	healthCtrl := controllers.NewHealthController()
	explorationCtrl := controllers.NewExplorationController(med)
	paperCtrl := controllers.NewPaperController(med)
	swaggerH := swagger.NewHandler()

	r := NewRouter(healthCtrl, explorationCtrl, paperCtrl, swaggerH, "")

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Swagger redirect",
			method:         http.MethodGet,
			url:            "/swagger",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			name:           "Swagger UI page",
			method:         http.MethodGet,
			url:            "/swagger/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Docs redirect",
			method:         http.MethodGet,
			url:            "/docs",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			name:           "OpenAPI JSON spec",
			method:         http.MethodGet,
			url:            "/api/v1/swagger.json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health check",
			method:         http.MethodGet,
			url:            "/api/v1/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Papers providers endpoint",
			method:         http.MethodGet,
			url:            "/api/v1/papers/providers",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Models endpoint",
			method:         http.MethodGet,
			url:            "/api/v1/models",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Exploration models endpoint",
			method:         http.MethodGet,
			url:            "/api/v1/exploration/models",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d for %s %s, got %d", tc.expectedStatus, tc.method, tc.url, rr.Code)
			}
		})
	}
}
