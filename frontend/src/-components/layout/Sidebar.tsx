import { Link } from '@tanstack/react-router'

import styled from '@emotion/styled'
import {
  MessageSquare,
  Network,
  PanelLeftClose,
  Plus,
  Search,
} from 'lucide-react'

import {
  Button,
  IconButton,
  NavItem,
  SectionLabel,
  StatusDot,
} from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ResearchPipeline } from '#/types/pipeline.type'

export const SidebarRoot = styled.aside`
  width: 260px;
  background-color: ${theme.colors.canvas.subtle};
  border-right: ${theme.colors.border.subtle};
  display: flex;
  flex-direction: column;
  height: 100%;
  flex-shrink: 0;
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  color: ${theme.colors.ink.primary};
  user-select: none;
`

export const SidebarHeader = styled.div`
  padding: 16px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(100, 111, 121, 0.12);
`

const BrandLink = styled(Link)`
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: ${theme.colors.ink.primary};
`

const BrandName = styled.span`
  font-weight: ${theme.typography.weights.medium};
  font-size: 0.98rem;
  letter-spacing: ${theme.typography.letterSpacing.tight};
  color: ${theme.colors.ink.primary};
  line-height: ${theme.typography.lineHeights.tight};
`

export const SidebarNewAction = styled.div`
  padding: 16px 14px 12px;
`

export const SidebarNavSection = styled.div`
  padding: 0 10px;
  display: flex;
  flex-direction: column;
  gap: 3px;
`

export const SidebarContent = styled.div`
  margin-top: 16px;
  padding: 0 10px;
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 3px;
`

export const SidebarFooter = styled.div`
  border-top: 1px solid rgba(100, 111, 121, 0.16);
  padding: 12px 10px;
  display: flex;
  flex-direction: column;
  gap: 3px;
`

export interface SidebarProps {
  activeTopicKey: string
  onSelectTopic: (key: string) => void
  onNewThread: () => void
  onOpenGraph: () => void
  onOpenSearch?: () => void
  pipelines?: Array<ResearchPipeline>
}

export const Sidebar = ({
  activeTopicKey,
  onSelectTopic,
  onNewThread,
  onOpenGraph,
  onOpenSearch,
  pipelines,
}: SidebarProps) => {
  return (
    <SidebarRoot>
      {/* Top Header & Brand Mark */}
      <SidebarHeader>
        <BrandLink to="/">
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <BrandName>Kuorum</BrandName>
          </div>
        </BrandLink>

        <IconButton
          variant="ghost"
          size="icon-sm"
          title="Close Sidebar"
          aria-label="Close Sidebar"
        >
          <PanelLeftClose size={16} />
        </IconButton>
      </SidebarHeader>

      {/* Primary Dark "New Thread" Button (Turns Coral on hover) */}
      <SidebarNewAction>
        <Button
          variant="primary"
          size="md"
          fullWidth
          startIcon={<Plus size={16} strokeWidth={2.4} />}
          onClick={onNewThread}
        >
          New Thread
        </Button>
      </SidebarNewAction>

      {/* Primary Navigation Menu */}
      {/* <SidebarNavSection>
        <NavItem icon={<Library size={15} />} label="My Library" />
        <NavItem icon={<History size={15} />} label="History" />
      </SidebarNavSection> */}

      {/* Recents / Chat Threads */}
      <SidebarContent>
        <SectionLabel style={{ padding: '6px 12px' }}>
          Chat History
        </SectionLabel>

        {/* Real Dynamic Pipelines / Chat History */}
        {pipelines && pipelines.length > 0 ? (
          pipelines.map((pipe) => {
            const isActive = activeTopicKey === pipe.id
            const dotVariant =
              pipe.status === 'COMPLETED'
                ? 'success'
                : pipe.status === 'WAITING_INPUT'
                  ? 'warning'
                  : 'purple'

            return (
              <NavItem
                key={pipe.id}
                isActive={isActive}
                onClick={() => onSelectTopic(pipe.id)}
                icon={<MessageSquare size={14} />}
                label={pipe.title}
                title={pipe.title}
                isPinned={pipe.pinned}
                indicator={
                  <StatusDot
                    variant={dotVariant}
                    size={7}
                    title={`Status: ${pipe.status}`}
                  />
                }
              />
            )
          })
        ) : (
          <div
            style={{
              padding: '8px 12px',
              fontSize: '0.8rem',
              color: theme.colors.ink.muted,
            }}
          >
            No recent threads
          </div>
        )}
      </SidebarContent>

      {/* Tools Section */}
      <SidebarFooter>
        <SectionLabel style={{ padding: '4px 12px' }}>Tools</SectionLabel>
        <NavItem
          icon={<Search size={14} />}
          label="Paper Search"
          onClick={onOpenSearch}
        />
        <NavItem
          variant="accent"
          icon={<Network size={14} />}
          label="Citation Graph"
          onClick={onOpenGraph}
        />
      </SidebarFooter>
    </SidebarRoot>
  )
}
