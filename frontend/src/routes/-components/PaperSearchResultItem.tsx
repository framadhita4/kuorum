import styled from '@emotion/styled'
import { Check, ExternalLink, Play } from 'lucide-react'

import { Badge, Button } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { PaperMetadata } from '#/types/paper.type'

interface PaperSearchResultItemProps {
  paper: PaperMetadata
  onSelectPaper?: (paper: PaperMetadata) => void
  onStartPipelineWithPaper?: (paper: PaperMetadata) => void
}

const PaperCard = styled.div`
  padding: 16px 18px;
  border-radius: ${theme.radius.card};
  background-color: ${theme.colors.canvas.primary};
  border: 1px solid rgba(100, 111, 121, 0.16);
  box-shadow: ${theme.shadows.sm};
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: ${theme.transitions.fast};
  font-family: ${theme.typography.fontFamily.sans};
`

const PaperCardHeader = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
`

const PaperInfo = styled.div`
  flex: 1;
`

const PaperTitle = styled.h4`
  margin: 0;
  font-size: ${theme.typography.sizes.body};
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
  line-height: ${theme.typography.lineHeights.snug};
`

const PaperMetaRow = styled.div`
  font-size: 12px;
  color: ${theme.colors.ink.secondary};
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
`

const VenueText = styled.span`
  font-style: italic;
  color: ${theme.colors.accent.purpleDark};
`

const PaperBadges = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
`

const PaperAbstract = styled.p`
  margin: 0;
  font-size: 12px;
  color: ${theme.colors.ink.muted};
  line-height: ${theme.typography.lineHeights.normal};
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
`

const PaperActions = styled.div`
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px solid rgba(100, 111, 121, 0.08);
`

const OpenPdfLink = styled.a`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: ${theme.colors.ink.secondary};
  text-decoration: none;
  padding: 4px 8px;
  border-radius: ${theme.radius.sm};
  border: 1px solid rgba(100, 111, 121, 0.16);
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.ink.primary};
    border-color: ${theme.colors.ink.secondary};
  }
`

export const PaperSearchResultItem = ({
  paper,
  onSelectPaper,
  onStartPipelineWithPaper,
}: PaperSearchResultItemProps) => {
  const authorsText =
    paper.authors.length > 0
      ? paper.authors
          .slice(0, 3)
          .map((a) => a.name)
          .join(', ') +
        (paper.authors.length > 3 ? ` et al. (${paper.authors.length})` : '')
      : 'Unknown Authors'

  return (
    <PaperCard>
      <PaperCardHeader>
        <PaperInfo>
          <PaperTitle>{paper.title}</PaperTitle>
          <PaperMetaRow>
            <span>{authorsText}</span>
            <span>•</span>
            <span>{paper.year || 'Year N/A'}</span>
            {paper.venue && (
              <>
                <span>•</span>
                <VenueText>{paper.venue}</VenueText>
              </>
            )}
          </PaperMetaRow>
        </PaperInfo>

        <PaperBadges>
          {paper.citationCount > 0 && (
            <Badge variant="neutral" size="sm" isMono>
              {paper.citationCount} Citations
            </Badge>
          )}
          {paper.openAccessPdfUrl && (
            <Badge variant="success" size="sm">
              PDF OpenAccess
            </Badge>
          )}
        </PaperBadges>
      </PaperCardHeader>

      {/* Abstract preview */}
      {paper.abstract && <PaperAbstract>{paper.abstract}</PaperAbstract>}

      {/* Action buttons */}
      <PaperActions>
        {paper.openAccessPdfUrl && (
          <OpenPdfLink
            href={paper.openAccessPdfUrl}
            target="_blank"
            rel="noreferrer"
          >
            <ExternalLink size={11} />
            Open PDF
          </OpenPdfLink>
        )}

        {onSelectPaper && (
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onSelectPaper(paper)}
            startIcon={<Check size={12} />}
          >
            Select Paper
          </Button>
        )}

        {onStartPipelineWithPaper && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => onStartPipelineWithPaper(paper)}
            startIcon={<Play size={12} />}
          >
            Start Autonomous Research
          </Button>
        )}
      </PaperActions>
    </PaperCard>
  )
}
