import styled from '@emotion/styled'
import { Check, Cpu, Layers } from 'lucide-react'

import { Badge } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ChatPaperSelectionData } from '#/types/pipeline.type'

interface ChatMessagePaperSelectionProps {
  paperSelection: ChatPaperSelectionData
}

const SelectionContainer = styled.div<{ $isDeep: boolean }>`
  background-color: ${({ $isDeep }) => ($isDeep ? '#F5F3FF' : '#F0FDFA')};
  border: 1px solid ${({ $isDeep }) => ($isDeep ? '#C4B5FD' : '#99F6E4')};
  border-radius: ${theme.radius.lg};
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: 0 4px 12px
    ${({ $isDeep }) =>
      $isDeep ? 'rgba(121, 110, 255, 0.08)' : 'rgba(13, 148, 136, 0.08)'};
  font-family: ${theme.typography.fontFamily.sans};
`

const HeaderRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const TitleText = styled.span<{ $isDeep: boolean }>`
  font-weight: ${theme.typography.weights.semibold};
  font-size: 0.9rem;
  color: ${({ $isDeep }) => ($isDeep ? '#4C1D95' : '#134E4A')};
`

const SummaryText = styled.p<{ $isDeep: boolean }>`
  font-size: 0.84rem;
  color: ${({ $isDeep }) => ($isDeep ? '#5B21B6' : '#115E59')};
  margin: 0;
  line-height: ${theme.typography.lineHeights.normal};
`

const PapersList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
`

const PaperItemCard = styled.div<{ $isDeep: boolean }>`
  background-color: ${theme.colors.canvas.primary};
  border: 1px solid ${({ $isDeep }) => ($isDeep ? '#DDD6FE' : '#A7F3D0')};
  border-radius: ${theme.radius.button};
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
`

const PaperMetaRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
`

const PaperTitle = styled.div`
  font-weight: ${theme.typography.weights.semibold};
  font-size: 0.85rem;
  color: #1e293b;
  margin-bottom: 2px;
`

const AbstractSnippet = styled.div`
  font-size: 0.75rem;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

export const ChatMessagePaperSelection = ({
  paperSelection,
}: ChatMessagePaperSelectionProps) => {
  const isDeep = paperSelection.approach === 'DEEP_RESEARCH'

  return (
    <SelectionContainer $isDeep={isDeep}>
      <HeaderRow>
        <HeaderLeft>
          {isDeep ? (
            <Cpu size={16} color={theme.colors.accent.purple} />
          ) : (
            <Layers size={16} color="#0D9488" />
          )}
          <TitleText $isDeep={isDeep}>
            {isDeep
              ? 'Deep Research: Rapid Cloudflare Clef Screening'
              : 'Normal Research: Academic Relevance Point Scoring'}
          </TitleText>
        </HeaderLeft>

        <Badge variant={isDeep ? 'violet' : 'emerald'} size="sm">
          {paperSelection.selectedCount} Papers Selected for Deep Dive
        </Badge>
      </HeaderRow>

      <SummaryText $isDeep={isDeep}>
        {isDeep
          ? `Abstract screening across ${paperSelection.totalRetrieved} papers completed via the Clef rapid decision model. Papers with the highest relevance probability were automatically selected for deep dive:`
          : `Composite evaluation across 20 candidate papers completed based on relevance, citations, recency, and open access. Papers with the highest scores were automatically selected for deep dive:`}
      </SummaryText>

      <PapersList>
        {paperSelection.selectedPapers.map((paper, idx) => {
          const score = paperSelection.scores?.find(
            (s) => s.paperId === paper.id,
          )

          return (
            <PaperItemCard key={paper.id} $isDeep={isDeep}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <PaperMetaRow>
                  <Badge variant={idx === 0 ? 'amber' : 'neutral'} size="sm">
                    {idx === 0 ? '★ Primary Anchor Paper' : `#${idx + 1}`}
                  </Badge>

                  {score && (
                    <Badge variant={isDeep ? 'violet' : 'emerald'} size="sm">
                      {isDeep && score.clefRelevanceScore !== undefined
                        ? `Clef Score: ${(score.clefRelevanceScore * 100).toFixed(1)}% (${score.clefLabel || 'directly_relevant'})`
                        : `Score: ${score.totalPoints.toFixed(1)} / 100 pts`}
                    </Badge>
                  )}

                  <span style={{ fontSize: '0.76rem', color: '#64748B' }}>
                    {paper.year} • {paper.citationCount} citations
                  </span>
                </PaperMetaRow>

                <PaperTitle>{paper.title}</PaperTitle>
                <AbstractSnippet>
                  {paper.abstract
                    ? `${paper.abstract.slice(0, 140)}...`
                    : 'Abstract analyzed within corpus.'}
                </AbstractSnippet>
              </div>

              <Badge variant="success" size="sm">
                <Check size={11} strokeWidth={2.5} style={{ marginRight: 3 }} />
                Deep Dive
              </Badge>
            </PaperItemCard>
          )
        })}
      </PapersList>
    </SelectionContainer>
  )
}
