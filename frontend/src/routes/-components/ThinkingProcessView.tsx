import { useEffect, useState } from 'react'

import { css, keyframes } from '@emotion/react'
import {
  ArrowUpRight,
  BookOpen,
  Brain,
  ChevronDown,
  ChevronRight,
  Compass,
  Eye,
  History,
  Info,
  Loader2,
  Network,
  Quote,
  Search,
  ShieldCheck,
  Sparkles,
} from 'lucide-react'

import type { Theme } from '#/-components/ui/theme'
import type {
  ResearchStatsSummary,
  ThinkingStep,
  ThinkingStepType,
} from '#/routes/-types/research.type'

// ─── CUSTOM ANIMATIONS ───────────────────────────────────────────────────────

const blinkCaretKeyframe = keyframes`
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0;
  }
`

const stepEntranceKeyframe = keyframes`
  from {
    opacity: 0;
    transform: translateY(3px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`

const pulseDotKeyframe = keyframes`
  0% {
    transform: scale(0.9);
    opacity: 0.7;
  }
  50% {
    transform: scale(1.15);
    opacity: 1;
  }
  100% {
    transform: scale(0.9);
    opacity: 0.7;
  }
`

// ─── STYLES ──────────────────

const containerStyle = (theme: Theme) => css`
  display: flex;
  flex-direction: column;
  background: transparent;
  border: none;
  box-shadow: none;
  border-radius: 0;
  padding: 0;
  margin: ${theme.spacing['2']} 0 ${theme.spacing['3']};
  font-family: ${theme.typography.fontFamily.sans};
`

const headerToggleBtnStyle = css`
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 4px 0 8px;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
`

const headerLeftStyle = css`
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
`

const headerIconWrapStyle = (isActive: boolean) => (theme: Theme) => css`
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.accent.purple};
  flex-shrink: 0;
  ${
    isActive &&
    css`
      animation: ${pulseDotKeyframe} 1.6s infinite ease-in-out;
    `
  }
`

const headerTitleStyle = (theme: Theme) => css`
  font-size: 0.88rem;
  font-weight: 500;
  color: ${theme.colors.ink.primary};
  letter-spacing: -0.01em;
  display: flex;
  align-items: center;
  gap: 4px;
`

const headerStepsCountStyle = (theme: Theme) => css`
  font-size: 0.84rem;
  font-weight: 400;
  color: ${theme.colors.ink.muted};
`

const headerRightStyle = css`
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
`

const liveBadgeStyle = (theme: Theme) => css`
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: ${theme.radius.pill};
  background-color: rgba(121, 110, 255, 0.1);
  border: 1px solid rgba(121, 110, 255, 0.25);
  font-size: 0.72rem;
  font-weight: 600;
  color: ${theme.colors.accent.purple};
  letter-spacing: 0.02em;
`

const pulsingDotStyle = css`
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
  animation: ${pulseDotKeyframe} 1.4s infinite ease-in-out;
`

const chevronStyle = (isExpanded: boolean) => (theme: Theme) => css`
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.ink.muted};
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
  transform: ${isExpanded ? 'rotate(0deg)' : 'rotate(-90deg)'};
  margin-left: 2px;
`

const durationTextStyle = (theme: Theme) => css`
  font-size: 0.74rem;
  color: ${theme.colors.ink.muted};
  font-family: ${theme.typography.fontFamily.mono};
`

const contentAreaStyle = (theme: Theme) => css`
  display: flex;
  flex-direction: column;
  padding: 4px 0 0;
  gap: ${theme.spacing['3']};
  background: transparent;
`

// ─── STATS SUMMARY BAR (PERPLEXITY/CONSENSUS STYLE) ──────────────────────────

const statsGridStyle = (theme: Theme) => css`
  display: grid;
  grid-template-columns: 1fr auto 1fr auto 1fr;
  align-items: center;
  padding: 6px 0 14px;
  border-bottom: 1px solid ${theme.colors.border.subtleColor};
  margin-bottom: 4px;
`

