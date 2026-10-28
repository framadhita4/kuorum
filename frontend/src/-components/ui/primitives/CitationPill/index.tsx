import type { ButtonHTMLAttributes, ReactNode } from 'react'
import * as React from 'react'

import styled from '@emotion/styled'

import { theme } from '../../theme'

export interface CitationPillProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  num?: number | string
  isActive?: boolean
  children?: ReactNode
}

const StyledCitationPill = styled.button<{ $isActive?: boolean }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0 3px;
  padding: 1px 7px;
  border-radius: ${theme.radius.pill};
  font-size: 0.78rem;
  font-weight: 600;
  font-family: ${theme.typography.fontFamily.mono};
  border: 1px solid
    ${({ $isActive }) =>
      $isActive ? theme.colors.accent.purple : 'rgba(121, 110, 255, 0.35)'};
  background-color: ${({ $isActive }) =>
    $isActive ? theme.colors.accent.purple : theme.colors.canvas.subtle};
  color: ${({ $isActive }) =>
    $isActive ? '#ffffff' : theme.colors.accent.purple};
  cursor: pointer;
  vertical-align: baseline;
  outline: none;
  user-select: none;
  transition: ${theme.transitions.fast};

  &:hover {
    background-color: ${theme.colors.accent.purple};
    border-color: ${theme.colors.accent.purple};
    color: #ffffff;
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 1px;
  }
`

export const CitationPill = React.forwardRef<
  HTMLButtonElement,
  CitationPillProps
>(({ num, isActive = false, children, type = 'button', ...props }, ref) => {
  return (
    <StyledCitationPill ref={ref} type={type} $isActive={isActive} {...props}>
      {num !== undefined ? `[${num}]` : children}
    </StyledCitationPill>
  )
})

CitationPill.displayName = 'CitationPill'
