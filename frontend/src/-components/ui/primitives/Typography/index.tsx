import type { HTMLAttributes, ReactNode } from 'react'

import styled from '@emotion/styled'

import { theme } from '../../theme'

export interface DisplayHeadingProps extends HTMLAttributes<HTMLHeadingElement> {
  as?: 'h1' | 'h2' | 'h3' | 'h4'
  children: ReactNode
}

export const GradientText = styled.span`
  color: ${theme.colors.ink.primary};
  font-weight: inherit;
`

export const DisplayHeading = styled.h1`
  font-family: ${theme.typography.fontFamily.sans};
  font-weight: ${theme.typography.weights.light};
  letter-spacing: ${theme.typography.letterSpacing.tightest};
  line-height: ${theme.typography.lineHeights.tight};
  color: ${theme.colors.ink.primary};
`

export const SectionHeading = styled.h2`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h2};
  font-weight: ${theme.typography.weights.light};
  letter-spacing: ${theme.typography.letterSpacing.tight};
  line-height: ${theme.typography.lineHeights.snug};
  color: ${theme.colors.ink.primary};
  margin: 0;
`

export const Text = styled.p<{
  $variant?: 'primary' | 'secondary' | 'muted'
  $size?: 'sm' | 'body' | 'subtitle'
}>`
  font-family: ${theme.typography.fontFamily.sans};
  line-height: ${theme.typography.lineHeights.normal};
  margin: 0;

  color: ${({ $variant }) => {
    switch ($variant) {
      case 'muted':
        return theme.colors.ink.muted
      case 'secondary':
        return theme.colors.ink.secondary
      case 'primary':
      default:
        return theme.colors.ink.primary
    }
  }};

  font-size: ${({ $size }) => {
    switch ($size) {
      case 'sm':
        return theme.typography.sizes.sm
      case 'subtitle':
        return theme.typography.sizes.subtitle
      case 'body':
      default:
        return theme.typography.sizes.body
    }
  }};
`

export const MonospaceToken = styled.code<{
  $highlight?: 'violet' | 'cyan' | 'emerald' | 'neutral'
}>`
  font-family: ${theme.typography.fontFamily.mono};
  font-size: ${theme.typography.sizes.caption};
  padding: 2px 7px;
  border-radius: ${theme.radius.sm};
  border: 1px solid rgba(100, 111, 121, 0.2);
  background: ${theme.colors.canvas.neutral};
  color: ${theme.colors.ink.primary};
  letter-spacing: 0;
`

export const SectionLabel = styled.span`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.micro};
  font-weight: ${theme.typography.weights.medium};
  text-transform: uppercase;
  letter-spacing: ${theme.typography.letterSpacing.widest};
  color: ${theme.colors.ink.muted};
  user-select: none;
`
