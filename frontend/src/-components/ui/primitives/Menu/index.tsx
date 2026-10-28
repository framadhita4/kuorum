import type { CSSProperties, ReactNode } from 'react'
import * as React from 'react'
import { createContext, useContext } from 'react'

import { Menu as BaseMenu } from '@base-ui/react'
import { keyframes } from '@emotion/react'
import styled from '@emotion/styled'
import { Check, ChevronRight } from 'lucide-react'

import { theme } from '../../theme'

export type MenuVariant = 'light' | 'dark'
export type MenuSize = 'sm' | 'md'

interface MenuContextValue {
  variant: MenuVariant
  size: MenuSize
}

const MenuThemeContext = createContext<MenuContextValue>({
  variant: 'light',
  size: 'md',
})

export const useMenuTheme = () => useContext(MenuThemeContext)

const menuPopupIn = keyframes`
  0% {
    opacity: 0;
    transform: scale(0.96) translateY(-4px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
`

const checkPop = keyframes`
  0% {
    opacity: 0;
    transform: scale(0.4);
  }
  70% {
    transform: scale(1.2);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
`

const AnimatedCheck = styled(Check)`
  animation: ${checkPop} 0.18s cubic-bezier(0.16, 1, 0.3, 1);
`

export interface MenuRootProps extends BaseMenu.Root.Props {
  variant?: MenuVariant
  size?: MenuSize
  children?: ReactNode
}

export const MenuRoot = ({
  variant = 'light',
  size = 'md',
  children,
  ...props
}: MenuRootProps) => {
  return (
    <MenuThemeContext.Provider value={{ variant, size }}>
      <BaseMenu.Root {...props}>{children}</BaseMenu.Root>
    </MenuThemeContext.Provider>
  )
}

export const MenuTrigger = BaseMenu.Trigger
export const MenuPortal = BaseMenu.Portal
export const MenuPositioner = BaseMenu.Positioner
export const MenuGroup = BaseMenu.Group

export interface MenuPopupProps extends BaseMenu.Popup.Props {
  variant?: MenuVariant
  minWidth?: string | number
  maxWidth?: string | number
  maxHeight?: string | number
}

const StyledPopup = styled(BaseMenu.Popup)<{
  $variant: MenuVariant
  $minWidth?: string | number
  $maxWidth?: string | number
  $maxHeight?: string | number
}>`
  min-width: ${({ $minWidth }) =>
    typeof $minWidth === 'number' ? `${$minWidth}px` : ($minWidth ?? '200px')};
  max-width: ${({ $maxWidth }) =>
    typeof $maxWidth === 'number' ? `${$maxWidth}px` : ($maxWidth ?? '340px')};
  max-height: ${({ $maxHeight }) =>
    typeof $maxHeight === 'number'
      ? `${$maxHeight}px`
      : ($maxHeight ?? '380px')};
  overflow-y: auto;
  border-radius: ${theme.radius.lg};
  padding: 6px;
  outline: none;
  font-family: ${theme.typography.fontFamily.sans};
  user-select: none;
  animation: ${menuPopupIn} 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  box-sizing: border-box;

  /* Custom scrollbar */
  scrollbar-width: thin;
  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-thumb {
    border-radius: 4px;
    background: ${({ $variant }) =>
      $variant === 'dark'
        ? 'rgba(255, 255, 255, 0.15)'
        : 'rgba(100, 111, 121, 0.2)'};
  }

  ${({ $variant }) =>
    $variant === 'dark'
      ? `
    background-color: #1e1f22;
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-shadow: 0 16px 40px rgba(13, 14, 16, 0.4), 0 4px 12px rgba(13, 14, 16, 0.2);
    color: #f5f5f7;
  `
      : `
    background-color: #ffffff;
    border: 1px solid rgba(100, 111, 121, 0.18);
    box-shadow: 0 16px 40px rgba(13, 14, 16, 0.12), 0 4px 12px rgba(13, 14, 16, 0.04);
    color: ${theme.colors.ink.primary};
  `}
`

