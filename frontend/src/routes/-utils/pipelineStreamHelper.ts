import type {
  ChatMessage,
  ResearchTopic,
  ThinkingStepType,
} from '#/routes/-types/research.type'
import { createInitialMessages } from '#/routes/-utils/chatAssistant'
import type { PaperMetadata } from '#/types/paper.type'
import type {
  PipelineEvent,
  ResearchAngle,
  ResearchPipeline,
} from '#/types/pipeline.type'

/**
 * Resolves the message list for an active topic or pipeline session.
 * Returns null if no state update is needed to avoid redundant re-renders.
 */
export function resolveTopicMessages(
  existing: Array<ChatMessage> | undefined,
  currentTopic: ResearchTopic,
  activePipeline?: ResearchPipeline | null,
): Array<ChatMessage> | null {
  if (activePipeline?.messages && activePipeline.messages.length > 0) {
    // If not initialized in local state yet
    if (!existing || existing.length === 0) {
      // If pipeline is running and only has the initial user message, synthesize streaming assistant placeholder
      if (
        activePipeline.status === 'RUNNING' &&
        activePipeline.messages.length === 1 &&
        activePipeline.messages[0].role === 'user'
      ) {
        const timestamp = new Date().toLocaleTimeString([], {
          hour: '2-digit',
          minute: '2-digit',
        })
        const initialAssistant: ChatMessage = {
          id: `ans-${Date.now()}`,
          role: 'assistant',
          content: '',
          timestamp,
          status: 'streaming',
          isLiveProgress: true,
          isThinkingActive: true,
          liveProgressMessage:
            activePipeline.approach === 'DEEP_RESEARCH'
              ? 'Starting Deep Research: Retrieving 100-200 papers & Cloudflare Clef classification...'
              : 'Deconstructing research concepts & searching journal corpus...',
          thinkingSteps: [
            {
              id: `step-init-thought-${Date.now()}`,
              type: 'thought',
              description: `Initializing research pipeline for: "${activePipeline.initialPrompt || activePipeline.title}".`,
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
        return [activePipeline.messages[0], initialAssistant]
      }

      return activePipeline.messages
    }

    // If local state is actively streaming and pipeline is still running, do not overwrite with stale snapshot
    const isCurrentlyStreaming = existing.some(
      (m) => m.status === 'streaming' || m.isLiveProgress,
    )
    if (isCurrentlyStreaming && activePipeline.status === 'RUNNING') {
      return null
    }

    // If pipeline completed or has newer messages from backend
    if (
      activePipeline.status === 'COMPLETED' ||
      activePipeline.messages.length > existing.length
    ) {
      return activePipeline.messages
    }

    return null
  }

  if (existing && existing.length > 0) {
    return null
  }

  return createInitialMessages(currentTopic, activePipeline)
}

/**
 * Pure reducer function that applies an incoming SSE PipelineEvent to the list of chat messages.
 */
export function reducePipelineStreamEvent(
  currentMessages: Array<ChatMessage>,
  event: PipelineEvent,
  activePipeline?: ResearchPipeline | null,
  fallbackSummary?: string,
): Array<ChatMessage> {
  const updatedList = [...currentMessages]

  // Defensively ensure an assistant turn exists to absorb progress events
  if (
    updatedList.length === 0 ||
    updatedList[updatedList.length - 1].role !== 'assistant'
  ) {
    const timestamp = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    })
    const newAst: ChatMessage = {
      id: `ans-${Date.now()}`,
      role: 'assistant',
      content: '',
      timestamp,
      status: 'streaming',
      isLiveProgress: true,
      isThinkingActive: true,
      steps: [],
      thinkingSteps: [],
    }
    updatedList.push(newAst)
  }

  const lastIdx = updatedList.length - 1
  const lastMsg = updatedList[lastIdx]

  const updatedSteps = [...(lastMsg.steps ?? [])]
  if (
    event.message &&
    !updatedSteps.includes(event.message) &&
    event.type !== 'awaiting_input'
  ) {
    updatedSteps.push(event.message)
  }

  const updatedThinkingSteps = [...(lastMsg.thinkingSteps ?? [])]
  if (event.step) {
    const sIdx = updatedThinkingSteps.findIndex((s) => s.id === event.step!.id)
    if (sIdx >= 0) {
      updatedThinkingSteps[sIdx] = event.step
    } else {
      updatedThinkingSteps.push(event.step)
    }
  } else if (
    event.message &&
    event.type !== 'awaiting_input' &&
    !updatedThinkingSteps.some((s) => s.description === event.message)
  ) {
    let stepType: ThinkingStepType = 'thought'
    let stepLabel: string | undefined = undefined
    if (event.stage === 'EXPLORATION') {
      stepType = 'deconstruct'
      stepLabel = 'Deconstruct'
    } else if (event.stage === 'RETRIEVAL') {
      stepType = 'search'
    } else if (event.stage === 'INGESTION') {
      stepType = 'read_fulltext'
      stepLabel = 'Read'
    } else if (event.stage === 'GRAPH') {
      stepType = 'references'
      stepLabel = 'References'
    } else if (event.stage === 'EXTRACTION') {
      stepType = 'audit'
      stepLabel = 'Grounding Audit'
    }

    updatedThinkingSteps.push({
      id: `ev-step-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
      type: stepType,
      label: stepLabel,
      description: event.message,
      status: event.type === 'stage_completed' ? 'completed' : 'in_progress',
    })
  }

  const isCompleted = event.type === 'pipeline_completed'
  if (isCompleted) {
    for (const s of updatedThinkingSteps) {
      s.status = 'completed'
    }
  }

  const updatedStats = event.stats ?? lastMsg.statsSummary

  let updatedCheckpoint = lastMsg.checkpoint
  if (event.type === 'awaiting_input') {
    if (event.checkpoint === 'ANGLE_SELECTION') {
      let angles: Array<ResearchAngle> = []
      if (Array.isArray(event.data)) {
        angles = event.data as Array<ResearchAngle>
      } else if (activePipeline?.stageData.decomposition?.angles) {
        angles = activePipeline.stageData.decomposition.angles
      }
      updatedCheckpoint = {
        type: 'ANGLE_SELECTION',
        angles,
      }
    } else if (event.checkpoint === 'ANCHOR_SELECTION') {
      let candidatePapers: Array<PaperMetadata> = []
      if (Array.isArray(event.data)) {
        candidatePapers = event.data as Array<PaperMetadata>
      } else if (event.data && typeof event.data === 'object') {
        const rawData = event.data as {
          candidates?: Array<{ paper: PaperMetadata } | PaperMetadata>
        }
        if (Array.isArray(rawData.candidates)) {
          candidatePapers = rawData.candidates.map((c) =>
            'paper' in c ? c.paper : c,
          )
        }
      }
      if (candidatePapers.length === 0) {
        candidatePapers = activePipeline?.stageData.candidatePapers ?? []
      }
      updatedCheckpoint = {
        type: 'ANCHOR_SELECTION',
        candidatePapers,
      }
    }
  }

  const updatedAssistant: ChatMessage = {
    ...lastMsg,
    steps: updatedSteps,
    thinkingSteps: updatedThinkingSteps,
    statsSummary: updatedStats,
    isThinkingActive: !isCompleted && event.type !== 'awaiting_input',
    isLiveProgress: !isCompleted,
    liveProgressMessage: isCompleted ? undefined : event.message,
    checkpoint: updatedCheckpoint,
    content: isCompleted ? fallbackSummary || event.message : lastMsg.content,
    status: isCompleted ? 'done' : 'streaming',
  }

  updatedList[lastIdx] = updatedAssistant
  return updatedList
}
