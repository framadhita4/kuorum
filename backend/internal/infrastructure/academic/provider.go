package academic

import (
	"kuorum/backend/internal/infrastructure/academic/providers"
)

// Re-export core types and interfaces from providers package for convenience and backwards compatibility.
type (
	ProviderName     = providers.ProviderName
	AcademicProvider = providers.AcademicProvider
	BaseProvider     = providers.BaseProvider
	RateLimiter      = providers.RateLimiter
)

const (
	ProviderOpenAlex        = providers.ProviderOpenAlex
	ProviderSemanticScholar = providers.ProviderSemanticScholar
	ProviderArxiv           = providers.ProviderArxiv
	ProviderCrossref        = providers.ProviderCrossref

	EnglishStopWordsRaw = providers.EnglishStopWordsRaw
)

var (
	AllProviders          = providers.AllProviders
	NormalizeProviderName = providers.NormalizeProviderName
	NewRateLimiter        = providers.NewRateLimiter
	NewBaseProvider       = providers.NewBaseProvider

	NewOpenAlexProvider        = providers.NewOpenAlexProvider
	NewSemanticScholarProvider = providers.NewSemanticScholarProvider
	NewArxivProvider           = providers.NewArxivProvider
	NewCrossrefProvider        = providers.NewCrossrefProvider

	NonAlphaNumRegex = providers.NonAlphaNumRegex
	EnglishStopWords = providers.EnglishStopWords
	IsStopWord       = providers.IsStopWord
)

type (
	OpenAlexProvider        = providers.OpenAlexProvider
	SemanticScholarProvider = providers.SemanticScholarProvider
	ArxivProvider           = providers.ArxivProvider
	CrossrefProvider        = providers.CrossrefProvider
)
