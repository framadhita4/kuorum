import { useState } from 'react'

import styled from '@emotion/styled'
import {
  ArrowUpRight,
  BookOpen,
  Check,
  ChevronRight,
  Copy,
  ExternalLink,
  Network,
  X,
} from 'lucide-react'

import { Button, IconButton, Pill } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import { CitationGraphVisualizer } from '#/routes/-components/CitationGraphVisualizer'
import type { CitationEdge, ReferenceItem } from '#/routes/-types/research.type'

interface CitationDrawerProps {
  references: Array<ReferenceItem>
  edges: Array<CitationEdge>
  activeRefId: string | null
  tab: 'detail' | 'graph'
  onChangeTab: (tab: 'detail' | 'graph') => void
  onSelectRef: (id: string) => void
  onClose: () => void
}

const DrawerBackdrop = styled.div`
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  justify-content: flex-end;
  background-color: ${theme.colors.canvas.overlay};
`

const BackdropClickArea = styled.div`
  flex: 1;
`

const DrawerPanel = styled.div`
  width: 100%;
  max-width: 500px;
  height: 100%;
  background-color: ${theme.colors.canvas.primary};
  border-left: ${theme.colors.border.subtle};
  box-shadow: ${theme.shadows.lg};
  display: flex;
  flex-direction: column;
  z-index: 60;
  font-family: ${theme.typography.fontFamily.sans};
`

const DrawerHeader = styled.div`
  padding: 16px 24px;
  border-bottom: 1px solid rgba(100, 111, 121, 0.12);
  display: flex;
  align-items: center;
  justify-content: space-between;
`

const HeaderTabs = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const DrawerContent = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 28px 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
`

const MetaHeader = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`

const MetaTopRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
`

const RefNumToken = styled.span`
  font-size: 0.74rem;
  font-weight: 700;
  color: ${theme.colors.accent.purple};
  font-family: ${theme.typography.fontFamily.mono};
`

const JournalInfo = styled.span`
  font-size: 0.76rem;
  font-weight: 600;
  color: ${theme.colors.ink.primary};
`

const QuartileTag = styled.span`
  font-size: 0.7rem;
  font-weight: 600;
  color: ${theme.colors.status.success};
  font-family: ${theme.typography.fontFamily.mono};
`

const PaperTitle = styled.h3`
  font-size: 1.25rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
  margin: 0;
  line-height: ${theme.typography.lineHeights.snug};
  letter-spacing: ${theme.typography.letterSpacing.tight};
`

const AuthorsMeta = styled.p`
  font-size: 0.78rem;
  color: ${theme.colors.ink.muted};
  margin: 0;
`

const EvidenceQuote = styled.div`
  border-left: 3px solid ${theme.colors.brand.coral};
  padding-left: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
`

const SectionHeadingLabel = styled.span`
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: ${theme.typography.letterSpacing.widest};
  color: ${theme.colors.brand.coral};
  font-family: ${theme.typography.fontFamily.mono};
`

const QuoteText = styled.p`
  font-size: 0.9rem;
  color: ${theme.colors.ink.primary};
  font-style: italic;
  line-height: ${theme.typography.lineHeights.relaxed};
  margin: 0;
`

const SummarySection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 6px;
`

const SummaryLabel = styled.span`
  font-size: 0.68rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: ${theme.typography.letterSpacing.widest};
  color: ${theme.colors.ink.muted};
  font-family: ${theme.typography.fontFamily.mono};
`

const SummaryText = styled.p`
  font-size: 0.88rem;
  color: ${theme.colors.ink.secondary};
  line-height: ${theme.typography.lineHeights.relaxed};
  margin: 0;
`

const ConnectedPapersSection = styled.div`
  padding-top: 16px;
  border-top: 1px solid rgba(100, 111, 121, 0.12);
  display: flex;
  flex-direction: column;
  gap: 10px;
`

const SectionHeaderRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`

const OpenGraphButton = styled.button`
  display: flex;
  align-items: center;
  gap: 4px;
  background: none;
  border: none;
  color: ${theme.colors.accent.purple};
  font-size: 0.74rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0;

  &:hover {
    text-decoration: underline;
  }
`

const ConnectedPaperRow = styled.div`
  padding: 8px 10px;
  border-radius: ${theme.radius.sm};
  background-color: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  transition: background-color 0.12s ease;

  &:hover {
    background-color: ${theme.colors.canvas.neutral};
  }
`

const DrawerFooter = styled.div`
  padding: 14px 24px;
  border-top: 1px solid rgba(100, 111, 121, 0.12);
  background-color: ${theme.colors.canvas.primary};
  display: flex;
  align-items: center;
  justify-content: space-between;
`

const ExternalPaperLink = styled.a`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: ${theme.radius.button};
  background-color: ${theme.colors.ink.primary};
  color: ${theme.colors.ink.inverse};
  font-size: 0.76rem;
  font-weight: 500;
  text-decoration: none;
  transition: ${theme.transitions.fast};

  &:hover {
    background-color: ${theme.colors.brand.coral};
    color: ${theme.colors.ink.primary};
  }
