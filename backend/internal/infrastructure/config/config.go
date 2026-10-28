package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        int
	Environment string
	LLM         LLMConfig
	Academic    AcademicConfig
	Database    DatabaseConfig
	Cloudflare  CloudflareConfig
}

type CloudflareConfig struct {
	AccountID string
	APIToken  string
	ClefModel string
}

type DatabaseConfig struct {
	DSN string
}

type LLMConfig struct {
	EndpointType    string
	BaseURL         string
	APIKey          string
	Model           string
	ReasoningEffort string
}

type AcademicConfig struct {
	SemanticScholarAPIKey  string
	PoliteEmail            string
	OpenAlexBaseURL        string
	SemanticScholarBaseURL string
	ArxivBaseURL           string
	CrossrefBaseURL        string
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadDotEnv() {
	candidates := []string{".env", "../.env", "../../.env", "backend/.env"}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func Load() *Config {
	loadDotEnv()
	port := 8080
	if val := os.Getenv("PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			port = p
		}
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	llmBaseURL := os.Getenv("LLM_BASE_URL")
	if llmBaseURL == "" {
		llmBaseURL = "https://api.openai.com/v1"
	}

	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		llmModel = "gpt-4o-mini"
	}

	llmEndpointType := getEnvOrDefault("LLM_ENDPOINT_TYPE", "openai")
	llmReasoningEffort := os.Getenv("LLM_REASONING_EFFORT")

	llmAPIKey := os.Getenv("LLM_API_KEY")
	if llmAPIKey == "" {
		// Auto-resolve API keys from specific provider environment variables
		candidates := []struct {
			envKey       string
			endpointType string
		}{
			{"GEMINI_API_KEY", "gemini"},
			{"GOOGLE_API_KEY", "gemini"},
			{"OPENAI_API_KEY", "openai"},
			{"ANTHROPIC_API_KEY", "anthropic"},
			{"CLAUDE_API_KEY", "anthropic"},
			{"GROQ_API_KEY", "groq"},
			{"DEEPSEEK_API_KEY", "deepseek"},
			{"OPENROUTER_API_KEY", "openrouter"},
			{"TOGETHER_API_KEY", "together"},
			{"MISTRAL_API_KEY", "mistral"},
			{"PERPLEXITY_API_KEY", "perplexity"},
		}
		for _, c := range candidates {
			if val := os.Getenv(c.envKey); val != "" {
				llmAPIKey = val
				if os.Getenv("LLM_ENDPOINT_TYPE") == "" {
					llmEndpointType = c.endpointType
				}
				break
			}
		}
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "kuorum.db"
	}

	return &Config{
		Port:        port,
		Environment: env,
		Database: DatabaseConfig{
			DSN: dbPath,
		},
		LLM: LLMConfig{
			EndpointType:    llmEndpointType,
			BaseURL:         llmBaseURL,
			APIKey:          llmAPIKey,
			Model:           llmModel,
			ReasoningEffort: llmReasoningEffort,
		},
		Academic: AcademicConfig{
			SemanticScholarAPIKey:  os.Getenv("S2_API_KEY"),
			PoliteEmail:            getEnvOrDefault("POLITE_EMAIL", "kuorum@example.com"),
			OpenAlexBaseURL:        getEnvOrDefault("OPENALEX_BASE_URL", "https://api.openalex.org"),
			SemanticScholarBaseURL: getEnvOrDefault("S2_BASE_URL", "https://api.semanticscholar.org/graph/v1"),
			ArxivBaseURL:           getEnvOrDefault("ARXIV_BASE_URL", "https://export.arxiv.org/api"),
			CrossrefBaseURL:        getEnvOrDefault("CROSSREF_BASE_URL", "https://api.crossref.org"),
		},
		Cloudflare: CloudflareConfig{
			AccountID: getEnvOrDefault("CLOUDFLARE_ACCOUNT_ID", os.Getenv("CF_ACCOUNT_ID")),
			APIToken:  getEnvOrDefault("CLOUDFLARE_API_TOKEN", os.Getenv("CF_API_TOKEN")),
			ClefModel: getEnvOrDefault("CLOUDFLARE_CLEF_MODEL", "@cf/cloudflare/clef-flash"),
		},
	}
}
