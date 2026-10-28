import type { HTMLAttributes, ReactNode } from 'react'

import { css } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type BadgeVariant =
  | 'neutral'
  | 'white'
  | 'success'
  | 'warning'
  | 'error'
  | 'violet'
  | 'indigo'
  | 'cyan'
  | 'emerald'
  | 'amber'
  | 'rose'

export type BadgeSize = 'sm' | 'md'

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  size?: BadgeSize
  dot?: boolean
  isMono?: boolean
  children: ReactNode
}

const getBadgeStyles = (variant: BadgeVariant) => {
  switch (variant) {
    case 'white':
      return css`
        background: ${theme.colors.canvas.elevated};
        color: ${theme.colors.ink.primary};
        border: 1px solid ${theme.colors.border.subtleColor};
      `
    case 'success':
    case 'emerald':
      return css`
        background: rgba(52, 199, 89, 0.12);
        color: #16a34a;
        border: 1px solid rgba(52, 199, 89, 0.25);
      `
    case 'warning':
    case 'amber':
      return css`
        background: rgba(234, 179, 8, 0.12);
        color: #ca8a04;
        border: 1px solid rgba(234, 179, 8, 0.25);
      `
    case 'error':
    case 'rose':
      return css`
        background: rgba(239, 68, 68, 0.12);
        color: #dc2626;
        border: 1px solid rgba(239, 68, 68, 0.25);
      `
    case 'cyan':
    case 'violet':
    case 'indigo':
      return css`
        background: rgba(121, 110, 255, 0.12);
        color: ${theme.colors.accent.purple};
        border: 1px solid rgba(121, 110, 255, 0.25);
      `
    case 'neutral':
    default:
      return css`
        background: ${theme.colors.canvas.neutral};
        color: ${theme.colors.ink.secondary};
        border: 1px solid rgba(100, 111, 121, 0.16);
      `
  }
}

const getDotColor = (variant: BadgeVariant) => {
  switch (variant) {
    case 'success':
    case 'emerald':
      return theme.colors.status.success
    case 'warning':
    case 'amber':
      return theme.colors.status.warning
    case 'error':
    case 'rose':
      return theme.colors.status.error
    case 'white':
      return theme.colors.ink.primary
    case 'violet':
    case 'indigo':
    case 'cyan':
      return theme.colors.accent.purple
    default:
      return theme.colors.ink.muted
  }
}

const StyledBadge = styled.span<{
  $variant: BadgeVariant
  $size: BadgeSize
  $isMono?: boolean
}>`
  display: inline-flex;
  align-items: center;
  gap: ${theme.spacing['1']};
  border-radius: ${theme.radius.pill};
  white-space: nowrap;
  font-weight: ${theme.typography.weights.medium};
  letter-spacing: ${({ $isMono }) => ($isMono ? '0' : '0.02em')};
  font-family: ${({ $isMono }) =>
    $isMono
      ? theme.typography.fontFamily.mono
      : theme.typography.fontFamily.sans};
  transition: ${theme.transitions.fast};

  ${({ $size }) =>
    $size === 'sm'
      ? css`
          font-size: ${theme.typography.sizes.micro};
          padding: 3px 8px;
        `
      : css`
          font-size: ${theme.typography.sizes.caption};
          padding: 4px 10px;
        `}

  ${({ $variant }) => getBadgeStyles($variant)}
`

const Dot = styled.span<{ $variant: BadgeVariant }>`
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
  background-color: ${({ $variant }) => getDotColor($variant)};
`

export const Badge = ({
  variant = 'neutral',
  size = 'md',
  dot = false,
  isMono = false,
  children,
  ...props
}: BadgeProps) => {
  return (
    <StyledBadge $variant={variant} $size={size} $isMono={isMono} {...props}>
      {dot && <Dot $variant={variant} />}
      {children}
    </StyledBadge>
  )
}