const statColStyle = css`
  display: flex;
  flex-direction: column;
  gap: 2px;
`

const statValueWrapStyle = (theme: Theme) => css`
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 1.35rem;
  font-weight: 600;
  color: ${theme.colors.ink.primary};
  letter-spacing: -0.02em;
  font-family: ${theme.typography.fontFamily.sans};
  line-height: 1.2;
`

const infoIconStyle = (theme: Theme) => css`
  color: ${theme.colors.ink.muted};
  cursor: help;
  display: inline-flex;
  align-items: center;
  transition: color 0.15s ease;
  flex-shrink: 0;

  &:hover {
    color: ${theme.colors.accent.purple};
  }
`

const statLabelStyle = (theme: Theme) => css`
  font-size: 0.78rem;
  font-weight: 400;
  color: ${theme.colors.ink.muted};
  text-transform: capitalize;
`

const dividerColStyle = (theme: Theme) => css`
  width: 1px;
  height: 32px;
  background-color: ${theme.colors.border.subtleColor};
  margin: 0 16px;
`

// ─── STEPS & THOUGHTS LIST ───────────────────────────────────────────────────

const stepsListStyle = css`
  display: flex;
  flex-direction: column;
  gap: 6px;
`

const thoughtBlockStyle = (theme: Theme) => css`
  font-size: 0.88rem;
  font-weight: 350;
  line-height: 1.6;
  color: ${theme.colors.ink.secondary};
  padding: 3px 0;
  white-space: pre-wrap;
  word-break: break-word;
  animation: ${stepEntranceKeyframe} 0.25s cubic-bezier(0.16, 1, 0.3, 1)
    forwards;
`

const blinkingCaretStyle = (theme: Theme) => css`
  display: inline-block;
  margin-left: 2px;
  font-size: 0.9em;
  color: ${theme.colors.accent.purple};
  font-weight: 700;
  animation: ${blinkCaretKeyframe} 0.9s infinite step-start;
  vertical-align: baseline;
`

const stepRowStyle = (isInProgress: boolean) => (theme: Theme) => css`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 5px 6px;
  border-radius: ${theme.radius.md};
  background-color: ${
    isInProgress ? 'rgba(121, 110, 255, 0.05)' : 'transparent'
  };
  border: none;
  font-size: 0.86rem;
  color: ${
    isInProgress ? theme.colors.ink.primary : theme.colors.ink.secondary
  };
  animation: ${stepEntranceKeyframe} 0.25s cubic-bezier(0.16, 1, 0.3, 1)
    forwards;
  transition: background-color 0.15s ease;

  &:hover {
    background-color: rgba(13, 14, 16, 0.035);
  }
`

const stepRowLeftStyle = css`
  display: flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
  flex: 1;
`

const stepIconStyle =
  (type: ThinkingStepType, isInProgress: boolean) => (theme: Theme) => {
    let color: string = theme.colors.ink.muted
    if (isInProgress) {
      color = theme.colors.accent.purple
    } else if (type === 'search') {
      color = theme.colors.accent.blue
    } else if (type === 'read_abstract' || type === 'read_fulltext') {
      color = theme.colors.accent.purple
    } else if (type === 'references' || type === 'citations') {
      color = theme.colors.brand.coral
    } else if (type === 'audit') {
      color = theme.colors.status.success
    } else if (type === 'deconstruct') {
      color = theme.colors.accent.purple
    }

    return css`
      color: ${color};
      flex-shrink: 0;
      display: flex;
      align-items: center;
      justify-content: center;
    `
  }

const stepLabelStyle = (theme: Theme) => css`
  font-weight: 600;
  color: ${theme.colors.ink.primary};
  flex-shrink: 0;
`

const stepTextStyle = (theme: Theme) => css`
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
  color: ${theme.colors.ink.secondary};
`

const stepRowRightStyle = (theme: Theme) => css`
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 0.8rem;
  font-weight: 600;
  color: ${theme.colors.ink.primary};
  flex-shrink: 0;
`

