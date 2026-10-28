import { useMemo, useState } from 'react'

import styled from '@emotion/styled'
import { Brain, Check, Cpu, Search, Sparkles, Zap } from 'lucide-react'

import { useSelectedModel } from '#/-components/providers/ModelContext'
import { Badge } from '#/-components/ui/primitives/Badge'
import {
  DialogBackdrop,
  DialogClose,
  DialogDescription,
  DialogPopup,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from '#/-components/ui/primitives/Dialog'
import { theme } from '#/-components/ui/theme'
import type { LLMModel } from '#/types/llm.type'

interface ModelSelectionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const SearchContainer = styled.div`
  display: flex;
  align-items: center;
  gap: ${theme.spacing['2']};
  padding: 8px 12px;
  background: ${theme.colors.canvas.subtle};
  border: 1px solid ${theme.colors.border.subtleColor};
  border-radius: ${theme.radius.button};
  margin-bottom: ${theme.spacing['3']};
  transition: ${theme.transitions.fast};

  &:focus-within {
    border-color: ${theme.colors.accent.purple};
    box-shadow: 0 0 0 2px rgba(121, 110, 255, 0.15);
  }
`

const SearchInput = styled.input`
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  color: ${theme.colors.ink.primary};

  &::placeholder {
    color: ${theme.colors.ink.muted};
  }
`

const ProviderTabs = styled.div`
  display: flex;
  gap: ${theme.spacing['1']};
  overflow-x: auto;
  padding-bottom: ${theme.spacing['2']};
  margin-bottom: ${theme.spacing['3']};
  scrollbar-width: none;
  &::-webkit-scrollbar {
    display: none;
  }
`

const TabButton = styled.button<{ $active: boolean }>`
  border: none;
  background: ${({ $active }) =>
    $active ? theme.colors.accent.midnight : theme.colors.canvas.neutral};
  color: ${({ $active }) =>
    $active ? theme.colors.ink.inverse : theme.colors.ink.secondary};
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.xs};
  font-weight: ${theme.typography.weights.medium};
  padding: 5px 12px;
  border-radius: ${theme.radius.pill};
  cursor: pointer;
  white-space: nowrap;
  transition: ${theme.transitions.fast};

  &:hover {
    background: ${({ $active }) =>
      $active ? theme.colors.accent.midnight : theme.colors.neutral['200']};
  }
`

const ModelsList = styled.div`
  display: flex;
  flex-direction: column;
  gap: ${theme.spacing['2']};
  max-height: 380px;
  overflow-y: auto;
  padding-right: 4px;

  /* Custom subtle scrollbar */
  &::-webkit-scrollbar {
    width: 5px;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(100, 111, 121, 0.2);
    border-radius: 4px;
  }
`

const ModelCard = styled.div<{ $selected: boolean }>`
  display: flex;
  flex-direction: column;
  gap: ${theme.spacing['1']};
  padding: ${theme.spacing['3']};
  background: ${({ $selected }) =>
    $selected ? 'rgba(121, 110, 255, 0.05)' : theme.colors.canvas.primary};
  border: 1px solid
    ${({ $selected }) =>
      $selected ? theme.colors.accent.purple : theme.colors.border.subtleColor};
  border-radius: ${theme.radius.md};
  cursor: pointer;
  transition: ${theme.transitions.fast};

  &:hover {
    background: ${({ $selected }) =>
      $selected ? 'rgba(121, 110, 255, 0.08)' : theme.colors.canvas.subtle};
    border-color: ${({ $selected }) =>
      $selected ? theme.colors.accent.purple : theme.colors.border.medium};
  }
`

const ModelHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: ${theme.spacing['2']};
`

const ModelTitleGroup = styled.div`
  display: flex;
  align-items: center;
  gap: ${theme.spacing['2']};
  min-width: 0;
`

const ModelName = styled.span`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  font-weight: ${theme.typography.weights.semibold};
  color: ${theme.colors.ink.primary};
  letter-spacing: -0.01em;
`

const ModelProviderLabel = styled.span`
  font-size: ${theme.typography.sizes.micro};
  color: ${theme.colors.ink.muted};
  font-weight: ${theme.typography.weights.regular};
`

const ModelDescription = styled.p`
  margin: 0;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.xs};
  color: ${theme.colors.ink.secondary};
  line-height: ${theme.typography.lineHeights.normal};
`

const BadgesRow = styled.div`
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: ${theme.spacing['1']};
  margin-top: 4px;
`

const DialogFooter = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid ${theme.colors.border.subtleColor};
  padding-top: ${theme.spacing['3']};
  margin-top: ${theme.spacing['3']};
  font-size: ${theme.typography.sizes.xs};
  color: ${theme.colors.ink.muted};
`

const formatContextWindow = (ctx: number): string => {
  if (ctx >= 1000000) {
    return `${(ctx / 1000000).toFixed(ctx % 1000000 === 0 ? 0 : 1)}M ctx`
  }
  return `${Math.round(ctx / 1000)}k ctx`
}

const getProviderIcon = (provider: string) => {
  switch (provider.toLowerCase()) {
    case 'gemini':
      return <Sparkles size={14} color={theme.colors.accent.purple} />
    case 'anthropic':
      return <Cpu size={14} color={theme.colors.brand.coral} />
    case 'openai':
      return <Zap size={14} color={theme.colors.status.success} />
    case 'deepseek':
      return <Brain size={14} color={theme.colors.accent.blue} />
    default:
      return <Cpu size={14} color={theme.colors.ink.muted} />
  }
}

