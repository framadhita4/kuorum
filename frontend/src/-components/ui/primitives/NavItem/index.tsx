import type { ButtonHTMLAttributes, ReactNode } from 'react'
import * as React from 'react'

import { css } from '@emotion/react'
import styled from '@emotion/styled'
import { Pin } from 'lucide-react'

import { theme } from '../../theme'

export type NavItemVariant = 'default' | 'accent'

export interface NavItemProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: ReactNode
  label?: ReactNode
  isActive?: boolean
  variant?: NavItemVariant
  isPinned?: boolean
  indicator?: ReactNode
  endAction?: ReactNode
  children?: ReactNode
}

const getVariantStyles = (variant: NavItemVariant, isActive: boolean) => {
  if (isActive) {
    return css`
      background: #ffffff;
      color: ${theme.colors.ink.primary};
      border-color: rgba(121, 110, 255, 0.3);
      box-shadow: ${theme.shadows.sm};
      font-weight: ${theme.typography.weights.medium};

      .nav-icon {
        color: ${theme.colors.accent.purple};
      }
    `
  }

  if (variant === 'accent') {
    return css`
      background: transparent;
      color: ${theme.colors.accent.purple};
      border-color: transparent;
      font-weight: ${theme.typography.weights.medium};

      .nav-icon {
        color: ${theme.colors.accent.purple};
      }

      &:hover:not(:disabled) {
        background: rgba(121, 110, 255, 0.08);
        color: ${theme.colors.accent.purple};
      }
    `
  }

  return css`
    background: transparent;
    color: ${theme.colors.ink.secondary};
    border-color: transparent;
    font-weight: ${theme.typography.weights.regular};

    .nav-icon {
      color: ${theme.colors.ink.muted};
      transition: ${theme.transitions.fast};
    }

    &:hover:not(:disabled) {
      background: ${theme.colors.canvas.neutral};
      color: ${theme.colors.ink.primary};

      .nav-icon {
        color: ${theme.colors.ink.primary};
      }
    }
  `
}

const StyledNavItem = styled.button<{
  $variant: NavItemVariant
  $isActive: boolean
}>`
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: ${theme.radius.button};
  border: 1px solid transparent;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 0.85rem;
  line-height: ${theme.typography.lineHeights.normal};
  text-align: left;
  cursor: pointer;
  outline: none;
  user-select: none;
  transition: ${theme.transitions.fast};

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  ${({ $variant, $isActive }) => getVariantStyles($variant, $isActive)}
`

const NavLabel = styled.span`
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

const IconContainer = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
`

export const NavItem = React.forwardRef<HTMLButtonElement, NavItemProps>(
  (
    {
      icon,
      label,
      isActive = false,
      variant = 'default',
      isPinned = false,
      indicator,
      endAction,
      children,
      type = 'button',
      ...props
    },
    ref,
  ) => {
    return (
      <StyledNavItem
        ref={ref}
        type={type}
        $variant={variant}
        $isActive={isActive}
        {...props}
      >
        {icon && <IconContainer className="nav-icon">{icon}</IconContainer>}
        {label ? <NavLabel>{label}</NavLabel> : children}
        {isPinned && (
          <Pin
            size={12}
            color={theme.colors.accent.purple}
            style={{ transform: 'rotate(45deg)', flexShrink: 0 }}
          />
        )}
        {indicator && (
          <span style={{ display: 'inline-flex', flexShrink: 0 }}>
            {indicator}
          </span>
        )}
        {endAction && (
          <span style={{ display: 'inline-flex', flexShrink: 0 }}>
            {endAction}
          </span>
        )}
      </StyledNavItem>
    )
  },
)

NavItem.displayName = 'NavItem'
