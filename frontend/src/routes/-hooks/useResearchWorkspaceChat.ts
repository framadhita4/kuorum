import { useEffect, useState } from 'react'

import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'

import { useChatMessages } from '#/-components/providers/ChatMessagesContext'
import type { ChatSubmissionMode } from '#/routes/-components/ChatInputBar'
import {
  NEW_RESEARCH_TOPIC,
  RESEARCH_TOPICS,
} from '#/routes/-data/researchTopics'
import type { ChatMessage } from '#/routes/-types/research.type'
import { generateFollowUpAnswer } from '#/routes/-utils/chatAssistant'
import { mapPipelineToTopic } from '#/routes/-utils/pipelineAdapter'
import {
  reducePipelineStreamEvent,
  resolveTopicMessages,
} from '#/routes/-utils/pipelineStreamHelper'
import { useCancelPipeline } from '#/services/pipeline/useCancelPipeline'
import { useGetPipeline } from '#/services/pipeline/useGetPipeline'
import { useListPipelines } from '#/services/pipeline/useListPipelines'
import { usePipelineStream } from '#/services/pipeline/usePipelineStream'
import { useSelectAnchor } from '#/services/pipeline/useSelectAnchor'
import { useSelectAngle } from '#/services/pipeline/useSelectAngle'
import { useStartPipeline } from '#/services/pipeline/useStartPipeline'
import type { PaperMetadata } from '#/types/paper.type'
import type { PipelineEvent, ResearchApproach } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export interface UseResearchWorkspaceChatOptions {
  chatId?: string
}

