import type { CSSProperties } from 'react'
import { useState } from 'react'

import styled from '@emotion/styled'
import { ChevronDown, KeyRound, Sparkles } from 'lucide-react'

import { useSelectedModel } from '#/-components/providers/ModelContext'
import { theme } from '#/-components/ui/theme'

import { ModelSelectionMenu } from './ModelSelectionMenu'

export type ModelSelectorVariant = 'pill' | 'compact' | 'standard'

export interface ModelSelectorProps {
  variant?: ModelSelectorVariant
  className?: string
  style?: CSSProperties
  side?: 'top' | 'bottom'
}

const ChevronIndicator = styled(ChevronDown)<{ $open: boolean }>`
  transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  transform: ${({ $open }) => ($open ? 'rotate(180deg)' : 'rotate(0deg)')};
  opacity: 0.65;
  flex-shrink: 0;
`

const PillButton = styled.button<{ $open: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background-color: ${({ $open }) =>
    $open ? '#ffffff' : 'var(--color-bg-neutral, #f3f3f3)'};
  border: 1px solid
    ${({ $open }) =>
      $open ? theme.colors.accent.purple : 'rgba(100, 111, 121, 0.16)'};
  box-shadow: ${({ $open }) =>
    $open ? '0 0 0 2px rgba(121, 110, 255, 0.15)' : 'none'};
  border-radius: var(--radius-pill, 9999px);
  padding: 4px 10px;
  font-family: var(--font-sans);
  font-size: 0.74rem;
  color: ${({ $open }) =>
    $open ? theme.colors.ink.primary : 'var(--color-ink-secondary, #4a5464)'};
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    background-color: ${({ $open }) =>
      $open ? '#ffffff' : theme.colors.neutral['200']};
    color: ${theme.colors.ink.primary};
    border-color: ${({ $open }) =>
      $open ? theme.colors.accent.purple : 'rgba(100, 111, 121, 0.3)'};
    transform: translateY(-1px);
  }
`

const CompactButton = styled.button<{ $open: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: ${({ $open }) =>
    $open ? '#ffffff' : theme.colors.canvas.subtle};
  border: 1px solid
    ${({ $open }) =>
      $open ? theme.colors.accent.purple : theme.colors.border.subtleColor};
  box-shadow: ${({ $open }) =>
    $open ? '0 0 0 2px rgba(121, 110, 255, 0.15)' : 'none'};
  border-radius: ${theme.radius.button};
  padding: 5px 10px;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 0.8rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${({ $open }) =>
    $open ? theme.colors.ink.primary : theme.colors.ink.secondary};
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    background: #ffffff;
    color: ${theme.colors.ink.primary};
    border-color: ${({ $open }) =>
      $open ? theme.colors.accent.purple : theme.colors.border.medium};
    transform: translateY(-1px);
  }
`

const StandardButton = styled.button<{ $open: boolean }>`
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  background: ${({ $open }) =>
    $open ? '#ffffff' : theme.colors.canvas.primary};
  border: 1px solid
    ${({ $open }) =>
      $open ? theme.colors.accent.purple : theme.colors.border.subtleColor};
  box-shadow: ${({ $open }) =>
    $open ? '0 0 0 2px rgba(121, 110, 255, 0.15)' : 'none'};
  border-radius: ${theme.radius.button};
  padding: 8px 12px;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    background: #ffffff;
    border-color: ${({ $open }) =>
      $open ? theme.colors.accent.purple : theme.colors.border.medium};
    transform: translateY(-1px);
  }
`

export const ModelSelector = ({
  variant = 'compact',
  className,
  style,
  side,
}: ModelSelectorProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const { selectedModel, isLoading, models } = useSelectedModel()

  const displayName =
    selectedModel?.displayName ||
    (isLoading
      ? 'Loading...'
      : models.length === 0
        ? 'No API Key Set'
        : 'Select Model')

  const resolvedSide = side ?? (variant === 'pill' ? 'top' : 'bottom')
  const hasConfiguredModels = models.length > 0

  const renderIcon = (size: number) => {
    if (!hasConfiguredModels && !isLoading) {
      return <KeyRound size={size} color={theme.colors.ink.muted} />
    }
    return <Sparkles size={size} color={theme.colors.accent.purple} />
  }

  const renderTrigger = () => {
    switch (variant) {
      case 'pill':
        return (
          <PillButton
            type="button"
            className={className}
            style={style}
            $open={isOpen}
            title={
              hasConfiguredModels
                ? 'Ganti Model AI'
                : 'No API Key Configured (Click to see instructions)'
            }
          >
            {renderIcon(12)}
            <span>{displayName}</span>
            <ChevronIndicator size={12} $open={isOpen} />
          </PillButton>
        )

      case 'standard':
        return (
          <StandardButton
            type="button"
            className={className}
            style={style}
            $open={isOpen}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              {renderIcon(14)}
              <span>{displayName}</span>
            </div>
            <ChevronIndicator size={14} $open={isOpen} />
          </StandardButton>
        )

      case 'compact':
      default:
        return (
          <CompactButton
            type="button"
            className={className}
            style={style}
            $open={isOpen}
            title={
              hasConfiguredModels
                ? 'Select Language Model'
                : 'No API Key Configured'
            }
          >
            {renderIcon(13)}
            <span>{displayName}</span>
            <ChevronIndicator size={12} $open={isOpen} />
          </CompactButton>
        )
    }
  }

  return (
    <ModelSelectionMenu
      open={isOpen}
      onOpenChange={setIsOpen}
      side={resolvedSide}
    >
      {renderTrigger()}
    </ModelSelectionMenu>
  )
}
