import { useState } from 'react'

import styled from '@emotion/styled'
import { BookOpen, FileText, Network, X } from 'lucide-react'

import { IconButton, Pill } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import { CitationGraphVisualizer } from '#/routes/-components/CitationGraphVisualizer'
import { DetailPanelReferencesTab } from '#/routes/-components/DetailPanelReferencesTab'
import { DetailPanelReportTab } from '#/routes/-components/DetailPanelReportTab'
import type {
  CitationEdge,
  ComparisonRow,
  ReferenceItem,
  ReportSection,
} from '#/routes/-types/research.type'

interface DetailPanelProps {
  title: string
  subtitle: string
  sections: Array<ReportSection>
  comparisonTable: Array<ComparisonRow>
  references: Array<ReferenceItem>
  edges: Array<CitationEdge>
  activeTab: 'report' | 'references' | 'graph'
  onChangeTab: (tab: 'report' | 'references' | 'graph') => void
  onClose: () => void
  isResizing?: boolean
  selectedRefId?: string
  onSelectRef?: (id: string) => void
  summary?: string
  reportMarkdown?: string
}

const DetailPanelContainer = styled.aside<{ $isResizing: boolean }>`
  flex: 1;
  min-width: 320px;
  background-color: ${theme.colors.canvas.primary};
  border-left: none;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  z-index: 20;
  font-family: ${theme.typography.fontFamily.sans};
  transition: ${({ $isResizing }) => ($isResizing ? 'none' : 'flex 0.2s ease')};
`

const PanelHeader = styled.div`
  height: 60px;
  border-bottom: ${theme.colors.border.subtle};
  padding: 0 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  background-color: ${theme.colors.canvas.primary};
  z-index: 10;
`

const TabsGroup = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const PanelBody = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 40px 56px 100px;
  display: flex;
  flex-direction: column;
  gap: 36px;
  background-color: ${theme.colors.canvas.primary};
`

const GraphWrapper = styled.div`
  max-width: 920px;
  margin: 0 auto;
  width: 100%;
`

export const DetailPanel = ({
  title,
  subtitle,
  sections,
  comparisonTable,
  references,
  edges,
  activeTab,
  onChangeTab,
  onClose,
  isResizing = false,
  selectedRefId: externalSelectedRefId,
  onSelectRef: externalOnSelectRef,
  summary,
  reportMarkdown,
}: DetailPanelProps) => {
  const [internalSelectedRefId, setInternalSelectedRefId] = useState<string>(
    references[0]?.id || '',
  )
  const selectedRefId = externalSelectedRefId ?? internalSelectedRefId

  const handleSelectRef = (id: string) => {
    setInternalSelectedRefId(id)
    externalOnSelectRef?.(id)
  }

  return (
    <DetailPanelContainer $isResizing={isResizing}>
      {/* Top Header Bar with Filter / Tab Pills */}
      <PanelHeader>
        <TabsGroup>
          <Pill
            variant={activeTab === 'report' ? 'active' : 'neutral'}
            size="md"
            icon={<FileText size={15} />}
            label="Synthesis Report"
            onClick={() => onChangeTab('report')}
          />
          <Pill
            variant={activeTab === 'references' ? 'active' : 'neutral'}
            size="md"
            icon={<BookOpen size={15} />}
            label={`Citations (${references.length})`}
            onClick={() => onChangeTab('references')}
          />
          <Pill
            variant={activeTab === 'graph' ? 'active' : 'neutral'}
            size="md"
            icon={<Network size={15} />}
            label="Citation Graph"
            onClick={() => onChangeTab('graph')}
          />
        </TabsGroup>

        {/* Close Button */}
        <IconButton
          size="icon-sm"
          variant="ghost"
          onClick={onClose}
          title="Close Detail Panel"
          aria-label="Close Detail Panel"
        >
          <X size={18} />
        </IconButton>
      </PanelHeader>

      {/* Main Panel Content Area */}
      <PanelBody>
        {activeTab === 'report' && (
          <DetailPanelReportTab
            title={title}
            subtitle={subtitle}
            summary={summary}
            reportMarkdown={reportMarkdown}
            sections={sections}
            comparisonTable={comparisonTable}
            references={references}
            onSelectRef={handleSelectRef}
            onChangeTab={onChangeTab}
          />
        )}

        {activeTab === 'references' && (
          <DetailPanelReferencesTab
            references={references}
            selectedRefId={selectedRefId}
            onSelectRef={handleSelectRef}
            onChangeTab={onChangeTab}
          />
        )}

        {activeTab === 'graph' && (
          <GraphWrapper>
            <CitationGraphVisualizer
              references={references}
              edges={edges}
              activeRefId={selectedRefId}
              onSelectRef={handleSelectRef}
            />
          </GraphWrapper>
        )}
      </PanelBody>
    </DetailPanelContainer>
  )
}
