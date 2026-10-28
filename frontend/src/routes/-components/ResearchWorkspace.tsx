import { useEffect, useState } from 'react'

import { useNavigate } from '@tanstack/react-router'

import styled from '@emotion/styled'

import { Sidebar } from '#/-components/layout/Sidebar'
import { useChatMessages } from '#/-components/providers/ChatMessagesContext'
import { theme } from '#/-components/ui/theme'
import { DetailPanel } from '#/routes/-components/DetailPanel'
import { MainChat } from '#/routes/-components/MainChat'
import { PaperSearchDialog } from '#/routes/-components/PaperSearchDialog'
import { NEW_RESEARCH_TOPIC } from '#/routes/-data/researchTopics'
import { useResearchWorkspaceChat } from '#/routes/-hooks/useResearchWorkspaceChat'
import { useWorkspaceResizer } from '#/routes/-hooks/useWorkspaceResizer'

// Re-export for external consumers / backwards compatibility
export { NEW_RESEARCH_TOPIC }

const WorkspaceRoot = styled.div<{ $isDragging: boolean }>`
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: ${theme.colors.canvas.primary};
  color: ${theme.colors.ink.primary};
  overflow: hidden;
  font-family: ${theme.typography.fontFamily.sans};
  cursor: ${({ $isDragging }) => ($isDragging ? 'col-resize' : 'default')};
  user-select: ${({ $isDragging }) => ($isDragging ? 'none' : 'auto')};
`

const MainContentArea = styled.div`
  flex: 1;
  display: flex;
  height: 100%;
  overflow: hidden;
  position: relative;
`

const ResizerSplitter = styled.div<{ $isDragging: boolean }>`
  width: 8px;
  cursor: col-resize;
  background-color: ${({ $isDragging }) =>
    $isDragging ? theme.colors.accent.purple : theme.colors.canvas.subtle};
  border-left: 1px solid rgba(100, 111, 121, 0.2);
  border-right: 1px solid rgba(100, 111, 121, 0.2);
  position: relative;
  flex-shrink: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: ${({ $isDragging }) =>
    $isDragging ? 'none' : 'background-color 0.15s ease'};

  &:hover {
    background-color: ${({ $isDragging }) =>
      $isDragging ? theme.colors.accent.purple : theme.colors.canvas.neutral};
  }
`

const ResizerHandle = styled.div<{ $isDragging: boolean }>`
  width: 2px;
  height: 28px;
  border-radius: ${theme.radius.pill};
  background-color: ${({ $isDragging }) =>
    $isDragging ? '#FFFFFF' : theme.colors.ink.muted};
  transition: background-color 0.15s ease;
`

export interface ResearchWorkspaceProps {
  chatId?: string
}

export const ResearchWorkspace = ({ chatId }: ResearchWorkspaceProps) => {
  const navigate = useNavigate()

  const {
    splitRatio,
    setSplitRatio,
    isRightPanelOpen,
    setIsRightPanelOpen,
    rightPanelTab,
    setRightPanelTab,
  } = useChatMessages()

  const [isSearchDialogOpen, setIsSearchDialogOpen] = useState(false)
  const [selectedRefId, setSelectedRefId] = useState<string | undefined>(
    undefined,
  )

  const {
    activeTopicKey,
    activePipeline,
    currentTopic,
    activeMessages,
    isResearching,
    pipelines,
    handlePromptSubmit,
    handleSelectAngle,
    handleSelectAnchor,
    handleCancelPipeline,
    handleStartPipelineFromPaper,
  } = useResearchWorkspaceChat({ chatId })

  const { containerRef, isDragging, handleMouseDown, handleResetRatio } =
    useWorkspaceResizer({
      onRatioChange: setSplitRatio,
    })

  // Clear selected citation when active chat topic switches
  useEffect(() => {
    setSelectedRefId(undefined)
  }, [activeTopicKey])

  const handleSelectTopic = (key: string) => {
    void navigate({
      to: '/chat/$chatId',
      params: { chatId: key },
    })
  }

  const handleNewThread = () => {
    void navigate({
      to: '/',
    })
  }

  const handleOpenDetails = () => {
    setIsRightPanelOpen(true)
    setRightPanelTab('report')
  }

  const handleOpenGraph = () => {
    setIsRightPanelOpen(true)
    setRightPanelTab('graph')
  }

  const handleSelectCitation = (citationId: string) => {
    setSelectedRefId(citationId)
    setIsRightPanelOpen(true)
    setRightPanelTab('references')
  }

  return (
    <WorkspaceRoot $isDragging={isDragging}>
      {/* 1st Column: Left Sidebar */}
      <Sidebar
        activeTopicKey={activeTopicKey}
        onSelectTopic={handleSelectTopic}
        onNewThread={handleNewThread}
        onOpenGraph={handleOpenGraph}
        onOpenSearch={() => setIsSearchDialogOpen(true)}
        pipelines={pipelines}
      />

      {/* Main Content Area: Chat + Resizer Splitter + Detail Panel */}
      <MainContentArea ref={containerRef}>
        {/* 2nd Column: Main Chat Stream */}
        <MainChat
          topic={currentTopic}
          messages={activeMessages}
          isResearching={isResearching}
          isRightPanelOpen={isRightPanelOpen}
          isNewThread={activeTopicKey === 'new'}
          onToggleRightPanel={() => setIsRightPanelOpen(!isRightPanelOpen)}
          onOpenDetails={handleOpenDetails}
          onOpenGraph={handleOpenGraph}
          onSelectCitation={handleSelectCitation}
          onSubmitPrompt={handlePromptSubmit}
          widthPercent={isRightPanelOpen ? splitRatio * 100 : 100}
          isResizing={isDragging}
          activePipeline={activePipeline}
          onSelectAngle={handleSelectAngle}
          onSelectAnchor={handleSelectAnchor}
          onCancelPipeline={handleCancelPipeline}
        />

        {/* Resizer Splitter Bar (Visible when right panel is open) */}
        {isRightPanelOpen && (
          <ResizerSplitter
            $isDragging={isDragging}
            onMouseDown={handleMouseDown}
            onDoubleClick={handleResetRatio}
            title="Drag to resize • Double-click to reset 50:50"
          >
            <ResizerHandle $isDragging={isDragging} />
          </ResizerSplitter>
        )}

        {/* 3rd Column: Right Detail Panel (Full Report, Citations, & Graph) */}
        {isRightPanelOpen && (
          <DetailPanel
            title={currentTopic.title}
            subtitle={currentTopic.subtitle}
            summary={currentTopic.summary}
            sections={currentTopic.sections}
            comparisonTable={currentTopic.comparisonTable}
            references={currentTopic.references}
            edges={currentTopic.edges}
            activeTab={rightPanelTab}
            onChangeTab={setRightPanelTab}
            onClose={() => setIsRightPanelOpen(false)}
            isResizing={isDragging}
            selectedRefId={selectedRefId}
            onSelectRef={setSelectedRefId}
          />
        )}
      </MainContentArea>

      {/* Global Academic Literature Search Dialog */}
      <PaperSearchDialog
        open={isSearchDialogOpen}
        onOpenChange={setIsSearchDialogOpen}
        onStartPipelineWithPaper={handleStartPipelineFromPaper}
      />
    </WorkspaceRoot>
  )
}
