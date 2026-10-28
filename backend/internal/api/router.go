package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"kuorum/backend/internal/api/controllers"
	"kuorum/backend/internal/api/middleware"
	"kuorum/backend/internal/api/swagger"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	healthCtrl *controllers.HealthController,
	explorationCtrl *controllers.ExplorationController,
	paperCtrl *controllers.PaperController,
	swaggerHandler *swagger.Handler,
	staticDir string,
	pipelineCtrl ...*controllers.PipelineController,
) http.Handler {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS())

	// Swagger Interactive Documentation
	r.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/", swaggerHandler.ServeUI)
	r.Get("/swagger/index.html", swaggerHandler.ServeUI)
	r.Get("/swagger/doc.json", swaggerHandler.ServeSpec)
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/docs/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})

	// API Routes under /api/v1
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/swagger.json", swaggerHandler.ServeSpec)
		api.Get("/health", healthCtrl.Check)
		api.Get("/models", explorationCtrl.GetModels)

		api.Route("/exploration", func(exp chi.Router) {
			exp.Post("/decompose", explorationCtrl.Decompose)
			exp.Post("/stream", explorationCtrl.Stream)
			exp.Post("/reformulate", explorationCtrl.Reformulate)
			exp.Get("/angles", explorationCtrl.GetAngles)
			exp.Get("/models", explorationCtrl.GetModels)
		})

		api.Route("/papers", func(pr chi.Router) {
			pr.Get("/search", paperCtrl.Search)
			pr.Post("/search", paperCtrl.Search)
			pr.Get("/providers", paperCtrl.GetProviders)
			pr.Get("/details", paperCtrl.GetDetails)
			pr.Get("/references", paperCtrl.GetReferences)
			pr.Get("/citations", paperCtrl.GetCitations)
			pr.Get("/{id}", paperCtrl.GetDetails)
			pr.Get("/{id}/references", paperCtrl.GetReferences)
			pr.Get("/{id}/citations", paperCtrl.GetCitations)
			pr.Post("/upload-pdf", paperCtrl.UploadPdf)
			pr.Get("/{id}/sections", paperCtrl.GetSections)
			pr.Get("/{id}/pdf", paperCtrl.StreamPdf)
		})

		if len(pipelineCtrl) > 0 && pipelineCtrl[0] != nil {
			ctrl := pipelineCtrl[0]
			setupChatRoutes := func(pipe chi.Router) {
				pipe.Post("/", ctrl.Start)
				pipe.Get("/", ctrl.List)
				pipe.Get("/{id}", ctrl.Get)
				pipe.Delete("/{id}", ctrl.Delete)
				pipe.Get("/{id}/stream", ctrl.Stream)
				pipe.Post("/{id}/select-angle", ctrl.SelectAngle)
				pipe.Post("/{id}/select-anchor", ctrl.SelectAnchor)
				pipe.Post("/{id}/resume", ctrl.Resume)
				pipe.Post("/{id}/cancel", ctrl.Cancel)
				pipe.Get("/{id}/messages", ctrl.GetMessages)
				pipe.Post("/{id}/messages", ctrl.AddMessage)
				pipe.Patch("/{id}/pin", ctrl.TogglePin)
			}
			api.Route("/pipelines", setupChatRoutes)
			api.Route("/chat-history", setupChatRoutes)
		}
	})

	// SPA Fallback for frontend (TanStack Router)
	if staticDir != "" {
		if _, err := os.Stat(staticDir); err == nil {
			fileServer := http.FileServer(http.Dir(staticDir))
			r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
				if strings.HasPrefix(req.URL.Path, "/api") ||
					strings.HasPrefix(req.URL.Path, "/swagger") ||
					strings.HasPrefix(req.URL.Path, "/docs") {
					http.NotFound(w, req)
					return
				}

				path := filepath.Join(staticDir, filepath.Clean(req.URL.Path))
				if info, err := os.Stat(path); os.IsNotExist(err) || info.IsDir() {
					http.ServeFile(w, req, filepath.Join(staticDir, "index.html"))
					return
				}
				fileServer.ServeHTTP(w, req)
			})
		}
	}

	return r
}
