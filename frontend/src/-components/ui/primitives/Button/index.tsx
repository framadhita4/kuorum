import type { ReactNode } from 'react'
import * as React from 'react'

import { Button as BaseButton } from '@base-ui/react'
import { css, keyframes } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type ButtonVariant =
  'primary' | 'secondary' | 'outline' | 'ghost' | 'violet' | 'cyan'

export type ButtonSize = 'sm' | 'md' | 'lg' | 'icon' | 'icon-sm'

export interface ButtonProps extends Omit<
  React.ComponentProps<typeof BaseButton>,
  'color'
> {
  variant?: ButtonVariant
  size?: ButtonSize
  isLoading?: boolean
  startIcon?: ReactNode
  endIcon?: ReactNode
  fullWidth?: boolean
  children?: ReactNode
  className?: string
  style?: React.CSSProperties
}

const spin = keyframes`
  to {
    transform: rotate(360deg);
  }
`

const Spinner = styled.span<{ $inverse?: boolean }>`
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid
    ${({ $inverse }) =>
      $inverse ? 'rgba(13, 14, 16, 0.25)' : 'rgba(255, 255, 255, 0.25)'};
  border-radius: 50%;
  border-top-color: ${({ $inverse }) =>
    $inverse ? theme.colors.ink.primary : theme.colors.ink.inverse};
  animation: ${spin} 0.75s linear infinite;
`

const getVariantStyles = (variant: ButtonVariant) => {
  switch (variant) {
    case 'primary':
      return css`
        background: ${theme.colors.ink.primary};
        color: ${theme.colors.ink.inverse};
        border: 1px solid transparent;

        &:hover:not(:disabled) {
          background: ${theme.colors.brand.coral};
          color: ${theme.colors.ink.primary};
        }

        &:active:not(:disabled) {
          background: ${theme.colors.brand.vibrant};
          color: ${theme.colors.ink.inverse};
        }
      `
    case 'violet':
      return css`
        background: ${theme.colors.accent.purple};
        color: #ffffff;
        border: 1px solid transparent;

        &:hover:not(:disabled) {
          background: ${theme.colors.accent.purpleDark};
          color: #ffffff;
        }

        &:active:not(:disabled) {
          background: ${theme.colors.accent.midnight};
        }
      `
    case 'cyan':
    case 'outline':
      return css`
        background: transparent;
        color: ${theme.colors.ink.primary};
        border: 1px solid ${theme.colors.ink.primary};

        &:hover:not(:disabled) {
          background: ${theme.colors.brand.coral};
          border-color: ${theme.colors.brand.coral};
          color: ${theme.colors.ink.primary};
        }

        &:active:not(:disabled) {
          background: ${theme.colors.brand.vibrant};
          border-color: ${theme.colors.brand.vibrant};
          color: #ffffff;
        }
      `
    case 'ghost':
      return css`
        background: transparent;
        color: ${theme.colors.ink.secondary};
        border: 1px solid transparent;

        &:hover:not(:disabled) {
          background: ${theme.colors.canvas.neutral};
          color: ${theme.colors.ink.primary};
        }
      `
    case 'secondary':
    default:
      // Neutral Soft Surface Button
      return css`
        background: ${theme.colors.canvas.neutral};
        color: ${theme.colors.ink.primary};
        border: 1px solid transparent;

        &:hover:not(:disabled) {
          background: #eaeaea;
          color: ${theme.colors.ink.primary};
        }

        &:active:not(:disabled) {
          background: #dfdfdf;
        }
      `
  }
}

const getSizeStyles = (size: ButtonSize) => {
  switch (size) {
    case 'sm':
      return css`
        padding: 6px 14px;
        font-size: ${theme.typography.sizes.sm};
        gap: ${theme.spacing['2']};
        border-radius: ${theme.radius.button};
      `
    case 'lg':
      return css`
        padding: 14px 28px;
        font-size: ${theme.typography.sizes.body};
        gap: ${theme.spacing['3']};
        border-radius: ${theme.radius.button};
      `
    case 'icon':
      return css`
        width: 36px;
        height: 36px;
        padding: 0;
        border-radius: ${theme.radius.button};
      `
    case 'icon-sm':
      return css`
        width: 28px;
        height: 28px;
        padding: 0;
        border-radius: ${theme.radius.sm};
      `
    case 'md':
    default:
      return css`
        padding: 10px 20px;
        font-size: ${theme.typography.sizes.sm};
        gap: ${theme.spacing['2']};
        border-radius: ${theme.radius.button};
      `
  }
}

const StyledBaseButton = styled(BaseButton)<{
  $variant: ButtonVariant
  $size: ButtonSize
  $fullWidth?: boolean
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: ${({ $fullWidth }) => ($fullWidth ? '100%' : 'auto')};
  font-family: ${theme.typography.fontFamily.sans};
  font-weight: ${theme.typography.weights.medium};
  letter-spacing: 0.01em;
  cursor: pointer;
  outline: none;
  text-decoration: none;
  white-space: nowrap;
  user-select: none;
  box-shadow: none;
  transition: ${theme.transitions.fast};

  &:focus-visible {
    border-color: ${theme.colors.accent.purple};
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  ${({ $variant }) => getVariantStyles($variant)}
  ${({ $size }) => getSizeStyles($size)}
`

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = 'secondary',
      size = 'md',
      isLoading = false,
      startIcon,
      endIcon,
      disabled,
      fullWidth = false,
      children,
      ...props
    },
    ref,
  ) => {
    const isLightButton = variant === 'primary' || variant === 'violet'

    return (
      <StyledBaseButton
        ref={ref}
        $variant={variant}
        $size={size}
        $fullWidth={fullWidth}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <Spinner $inverse={!isLightButton} />
        ) : (
          startIcon && (
            <span style={{ display: 'inline-flex', alignItems: 'center' }}>
              {startIcon}
            </span>
          )
        )}
        {children}
        {!isLoading && endIcon && (
          <span style={{ display: 'inline-flex', alignItems: 'center' }}>
            {endIcon}
          </span>
        )}
      </StyledBaseButton>
    )
  },
)

Button.displayName = 'Button'

export interface IconButtonProps extends Omit<
  ButtonProps,
  'startIcon' | 'endIcon'
> {
  'aria-label'?: string
}

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ size = 'icon', variant = 'ghost', children, ...props }, ref) => {
    return (
      <Button ref={ref} size={size} variant={variant} {...props}>
        {children}
      </Button>
    )
  },
)

IconButton.displayName = 'IconButton'
