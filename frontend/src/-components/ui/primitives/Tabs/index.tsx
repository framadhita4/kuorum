import { Tabs as BaseTabs } from '@base-ui/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type TabsVariant = 'pill' | 'underline' | 'segment'

export const TabsRoot = styled(BaseTabs.Root)`
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: ${theme.spacing.md};
`

export const TabsList = styled(BaseTabs.List)<{ $variant?: TabsVariant }>`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  position: relative;
  background: ${theme.colors.canvas.neutral};
  border: 1px solid rgba(100, 111, 121, 0.16);
  padding: 4px;
  border-radius: ${theme.radius.pill};
`

export const TabsTab = styled(BaseTabs.Tab)<{ $variant?: TabsVariant }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: ${theme.spacing['2']};
  padding: 8px 18px;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.secondary};
  border: 1px solid transparent;
  border-radius: ${theme.radius.pill};
  background: transparent;
  cursor: pointer;
  outline: none;
  user-select: none;
  transition: ${theme.transitions.fast};

  &:hover:not([data-disabled]) {
    color: ${theme.colors.ink.primary};
    background: rgba(0, 0, 0, 0.04);
  }

  &[data-active] {
    color: #ffffff;
    background: ${theme.colors.accent.midnight};
    font-weight: ${theme.typography.weights.medium};
  }

  &[data-disabled] {
    opacity: 0.35;
    cursor: not-allowed;
  }
`

export const TabsPanel = styled(BaseTabs.Panel)`
  outline: none;
  width: 100%;
  display: flex;
  flex-direction: column;
`

export const TabsIndicator = styled(BaseTabs.Indicator)`
  position: absolute;
`

export const Tabs = {
  Root: TabsRoot,
  List: TabsList,
  Tab: TabsTab,
  Panel: TabsPanel,
  Indicator: TabsIndicator,
}
