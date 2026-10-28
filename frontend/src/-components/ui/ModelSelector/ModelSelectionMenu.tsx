import type { ReactNode } from 'react'
import * as React from 'react'
import { useEffect, useState } from 'react'

import { keyframes } from '@emotion/react'
import styled from '@emotion/styled'
import {
  Brain,
  Check,
  ChevronLeft,
  ChevronRight,
  Cpu,
  KeyRound,
  Layers,
  Sparkles,
  Zap,
} from 'lucide-react'

import { useSelectedModel } from '#/-components/providers/ModelContext'
import { Menu } from '#/-components/ui/primitives/Menu'
import { theme } from '#/-components/ui/theme'
import type { LLMModel, LLMProvider, ThinkingLevel } from '#/types/llm.type'

interface ModelSelectionMenuProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  side?: 'top' | 'bottom'
  children: ReactNode
}

/* -------------------------------------------------------------------------- */
/*                                 Animations                                 */
/* -------------------------------------------------------------------------- */

const slideInForward = keyframes`
  0% {
    opacity: 0;
    transform: translateX(8px);
  }
  100% {
    opacity: 1;
    transform: translateX(0);
  }
`

const slideInBackward = keyframes`
  0% {
    opacity: 0;
    transform: translateX(-8px);
  }
  100% {
    opacity: 1;
    transform: translateX(0);
  }
`

const ViewWrapper = styled.div<{ $direction: 'forward' | 'backward' }>`
  animation: ${({ $direction }) =>
      $direction === 'forward' ? slideInForward : slideInBackward}
    0.18s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
  gap: 2px;
`

const ProviderIconBox = styled.div<{ $bg?: string }>`
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: ${theme.radius.sm};
  background-color: ${({ $bg }) => $bg ?? theme.colors.canvas.subtle};
  border: 1px solid rgba(100, 111, 121, 0.14);
  flex-shrink: 0;
  transition: transform 0.15s cubic-bezier(0.16, 1, 0.3, 1);
`

const THINKING_LEVEL_OPTIONS: Array<{
  id: ThinkingLevel
  title: string
  description: string
}> = [
  {
    id: 'low',
    title: 'Low',
    description: 'Quick and efficient',
  },
  {
    id: 'medium',
    title: 'Medium',
    description: 'Balanced depth',
  },
  {
    id: 'high',
    title: 'High',
    description: 'Extra thorough',
  },
]

const getModelTagline = (model: LLMModel): string => {
  const customTaglines: Record<string, string> = {
    'gemini-2.0-flash': 'Fastest answers',
    'gemini-2.0-pro-exp-02-05': 'Advanced reasoning',
    'gemini-1.5-pro': 'All-around help & 2M context',
    'claude-3-7-sonnet-20250219': 'Hybrid reasoning & thinking',
    'claude-3-5-sonnet-20241022': 'All-around help',
    'claude-3-5-haiku-20241022': 'Fastest answers',
    'gpt-4o-mini': 'Fastest answers',
    'gpt-4o': 'All-around help',
    'o3-mini': 'STEM & math reasoning',
    o1: 'Complex tasks',
    'deepseek-chat': 'General flagship conversation',
    'deepseek-reasoner': 'Reinforced multi-step proofs',
    'llama-3.3-70b-versatile': 'Fast open-source answers',
  }

  if (customTaglines[model.id]) {
    return customTaglines[model.id]
  }

  if (model.description) {
    return model.description.length > 34
      ? `${model.description.slice(0, 31)}...`
      : model.description
  }

  return 'General intelligence'
}