const stepArrowStyle = (theme: Theme) => css`
  color: ${theme.colors.ink.muted};
  display: inline-flex;
  align-items: center;
`

const loaderIconStyle = (theme: Theme) => css`
  color: ${theme.colors.accent.purple};
`

// ─── HELPERS ─────────────────────────────────────────────────────────────────

const renderStepIcon = (type: ThinkingStepType, label?: string) => {
  if (label?.toLowerCase().includes('thread')) {
    return <History size={14} />
  }

  switch (type) {
    case 'search':
      return <Search size={14} />
    case 'read_abstract':
      return <Eye size={14} />
    case 'read_fulltext':
      return <BookOpen size={14} />
    case 'references':
      return <Network size={14} />
    case 'citations':
      return <Quote size={14} />
    case 'deconstruct':
      return <Compass size={14} />
    case 'audit':
      return <ShieldCheck size={14} />
    case 'thought':
    default:
      return <Sparkles size={14} />
  }
}

function formatStatNumber(val: number | string | undefined): string {
  if (val === undefined) return '0'
  if (typeof val === 'string') return val
  if (val >= 1_000_000) {
    const formatted = (val / 1_000_000).toFixed(1)
    return formatted.endsWith('.0') ? `${val / 1_000_000}M` : `${formatted}M`
  }
  if (val >= 1_000) {
    const formatted = (val / 1_000).toFixed(1)
    return formatted.endsWith('.0') ? `${val / 1_000}K` : `${formatted}K`
  }
  return val.toString()
}

// ─── COMPONENT PROPS ─────────────────────────────────────────────────────────

export interface ThinkingProcessViewProps {
  title?: string
  thinkingSteps?: Array<ThinkingStep>
  statsSummary?: ResearchStatsSummary
  isThinkingActive?: boolean
  thinkingDuration?: string
  liveProgressMessage?: string
  defaultExpanded?: boolean
  activeThinkingText?: string
}