export const MenuPopup = ({
  variant,
  minWidth,
  maxWidth,
  maxHeight,
  style,
  ...props
}: MenuPopupProps) => {
  const context = useMenuTheme()
  const activeVariant = variant ?? context.variant

  return (
    <StyledPopup
      $variant={activeVariant}
      $minWidth={minWidth}
      $maxWidth={maxWidth}
      $maxHeight={maxHeight}
      style={style}
      {...props}
    />
  )
}

export interface MenuItemProps extends Omit<BaseMenu.Item.Props, 'title'> {
  title?: ReactNode
  subtitle?: ReactNode
  startIcon?: ReactNode
  endIcon?: ReactNode
  selected?: boolean
  inset?: boolean
  variant?: 'default' | 'danger'
}

const StyledItem = styled(BaseMenu.Item)<{
  $variant: MenuVariant
  $size: MenuSize
  $itemVariant: 'default' | 'danger'
  $selected?: boolean
}>`
  display: flex;
  align-items: center;
  width: 100%;
  padding: ${({ $size }) => ($size === 'sm' ? '6px 8px' : '8px 10px')};
  border-radius: ${theme.radius.md};
  cursor: pointer;
  outline: none;
  text-decoration: none;
  box-sizing: border-box;
  gap: 8px;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);

  &[data-disabled] {
    opacity: 0.45;
    cursor: not-allowed;
    pointer-events: none;
  }

  /* Colors based on theme variant */
  ${({ $variant, $itemVariant, $selected }) => {
    if ($itemVariant === 'danger') {
      return `
        color: ${theme.colors.status.error};
        &:hover, &[data-highlighted] {
          background-color: rgba(239, 68, 68, 0.08);
          color: ${theme.colors.status.error};
        }
      `
    }

    if ($variant === 'dark') {
      return `
        color: #f5f5f7;
        ${$selected ? 'background-color: rgba(121, 110, 255, 0.12);' : ''}
        &:hover, &[data-highlighted] {
          background-color: rgba(255, 255, 255, 0.08);
          color: #ffffff;
        }
      `
    }

    return `
      color: ${theme.colors.ink.primary};
      ${$selected ? 'background-color: rgba(121, 110, 255, 0.08);' : ''}
      &:hover, &[data-highlighted] {
        background-color: ${theme.colors.canvas.subtle};
        color: ${theme.colors.ink.primary};
      }
    `
  }}
`

const ItemContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
  text-align: left;
`

const ItemTitle = styled.div<{ $size: MenuSize }>`
  font-size: ${({ $size }) => ($size === 'sm' ? '0.78rem' : '0.84rem')};
  font-weight: ${theme.typography.weights.medium};
  line-height: 1.25;
`

const ItemSubtitle = styled.div<{ $variant: MenuVariant; $size: MenuSize }>`
  font-size: ${({ $size }) => ($size === 'sm' ? '0.68rem' : '0.73rem')};
  color: ${({ $variant }) =>
    $variant === 'dark' ? '#9aa0a6' : theme.colors.ink.secondary};
  line-height: 1.25;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`

const SlotContainer = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
`