`

export const CitationDrawer = ({
  references,
  edges,
  activeRefId,
  tab,
  onChangeTab,
  onSelectRef,
  onClose,
}: CitationDrawerProps) => {
  const [copiedBibtex, setCopiedBibtex] = useState(false)

  if (!activeRefId) return null
  const ref = references.find((r) => r.id === activeRefId) ?? references[0]

  const handleCopyBibtex = () => {
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
        setCopiedBibtex(true)
        setTimeout(() => setCopiedBibtex(false), 2000)
      })
      .catch((err) => console.error(err))
  }

  return (
    <DrawerBackdrop>
      {/* Click outside backdrop to close */}
      <BackdropClickArea onClick={onClose} />

      {/* Slide-over Drawer Panel */}
      <DrawerPanel>
        {/* Drawer Header Toolbar */}
        <DrawerHeader>
          {/* Tab Switcher */}
          <HeaderTabs>
            <Pill
              variant={tab === 'detail' ? 'active' : 'neutral'}
              size="sm"
              icon={<BookOpen size={13} />}
              label="Paper Details"
              onClick={() => onChangeTab('detail')}
            />
            <Pill
              variant={tab === 'graph' ? 'active' : 'neutral'}
              size="sm"
              icon={<Network size={13} />}
              label="Citation Graph"
              onClick={() => onChangeTab('graph')}
            />
          </HeaderTabs>

          <IconButton
            size="icon-sm"
            variant="ghost"
            onClick={onClose}
            title="Close panel"
            aria-label="Close panel"
          >
            <X size={16} />
          </IconButton>
        </DrawerHeader>

        {/* Scrollable Drawer Content */}
        <DrawerContent>
          {/* Paper Identity Meta */}
          <MetaHeader>
            <MetaTopRow>
              <RefNumToken>[{ref.num}]</RefNumToken>
              <JournalInfo>
                {ref.journal} ({ref.year})
              </JournalInfo>
              <QuartileTag>{ref.quartile}</QuartileTag>
            </MetaTopRow>

            <PaperTitle>{ref.title}</PaperTitle>

            <AuthorsMeta>
              {ref.authors} •{' '}
              <strong style={{ color: theme.colors.ink.primary }}>
                {ref.citationsCount.toLocaleString()} citations
              </strong>
            </AuthorsMeta>
          </MetaHeader>

          {/* TAB 1: PAPER DETAIL & VERBATIM QUOTE */}
          {tab === 'detail' && (
            <div
              style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}
            >
              {/* Verbatim Snippet */}
              <EvidenceQuote>
                <SectionHeadingLabel>
                  Verbatim Evidence Quote
                </SectionHeadingLabel>
                <QuoteText>{ref.keySnippet}</QuoteText>
              </EvidenceQuote>

              {/* Abstract / Contribution Summary */}
              <SummarySection>
                <SummaryLabel>Contribution Summary</SummaryLabel>
                <SummaryText>{ref.abstract}</SummaryText>
              </SummarySection>

              {/* Connected Papers */}
              <ConnectedPapersSection>
                <SectionHeaderRow>
                  <SummaryLabel>Related Citation Connections</SummaryLabel>
                  <OpenGraphButton
                    type="button"
                    onClick={() => onChangeTab('graph')}
                  >
                    <span>Open Graph</span>
                    <ArrowUpRight size={12} />
                  </OpenGraphButton>
                </SectionHeaderRow>

                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '4px',
                  }}
                >
                  {edges
                    .filter((e) => e.from === ref.id || e.to === ref.id)
                    .map((edge, idx) => {
                      const isOutgoing = edge.from === ref.id
                      const otherId = isOutgoing ? edge.to : edge.from
                      const otherRef = references.find((r) => r.id === otherId)
                      if (!otherRef) return null

                      return (
                        <ConnectedPaperRow
                          key={idx}
                          onClick={() => onSelectRef(otherRef.id)}
                        >
                          <div style={{ fontSize: '0.76rem' }}>
                            <span style={{ color: theme.colors.ink.muted }}>
                              {isOutgoing ? 'Cites ' : 'Cited by '}
                            </span>
                            <strong style={{ color: theme.colors.ink.primary }}>
                              [{otherRef.num}] {otherRef.authors.split(',')[0]}{' '}
                              ({otherRef.year})
                            </strong>
                            <p
                              style={{
                                color: theme.colors.ink.secondary,
                                margin: '2px 0 0',
                                fontSize: '0.72rem',
                              }}
                            >
                              {edge.relation}
                            </p>
                          </div>
                          <ChevronRight
                            size={13}
                            color={theme.colors.ink.muted}
                            style={{ flexShrink: 0 }}
                          />
                        </ConnectedPaperRow>
                      )
                    })}
                </div>
              </ConnectedPapersSection>
            </div>
          )}

          {/* TAB 2: INTERACTIVE CITATION GRAPH */}
          {tab === 'graph' && (
            <CitationGraphVisualizer
              references={references}
              edges={edges}
              activeRefId={ref.id}
              onSelectRef={onSelectRef}
            />
          )}
        </DrawerContent>

        {/* Drawer Footer Actions */}
        <DrawerFooter>
          {/* BibTeX Copy Button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopyBibtex}
            startIcon={
              copiedBibtex ? (
                <Check size={13} color={theme.colors.status.success} />
              ) : (
                <Copy size={13} />
              )
            }
          >
            {copiedBibtex ? 'BibTeX Copied' : 'Copy BibTeX'}
          </Button>

          {/* Direct DOI External Link */}
          <ExternalPaperLink
            href={`https://doi.org/${ref.doi}`}
            target="_blank"
            rel="noreferrer"
          >
            <span>Open Original Paper</span>
            <ExternalLink size={12} />
          </ExternalPaperLink>
        </DrawerFooter>
      </DrawerPanel>
    </DrawerBackdrop>
  )
}