export function useResearchWorkspaceChat({
  chatId,
}: UseResearchWorkspaceChatOptions) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { messagesMap, setMessagesMap, setIsRightPanelOpen } = useChatMessages()

  const activeTopicKey = chatId && chatId !== 'new' ? chatId : 'new'
  const isPipelineKey = activeTopicKey.startsWith('pipe_')

  const [isResearching, setIsResearching] = useState(false)

  // Clear research status when switching topic
  useEffect(() => {
    setIsResearching(false)
  }, [activeTopicKey])

  // Real backend pipeline hooks
  const { data: pipelines } = useListPipelines(20)
  const { data: activePipeline } = useGetPipeline(
    isPipelineKey ? activeTopicKey : null,
  )

  const startPipelineMutation = useStartPipeline()
  const selectAngleMutation = useSelectAngle()
  const selectAnchorMutation = useSelectAnchor()
  const cancelPipelineMutation = useCancelPipeline()

  // Map real pipeline data to visualizer topic if available
  const mappedTopic = mapPipelineToTopic(activePipeline)
  const fallbackTopic =
    activeTopicKey === 'new'
      ? NEW_RESEARCH_TOPIC
      : (RESEARCH_TOPICS[activeTopicKey] ?? NEW_RESEARCH_TOPIC)

  const currentTopic = mappedTopic ?? fallbackTopic

  // Ensure conversation exists and is kept in sync with backend pipeline state
  useEffect(() => {
    if (activeTopicKey === 'new') {
      return
    }

    setMessagesMap((prev) => {
      const existing = prev[activeTopicKey]
      const resolved = resolveTopicMessages(
        existing,
        currentTopic,
        activePipeline,
      )
      if (!resolved) {
        return prev
      }
      return {
        ...prev,
        [activeTopicKey]: resolved,
      }
    })
  }, [activeTopicKey, currentTopic, activePipeline, setMessagesMap])

  const streamPipelineId = isPipelineKey ? activeTopicKey : activePipeline?.id

  // Real-time SSE Stream Listener
  const handleStreamEvent = (event: PipelineEvent) => {
    // Invalidate pipeline query cache on stages or checkpoints
    if (streamPipelineId) {
      void queryClient.invalidateQueries({
        queryKey: ['pipeline', streamPipelineId],
      })
    }

    setMessagesMap((prev) => {
      const currentMessages = prev[activeTopicKey] ?? []
      const updatedList = reducePipelineStreamEvent(
        currentMessages,
        event,
        activePipeline,
        mappedTopic?.summary,
      )
      return {
        ...prev,
        [activeTopicKey]: updatedList,
      }
    })
  }

  const { isStreaming } = usePipelineStream(streamPipelineId, {
    onEvent: handleStreamEvent,
    onCompleted: () => {
      setIsResearching(false)
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
      if (streamPipelineId) {
        void queryClient.invalidateQueries({
          queryKey: ['pipeline', streamPipelineId],
        })
      }
    },
  })

  const handlePromptSubmit = async (
    prompt: string,
    submissionMode: ChatSubmissionMode,
    approach: ResearchApproach = 'NORMAL',
  ) => {
    const timestamp = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    })

    // If starting a new pipeline session or currently on 'new' canvas
    if (
      submissionMode === 'new_pipeline' ||
      activeTopicKey === 'new' ||
      !isPipelineKey
    ) {
      setIsResearching(true)
      try {
        const created = await startPipelineMutation.mutateAsync({
          prompt,
          mode: 'AUTONOMOUS',
          approach,
        })

        const userMessage: ChatMessage = {
          id: `usr-${Date.now()}`,
          role: 'user',
          content: prompt,
          timestamp,
        }

        const initialAssistant: ChatMessage = {
          id: `ans-${Date.now()}`,
          role: 'assistant',
          content: '',
          timestamp,
          status: 'streaming',
          isLiveProgress: true,
          isThinkingActive: true,
          liveProgressMessage:
            approach === 'DEEP_RESEARCH'
              ? 'Starting Deep Research: Retrieving 100-200 papers & Cloudflare Clef classification...'
              : 'Contacting AI to deconstruct research angles & score candidate papers...',
          statsSummary: {
            retrieved: 0,
            eligible: 0,
            included: 0,
          },
          thinkingSteps: [
            {
              id: `step-init-thought-${Date.now()}`,
              type: 'thought',
              description:
                approach === 'DEEP_RESEARCH'
                  ? `Initializing Deep Research for: "${prompt}". Targeting 100-200 papers with Cloudflare Clef rapid classifier.`
                  : `Initializing autonomous research pipeline for: "${prompt}". Targeting candidate papers with composite scoring.`,
              status: 'in_progress',
            },
            {
              id: `step-init-decomp-${Date.now()}`,
              type: 'deconstruct',
              label: 'Deconstruct',
              description:
                'Academic concept deconstruction & research angle mapping',
              status: 'pending',
            },
          ],
          steps: [
            'Initializing autonomous research pipeline session',
            'Analyzing scientific domain & ontological terminology',
          ],
        }

        setMessagesMap((prev) => ({
          ...prev,
          [created.id]: [userMessage, initialAssistant],
        }))

        setIsRightPanelOpen(true)
        await queryClient.invalidateQueries({ queryKey: ['pipelines'] })

        // Redirect to the dedicated chat route
        void navigate({
          to: '/chat/$chatId',
          params: { chatId: created.id },
        })
      } catch (err) {
        console.error('Failed to start research pipeline:', err)
        setIsResearching(false)
      }
      return
    }

    // Follow-up chat within active research topic context
    const userMessage: ChatMessage = {
      id: `usr-${Date.now()}`,
      role: 'user',
      content: prompt,
      timestamp,
    }

    const pendingAssistantMessage: ChatMessage = {
      id: `ans-${Date.now()}`,
      role: 'assistant',
      content: '',
      timestamp,
      status: 'streaming',
      isLiveProgress: true,
      isThinkingActive: true,
      liveProgressMessage:
        'Searching active research corpus & verifying citations...',
      statsSummary: {
        retrieved: currentTopic.references.length * 3 || 18,
        eligible: currentTopic.references.length || 6,
        included: Math.min(3, currentTopic.references.length),
      },
      thinkingSteps: [
        {
          id: `step-followup-init-${Date.now()}`,
          type: 'thought',
          description: `Analyzing focus of question "${prompt}" against verified journal literature.`,
          status: 'in_progress',
        },
      ],
    }

    setMessagesMap((prev) => ({
      ...prev,
      [activeTopicKey]: [
        ...(prev[activeTopicKey] ?? []),
        userMessage,
        pendingAssistantMessage,
      ],
    }))

    setIsResearching(true)
    try {
      const answered = await generateFollowUpAnswer(
        prompt,
        currentTopic,
        (partial) => {
          setMessagesMap((prev) => {
            const currentList = prev[activeTopicKey] ?? []
            const updated = currentList.map((m) =>
              m.id === pendingAssistantMessage.id ? { ...m, ...partial } : m,
            )
            return {
              ...prev,
              [activeTopicKey]: updated,
            }
          })
        },
      )

      setMessagesMap((prev) => {
        const currentList = prev[activeTopicKey] ?? []
        const updated = currentList.map((m) =>
          m.id === pendingAssistantMessage.id
            ? {
                ...answered,
                id: pendingAssistantMessage.id,
                timestamp: new Date().toLocaleTimeString([], {
                  hour: '2-digit',
                  minute: '2-digit',
                }),
                status: 'done' as const,
                isLiveProgress: false,
                isThinkingActive: false,
              }
            : m,
        )
        return {
          ...prev,
          [activeTopicKey]: updated,
        }
      })

      // Persist user and assistant exchange to backend SQLite store
      if (activePipeline?.id) {
        void fetchHelper(`/api/v1/pipelines/${activePipeline.id}/messages`, {
          method: 'POST',
          body: {
            role: 'user',
            content: prompt,
          },
        })
        void fetchHelper(`/api/v1/pipelines/${activePipeline.id}/messages`, {
          method: 'POST',
          body: {
            role: 'assistant',
            content: answered.content,
            steps: answered.steps,
            suggestedFollowUps: answered.suggestedFollowUps,
          },
        })
      }
    } catch (err) {
      console.error('Failed to answer follow-up query:', err)
      setMessagesMap((prev) => {
        const currentList = prev[activeTopicKey] ?? []
        const updated = currentList.map((m) =>
          m.id === pendingAssistantMessage.id
            ? {
                ...m,
                content:
                  'Sorry, an error occurred while processing the response from the research corpus.',
                status: 'error' as const,
                isLiveProgress: false,
              }
            : m,
        )
        return {
          ...prev,
          [activeTopicKey]: updated,
        }
      })
    } finally {
      setIsResearching(false)
    }
  }

  const handleSelectAngle = async (angleId: string) => {
    if (!activePipeline) return
    setIsResearching(true)
    try {
      await selectAngleMutation.mutateAsync({
        pipelineId: activePipeline.id,
        angleId,
      })
      await queryClient.invalidateQueries({
        queryKey: ['pipeline', activePipeline.id],
      })
    } catch (err) {
      console.error('Failed to select angle:', err)
      setIsResearching(false)
    }
  }

  const handleSelectAnchor = async (anchorPaperId: string) => {
    if (!activePipeline) return
    setIsResearching(true)
    try {
      await selectAnchorMutation.mutateAsync({
        pipelineId: activePipeline.id,
        anchorPaperId,
      })
      await queryClient.invalidateQueries({
        queryKey: ['pipeline', activePipeline.id],
      })
    } catch (err) {
      console.error('Failed to select anchor paper:', err)
      setIsResearching(false)
    }
  }

  const handleCancelPipeline = async () => {
    if (!activePipeline) return
    try {
      await cancelPipelineMutation.mutateAsync({
        pipelineId: activePipeline.id,
      })
      await queryClient.invalidateQueries({
        queryKey: ['pipeline', activePipeline.id],
      })
      setIsResearching(false)
    } catch (err) {
      console.error('Failed to cancel pipeline:', err)
    }
  }

  const handleStartPipelineFromPaper = async (paper: PaperMetadata) => {
    setIsResearching(true)
    const timestamp = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    })
    try {
      const prompt = `Deep analysis of paper "${paper.title}" (${paper.year || 2024}) including citation network mapping and methodological evidence.`
      const created = await startPipelineMutation.mutateAsync({
        prompt,
        mode: 'AUTONOMOUS',
      })

      const userMessage: ChatMessage = {
        id: `usr-${Date.now()}`,
        role: 'user',
        content: prompt,
        timestamp,
      }

      const initialAssistant: ChatMessage = {
        id: `ans-${Date.now()}`,
        role: 'assistant',
        content: '',
        timestamp,
        status: 'streaming',
        isLiveProgress: true,
        isThinkingActive: true,
        liveProgressMessage: `Analyzing paper "${paper.title}" & preparing citation traversal...`,
        thinkingSteps: [
          {
            id: `step-paper-init-${Date.now()}`,
            type: 'thought',
            description: `Initializing deep dive analysis for paper: "${paper.title}".`,
            status: 'in_progress',
          },
        ],
      }

      setMessagesMap((prev) => ({
        ...prev,
        [created.id]: [userMessage, initialAssistant],
      }))

      await queryClient.invalidateQueries({ queryKey: ['pipelines'] })
      setIsRightPanelOpen(true)

      // Redirect to newly started chat
      void navigate({
        to: '/chat/$chatId',
        params: { chatId: created.id },
      })
    } catch (err) {
      console.error('Failed to start research from paper:', err)
      setIsResearching(false)
    }
  }

  const activeMessages = messagesMap[activeTopicKey] ?? []
  const isResearchingStatus =
    isResearching || isStreaming || activePipeline?.status === 'RUNNING'

  return {
    activeTopicKey,
    activePipeline,
    currentTopic,
    activeMessages,
    isResearching: isResearchingStatus,
    pipelines,
    handlePromptSubmit,
    handleSelectAngle,
    handleSelectAnchor,
    handleCancelPipeline,
    handleStartPipelineFromPaper,
  }
}