const getProviderIcon = (providerId: string) => {
  switch (providerId.toLowerCase()) {
    case 'gemini':
      return (
        <ProviderIconBox $bg="rgba(121, 110, 255, 0.08)">
          <Sparkles size={13} color={theme.colors.accent.purple} />
        </ProviderIconBox>
      )
    case 'anthropic':
      return (
        <ProviderIconBox $bg="rgba(240, 106, 106, 0.08)">
          <Cpu size={13} color={theme.colors.brand.coral} />
        </ProviderIconBox>
      )
    case 'openai':
      return (
        <ProviderIconBox $bg="rgba(52, 199, 89, 0.08)">
          <Zap size={13} color={theme.colors.status.success} />
        </ProviderIconBox>
      )
    case 'deepseek':
      return (
        <ProviderIconBox $bg="rgba(0, 178, 255, 0.08)">
          <Brain size={13} color={theme.colors.accent.blue} />
        </ProviderIconBox>
      )
    case 'groq':
      return (
        <ProviderIconBox $bg="rgba(252, 189, 1, 0.08)">
          <Layers size={13} color={theme.colors.accent.gold} />
        </ProviderIconBox>
      )
    default:
      return (
        <ProviderIconBox>
          <Cpu size={13} color={theme.colors.ink.muted} />
        </ProviderIconBox>
      )
  }
}

const getProviderSummary = (
  provider: LLMProvider,
  models: Array<LLMModel>,
): string => {
  const count = models.filter((m) => m.provider === provider.id).length
  switch (provider.id.toLowerCase()) {
    case 'gemini':
      return `${count} models • Flash & Pro`
    case 'anthropic':
      return `${count} models • Claude 3.7 & 3.5`
    case 'openai':
      return `${count} models • GPT-4o, o3-mini, o1`
    case 'deepseek':
      return `${count} models • V3 & R1 Reasoner`
    case 'groq':
      return `${count} models • LPU Inference`
    default:
      return `${count} models available`
  }
}

