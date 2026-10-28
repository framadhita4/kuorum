import { useEffect, useState } from 'react'

import styled from '@emotion/styled'
import { BookOpen, Filter, Loader2, Search, Sparkles, X } from 'lucide-react'

import { Button, Dialog, Pill } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import { PaperSearchResultItem } from '#/routes/-components/PaperSearchResultItem'
import { useGetPaperProviders } from '#/services/paper/useGetPaperProviders'
import { useSearchPapers } from '#/services/paper/useSearchPapers'
import type { PaperMetadata } from '#/types/paper.type'

interface PaperSearchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialQuery?: string
  onSelectPaper?: (paper: PaperMetadata) => void
  onStartPipelineWithPaper?: (paper: PaperMetadata) => void
}

const DialogHeader = styled.div`
  padding: 20px 24px 16px;
  border-bottom: ${theme.colors.border.subtle};
  background-color: ${theme.colors.canvas.primary};
  position: relative;
`

const DialogTitleStyled = styled(Dialog.Title)`
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: ${theme.typography.sizes.h5};
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
`

const SearchForm = styled.form`
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
`

const InputWrapper = styled.div`
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
`

const SearchIconWrapper = styled.div`
  position: absolute;
  left: 12px;
  display: flex;
  align-items: center;
  pointer-events: none;
`

const SearchInputField = styled.input`
  width: 100%;
  padding: 10px 14px 10px 38px;
  border-radius: ${theme.radius.button};
  border: 1px solid rgba(100, 111, 121, 0.25);
  background-color: ${theme.colors.canvas.subtle};
  color: ${theme.colors.ink.primary};
  font-size: ${theme.typography.sizes.sm};
  outline: none;
  font-family: ${theme.typography.fontFamily.sans};
  transition: ${theme.transitions.fast};

  &:focus {
    border-color: ${theme.colors.accent.purple};
    background-color: ${theme.colors.canvas.primary};
  }

  &::placeholder {
    color: ${theme.colors.ink.muted};
  }
`

const ProviderFilterRow = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  flex-wrap: wrap;
`

const ProviderLabel = styled.span`
  font-size: 11px;
  color: ${theme.colors.ink.muted};
  margin-right: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
`

const ResultsContainer = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background-color: ${theme.colors.canvas.subtle};
`

const SuggestionContainer = styled.div`
  padding: 32px 16px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
`

const SuggestionIcon = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: rgba(121, 110, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.accent.purple};
`

const SuggestionTitle = styled.div`
  font-weight: ${theme.typography.weights.medium};
  font-size: ${theme.typography.sizes.title};
  color: ${theme.colors.ink.primary};
`

const SuggestionTags = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
  max-width: 540px;
`

const LoadingBox = styled.div`
  padding: 40px 16px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: ${theme.colors.ink.muted};
  font-size: ${theme.typography.sizes.sm};
`

const ErrorBox = styled.div`
  padding: 16px;
  border-radius: ${theme.radius.card};
  background-color: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: ${theme.colors.status.error};
  font-size: ${theme.typography.sizes.sm};
`

const ResultsCount = styled.span`
  font-size: 12px;
  font-weight: 500;
  color: ${theme.colors.ink.muted};
  padding: 0 4px;
`

const EmptyResultsBox = styled.div`
  padding: 32px 16px;
  text-align: center;
  color: ${theme.colors.ink.muted};
  font-size: ${theme.typography.sizes.sm};
`

