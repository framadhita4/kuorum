import { useEffect, useRef, useState } from 'react'

import styled from '@emotion/styled'
import { Bookmark, Check, PanelRight, Share2, Sparkles } from 'lucide-react'

import { Badge, Button } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ChatSubmissionMode } from '#/routes/-components/ChatInputBar'
import { ChatInputBar } from '#/routes/-components/ChatInputBar'
import { ChatMessageItem } from '#/routes/-components/ChatMessageItem'
import type { ChatMessage, ResearchTopic } from '#/routes/-types/research.type'
import type { ResearchApproach, ResearchPipeline } from '#/types/pipeline.type'

interface MainChatProps {
  topic: ResearchTopic
  messages: Array<ChatMessage>
  isResearching: boolean
  isRightPanelOpen: boolean
  isNewThread?: boolean
  onToggleRightPanel: () => void
  onOpenDetails: () => void
  onOpenGraph: () => void
  onSelectCitation?: (citationId: string) => void
  onSubmitPrompt: (
    prompt: string,
    submissionMode: ChatSubmissionMode,
    approach: ResearchApproach,
  ) => void
  widthPercent?: number
  isResizing?: boolean
  activePipeline?: ResearchPipeline | null
  onSelectAngle?: (angleId: string) => void
  onSelectAnchor?: (anchorId: string) => void
  onCancelPipeline?: () => void
}

const MainChatContainer = styled.div<{
  $isRightPanelOpen: boolean
  $widthPercent: number
  $isResizing: boolean
}>`
  flex: ${({ $isRightPanelOpen, $widthPercent }) =>
    $isRightPanelOpen ? `0 0 ${$widthPercent}%` : 1};
  width: ${({ $isRightPanelOpen, $widthPercent }) =>
    $isRightPanelOpen ? `${$widthPercent}%` : '100%'};
  min-width: ${({ $isRightPanelOpen }) =>
    $isRightPanelOpen ? '320px' : 'auto'};
  background-color: ${theme.colors.canvas.primary};
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  position: relative;
  font-family: ${theme.typography.fontFamily.sans};
  border-right: none;
  transition: ${({ $isResizing }) =>
    $isResizing ? 'none' : 'flex 0.2s ease, width 0.2s ease'};
`

const ChatHeader = styled.div`
  height: 60px;
  border-bottom: ${theme.colors.border.subtle};
  padding: 0 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  background-color: ${theme.colors.canvas.primary};
  z-index: 10;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
  overflow: hidden;
`

const TopicTitle = styled.h2`
  font-size: 1.05rem;
  font-weight: ${theme.typography.weights.regular};
  color: ${theme.colors.ink.primary};
  margin: 0;
  letter-spacing: ${theme.typography.letterSpacing.tight};
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const PanelToggleButton = styled.button<{ $isOpen: boolean }>`
  background: ${({ $isOpen }) =>
    $isOpen ? theme.colors.accent.midnight : theme.colors.canvas.neutral};
  border: none;
  border-radius: ${theme.radius.button};
  padding: 8px 12px;
  color: ${({ $isOpen }) => ($isOpen ? '#ffffff' : theme.colors.ink.secondary)};
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  font-weight: ${theme.typography.weights.medium};
  outline: none;
  transition: ${theme.transitions.fast};

  &:hover {
    background: ${({ $isOpen }) =>
      $isOpen ? theme.colors.accent.midnight : theme.colors.neutral['200']};
    color: ${({ $isOpen }) => ($isOpen ? '#ffffff' : theme.colors.ink.primary)};
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }
`

const BookmarkButton = styled.button<{ $isBookmarked: boolean }>`
  background: ${({ $isBookmarked }) =>
    $isBookmarked ? 'rgba(240, 106, 106, 0.12)' : 'transparent'};
  border: 1px solid
    ${({ $isBookmarked }) =>
      $isBookmarked ? theme.colors.brand.coral : 'rgba(100, 111, 121, 0.2)'};
  border-radius: ${theme.radius.button};
  padding: 8px;
  color: ${({ $isBookmarked }) =>
    $isBookmarked ? theme.colors.brand.coral : theme.colors.ink.secondary};
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  outline: none;
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.brand.coral};
    border-color: ${theme.colors.brand.coral};
    background: rgba(240, 106, 106, 0.08);
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.brand.coral};
    outline-offset: 2px;
  }
`

const MessagesStream = styled.div<{ $isRightPanelOpen: boolean }>`
  flex: 1;
  overflow-y: auto;
  padding: ${({ $isRightPanelOpen }) =>
    $isRightPanelOpen ? '24px 28px 160px' : '36px 48px 160px'};
  display: flex;
  flex-direction: column;
  gap: 28px;
  max-width: ${({ $isRightPanelOpen }) => ($isRightPanelOpen ? '100%' : '840px')};
  margin: 0 auto;
  width: 100%;