export const ModelSelectionMenu = ({
  open,
  onOpenChange,
  side = 'top',
  children,
}: ModelSelectionMenuProps) => {
  const {
    models,
    providers,
    selectedModelId,
    selectedModel,
    setSelectedModelId,
    thinkingLevel,
    setThinkingLevel,
    isLoading,
  } = useSelectedModel()

  const [view, setView] = useState<'providers' | 'models'>('providers')
  const [activeProviderId, setActiveProviderId] = useState<string>('gemini')
  const [navDirection, setNavDirection] = useState<'forward' | 'backward'>(
    'forward',
  )

  // Reset view to 'providers' each time menu opens, and set activeProvider to current selection if exists
  useEffect(() => {
    if (open) {
      setView('providers')
      setNavDirection('forward')
      if (selectedModel?.provider) {
        setActiveProviderId(selectedModel.provider)
      } else if (
        providers.length > 0 &&
        !providers.some((p) => p.id === activeProviderId)
      ) {
        setActiveProviderId(providers[0].id)
      }
    }
  }, [open, selectedModel?.provider, providers, activeProviderId])

  const handleProviderClick = (providerId: string) => {
    setActiveProviderId(providerId)
    setNavDirection('forward')
    setView('models')
  }

  const handleBackToProviders = () => {
    setNavDirection('backward')
    setView('providers')
  }

  const handleModelSelect = (model: LLMModel) => {
    setSelectedModelId(model.id)
  }

  const handleThinkingLevelSelect = (level: ThinkingLevel) => {
    setThinkingLevel(level)
    onOpenChange(false)
  }

  const activeProvider =
    providers.find((p) => p.id === activeProviderId) ?? providers[0]
  const providerModels = models.filter((m) => m.provider === activeProvider.id)

  return (
    <Menu.Root
      open={open}
      onOpenChange={onOpenChange}
      variant="light"
      modal={false}
    >
      <Menu.Trigger
        render={
          React.isValidElement(children) ? (
            children
          ) : (
            <button type="button">{children}</button>
          )
        }
      />

      <Menu.Portal>
        <Menu.Positioner
          side={side}
          align="start"
          sideOffset={8}
          style={{ zIndex: 1000 }}
        >
          <Menu.Popup minWidth={260} maxWidth={290}>
            {view === 'providers' ? (
              <ViewWrapper key="providers-view" $direction={navDirection}>
                <Menu.Header>Select Provider</Menu.Header>
                {isLoading ? (
                  <div
                    style={{
                      padding: '24px 16px',
                      textAlign: 'center',
                      color: theme.colors.ink.secondary,
                      fontSize: '0.8rem',
                    }}
                  >
                    Loading models...
                  </div>
                ) : providers.length === 0 ? (
                  <div
                    style={{
                      padding: '20px 16px',
                      textAlign: 'center',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: '10px',
                    }}
                  >
                    <div
                      style={{
                        width: '36px',
                        height: '36px',
                        borderRadius: '50%',
                        backgroundColor: 'rgba(121, 110, 255, 0.08)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: theme.colors.accent.purple,
                      }}
                    >
                      <KeyRound size={18} />
                    </div>
                    <div
                      style={{
                        fontWeight: 600,
                        fontSize: '0.85rem',
                        color: theme.colors.ink.primary,
                      }}
                    >
                      No API Key Configured
                    </div>
                    <div
                      style={{
                        fontSize: '0.74rem',
                        color: theme.colors.ink.secondary,
                        lineHeight: 1.45,
                      }}
                    >
                      Set <code>GEMINI_API_KEY</code>,{' '}
                      <code>OPENAI_API_KEY</code>, or{' '}
                      <code>ANTHROPIC_API_KEY</code> in your environment or{' '}
                      <code>backend/.env</code>.
                    </div>
                  </div>
                ) : (
                  <Menu.Group>
                    {providers.map((provider) => {
                      const isCurrent = selectedModel?.provider === provider.id
                      return (
                        <Menu.Item
                          key={provider.id}
                          startIcon={getProviderIcon(provider.id)}
                          title={provider.displayName}
                          subtitle={getProviderSummary(provider, models)}
                          endIcon={
                            <div
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px',
                              }}
                            >
                              {isCurrent && (
                                <Check
                                  size={13}
                                  strokeWidth={2.5}
                                  color={theme.colors.accent.purple}
                                />
                              )}
                              <ChevronRight
                                size={14}
                                color={theme.colors.ink.muted}
                              />
                            </div>
                          }
                          onClick={() => handleProviderClick(provider.id)}
                          closeOnClick={false}
                        />
                      )
                    })}
                  </Menu.Group>
                )}
              </ViewWrapper>
            ) : (
              <ViewWrapper key="models-view" $direction={navDirection}>
                <Menu.Item
                  startIcon={<ChevronLeft size={14} />}
                  title={activeProvider.displayName}
                  subtitle="Back to all providers"
                  onClick={handleBackToProviders}
                  closeOnClick={false}
                />
                <Menu.Separator />
                <Menu.Group>
                  {providerModels.length === 0 ? (
                    <div
                      style={{
                        padding: '16px',
                        textAlign: 'center',
                        fontSize: '0.75rem',
                        color: theme.colors.ink.secondary,
                      }}
                    >
                      No models configured for this provider.
                    </div>
                  ) : (
                    providerModels.map((model) => {
                      const isSelected = model.id === selectedModelId
                      return (
                        <Menu.Item
                          key={model.id}
                          selected={isSelected}
                          title={model.displayName}
                          subtitle={getModelTagline(model)}
                          onClick={() => handleModelSelect(model)}
                          closeOnClick={false}
                        />
                      )
                    })
                  )}
                </Menu.Group>
                <Menu.Separator />
                <Menu.Submenu
                  label="Thinking level"
                  startIcon={
                    <Brain size={14} color={theme.colors.accent.purple} />
                  }
                  subtitle={`${thinkingLevel.charAt(0).toUpperCase() + thinkingLevel.slice(1)} intensity`}
                  minWidth={205}
                >
                  <Menu.Header>Reasoning Effort</Menu.Header>
                  {THINKING_LEVEL_OPTIONS.map((opt) => {
                    const isLevelSelected = thinkingLevel === opt.id
                    return (
                      <Menu.Item
                        key={opt.id}
                        selected={isLevelSelected}
                        title={opt.title}
                        subtitle={opt.description}
                        onClick={() => handleThinkingLevelSelect(opt.id)}
                      />
                    )
                  })}
                </Menu.Submenu>
              </ViewWrapper>
            )}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}