export const PaperSearchDialog = ({
  open,
  onOpenChange,
  initialQuery = '',
  onSelectPaper,
  onStartPipelineWithPaper,
}: PaperSearchDialogProps) => {
  const [searchTerm, setSearchTerm] = useState(initialQuery)
  const [activeQuery, setActiveQuery] = useState(initialQuery)
  const [selectedProvider, setSelectedProvider] = useState<string>('all')

  useEffect(() => {
    if (initialQuery && open) {
      setSearchTerm(initialQuery)
      setActiveQuery(initialQuery)
    }
  }, [initialQuery, open])

  const { data: providersData } = useGetPaperProviders()
  const availableProviders = providersData?.providers ?? [
    'semanticscholar',
    'openalex',
    'arxiv',
    'crossref',
  ]

  const {
    data: searchResults,
    isLoading,
    isFetching,
    error,
  } = useSearchPapers(
    {
      query: activeQuery,
      limit: 15,
      providers: selectedProvider === 'all' ? undefined : [selectedProvider],
    },
    Boolean(activeQuery.trim()),
  )

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!searchTerm.trim()) return
    setActiveQuery(searchTerm.trim())
  }

  const handleQuickTagClick = (tag: string) => {
    setSearchTerm(tag)
    setActiveQuery(tag)
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Popup
          style={{
            maxWidth: '780px',
            width: '92vw',
            maxHeight: '85vh',
            display: 'flex',
            flexDirection: 'column',
            padding: 0,
            overflow: 'hidden',
          }}
        >
          {/* Header */}
          <DialogHeader>
            <DialogTitleStyled>
              <BookOpen size={18} color={theme.colors.accent.purple} />
              Search Papers & Academic Literature
            </DialogTitleStyled>
            <Dialog.Description style={{ margin: '4px 0 0' }}>
              Search across scientific literature from Semantic Scholar,
              OpenAlex, arXiv, and Crossref simultaneously.
            </Dialog.Description>
            <Dialog.Close>
              <X size={15} />
            </Dialog.Close>

            {/* Search Input Bar */}
            <SearchForm onSubmit={handleSearchSubmit}>
              <InputWrapper>
                <SearchIconWrapper>
                  <Search size={16} color={theme.colors.ink.muted} />
                </SearchIconWrapper>
                <SearchInputField
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Type research topic, paper title, DOI, or keywords..."
                />
              </InputWrapper>

              <Button
                type="submit"
                variant="primary"
                size="md"
                disabled={isLoading || isFetching}
                startIcon={
                  isLoading || isFetching ? (
                    <Loader2 size={14} className="animate-spin" />
                  ) : (
                    <Search size={14} />
                  )
                }
              >
                {isLoading || isFetching ? 'Searching...' : 'Search'}
              </Button>
            </SearchForm>

            {/* Provider Filter Chips */}
            <ProviderFilterRow>
              <ProviderLabel>
                <Filter size={11} /> Provider:
              </ProviderLabel>

              <Pill
                variant={selectedProvider === 'all' ? 'purple' : 'neutral'}
                size="sm"
                onClick={() => setSelectedProvider('all')}
                label="All Providers"
              />

              {availableProviders.map((prov) => (
                <Pill
                  key={prov}
                  variant={selectedProvider === prov ? 'purple' : 'neutral'}
                  size="sm"
                  onClick={() => setSelectedProvider(prov)}
                  label={prov.charAt(0).toUpperCase() + prov.slice(1)}
                />
              ))}
            </ProviderFilterRow>
          </DialogHeader>

          {/* Results List */}
          <ResultsContainer>
            {/* Quick Suggestion when no active query */}
            {!activeQuery && (
              <SuggestionContainer>
                <SuggestionIcon>
                  <Sparkles size={20} />
                </SuggestionIcon>
                <SuggestionTitle>Try popular research topics</SuggestionTitle>
                <SuggestionTags>
                  {[
                    'Solid state electrolyte dendrite lithium',
                    'CRISPR Cas9 base editing off-target',
                    'Perovskite tandem solar cell stability',
                    'Attention Is All You Need transformer',
                  ].map((sample) => (
                    <Pill
                      key={sample}
                      variant="neutral"
                      size="sm"
                      onClick={() => handleQuickTagClick(sample)}
                      label={sample}
                    />
                  ))}
                </SuggestionTags>
              </SuggestionContainer>
            )}

            {/* Loading Indicator */}
            {(isLoading || isFetching) && (
              <LoadingBox>
                <Loader2
                  size={24}
                  className="animate-spin"
                  color={theme.colors.accent.purple}
                />
                <span>
                  Connecting to Semantic Scholar, OpenAlex, arXiv & Crossref...
                </span>
              </LoadingBox>
            )}

            {/* Error Display */}
            {error && (
              <ErrorBox>
                Failed to load search results: {error.message}
              </ErrorBox>
            )}

            {/* Results Found Header */}
            {searchResults && (
              <ResultsCount>
                Found {searchResults.total} scientific publications for "
                {searchResults.query}"
              </ResultsCount>
            )}

            {/* Empty Results */}
            {searchResults && searchResults.papers.length === 0 && (
              <EmptyResultsBox>
                No papers match this query. Try using broader keywords.
              </EmptyResultsBox>
            )}

            {/* Paper Cards */}
            {searchResults?.papers.map((paper) => (
              <PaperSearchResultItem
                key={paper.id}
                paper={paper}
                onSelectPaper={(p) => {
                  onSelectPaper?.(p)
                  onOpenChange(false)
                }}
                onStartPipelineWithPaper={(p) => {
                  onStartPipelineWithPaper?.(p)
                  onOpenChange(false)
                }}
              />
            ))}
          </ResultsContainer>
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
