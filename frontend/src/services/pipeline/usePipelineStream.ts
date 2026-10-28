import { useEffect, useRef, useState } from 'react'

import { useQueryClient } from '@tanstack/react-query'

import type { PipelineEvent } from '#/types/pipeline.type'

interface UsePipelineStreamOptions {
  enabled?: boolean
  onEvent?: (event: PipelineEvent) => void
  onCompleted?: () => void
  onAwaitingInput?: (event: PipelineEvent) => void
}

export const usePipelineStream = (
  pipelineId: string | undefined | null,
  options: UsePipelineStreamOptions = {},
) => {
  const { enabled = true, onEvent, onCompleted, onAwaitingInput } = options
  const queryClient = useQueryClient()
  const [events, setEvents] = useState<Array<PipelineEvent>>([])
  const [latestEvent, setLatestEvent] = useState<PipelineEvent | null>(null)
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    if (!enabled || !pipelineId) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      setIsStreaming(false)
      return
    }

    setIsStreaming(true)
    setError(null)

    const url = `/api/v1/pipelines/${encodeURIComponent(pipelineId)}/stream`
    const es = new EventSource(url)
    eventSourceRef.current = es

    const handleEvent = (e: MessageEvent) => {
      try {
        const parsed = JSON.parse(e.data as string) as PipelineEvent
        setLatestEvent(parsed)
        setEvents((prev) => [...prev, parsed])
        onEvent?.(parsed)

        // Sync query cache
        void queryClient.invalidateQueries({
          queryKey: ['pipeline', pipelineId],
        })

        if (parsed.type === 'awaiting_input') {
          onAwaitingInput?.(parsed)
        } else if (parsed.type === 'pipeline_completed') {
          onCompleted?.()
          es.close()
          setIsStreaming(false)
        } else if (parsed.type === 'pipeline_failed') {
          es.close()
          setIsStreaming(false)
        }
      } catch {
        // Non-JSON ping or heartbeat
      }
    }

    es.addEventListener('pipeline_progress', handleEvent)
    es.addEventListener('stage_completed', handleEvent)
    es.addEventListener('awaiting_input', handleEvent)
    es.addEventListener('pipeline_completed', handleEvent)
    es.addEventListener('pipeline_failed', handleEvent)

    es.onerror = () => {
      if (es.readyState === EventSource.CLOSED) {
        setError('Event stream disconnected')
        setIsStreaming(false)
      }
    }

    return () => {
      es.close()
      eventSourceRef.current = null
      setIsStreaming(false)
    }
  }, [pipelineId, enabled, queryClient])

  return {
    events,
    latestEvent,
    isStreaming,
    error,
  }
}