const CheckIndicatorSlot = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  flex-shrink: 0;
`

export const MenuItem = ({
  title,
  subtitle,
  startIcon,
  endIcon,
  selected,
  inset,
  variant = 'default',
  children,
  ...props
}: MenuItemProps) => {
  const context = useMenuTheme()

  const hasCheckSlot = selected !== undefined || inset

  return (
    <StyledItem
      $variant={context.variant}
      $size={context.size}
      $itemVariant={variant}
      $selected={selected}
      {...props}
    >
      {hasCheckSlot && (
        <CheckIndicatorSlot>
          {selected && (
            <AnimatedCheck
              size={14}
              strokeWidth={2.5}
              color={
                context.variant === 'dark'
                  ? '#ffffff'
                  : theme.colors.accent.purple
              }
            />
          )}
        </CheckIndicatorSlot>
      )}

      {startIcon && <SlotContainer>{startIcon}</SlotContainer>}

      {title !== undefined ? (
        <ItemContent>
          <ItemTitle $size={context.size}>{title}</ItemTitle>
          {subtitle !== undefined && (
            <ItemSubtitle $variant={context.variant} $size={context.size}>
              {subtitle}
            </ItemSubtitle>
          )}
        </ItemContent>
      ) : (
        <ItemContent>{children}</ItemContent>
      )}

      {endIcon && (
        <SlotContainer style={{ marginLeft: 'auto' }}>{endIcon}</SlotContainer>
      )}
    </StyledItem>
  )
}

const StyledSeparator = styled(BaseMenu.Separator)<{ $variant: MenuVariant }>`
  height: 1px;
  margin: 4px 0;
  background-color: ${({ $variant }) =>
    $variant === 'dark'
      ? 'rgba(255, 255, 255, 0.08)'
      : 'rgba(100, 111, 121, 0.15)'};
`

export const MenuSeparator = (props: BaseMenu.Separator.Props) => {
  const context = useMenuTheme()
  return <StyledSeparator $variant={context.variant} {...props} />
}

const StyledGroupLabel = styled.div<{
  $variant: MenuVariant
  $size: MenuSize
}>`
  padding: 6px 10px 4px;
  font-size: ${({ $size }) => ($size === 'sm' ? '0.66rem' : '0.7rem')};
  font-weight: ${theme.typography.weights.semibold};
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: ${({ $variant }) =>
    $variant === 'dark' ? '#8e8e93' : theme.colors.ink.muted};
  user-select: none;
`

export type MenuGroupLabelProps = React.HTMLAttributes<HTMLDivElement>

export const MenuGroupLabel = (props: MenuGroupLabelProps) => {
  const context = useMenuTheme()
  return (
    <StyledGroupLabel
      $variant={context.variant}
      $size={context.size}
      aria-hidden="true"
      {...props}
    />
  )
}

export const MenuHeader = MenuGroupLabel

const StyledShortcut = styled.span<{ $variant: MenuVariant }>`
  margin-left: auto;
  font-size: 0.72rem;
  font-family: ${theme.typography.fontFamily.mono};
  color: ${({ $variant }) =>
    $variant === 'dark' ? '#8e8e93' : theme.colors.ink.muted};
  letter-spacing: 0.05em;
  padding-left: 12px;
`

export const MenuShortcut = ({ children }: { children: ReactNode }) => {
  const context = useMenuTheme()
  return <StyledShortcut $variant={context.variant}>{children}</StyledShortcut>
}

export const MenuSubmenuRoot = BaseMenu.SubmenuRoot

const StyledSubmenuTrigger = styled(BaseMenu.SubmenuTrigger)<{
  $variant: MenuVariant
  $size: MenuSize
}>`
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: ${({ $size }) => ($size === 'sm' ? '6px 8px' : '8px 10px')};
  border-radius: ${theme.radius.md};
  cursor: pointer;
  outline: none;
  box-sizing: border-box;
  gap: 8px;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);

  &[data-disabled] {
    opacity: 0.45;
    cursor: not-allowed;
    pointer-events: none;
  }

  ${({ $variant }) =>
    $variant === 'dark'
      ? `
    color: #f5f5f7;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.08);
    &:hover, &[data-highlighted], &[data-open] {
      background-color: rgba(255, 255, 255, 0.12);
      color: #ffffff;
    }
  `
      : `
    color: ${theme.colors.ink.primary};
    background-color: ${theme.colors.canvas.neutral};
    border: 1px solid rgba(100, 111, 121, 0.16);
    &:hover, &[data-highlighted], &[data-open] {
      background-color: #ffffff;
      border-color: ${theme.colors.accent.purple};
      box-shadow: 0 2px 10px rgba(121, 110, 255, 0.14);
    }
  `}
