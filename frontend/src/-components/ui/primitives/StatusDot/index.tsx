import type { HTMLAttributes } from 'react'

import { css, keyframes } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type StatusDotVariant =
  'success' | 'warning' | 'error' | 'purple' | 'neutral' | 'emerald' | 'amber'

export interface StatusDotProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: StatusDotVariant
  color?: string
  size?: number | string
  pulse?: boolean
}

const pulseKeyframes = keyframes`
  0% {
    transform: scale(0.95);
    opacity: 0.8;
  }
  50% {
    transform: scale(1.2);
    opacity: 1;
  }
  100% {
    transform: scale(0.95);
    opacity: 0.8;
  }
`

const getDotColor = (variant?: StatusDotVariant, customColor?: string) => {
  if (customColor) return customColor

  switch (variant) {
    case 'success':
    case 'emerald':
      return theme.colors.status.success
    case 'warning':
    case 'amber':
      return theme.colors.status.warning
    case 'error':
      return theme.colors.status.error
    case 'purple':
      return theme.colors.accent.purple
    case 'neutral':
    default:
      return theme.colors.ink.muted
  }
}

export const StatusDot = styled.span<
  StatusDotProps & {
    $variant?: StatusDotVariant
    $color?: string
    $size?: number | string
    $pulse?: boolean
  }
>`
  display: inline-block;
  width: ${({ $size, size }) => {
    const s = $size ?? size ?? 6
    return typeof s === 'number' ? `${s}px` : s
  }};
  height: ${({ $size, size }) => {
    const s = $size ?? size ?? 6
    return typeof s === 'number' ? `${s}px` : s
  }};
  border-radius: 50%;
  background-color: ${({ $variant, variant, $color, color }) =>
    getDotColor($variant ?? variant, $color ?? color)};
  flex-shrink: 0;
  transition: ${theme.transitions.fast};

  ${({ $pulse, pulse }) =>
    ($pulse || pulse) &&
    css`
      animation: ${pulseKeyframes} 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
    `}
`
