import styled from '@emotion/styled'
import { ArrowRight, Compass } from 'lucide-react'

import { Button } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ChatCheckpoint } from '#/routes/-types/research.type'

type CheckpointAngle = NonNullable<ChatCheckpoint['angles']>[number]

interface ChatMessageAngleCheckpointProps {
  angles: Array<CheckpointAngle>
  onSelectAngle?: (angleId: string) => void
}

const CheckpointCard = styled.div`
  background-color: ${theme.colors.canvas.subtle};
  border: 1px solid rgba(121, 110, 255, 0.3);
  border-radius: ${theme.radius.lg};
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: 0 4px 12px rgba(121, 110, 255, 0.08);
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
  color: ${theme.colors.ink.primary};
`

const DescriptionText = styled.p`
  font-size: 0.84rem;
  color: ${theme.colors.ink.secondary};
  margin: 0;
  line-height: ${theme.typography.lineHeights.normal};
`

const AnglesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 10px;
  margin-top: 4px;
`

const AngleCard = styled.div`
  background-color: ${theme.colors.canvas.primary};
  border: 1px solid rgba(100, 111, 121, 0.18);
  border-radius: ${theme.radius.button};
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  transition: ${theme.transitions.fast};
`

const AngleTitle = styled.div`
  font-weight: ${theme.typography.weights.semibold};
  font-size: 0.86rem;
  color: ${theme.colors.ink.primary};
  margin-bottom: 4px;
`

const AngleDesc = styled.div`
  font-size: 0.78rem;
  color: ${theme.colors.ink.secondary};
  line-height: ${theme.typography.lineHeights.normal};
`

export const ChatMessageAngleCheckpoint = ({
  angles,
  onSelectAngle,
}: ChatMessageAngleCheckpointProps) => {
  return (
    <CheckpointCard>
      <HeaderRow>
        <Compass size={16} color={theme.colors.accent.purple} />
        <HeaderTitle>Checkpoint: Select Research Angle</HeaderTitle>
      </HeaderRow>
      <DescriptionText>
        Scholar AI identified the following academic perspectives. Select an
        angle to guide literature retrieval:
      </DescriptionText>

      <AnglesGrid>
        {angles.map((ang) => (
          <AngleCard key={ang.id}>
            <div>
              <AngleTitle>{ang.title}</AngleTitle>
              <AngleDesc>{ang.description}</AngleDesc>
            </div>

            <Button
              variant="violet"
              size="sm"
              onClick={() => onSelectAngle?.(ang.id)}
              endIcon={<ArrowRight size={13} />}
            >
              Select This Angle
            </Button>
          </AngleCard>
        ))}
      </AnglesGrid>
    </CheckpointCard>
  )
}
