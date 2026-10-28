package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kuorum/backend/internal/api"
	"kuorum/backend/internal/api/controllers"
	"kuorum/backend/internal/api/swagger"
	"kuorum/backend/internal/application/common/behaviors"
	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/exploration/commands/decompose_concept"
	"kuorum/backend/internal/application/exploration/commands/reformulate_query"
	"kuorum/backend/internal/application/exploration/queries/get_angles"
	"kuorum/backend/internal/application/exploration/queries/get_models"
	"kuorum/backend/internal/application/extraction/commands/synthesize_concept"
	"kuorum/backend/internal/application/extraction/queries/get_synthesis"
	"kuorum/backend/internal/application/graph/commands/build_graph"
	"kuorum/backend/internal/application/graph/queries/get_graph"
	"kuorum/backend/internal/application/pipeline/orchestrator"
	"kuorum/backend/internal/application/retrieval/commands/ingest_pdf"
	"kuorum/backend/internal/application/retrieval/queries/get_citations"
	"kuorum/backend/internal/application/retrieval/queries/get_paper_details"
	"kuorum/backend/internal/application/retrieval/queries/get_pdf_sections"
	"kuorum/backend/internal/application/retrieval/queries/get_providers"
	"kuorum/backend/internal/application/retrieval/queries/get_references"
	"kuorum/backend/internal/application/retrieval/queries/search_papers"
	"kuorum/backend/internal/application/retrieval/services"
	"kuorum/backend/internal/infrastructure/academic"
	"kuorum/backend/internal/infrastructure/classifier"
	"kuorum/backend/internal/infrastructure/config"
	"kuorum/backend/internal/infrastructure/llm"
	"kuorum/backend/internal/infrastructure/pdf"
	"kuorum/backend/internal/infrastructure/persistence"
)

