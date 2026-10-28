import type { HTMLAttributes } from 'react'

import { css } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export interface StepItem {
  id: string | number
  label: string
  description?: string
}

export interface StepperProps extends HTMLAttributes<HTMLDivElement> {
  steps: Array<StepItem>
  currentStep: number // 0-indexed
  onStepClick?: (stepIndex: number) => void
  variant?: 'compact' | 'linear'
}

const CompactContainer = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: ${theme.colors.canvas.neutral};
  border: ${theme.colors.border.subtle};
  border-radius: ${theme.radius.sm};
  padding: 4px 10px;
`

const StepCounter = styled.span`
  font-family: ${theme.typography.fontFamily.mono};
  font-size: ${theme.typography.sizes.micro};
  font-weight: ${theme.typography.weights.semibold};
  color: ${theme.colors.ink.primary};
`

const StepLabel = styled.span`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.caption};
  color: ${theme.colors.ink.secondary};
`

const LinearContainer = styled.div`
  display: flex;
  align-items: center;
  width: 100%;
  gap: 8px;
`

const LinearStep = styled.button<{
  $isActive: boolean
  $isCompleted: boolean
  $isClickable: boolean
}>`
  display: flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: none;
  padding: 6px 10px;
  border-radius: ${theme.radius.button};
  cursor: ${({ $isClickable }) => ($isClickable ? 'pointer' : 'default')};
  outline: none;
  transition: ${theme.transitions.fast};

  &:hover {
    ${({ $isClickable }) =>
      $isClickable && `background: ${theme.colors.canvas.neutral};`}
  }
`

const StepNumberCircle = styled.span<{
  $isActive: boolean
  $isCompleted: boolean
}>`
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: ${theme.typography.fontFamily.mono};
  font-size: 11px;
  font-weight: ${theme.typography.weights.semibold};
  transition: ${theme.transitions.fast};

  ${({ $isActive, $isCompleted }) => {
    if ($isActive) {
      return css`
        background: ${theme.colors.ink.primary};
        color: ${theme.colors.ink.inverse};
      `
    }
    if ($isCompleted) {
      return css`
        background: ${theme.colors.accent.purple};
        color: #ffffff;
      `
    }
    return css`
      background: transparent;
      color: ${theme.colors.ink.muted};
      border: 1px solid rgba(100, 111, 121, 0.2);
    `
  }}
`

const StepText = styled.span<{ $isActive: boolean }>`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.caption};
  font-weight: ${({ $isActive }) =>
    $isActive
      ? theme.typography.weights.medium
      : theme.typography.weights.regular};
  color: ${({ $isActive }) =>
    $isActive ? theme.colors.ink.primary : theme.colors.ink.muted};
`

const StepDivider = styled.div<{ $isCompleted: boolean }>`
  flex: 1;
  height: 1px;
  background: ${({ $isCompleted }) =>
    $isCompleted ? theme.colors.accent.purple : 'rgba(100, 111, 121, 0.2)'};
`

export const Stepper = ({
  steps,
  currentStep,
  onStepClick,
  variant = 'compact',
  ...props
}: StepperProps) => {
  const current = steps[currentStep] || steps[0]

  if (variant === 'compact') {
    return (
      <CompactContainer {...props}>
        <StepCounter>
          0{currentStep + 1} / 0{steps.length}
        </StepCounter>
        {current.id && <StepLabel>{current.label}</StepLabel>}
      </CompactContainer>
    )
  }

  return (
    <LinearContainer {...props}>
      {steps.map((step, idx) => {
        const isActive = idx === currentStep
        const isCompleted = idx < currentStep
        const isClickable = !!onStepClick

        return (
          <div
            key={step.id}
            style={{ display: 'flex', alignItems: 'center', flex: 1 }}
          >
            <LinearStep
              type="button"
              $isActive={isActive}
              $isCompleted={isCompleted}
              $isClickable={isClickable}
              onClick={() => isClickable && onStepClick(idx)}
            >
              <StepNumberCircle $isActive={isActive} $isCompleted={isCompleted}>
                {isCompleted ? '✓' : `0${idx + 1}`}
              </StepNumberCircle>
              <StepText $isActive={isActive}>{step.label}</StepText>
            </LinearStep>
            {idx < steps.length - 1 && (
              <StepDivider $isCompleted={isCompleted} />
            )}
          </div>
        )
      })}
    </LinearContainer>
  )
}
