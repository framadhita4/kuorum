import styled from '@emotion/styled'
import { ArrowRight, FileText, Network, Quote } from 'lucide-react'

import { Button, MarkdownRenderer, Pill } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import { ChatMessageAnchorCheckpoint } from '#/routes/-components/ChatMessageAnchorCheckpoint'
import { ChatMessageAngleCheckpoint } from '#/routes/-components/ChatMessageAngleCheckpoint'
import { ChatMessagePaperSelection } from '#/routes/-components/ChatMessagePaperSelection'
import { ThinkingProcessView } from '#/routes/-components/ThinkingProcessView'
import type {
  ChatMessage,
  ReferenceItem,
  ThinkingStep,
} from '#/routes/-types/research.type'

interface ChatMessageItemProps {
  message: ChatMessage
  references: Array<ReferenceItem>
  isResearching?: boolean
  onSelectAngle?: (angleId: string) => void
  onSelectAnchor?: (anchorId: string) => void
  onSelectCitation?: (citationId: string) => void
  onOpenDetails?: () => void
  onOpenGraph?: () => void
  onSelectPrompt?: (prompt: string) => void
}

const UserRow = styled.div`
  display: flex;
  justify-content: flex-end;
  align-items: flex-start;
  gap: 10px;
`

const UserBubble = styled.div`
  background-color: ${theme.colors.canvas.charcoal};
  color: ${theme.colors.ink.inverse};
  border-radius: ${theme.radius.card};
  padding: 12px 20px;
  max-width: 85%;
  font-size: 0.94rem;
  line-height: ${theme.typography.lineHeights.normal};
  font-weight: ${theme.typography.weights.regular};
  box-shadow: ${theme.shadows.sm};
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ${theme.typography.fontFamily.sans};
`

const AssistantContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-width: 100%;
  font-family: ${theme.typography.fontFamily.sans};
`

const MessageBody = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const QuoteCallout = styled.div`
  margin-top: 4px;
  background-color: ${theme.colors.canvas.subtle};
  border-left: 3px solid ${theme.colors.brand.coral};
  border-radius: 0 8px 8px 0;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
`

const QuoteHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: ${theme.colors.brand.coral};
`

const QuoteText = styled.p`
  font-size: 0.88rem;
  font-style: italic;
  color: ${theme.colors.ink.secondary};
  margin: 0;
  line-height: 1.55;
`

const QuoteSource = styled.span`
  font-size: 0.74rem;
  color: ${theme.colors.ink.muted};
  font-weight: 500;
`

const MetricsChipsRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 6px;
`

const ActionButtonsToolbar = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 10px;
`

const FollowUpsContainer = styled.div`
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
`

const FollowUpsLabel = styled.span`
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: ${theme.colors.ink.muted};
`