func main() {
	cfg := config.Load()
	log.Printf("[Init] Starting Research Helper Backend on port %d (%s)...", cfg.Port, cfg.Environment)

	// 1. Infrastructure Layer Initialization
	var llmClient interfaces.ILLMClient
	if cfg.LLM.APIKey != "" {
		fedLLM := llm.NewFederatedFromConfig(cfg.LLM)
		llmClient = fedLLM
		log.Printf("[Init] Federated LLM Client initialized with providers: %v (endpoint: %s, model: %s)",
			fedLLM.AvailableProviders(), cfg.LLM.EndpointType, cfg.LLM.Model)
	} else {
		log.Println("[Init] LLM_API_KEY not set. Using resilient default decomposition handler.")
	}

	academicClient := academic.NewFederatedFromConfig(cfg.Academic)
	log.Printf("[Init] Federated Academic Client initialized with providers: %v", academicClient.AvailableProviders())

	cacheRepo := persistence.NewMemoryCache()

	// Initialize SQLite Database with automatic schema migrations
	sqlDB, err := persistence.NewSqliteDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("[Init] Failed to initialize SQLite database: %v", err)
	}
	defer sqlDB.Close()

	explorationRepo := persistence.NewSqliteExplorationRepository(sqlDB)
	paperRepo := persistence.NewSqlitePaperRepository(sqlDB)
	graphRepo := persistence.NewSqliteGraphRepository(sqlDB)
	extractionRepo := persistence.NewSqliteExtractionRepository(sqlDB)
	pipelineRepo := persistence.NewSqlitePipelineRepository(sqlDB)

	// Initialize PDF Parser, Storage & Downloader
	pdfParser := pdf.NewPopplerPdfParser()
	pdfStorage, err := pdf.NewLocalPdfStorage("data/papers")
	if err != nil {
		log.Printf("[Init] Warning: Failed to initialize PDF storage: %v", err)
	}
	pdfDownloader := pdf.NewPaperDownloader(pdfStorage, http.DefaultClient)

	// 2. Application Layer (CQRS Mediator & Pipeline)
	med := mediator.NewInMemoryMediator()
	med.AddBehavior(behaviors.LoggingBehavior())
	med.AddBehavior(behaviors.ValidationBehavior())

	// Register Exploration Handlers
	decomposeHandler := decompose_concept.NewDecomposeConceptCommandHandler(llmClient, explorationRepo)
	med.RegisterCommandHandler(decompose_concept.DecomposeConceptCommand{}, decomposeHandler)

	reformulateHandler := reformulate_query.NewReformulateQueryCommandHandler(llmClient)
	med.RegisterCommandHandler(reformulate_query.ReformulateQueryCommand{}, reformulateHandler)

	getAnglesHandler := get_angles.NewGetAnglesQueryHandler(cacheRepo, explorationRepo)
	med.RegisterQueryHandler(get_angles.GetAnglesQuery{}, getAnglesHandler)

	getModelsHandler := get_models.NewGetModelsQueryHandler(llm.DefaultCatalog())
	med.RegisterQueryHandler(get_models.GetModelsQuery{}, getModelsHandler)

	// Register Retrieval Handlers (Federated Academic Literature & PDF Ingestion)
	ingestPdfHandler := ingest_pdf.NewIngestPdfCommandHandler(pdfParser, pdfStorage, paperRepo, cacheRepo, pdfDownloader)
	med.RegisterCommandHandler(ingest_pdf.IngestPdfCommand{}, ingestPdfHandler)

	getPdfSectionsHandler := get_pdf_sections.NewGetPdfSectionsQueryHandler(cacheRepo, pdfParser, pdfStorage)
	med.RegisterQueryHandler(get_pdf_sections.GetPdfSectionsQuery{}, getPdfSectionsHandler)

	searchPapersHandler := search_papers.NewSearchPapersQueryHandler(academicClient, paperRepo)
	med.RegisterQueryHandler(search_papers.SearchPapersQuery{}, searchPapersHandler)

	getPaperDetailsHandler := get_paper_details.NewGetPaperDetailsQueryHandler(academicClient, paperRepo)
	med.RegisterQueryHandler(get_paper_details.GetPaperDetailsQuery{}, getPaperDetailsHandler)

	getReferencesHandler := get_references.NewGetReferencesQueryHandler(academicClient)
	med.RegisterQueryHandler(get_references.GetReferencesQuery{}, getReferencesHandler)

	getCitationsHandler := get_citations.NewGetCitationsQueryHandler(academicClient)
	med.RegisterQueryHandler(get_citations.GetCitationsQuery{}, getCitationsHandler)

	getProvidersHandler := get_providers.NewGetProvidersQueryHandler(academicClient)
	med.RegisterQueryHandler(get_providers.GetProvidersQuery{}, getProvidersHandler)

	// Register Graph Handlers
	buildGraphHandler := build_graph.NewBuildCitationGraphCommandHandler(academicClient, graphRepo, paperRepo, cacheRepo)
	med.RegisterCommandHandler(build_graph.BuildCitationGraphCommand{}, buildGraphHandler)

	getGraphHandler := get_graph.NewGetCitationGraphQueryHandler(graphRepo)
	med.RegisterQueryHandler(get_graph.GetCitationGraphQuery{}, getGraphHandler)

	// Register Extraction Handlers
	synthesizeHandler := synthesize_concept.NewSynthesizeConceptCommandHandler(llmClient, extractionRepo, cacheRepo, pdfParser, pdfStorage, paperRepo)
	med.RegisterCommandHandler(synthesize_concept.SynthesizeConceptCommand{}, synthesizeHandler)

	getSynthesisHandler := get_synthesis.NewGetSynthesisQueryHandler(extractionRepo)
	med.RegisterQueryHandler(get_synthesis.GetSynthesisQuery{}, getSynthesisHandler)

	// Initialize Research Pipeline Orchestrator with Clef Classifier
	ranker := services.NewAnchorPaperRanker()
	clefClassifier := classifier.NewClefClassifier(cfg.Cloudflare.AccountID, cfg.Cloudflare.APIToken, cfg.Cloudflare.ClefModel)
	orchestratorEngine := orchestrator.NewPipelineOrchestrator(med, pipelineRepo, ranker, clefClassifier)

	// 3. API Layer (Controllers, Swagger & Router)
	healthCtrl := controllers.NewHealthController()
	explorationCtrl := controllers.NewExplorationController(med)
	paperCtrl := controllers.NewPaperController(med, pdfStorage)
	pipelineCtrl := controllers.NewPipelineController(orchestratorEngine, pipelineRepo)
	swaggerHandler := swagger.NewHandler()

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../frontend/dist"
	}

	router := api.NewRouter(healthCtrl, explorationCtrl, paperCtrl, swaggerHandler, staticDir, pipelineCtrl)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[Server] HTTP API listening on http://localhost:%d", cfg.Port)
		log.Printf("[Server] Swagger UI documentation at http://localhost:%d/swagger", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Listen error: %v", err)
		}
	}()

	<-stopChan
	log.Println("[Server] Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[Server] Forced shutdown error: %v", err)
	}

	log.Println("[Server] Server stopped successfully.")
}
