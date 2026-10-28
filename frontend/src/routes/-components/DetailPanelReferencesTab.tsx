import { useState } from 'react'

import styled from '@emotion/styled'
import { Check, Copy, ExternalLink, Network } from 'lucide-react'

import { Badge } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ReferenceItem } from '#/routes/-types/research.type'

interface DetailPanelReferencesTabProps {
  references: Array<ReferenceItem>
  selectedRefId: string
  onSelectRef: (id: string) => void
  onChangeTab: (tab: 'report' | 'references' | 'graph') => void
}

const ReferencesContainer = styled.div`
  max-width: 920px;
  margin: 0 auto;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
  font-family: ${theme.typography.fontFamily.sans};
`

const TabHeaderRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`

const SummaryCount = styled.span`
  font-size: 0.88rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.secondary};
`

const GraphLinkButton = styled.button`
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  color: ${theme.colors.accent.purple};
  font-size: 0.84rem;
  font-weight: ${theme.typography.weights.medium};
  cursor: pointer;
  padding: 4px 8px;
  border-radius: ${theme.radius.button};
  transition: background-color 0.15s ease;

  &:hover {
    background-color: rgba(121, 110, 255, 0.08);
  }
`

const PapersList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 20px;
`

const PaperCard = styled.div<{ $isSelected: boolean }>`
  border: ${({ $isSelected }) =>
    $isSelected
      ? `1.5px solid ${theme.colors.accent.purple}`
      : `1px solid rgba(100, 111, 121, 0.2)`};
  border-radius: ${theme.radius.card};
  padding: 24px 28px;
  background-color: ${({ $isSelected }) =>
    $isSelected ? theme.colors.canvas.subtle : theme.colors.canvas.primary};
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  box-shadow: ${({ $isSelected }) =>
    $isSelected ? theme.shadows.md : theme.shadows.sm};
  transition: ${theme.transitions.fast};
`

const PaperTitleRow = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 14px;
`

const RefNumberBadge = styled.span<{ $isSelected: boolean }>`
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: ${({ $isSelected }) =>
    $isSelected ? theme.colors.accent.purple : theme.colors.ink.primary};
  color: #ffffff;
  font-size: 0.8rem;
  font-weight: ${theme.typography.weights.medium};
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-family: ${theme.typography.fontFamily.mono};
  margin-top: 2px;
`

const PaperTitle = styled.h4`
  font-size: 1.125rem;
  font-weight: ${theme.typography.weights.regular};
  color: ${theme.colors.ink.primary};
  margin: 0;
  line-height: ${theme.typography.lineHeights.snug};
  letter-spacing: ${theme.typography.letterSpacing.tight};
`

const AuthorsMetaRow = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.84rem;
  color: ${theme.colors.ink.secondary};
  flex-wrap: wrap;
  padding-left: 42px;
`

const BadgesRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
  padding-left: 42px;
`

const QuoteBox = styled.p`
  font-size: 0.9rem;
  color: ${theme.colors.ink.secondary};
  font-style: italic;
  margin: 6px 0 0 42px;
  line-height: ${theme.typography.lineHeights.relaxed};
  background-color: ${theme.colors.canvas.primary};
  padding: 12px 16px;
  border-radius: ${theme.radius.button};
  border-left: 3px solid ${theme.colors.brand.coral};
  border-top: 1px solid rgba(100, 111, 121, 0.1);
  border-right: 1px solid rgba(100, 111, 121, 0.1);
  border-bottom: 1px solid rgba(100, 111, 121, 0.1);
`

const ActionsFooter = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding-top: 10px;
  border-top: 1px solid rgba(100, 111, 121, 0.12);
  padding-left: 42px;
`

const CopyButton = styled.button<{ $isCopied: boolean }>`
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  color: ${({ $isCopied }) =>
    $isCopied ? theme.colors.accent.purple : theme.colors.ink.muted};
  font-size: 0.78rem;
  font-weight: ${theme.typography.weights.medium};
  cursor: pointer;
  transition: color 0.15s ease;

  &:hover {
    color: ${theme.colors.accent.purple};
  }
`