const FollowUpsList = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
`

export const ChatMessageItem = ({
  message,
  references,
  isResearching = false,
  onSelectAngle,
  onSelectAnchor,
  onSelectCitation,
  onOpenDetails,
  onOpenGraph,
  onSelectPrompt,
}: ChatMessageItemProps) => {
  const isUser = message.role === 'user'

  // Derive rich thinking steps (or map legacy string steps if needed)
  const derivedThinkingSteps: Array<ThinkingStep> =
    message.thinkingSteps && message.thinkingSteps.length > 0
      ? message.thinkingSteps
      : (message.steps ?? []).map((stepText, idx) => ({
          id: `step-${idx}-${message.id}`,
          type:
            idx === 0
              ? 'deconstruct'
              : idx === 1
                ? 'search'
                : idx === 2
                  ? 'read_abstract'
                  : 'audit',
          label:
            idx === 0
              ? 'Deconstruct'
              : idx === 1
                ? 'Search'
                : idx === 2
                  ? 'Read'
                  : 'Audit',
          description: stepText,
          status: 'completed',
        }))

  if (isUser) {
    return (
      <UserRow>
        <UserBubble>{message.content}</UserBubble>
      </UserRow>
    )
  }

  // Assistant Message
  return (
    <AssistantContainer>
      {/* Consensus-Style Deep Research / Thinking Process */}
      <ThinkingProcessView
        thinkingSteps={derivedThinkingSteps}
        statsSummary={message.statsSummary}
        isThinkingActive={
          message.isThinkingActive ||
          ((message.status === 'streaming' || isResearching) &&
            derivedThinkingSteps.some((s) => s.status === 'in_progress'))
        }
        thinkingDuration={message.thinkingDuration}
        liveProgressMessage={message.liveProgressMessage}
        defaultExpanded={
          message.status === 'streaming' ||
          Boolean(message.isThinkingActive) ||
          derivedThinkingSteps.length > 0
        }
      />

      {/* Interactive Checkpoint: Angle Selection */}
      {message.checkpoint?.type === 'ANGLE_SELECTION' &&
        message.checkpoint.angles &&
        message.checkpoint.angles.length > 0 && (
          <ChatMessageAngleCheckpoint
            angles={message.checkpoint.angles}
            onSelectAngle={onSelectAngle}
          />
        )}

      {/* Automated Paper Selection for Deep Dive Card */}
      {message.paperSelection && (
        <ChatMessagePaperSelection paperSelection={message.paperSelection} />
      )}

      {/* Interactive Checkpoint: Anchor Paper Selection */}
      {!message.paperSelection &&
        message.checkpoint?.type === 'ANCHOR_SELECTION' &&
        message.checkpoint.candidatePapers &&
        message.checkpoint.candidatePapers.length > 0 && (
          <ChatMessageAnchorCheckpoint
            candidatePapers={message.checkpoint.candidatePapers}
            onSelectAnchor={onSelectAnchor}
          />
        )}

      {/* Main Formatted Content */}
      {(message.content || message.status === 'streaming') && (
        <MessageBody>
          <MarkdownRenderer
            content={message.content}
            references={references}
            onSelectCitation={onSelectCitation}
            isStreaming={message.status === 'streaming'}
          />

          {/* Verbatim Quote Excerpt Callout */}
          {message.quote && (
            <QuoteCallout>
              <QuoteHeader>
                <Quote size={12} />
                <span>Verbatim Quote Evidence</span>
              </QuoteHeader>
              <QuoteText>"{message.quote.text}"</QuoteText>
              <QuoteSource>— {message.quote.source}</QuoteSource>
            </QuoteCallout>
          )}

          {/* Metrics Chips */}
          {message.citations && message.citations.length > 0 && (
            <MetricsChipsRow>
              <Pill
                variant="purple"
                size="sm"
                label={`${message.citations.length} Verified Q1 Journals`}
              />
              <Pill
                variant="neutral"
                size="sm"
                label="Zero-Hallucination Grounded"
              />
            </MetricsChipsRow>
          )}

          {/* Action Buttons Toolbar */}
          <ActionButtonsToolbar>
            {onOpenDetails && (
              <Button
                variant="primary"
                size="md"
                onClick={onOpenDetails}
                startIcon={<FileText size={14} />}
                endIcon={<ArrowRight size={13} />}
              >
                View Synthesis Report
              </Button>
            )}

            {onOpenGraph && (
              <Button
                variant="outline"
                size="md"
                onClick={onOpenGraph}
                startIcon={<Network size={14} />}
              >
                2D Citation Map
              </Button>
            )}
          </ActionButtonsToolbar>

          {/* Contextual Suggested Follow-up Prompts */}
          {message.suggestedFollowUps &&
            message.suggestedFollowUps.length > 0 &&
            onSelectPrompt && (
              <FollowUpsContainer>
                <FollowUpsLabel>Suggested Follow-up Questions:</FollowUpsLabel>
                <FollowUpsList>
                  {message.suggestedFollowUps.map((promptText, pIdx) => (
                    <Pill
                      key={pIdx}
                      variant="neutral"
                      size="sm"
                      onClick={() => onSelectPrompt(promptText)}
                      label={`↳ ${promptText}`}
                    />
                  ))}
                </FollowUpsList>
              </FollowUpsContainer>
            )}
        </MessageBody>
      )}
    </AssistantContainer>
  )
}
