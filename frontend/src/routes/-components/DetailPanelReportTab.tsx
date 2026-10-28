import styled from '@emotion/styled'
import { Sparkles } from 'lucide-react'

import { CitationPill, MarkdownRenderer } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type {
  ComparisonRow,
  ReferenceItem,
  ReportSection,
} from '#/routes/-types/research.type'

interface DetailPanelReportTabProps {
  title: string
  subtitle: string
  sections: Array<ReportSection>
  comparisonTable: Array<ComparisonRow>
  references: Array<ReferenceItem>
  onSelectRef: (id: string) => void
  onChangeTab: (tab: 'report' | 'references' | 'graph') => void
  summary?: string
  reportMarkdown?: string
}

const ReportContainer = styled.div`
  max-width: 920px;
  margin: 0 auto;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 36px;
  font-family: ${theme.typography.fontFamily.sans};
`

const ReportTitle = styled.h3`
  font-size: clamp(2rem, 3.4vw, 2.5rem);
  font-weight: ${theme.typography.weights.light};
  color: ${theme.colors.ink.primary};
  margin: 0 0 12px;
  line-height: ${theme.typography.lineHeights.tight};
  letter-spacing: ${theme.typography.letterSpacing.tightest};
`

const ReportSubtitleWrapper = styled.div`
  font-size: 1.125rem;
  font-weight: ${theme.typography.weights.light};
  color: ${theme.colors.ink.secondary};
  margin: 0;
  line-height: ${theme.typography.lineHeights.normal};
  max-width: 65ch;

  p {
    font-size: 1.125rem;
    color: ${theme.colors.ink.secondary};
    margin: 0;
  }
`

const SummaryCard = styled.div`
  background-color: ${theme.colors.canvas.subtle};
  border: 1px solid rgba(100, 111, 121, 0.16);
  border-radius: ${theme.radius.card};
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  box-shadow: ${theme.shadows.sm};
`

const SummaryHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.76rem;
  font-weight: ${theme.typography.weights.semibold};
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: ${theme.colors.accent.purple};
`

const SectionsWrapper = styled.div`
  display: flex;
  flex-direction: column;
  gap: 32px;
`

const SectionBlock = styled.div`
  display: flex;
  flex-direction: column;
  gap: 14px;
`

const SectionHeading = styled.h4`
  font-size: 1.45rem;
  font-weight: ${theme.typography.weights.regular};
  color: ${theme.colors.ink.primary};
  margin: 0;
  letter-spacing: ${theme.typography.letterSpacing.tight};
`

const ParagraphsWrapper = styled.div`
  display: flex;
  flex-direction: column;
  gap: 18px;
`

const ParagraphBlock = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`

const AttachedCitationsRow = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 2px;
`

const AttachedCitationsLabel = styled.span`
  font-size: 0.76rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.muted};
  font-family: ${theme.typography.fontFamily.sans};
`

const ComparisonSection = styled.div`
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
`

const TableCard = styled.div`
  border: ${theme.colors.border.subtle};
  border-radius: ${theme.radius.card};
  overflow: hidden;
  box-shadow: ${theme.shadows.sm};
`

const StyledTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  font-size: 0.92rem;
  text-align: left;
`

const TableHeaderRow = styled.tr`
  background-color: ${theme.colors.canvas.subtle};
  border-bottom: ${theme.colors.border.subtle};
  color: ${theme.colors.ink.primary};
`

const TableHeaderCell = styled.th`
  padding: 14px 20px;
  font-weight: ${theme.typography.weights.medium};
`

const TableRow = styled.tr<{ $isLast: boolean }>`
  border-bottom: ${({ $isLast }) =>
    $isLast ? 'none' : '1px solid rgba(100, 111, 121, 0.12)'};
  background-color: ${theme.colors.canvas.primary};
`

const TableCellPrimary = styled.td`
  padding: 14px 20px;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
`

const TableCellSecondary = styled.td`
  padding: 14px 20px;
  color: ${theme.colors.ink.secondary};
  font-weight: ${theme.typography.weights.light};
`

