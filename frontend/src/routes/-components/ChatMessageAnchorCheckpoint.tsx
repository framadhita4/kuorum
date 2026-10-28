import styled from '@emotion/styled'
import { ArrowRight, FileText } from 'lucide-react'

import { Badge, Button } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ChatCheckpoint } from '#/routes/-types/research.type'

type CandidatePaper = NonNullable<ChatCheckpoint['candidatePapers']>[number]

interface ChatMessageAnchorCheckpointProps {
  candidatePapers: Array<CandidatePaper>
  onSelectAnchor?: (anchorId: string) => void
}

const CheckpointContainer = styled.div`
  background-color: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: ${theme.radius.lg};
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: 0 4px 12px rgba(217, 119, 6, 0.08);
  font-family: ${theme.typography.fontFamily.sans};
`

const HeaderRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const HeaderTitle = styled.span`
  font-weight: ${theme.typography.weights.semibold};
  font-size: 0.9rem;
  color: #78350f;
`

const DescriptionText = styled.p`
  font-size: 0.84rem;
  color: #92400e;
  margin: 0;
  line-height: ${theme.typography.lineHeights.normal};
`

const PapersList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
`

const PaperCard = styled.div`
  background-color: ${theme.colors.canvas.primary};
  border: 1px solid #fde68a;
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
  gap: 6px;
  margin-bottom: 3px;
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

export const ChatMessageAnchorCheckpoint = ({
  candidatePapers,
  onSelectAnchor,
}: ChatMessageAnchorCheckpointProps) => {
  return (
    <CheckpointContainer>
      <HeaderRow>
        <FileText size={16} color="#D97706" />
        <HeaderTitle>Checkpoint: Confirm Primary Anchor Paper</HeaderTitle>
      </HeaderRow>
      <DescriptionText>
        The following are top-ranked paper candidates from the federated search.
        Select an anchor paper to download and extract in depth:
      </DescriptionText>

      <PapersList>
        {candidatePapers.slice(0, 4).map((paper, idx) => (
          <PaperCard key={paper.id}>
            <div style={{ flex: 1, minWidth: 0 }}>
              <PaperMetaRow>
                <Badge variant={idx === 0 ? 'amber' : 'neutral'} size="sm">
                  {idx === 0 ? '★ Top Recommendation' : `#${idx + 1}`}
                </Badge>
                <span style={{ fontSize: '0.76rem', color: '#64748B' }}>
                  {paper.year} • {paper.citationCount} citations
                </span>
              </PaperMetaRow>
              <PaperTitle>{paper.title}</PaperTitle>
              <AbstractSnippet>
                {paper.abstract?.slice(0, 120)}...
              </AbstractSnippet>
            </div>

            <Button
              variant={idx === 0 ? 'primary' : 'secondary'}
              size="sm"
              onClick={() => onSelectAnchor?.(paper.id)}
              endIcon={<ArrowRight size={13} />}
            >
              Select Anchor
            </Button>
          </PaperCard>
        ))}
      </PapersList>
    </CheckpointContainer>
  )
}