const DoiLink = styled.a`
  display: flex;
  align-items: center;
  gap: 5px;
  color: ${theme.colors.ink.primary};
  font-size: 0.78rem;
  font-weight: ${theme.typography.weights.medium};
  text-decoration: none;
  transition: color 0.15s ease;

  &:hover {
    color: ${theme.colors.brand.coral};
  }
`

export const DetailPanelReferencesTab = ({
  references,
  selectedRefId,
  onSelectRef,
  onChangeTab,
}: DetailPanelReferencesTabProps) => {
  const [copiedBibtexId, setCopiedBibtexId] = useState<string | null>(null)

  const handleCopyBibtex = (ref: ReferenceItem) => {
    const bibtex = `@article{${ref.authors
      .split(',')[0]
      .toLowerCase()
      .replace(
        /\s/g,
        '',
      )}${ref.year},\n  title={${ref.title}},\n  author={${ref.authors}},\n  journal={${ref.journal}},\n  year={${ref.year}},\n  doi={${ref.doi}}\n}`
    navigator.clipboard
      .writeText(bibtex)
      .then(() => {
        setCopiedBibtexId(ref.id)
        setTimeout(() => setCopiedBibtexId(null), 2000)
      })
      .catch((err) => console.error(err))
  }

  return (
    <ReferencesContainer>
      <TabHeaderRow>
        <SummaryCount>
          Review Results • {references.length} Key Scientific Publications
        </SummaryCount>
        <GraphLinkButton type="button" onClick={() => onChangeTab('graph')}>
          <Network size={15} />
          <span>View Citation Graph</span>
        </GraphLinkButton>
      </TabHeaderRow>

      {/* List of Papers */}
      <PapersList>
        {references.map((ref) => {
          const isSelected = selectedRefId === ref.id
          return (
            <PaperCard
              key={ref.id}
              $isSelected={isSelected}
              onClick={() => onSelectRef(ref.id)}
            >
              <PaperTitleRow>
                <RefNumberBadge $isSelected={isSelected}>
                  {ref.num}
                </RefNumberBadge>
                <PaperTitle>{ref.title}</PaperTitle>
              </PaperTitleRow>

              {/* Authors & Metrics */}
              <AuthorsMetaRow>
                <span>{ref.year}</span>
                <span>•</span>
                <strong style={{ color: theme.colors.ink.primary }}>
                  {ref.citationsCount.toLocaleString()} citations
                </strong>
                <span>•</span>
                <span>{ref.authors}</span>
              </AuthorsMetaRow>

              {/* Journal badge & Quartile */}
              <BadgesRow>
                <Badge variant="neutral" size="sm">
                  {ref.journal}
                </Badge>
                <Badge variant="rose" size="sm">
                  {ref.quartile}
                </Badge>
              </BadgesRow>

              {/* Verbatim Quote Snippet */}
              <QuoteBox>"{ref.keySnippet}"</QuoteBox>

              {/* Action buttons (Copy BibTeX & DOI) */}
              <ActionsFooter>
                <CopyButton
                  type="button"
                  $isCopied={copiedBibtexId === ref.id}
                  onClick={(e) => {
                    e.stopPropagation()
                    handleCopyBibtex(ref)
                  }}
                >
                  {copiedBibtexId === ref.id ? (
                    <>
                      <Check size={14} />
                      <span>Copied to Clipboard</span>
                    </>
                  ) : (
                    <>
                      <Copy size={14} />
                      <span>Copy BibTeX</span>
                    </>
                  )}
                </CopyButton>

                <DoiLink
                  href={`https://doi.org/${ref.doi}`}
                  target="_blank"
                  rel="noreferrer"
                  onClick={(e) => e.stopPropagation()}
                >
                  <span>Open DOI Paper</span>
                  <ExternalLink size={12} />
                </DoiLink>
              </ActionsFooter>
            </PaperCard>
          )
        })}
      </PapersList>
    </ReferencesContainer>
  )
}