`

const EmptyStateContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  margin-top: 40px;
  gap: 12px;
`

const EmptyStateIcon = styled.div`
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background-color: rgba(121, 110, 255, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.accent.purple};
`

const EmptyStateTitle = styled.h3`
  font-size: 1.25rem;
  font-weight: ${theme.typography.weights.regular};
  color: ${theme.colors.ink.primary};
  margin: 0;
`

const EmptyStateDescription = styled.p`
  font-size: 0.92rem;
  color: ${theme.colors.ink.secondary};
  max-width: 460px;
  margin: 0;
  line-height: ${theme.typography.lineHeights.normal};
`

export const MainChat = ({
  topic,
  messages,
  isResearching,
  isRightPanelOpen,
  isNewThread = false,
  onToggleRightPanel,
  onOpenDetails,
  onOpenGraph,
  onSelectCitation,
  onSubmitPrompt,
  widthPercent = 50,
  isResizing = false,
  activePipeline,
  onSelectAngle,
  onSelectAnchor,
  onCancelPipeline,
}: MainChatProps) => {
  const [isBookmarked, setIsBookmarked] = useState(false)
  const [isCopiedShare, setIsCopiedShare] = useState(false)

  const messagesEndRef = useRef<HTMLDivElement | null>(null)
  const scrollContainerRef = useRef<HTMLDivElement | null>(null)

  // Auto-scroll to bottom when new messages arrive or research status changes
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages.length, isResearching])

  const handleShare = () => {
    void navigator.clipboard.writeText(window.location.href).then(() => {
      setIsCopiedShare(true)
      setTimeout(() => setIsCopiedShare(false), 2000)
    })
  }

  return (
    <MainChatContainer
      $isRightPanelOpen={isRightPanelOpen}
      $widthPercent={widthPercent}
      $isResizing={isResizing}
    >
      {/* Top Header Bar */}
      <ChatHeader>
        <HeaderLeft>
          <TopicTitle title={topic.title}>{topic.title}</TopicTitle>
          {activePipeline && (
            <Badge
              variant={
                activePipeline.status === 'COMPLETED' ? 'success' : 'violet'
              }
              size="sm"
            >
              {activePipeline.status}
            </Badge>
          )}
        </HeaderLeft>

        <HeaderRight>
          <PanelToggleButton
            type="button"
            onClick={onToggleRightPanel}
            $isOpen={isRightPanelOpen}
            title={isRightPanelOpen ? 'Hide Detail Panel' : 'Show Detail Panel'}
          >
            <PanelRight size={15} />
          </PanelToggleButton>

          <BookmarkButton
            type="button"
            onClick={() => setIsBookmarked(!isBookmarked)}
            $isBookmarked={isBookmarked}
            title={isBookmarked ? 'Remove from Bookmarks' : 'Save to Bookmarks'}
          >
            <Bookmark size={15} fill={isBookmarked ? 'currentColor' : 'none'} />
          </BookmarkButton>

          <Button
            variant="primary"
            size="sm"
            onClick={handleShare}
            startIcon={
              isCopiedShare ? <Check size={13} /> : <Share2 size={13} />
            }
          >
            {isCopiedShare ? 'Tautan Tersalin' : 'Share'}
          </Button>
        </HeaderRight>
      </ChatHeader>

      {/* Main Chat Scrollable Stream */}
      <MessagesStream
        ref={scrollContainerRef}
        $isRightPanelOpen={isRightPanelOpen}
      >
        {/* Welcome / Empty state if no messages */}
        {messages.length === 0 && (
          <EmptyStateContainer>
            <EmptyStateIcon>
              <Sparkles size={22} />
            </EmptyStateIcon>
            <EmptyStateTitle>Scholar Research Chat</EmptyStateTitle>
            <EmptyStateDescription>
              Propose scientific hypotheses, deconstruct research concepts, or
              dive deep with verified journal citations.
            </EmptyStateDescription>
          </EmptyStateContainer>
        )}

        {/* Message Stream */}
        {messages.map((msg) => (
          <ChatMessageItem
            key={msg.id}
            message={msg}
            references={topic.references}
            isResearching={isResearching}
            onSelectAngle={onSelectAngle}
            onSelectAnchor={onSelectAnchor}
            onSelectCitation={onSelectCitation}
            onOpenDetails={onOpenDetails}
            onOpenGraph={onOpenGraph}
            onSelectPrompt={(p) => onSubmitPrompt(p, 'follow_up', 'NORMAL')}
          />
        ))}

        <div ref={messagesEndRef} style={{ height: '1px' }} />
      </MessagesStream>

      {/* Floating Prompt Input Bar at Bottom */}
      <ChatInputBar
        isRightPanelOpen={isRightPanelOpen}
        isResearching={isResearching}
        isNewThread={isNewThread}
        onSubmit={onSubmitPrompt}
        onCancel={onCancelPipeline}
      />
    </MainChatContainer>
  )
}