`

export interface MenuSubmenuProps {
  label: ReactNode
  subtitle?: ReactNode
  startIcon?: ReactNode
  endIcon?: ReactNode
  children: ReactNode
  disabled?: boolean
  sideOffset?: number
  minWidth?: string | number
}

export const MenuSubmenu = ({
  label,
  subtitle,
  startIcon,
  endIcon,
  children,
  disabled,
  sideOffset = 6,
  minWidth = 190,
}: MenuSubmenuProps) => {
  const context = useMenuTheme()

  return (
    <BaseMenu.SubmenuRoot>
      <StyledSubmenuTrigger
        $variant={context.variant}
        $size={context.size}
        disabled={disabled}
      >
        <div
          style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0 }}
        >
          {startIcon && <SlotContainer>{startIcon}</SlotContainer>}
          <ItemContent>
            <ItemTitle $size={context.size}>{label}</ItemTitle>
            {subtitle && (
              <ItemSubtitle $variant={context.variant} $size={context.size}>
                {subtitle}
              </ItemSubtitle>
            )}
          </ItemContent>
        </div>

        <SlotContainer style={{ marginLeft: 'auto' }}>
          {endIcon ?? <ChevronRight size={14} style={{ opacity: 0.7 }} />}
        </SlotContainer>
      </StyledSubmenuTrigger>

      <BaseMenu.Portal>
        <BaseMenu.Positioner
          side="right"
          align="start"
          sideOffset={sideOffset}
          style={{ zIndex: 1010 }}
        >
          <MenuPopup minWidth={minWidth}>{children}</MenuPopup>
        </BaseMenu.Positioner>
      </BaseMenu.Portal>
    </BaseMenu.SubmenuRoot>
  )
}

export interface MenuProps {
  trigger: ReactNode
  children: ReactNode
  open?: boolean
  defaultOpen?: boolean
  onOpenChange?: (open: boolean) => void
  side?: 'top' | 'bottom' | 'left' | 'right'
  align?: 'start' | 'center' | 'end'
  sideOffset?: number
  alignOffset?: number
  variant?: MenuVariant
  size?: MenuSize
  minWidth?: string | number
  maxWidth?: string | number
  modal?: boolean
  className?: string
  style?: CSSProperties
}

export const MenuComponent = ({
  trigger,
  children,
  open,
  defaultOpen,
  onOpenChange,
  side = 'bottom',
  align = 'start',
  sideOffset = 6,
  alignOffset = 0,
  variant = 'light',
  size = 'md',
  minWidth,
  maxWidth,
  modal = false,
  className,
  style,
}: MenuProps) => {
  return (
    <MenuRoot
      open={open}
      defaultOpen={defaultOpen}
      onOpenChange={onOpenChange}
      variant={variant}
      size={size}
      modal={modal}
    >
      <BaseMenu.Trigger
        render={
          React.isValidElement(trigger) ? (
            trigger
          ) : (
            <button type="button" className={className} style={style}>
              {trigger}
            </button>
          )
        }
      />

      <BaseMenu.Portal>
        <BaseMenu.Positioner
          side={side}
          align={align}
          sideOffset={sideOffset}
          alignOffset={alignOffset}
          style={{ zIndex: 1000 }}
        >
          <MenuPopup
            minWidth={minWidth}
            maxWidth={maxWidth}
            className={className}
            style={style}
          >
            {children}
          </MenuPopup>
        </BaseMenu.Positioner>
      </BaseMenu.Portal>
    </MenuRoot>
  )
}

export const Menu = Object.assign(MenuComponent, {
  Root: MenuRoot,
  Trigger: BaseMenu.Trigger,
  Portal: BaseMenu.Portal,
  Positioner: BaseMenu.Positioner,
  Popup: MenuPopup,
  Item: MenuItem,
  Submenu: MenuSubmenu,
  SubmenuRoot: MenuSubmenuRoot,
  SubmenuTrigger: BaseMenu.SubmenuTrigger,
  Separator: MenuSeparator,
  Group: MenuGroup,
  GroupLabel: MenuGroupLabel,
  Header: MenuHeader,
  Shortcut: MenuShortcut,
})
