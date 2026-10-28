import type { HTMLAttributes, ReactNode } from 'react'

import { css } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type CardVariant =
  'default' | 'interactive' | 'flat' | 'outline' | 'glow'
export type CardPadding = 'none' | 'sm' | 'md' | 'lg'
export type CardGlowColor = 'violet' | 'cyan' | 'emerald' | 'rose' | 'amber'

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant
  padding?: CardPadding
  glowColor?: CardGlowColor
  children: ReactNode
}

const getVariantStyles = (variant: CardVariant) => {
  switch (variant) {
    case 'interactive':
      return css`
        background: ${theme.colors.canvas.primary};
        border: ${theme.colors.border.subtle};
        box-shadow: ${theme.shadows.sm};
        cursor: pointer;

        &:hover {
          background: ${theme.colors.canvas.subtle};
          border-color: rgba(100, 111, 121, 0.4);
          transform: translateY(-2px);
          box-shadow: ${theme.shadows.md};
        }

        &:active {
          transform: translateY(0);
        }
      `
    case 'outline':
      return css`
        background: ${theme.colors.canvas.primary};
        border: ${theme.colors.border.subtle};
      `
    case 'flat':
      return css`
        background: ${theme.colors.canvas.subtle};
        border: 1px solid transparent;
      `
    case 'glow':
    case 'default':
    default:
      return css`
        background: ${theme.colors.canvas.primary};
        border: ${theme.colors.border.subtle};
        box-shadow: ${theme.shadows.sm};
      `
  }
}

const StyledCard = styled.div<{
  $variant: CardVariant
  $padding: CardPadding
}>`
  border-radius: ${theme.radius.card};
  position: relative;
  transition: ${theme.transitions.fast};

  /* Internal component padding */
  ${({ $padding }) => {
    switch ($padding) {
      case 'none':
        return 'padding: 0;'
      case 'sm':
        return `padding: ${theme.spacing['4']};`
      case 'lg':
        return `padding: ${theme.spacing['8']};`
      case 'md':
      default:
        return `padding: ${theme.spacing['6']};`
    }
  }}

  ${({ $variant }) => getVariantStyles($variant)}
`

export const CardHeader = styled.div`
  display: flex;
  flex-direction: column;
  gap: ${theme.spacing['1']};
  margin-bottom: ${theme.spacing['4']};
`

export const CardTitle = styled.h3`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h4};
  font-weight: ${theme.typography.weights.medium};
  letter-spacing: -0.01em;
  color: ${theme.colors.ink.primary};
  margin: 0;
  display: flex;
  align-items: center;
  gap: ${theme.spacing['2']};
`

export const CardDescription = styled.p`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  color: ${theme.colors.ink.secondary};
  line-height: ${theme.typography.lineHeights.normal};
  margin: 0;
`

export const CardContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: ${theme.spacing['4']};
`

export const CardFooter = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: ${theme.spacing['6']};
  padding-top: ${theme.spacing['4']};
  border-top: ${theme.colors.border.subtle};
`

export const Card = ({
  variant = 'default',
  padding = 'md',
  children,
  ...props
}: CardProps) => {
  return (
    <StyledCard $variant={variant} $padding={padding} {...props}>
      {children}
    </StyledCard>
  )
}
