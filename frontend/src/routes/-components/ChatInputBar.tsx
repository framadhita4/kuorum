import { useEffect, useState } from 'react'

import styled from '@emotion/styled'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  ArrowUp,
  Cpu,
  Filter,
  Layers,
  Loader2,
  MessageSquare,
  Sparkles,
  StopCircle,
} from 'lucide-react'
import { useForm } from 'react-hook-form'

import { ModelSelector, Pill } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ChatValues } from '#/routes/-schema/chat.schema'
import { chatSchema } from '#/routes/-schema/chat.schema'
import type { ResearchApproach } from '#/types/pipeline.type'

export type ChatSubmissionMode = 'follow_up' | 'new_pipeline'

interface ChatInputBarProps {
  isRightPanelOpen: boolean
  isResearching: boolean
  isNewThread?: boolean
  onSubmit: (
    prompt: string,
    subMode: ChatSubmissionMode,
    approach: ResearchApproach,
  ) => void
  onCancel?: () => void
}

const ChatInputBarContainer = styled.div<{ $isRightPanelOpen: boolean }>`
  position: absolute;
  bottom: 24px;
  left: ${({ $isRightPanelOpen }) => ($isRightPanelOpen ? '20px' : '50%')};
  transform: ${({ $isRightPanelOpen }) =>
    $isRightPanelOpen ? 'none' : 'translateX(-50%)'};
  width: ${({ $isRightPanelOpen }) =>
    $isRightPanelOpen ? 'calc(100% - 40px)' : '90%'};
  max-width: ${({ $isRightPanelOpen }) =>
    $isRightPanelOpen ? '100%' : '760px'};
  background-color: ${theme.colors.canvas.primary};
  border-radius: ${theme.radius.card};
  border: ${theme.colors.border.subtle};
  box-shadow: ${theme.shadows.lg};
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 30;
  transition: ${theme.transitions.fast};
`

const InputForm = styled.form`
  display: flex;
  align-items: center;
  gap: 10px;
`

const TextInput = styled.input`
  flex: 1;
  border: none;
  outline: none;
  font-size: 0.94rem;
  color: ${theme.colors.ink.primary};
  background-color: transparent;
  padding: 6px 4px;
  font-family: ${theme.typography.fontFamily.sans};
  line-height: ${theme.typography.lineHeights.normal};

  &::placeholder {
    color: ${theme.colors.ink.muted};
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`

const SubmitButton = styled.button`
  width: 36px;
  height: 36px;
  border-radius: ${theme.radius.button};
  background-color: ${theme.colors.ink.primary};
  color: ${theme.colors.ink.inverse};
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  box-shadow: ${theme.shadows.sm};
  transition: ${theme.transitions.fast};

  &:hover:not(:disabled) {
    background-color: ${theme.colors.brand.coral};
    color: ${theme.colors.ink.primary};
  }

  &:disabled {
    background-color: ${theme.colors.canvas.neutral};
    color: ${theme.colors.ink.muted};
    cursor: not-allowed;
    box-shadow: none;
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }
`

const CancelButton = styled.button`
  width: 36px;
  height: 36px;
  border-radius: ${theme.radius.button};
  background-color: rgba(239, 68, 68, 0.12);
  color: ${theme.colors.status.error};
  border: 1px solid rgba(239, 68, 68, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: ${theme.transitions.fast};

  &:hover {
    background-color: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.5);
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.status.error};
    outline-offset: 2px;
  }
`

const ErrorMessage = styled.span`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 0.72rem;
  color: ${theme.colors.status.error};
  padding-left: 4px;
`

const PillsContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
`

export const ChatInputBar = ({
  isRightPanelOpen,
  isResearching,
  isNewThread = false,
  onSubmit,
  onCancel,
}: ChatInputBarProps) => {
  const [submissionMode, setSubmissionMode] = useState<ChatSubmissionMode>(
    isNewThread ? 'new_pipeline' : 'follow_up',
  )
  const [approach, setApproach] = useState<ResearchApproach>('NORMAL')

  useEffect(() => {
    if (isNewThread) {
      setSubmissionMode('new_pipeline')
    } else {
      setSubmissionMode('follow_up')
    }
  }, [isNewThread])

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<ChatValues>({
    resolver: zodResolver(chatSchema),
    defaultValues: { prompt: '' },
  })

  const promptValue = watch('prompt')

  const onFormSubmit = (values: ChatValues) => {
    if (!values.prompt.trim()) return
    const activeSubMode = isNewThread ? 'new_pipeline' : submissionMode
    onSubmit(values.prompt.trim(), activeSubMode, approach)
    reset()
  }

  return (
    <ChatInputBarContainer $isRightPanelOpen={isRightPanelOpen}>
      <InputForm onSubmit={handleSubmit(onFormSubmit)}>
        <TextInput
          {...register('prompt')}
          type="text"
          disabled={isResearching}
          placeholder={
            isNewThread || submissionMode === 'new_pipeline'
              ? 'Type a research hypothesis or topic to start autonomous pipeline...'
              : 'Ask a follow-up question or compare methodologies across papers...'
          }
        />

        {/* Submit or Cancel Button */}
        {isResearching && onCancel ? (
          <CancelButton
            type="button"
            onClick={onCancel}
            title="Cancel research process"
          >
            <StopCircle size={18} />
          </CancelButton>
        ) : (
          <SubmitButton
            type="submit"
            disabled={!promptValue || isResearching}
            title="Kirim pesan"
          >
            {isResearching ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <ArrowUp size={16} strokeWidth={2.4} />
            )}
          </SubmitButton>
        )}
      </InputForm>

      {/* Error feedback if any */}
      {errors.prompt && <ErrorMessage>{errors.prompt.message}</ErrorMessage>}

      {/* Input Bar Pills*/}
      <PillsContainer>
        <ModelSelector variant="pill" />

        {/* Submission Mode: On new thread, always starting new research */}
        {isNewThread ? (
          <Pill
            variant="purple"
            icon={<Sparkles size={12} />}
            label="Start New Research"
            title="New Research Mode: Start an autonomous research pipeline from concept deconstruction to synthesis"
          />
        ) : (
          <Pill
            variant={submissionMode === 'new_pipeline' ? 'purple' : 'neutral'}
            onClick={() =>
              setSubmissionMode(
                submissionMode === 'follow_up' ? 'new_pipeline' : 'follow_up',
              )
            }
            icon={
              submissionMode === 'follow_up' ? (
                <MessageSquare size={12} />
              ) : (
                <Sparkles size={12} />
              )
            }
            label={
              submissionMode === 'follow_up'
                ? 'Topic Follow-up'
                : 'Start New Research'
            }
            title={
              submissionMode === 'follow_up'
                ? 'Follow-up Mode: Deep dive into the synthesis results of the active topic'
                : 'New Research Mode: Start a new research saga from concept deconstruction to synthesis'
            }
          />
        )}

        {/* Research Approach Selector: Normal Mode vs Deep Research */}
        {(isNewThread || submissionMode === 'new_pipeline') && (
          <Pill
            variant={approach === 'DEEP_RESEARCH' ? 'purple' : 'teal'}
            onClick={() =>
              setApproach(approach === 'NORMAL' ? 'DEEP_RESEARCH' : 'NORMAL')
            }
            icon={
              approach === 'DEEP_RESEARCH' ? (
                <Cpu size={12} />
              ) : (
                <Layers size={12} />
              )
            }
            label={
              approach === 'NORMAL'
                ? 'Normal Mode (20 Papers)'
                : 'Deep Research (100-200 Papers • Clef)'
            }
            title={
              approach === 'NORMAL'
                ? 'Normal Mode: Retrieves 20 papers, calculates relevance scores, and triggers an AI deep dive on selected papers'
                : 'Deep Research: Retrieves 100-200 papers, rapidly classifies with Cloudflare Clef, and conducts an AI deep dive'
            }
          />
        )}

        <Pill
          variant="neutral"
          icon={<Filter size={12} />}
          label="Q1 Journals Only"
        />
      </PillsContainer>
    </ChatInputBarContainer>
  )
}