export const ThinkingProcessView = ({
  title = 'Deep',
  thinkingSteps = [],
  statsSummary,
  isThinkingActive = false,
  thinkingDuration,
  liveProgressMessage,
  defaultExpanded = true,
  activeThinkingText,
}: ThinkingProcessViewProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded)

  // Keep expanded while thinking is in progress
  useEffect(() => {
    if (isThinkingActive) {
      setIsExpanded(true)
    }
  }, [isThinkingActive])

  // Count actionable steps
  const totalStepsCount = thinkingSteps.length

  if (thinkingSteps.length === 0 && !isThinkingActive && !statsSummary) {
    return null
  }

  const hasStats = Boolean(statsSummary)

  return (
    <div css={containerStyle}>
      {/* Header Bar */}
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        css={headerToggleBtnStyle}
        title={isExpanded ? 'Hide thinking process' : 'Show thinking process'}
      >
        <div css={headerLeftStyle}>
          <div css={headerIconWrapStyle(isThinkingActive)}>
            {isThinkingActive ? <Brain size={14} /> : <Sparkles size={14} />}
          </div>

          <div css={headerTitleStyle}>
            <span>{title}</span>
            <span css={headerStepsCountStyle}>
              ·{' '}
              {totalStepsCount > 0
                ? `${totalStepsCount} ${totalStepsCount === 1 ? 'step' : 'steps'}`
                : 'Analyzing'}
            </span>
          </div>

          <div css={chevronStyle(isExpanded)}>
            <ChevronDown size={14} />
          </div>
        </div>

        <div css={headerRightStyle}>
          {isThinkingActive && (
            <div css={liveBadgeStyle}>
              <span css={pulsingDotStyle} />
              <span>{liveProgressMessage || 'Thinking...'}</span>
            </div>
          )}

          {!isThinkingActive && thinkingDuration && (
            <span css={durationTextStyle}>{thinkingDuration}</span>
          )}
        </div>
      </button>

      {/* Expanded Details */}
      {isExpanded && (
        <div css={contentAreaStyle}>
          {/* Summary Stats Bar */}
          {hasStats && (
            <div css={statsGridStyle}>
              {/* Retrieved */}
              <div css={statColStyle}>
                <div css={statValueWrapStyle}>
                  <span>{formatStatNumber(statsSummary?.retrieved)}</span>
                  <span
                    css={infoIconStyle}
                    title="Total papers and preprints retrieved from Semantic Scholar, arXiv, OpenAlex, and Crossref."
                  >
                    <Info size={12} />
                  </span>
                </div>
                <div css={statLabelStyle}>Retrieved</div>
              </div>

              <div css={dividerColStyle} />

              {/* Eligible */}
              <div css={statColStyle}>
                <div css={statValueWrapStyle}>
                  <span>{formatStatNumber(statsSummary?.eligible)}</span>
                  <span
                    css={infoIconStyle}
                    title="Papers meeting peer-review screening, research angle relevance, and full-text availability criteria."
                  >
                    <Info size={12} />
                  </span>
                </div>
                <div css={statLabelStyle}>Eligible</div>
              </div>

              <div css={dividerColStyle} />

              {/* Included */}
              <div css={statColStyle}>
                <div css={statValueWrapStyle}>
                  <span>{formatStatNumber(statsSummary?.included)}</span>
                  <span
                    css={infoIconStyle}
                    title="Key anchor and foundational papers verified and explicitly cited in the grounded synthesis."
                  >
                    <Info size={12} />
                  </span>
                </div>
                <div css={statLabelStyle}>Included</div>
              </div>
            </div>
          )}

          {/* Sequential Thoughts & Action Milestones */}
          <div css={stepsListStyle}>
            {thinkingSteps.map((step) => {
              if (step.type === 'thought') {
                return (
                  <div key={step.id} css={thoughtBlockStyle}>
                    {step.description}
                    {step.status === 'in_progress' && (
                      <span css={blinkingCaretStyle}>▌</span>
                    )}
                  </div>
                )
              }

              const isInProgress = step.status === 'in_progress'
              const displayText = step.target || step.description
              const tooltipText =
                step.target && step.description
                  ? `${step.target}: ${step.description}`
                  : step.description

              const cleanCount =
                typeof step.count === 'string'
                  ? step.count.replace(/[>↗]/g, '').trim()
                  : step.count

              const isExternal =
                step.isExternalLink ||
                step.type === 'search' ||
                step.type === 'references' ||
                step.type === 'citations' ||
                (typeof step.count === 'string' && step.count.includes('↗'))

              const isChevron =
                step.type === 'read_abstract' ||
                step.type === 'read_fulltext' ||
                step.type === 'deconstruct' ||
                step.type === 'audit' ||
                (typeof step.count === 'string' && step.count.includes('>'))

              return (
                <div key={step.id} css={stepRowStyle(isInProgress)}>
                  <div css={stepRowLeftStyle}>
                    <div css={stepIconStyle(step.type, isInProgress)}>
                      {renderStepIcon(step.type, step.label)}
                    </div>

                    {step.label && (
                      <span css={stepLabelStyle}>{step.label}</span>
                    )}

                    <span css={stepTextStyle} title={tooltipText}>
                      {displayText}
                    </span>
                  </div>

                  <div css={stepRowRightStyle}>
                    {isInProgress ? (
                      <Loader2
                        size={13}
                        className="animate-spin"
                        css={loaderIconStyle}
                      />
                    ) : (
                      <>
                        {cleanCount !== undefined && cleanCount !== '' && (
                          <span>{cleanCount}</span>
                        )}
                        <span css={stepArrowStyle}>
                          {isExternal ? (
                            <ArrowUpRight size={13} />
                          ) : isChevron ? (
                            <ChevronRight size={13} />
                          ) : null}
                        </span>
                      </>
                    )}
                  </div>
                </div>
              )
            })}

            {/* Currently streaming token buffer in thought mode */}
            {activeThinkingText && (
              <div css={thoughtBlockStyle}>
                {activeThinkingText}
                <span css={blinkingCaretStyle}>▌</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