export const DetailPanelReportTab = ({
  title,
  subtitle,
  sections,
  comparisonTable,
  references,
  onSelectRef,
  onChangeTab,
  summary,
  reportMarkdown,
}: DetailPanelReportTabProps) => {
  const activeMarkdown = reportMarkdown || summary

  return (
    <ReportContainer>
      {/* Header Meta */}
      <div>
        <ReportTitle>{title}</ReportTitle>
        {subtitle && (
          <ReportSubtitleWrapper>
            <MarkdownRenderer
              content={subtitle}
              references={references}
              onSelectCitation={(id) => {
                onSelectRef(id)
                onChangeTab('references')
              }}
            />
          </ReportSubtitleWrapper>
        )}
      </div>

      {/* Synthesis Executive Summary Callout if available */}
      {activeMarkdown && (
        <SummaryCard>
          <SummaryHeader>
            <Sparkles size={15} />
            <span>Executive Synthesis Summary</span>
          </SummaryHeader>
          <MarkdownRenderer
            content={activeMarkdown}
            references={references}
            onSelectCitation={(id) => {
              onSelectRef(id)
              onChangeTab('references')
            }}
          />
        </SummaryCard>
      )}

      {/* Sections with Clickable Citation Pills & Markdown formatting */}
      {sections.length > 0 && (
        <SectionsWrapper>
          {sections.map((sec) => (
            <SectionBlock key={sec.id}>
              <SectionHeading>{sec.heading}</SectionHeading>
              <ParagraphsWrapper>
                {sec.paragraphs.map((para, pIdx) => {
                  // Filter out citations that are already cited inline as [num]
                  const extraCitations = para.citations
                    .map((refId) => references.find((r) => r.id === refId))
                    .filter((refObj): refObj is ReferenceItem => {
                      if (!refObj) return false
                      return !para.text.includes(`[${refObj.num}]`)
                    })

                  return (
                    <ParagraphBlock key={pIdx}>
                      <MarkdownRenderer
                        content={para.text}
                        references={references}
                        onSelectCitation={(id) => {
                          onSelectRef(id)
                          onChangeTab('references')
                        }}
                      />
                      {extraCitations.length > 0 && (
                        <AttachedCitationsRow>
                          <AttachedCitationsLabel>
                            Grounded in:
                          </AttachedCitationsLabel>
                          {extraCitations.map((refObj) => (
                            <CitationPill
                              key={refObj.id}
                              num={refObj.num}
                              onClick={() => {
                                onSelectRef(refObj.id)
                                onChangeTab('references')
                              }}
                              title={`View paper: ${refObj.title}`}
                            />
                          ))}
                        </AttachedCitationsRow>
                      )}
                    </ParagraphBlock>
                  )
                })}
              </ParagraphsWrapper>
            </SectionBlock>
          ))}
        </SectionsWrapper>
      )}

      {/* Comparative Summary Table */}
      {comparisonTable.length > 0 && (
        <ComparisonSection>
          <SectionHeading as="h4">
            Comparison Table of Characteristics
          </SectionHeading>
          <TableCard>
            <StyledTable>
              <thead>
                <TableHeaderRow>
                  <TableHeaderCell>Material</TableHeaderCell>
                  <TableHeaderCell>performance</TableHeaderCell>
                  <TableHeaderCell>Stabilities</TableHeaderCell>
                </TableHeaderRow>
              </thead>
              <tbody>
                {comparisonTable.map((row, idx) => (
                  <TableRow
                    key={idx}
                    $isLast={idx === comparisonTable.length - 1}
                  >
                    <TableCellPrimary>{row.type}</TableCellPrimary>
                    <TableCellSecondary>{row.conductivity}</TableCellSecondary>
                    <TableCellSecondary>{row.stability}</TableCellSecondary>
                  </TableRow>
                ))}
              </tbody>
            </StyledTable>
          </TableCard>
        </ComparisonSection>
      )}
    </ReportContainer>
  )
}
