import type { ButtonHTMLAttributes, ReactNode } from 'react'
import * as React from 'react'

import { css } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type PillVariant = 'neutral' | 'purple' | 'teal' | 'active'
export type PillSize = 'sm' | 'md'

export interface PillProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: PillVariant
  size?: PillSize
  icon?: ReactNode
  label?: ReactNode
  isClickable?: boolean
  children?: ReactNode
}

const getVariantStyles = (variant: PillVariant, isClickable: boolean) => {
  switch (variant) {
    case 'purple':
      return css`
        background-color: rgba(121, 110, 255, 0.12);
        border-color: rgba(121, 110, 255, 0.4);
        color: ${theme.colors.accent.purple};

        .pill-icon {
          color: ${theme.colors.accent.purple};
        }

        ${
          isClickable &&
          css`
            &:hover:not(:disabled) {
              background-color: rgba(121, 110, 255, 0.18);
              border-color: rgba(121, 110, 255, 0.55);
            }
          `
        }
      `
    case 'teal':
      return css`
        background-color: rgba(13, 148, 136, 0.12);
        border-color: rgba(13, 148, 136, 0.4);
        color: #0d9488;

        .pill-icon {
          color: #0d9488;
        }

        ${
          isClickable &&
          css`
            &:hover:not(:disabled) {
              background-color: rgba(13, 148, 136, 0.18);
              border-color: rgba(13, 148, 136, 0.55);
            }
          `
        }
      `
    case 'active':
      return css`
        background-color: ${theme.colors.accent.midnight};
        border-color: transparent;
        color: #ffffff;

        .pill-icon {
          color: #ffffff;
        }

        ${
          isClickable &&
          css`
            &:hover:not(:disabled) {
              background-color: ${theme.colors.accent.purple};
            }
          `
        }
      `
    case 'neutral':
    default:
      return css`
        background-color: ${theme.colors.canvas.neutral};
        border-color: rgba(100, 111, 121, 0.16);
        color: ${theme.colors.ink.secondary};

        .pill-icon {
          color: ${theme.colors.ink.muted};
        }

        ${
          isClickable &&
          css`
            &:hover:not(:disabled) {
              background-color: ${theme.colors.neutral['200']};
              color: ${theme.colors.ink.primary};
              border-color: rgba(100, 111, 121, 0.3);

              .pill-icon {
                color: ${theme.colors.ink.primary};
              }
            }
          `
        }
      `
  }
}

const getSizeStyles = (size: PillSize) => {
  switch (size) {
    case 'md':
      return css`
        padding: 8px 16px;
        font-size: 0.84rem;
        gap: ${theme.spacing['2']};
      `
    case 'sm':
    default:
      return css`
        padding: 4px 10px;
        font-size: 0.74rem;
        gap: 5px;
      `
  }
}

const StyledPillButton = styled.button<{
  $variant: PillVariant
  $size: PillSize
  $isClickable: boolean
}>`
  display: inline-flex;
  align-items: center;
  border-radius: ${theme.radius.pill};
  border: 1px solid transparent;
  font-family: ${theme.typography.fontFamily.sans};
  font-weight: ${theme.typography.weights.medium};
  cursor: ${({ $isClickable }) => ($isClickable ? 'pointer' : 'default')};
  outline: none;
  user-select: none;
  white-space: nowrap;
  transition: ${theme.transitions.fast};

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  ${({ $variant, $isClickable }) => getVariantStyles($variant, $isClickable)}
  ${({ $size }) => getSizeStyles($size)}
`

const PillIconWrapper = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: ${theme.transitions.fast};
`

const PillLabel = styled.span`
  display: inline-block;
`

export const Pill = React.forwardRef<HTMLButtonElement, PillProps>(
  (
    {
      variant = 'neutral',
      size = 'sm',
      icon,
      label,
      isClickable,
      onClick,
      children,
      type = 'button',
      ...props
    },
    ref,
  ) => {
    const clickable = isClickable ?? Boolean(onClick)

    return (
      <StyledPillButton
        ref={ref}
        type={type}
        $variant={variant}
        $size={size}
        $isClickable={clickable}
        onClick={onClick}
        {...props}
      >
        {icon && (
          <PillIconWrapper className="pill-icon">{icon}</PillIconWrapper>
        )}
        {label ? <PillLabel>{label}</PillLabel> : children}
      </StyledPillButton>
    )
  },
)

Pill.displayName = 'Pill'