export const ModelSelectionDialog = ({
  open,
  onOpenChange,
}: ModelSelectionDialogProps) => {
  const { models, selectedModelId, setSelectedModelId } = useSelectedModel()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedProviderTab, setSelectedProviderTab] = useState<string>('all')

  const availableProviderTabs = useMemo(() => {
    const set = new Set<string>()
    models.forEach((m) => set.add(m.provider))
    return ['all', ...Array.from(set)]
  }, [models])

  const filteredModels = useMemo(() => {
    return models.filter((m) => {
      const matchesProvider =
        selectedProviderTab === 'all' || m.provider === selectedProviderTab

      if (!matchesProvider) return false

      if (!searchQuery.trim()) return true
      const q = searchQuery.toLowerCase()
      return (
        m.displayName.toLowerCase().includes(q) ||
        m.id.toLowerCase().includes(q) ||
        m.providerDisplayName.toLowerCase().includes(q) ||
        m.description.toLowerCase().includes(q)
      )
    })
  }, [models, selectedProviderTab, searchQuery])

  const handleSelect = (model: LLMModel) => {
    setSelectedModelId(model.id)
    onOpenChange(false)
  }

  return (
    <DialogRoot open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        <DialogBackdrop />
        <DialogPopup style={{ maxWidth: '580px' }}>
          <DialogClose aria-label="Close" />
          <DialogTitle>Select Language Model</DialogTitle>
          <DialogDescription>
            Choose the AI model for scientific concept decomposition and
            literature synthesis.
          </DialogDescription>

          {/* Search Bar */}
          <SearchContainer>
            <Search size={14} color={theme.colors.ink.muted} />
            <SearchInput
              type="text"
              placeholder="Search by model name, provider, or capability..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              autoFocus
            />
          </SearchContainer>

          {/* Provider Filter Tabs */}
          <ProviderTabs>
            {availableProviderTabs.map((prov) => (
              <TabButton
                key={prov}
                type="button"
                $active={selectedProviderTab === prov}
                onClick={() => setSelectedProviderTab(prov)}
              >
                {prov === 'all'
                  ? 'All Providers'
                  : prov === 'gemini'
                    ? 'Google Gemini'
                    : prov === 'anthropic'
                      ? 'Anthropic Claude'
                      : prov === 'openai'
                        ? 'OpenAI'
                        : prov === 'deepseek'
                          ? 'DeepSeek'
                          : prov === 'groq'
                            ? 'Groq'
                            : prov.toUpperCase()}
              </TabButton>
            ))}
          </ProviderTabs>

          {/* Models Scrollable List */}
          <ModelsList>
            {filteredModels.length === 0 ? (
              <div
                style={{
                  padding: '32px 16px',
                  textAlign: 'center',
                  color: theme.colors.ink.muted,
                  fontSize: theme.typography.sizes.sm,
                }}
              >
                No models found matching &quot;{searchQuery}&quot;.
              </div>
            ) : (
              filteredModels.map((model) => {
                const isSelected = model.id === selectedModelId
                return (
                  <ModelCard
                    key={model.id}
                    $selected={isSelected}
                    onClick={() => handleSelect(model)}
                  >
                    <ModelHeader>
                      <ModelTitleGroup>
                        {getProviderIcon(model.provider)}
                        <ModelName>{model.displayName}</ModelName>
                        <ModelProviderLabel>
                          • {model.providerDisplayName}
                        </ModelProviderLabel>
                      </ModelTitleGroup>

                      {isSelected && (
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px',
                            color: theme.colors.brand.coral,
                            fontWeight: 600,
                            fontSize: '0.78rem',
                          }}
                        >
                          <Check size={15} strokeWidth={2.5} />
                          <span>Active</span>
                        </div>
                      )}
                    </ModelHeader>

                    {model.description && (
                      <ModelDescription>{model.description}</ModelDescription>
                    )}

                    <BadgesRow>
                      <Badge variant="neutral" isMono size="sm">
                        {model.id}
                      </Badge>

                      {model.contextWindow > 0 && (
                        <Badge variant="neutral" size="sm">
                          {formatContextWindow(model.contextWindow)}
                        </Badge>
                      )}

                      {model.thinkingCapability === 'required' && (
                        <Badge variant="rose" size="sm">
                          <Brain size={11} style={{ marginRight: 3 }} />
                          Reasoning Required
                        </Badge>
                      )}

                      {model.thinkingCapability === 'optional' && (
                        <Badge variant="violet" size="sm">
                          <Brain size={11} style={{ marginRight: 3 }} />
                          Thinking Supported
                        </Badge>
                      )}

                      {model.isDefault && (
                        <Badge variant="success" size="sm">
                          Default
                        </Badge>
                      )}
                    </BadgesRow>
                  </ModelCard>
                )
              })
            )}
          </ModelsList>

          <DialogFooter>
            <span>
              Active:{' '}
              <strong>
                {models.find((m) => m.id === selectedModelId)?.displayName ??
                  selectedModelId}
              </strong>
            </span>
            <span style={{ fontSize: '0.72rem' }}>
              Persisted across browser sessions
            </span>
          </DialogFooter>
        </DialogPopup>
      </DialogPortal>
    </DialogRoot>
  )
}
