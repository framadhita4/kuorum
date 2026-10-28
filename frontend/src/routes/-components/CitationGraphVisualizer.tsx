import { useRef, useState } from 'react'

import styled from '@emotion/styled'
import { RotateCcw } from 'lucide-react'

import { StatusDot } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { CitationEdge, ReferenceItem } from '#/routes/-types/research.type'

interface CitationGraphVisualizerProps {
  references: Array<ReferenceItem>
  edges: Array<CitationEdge>
  activeRefId: string
  onSelectRef: (id: string) => void
}

const GraphContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
`

const GraphToolbar = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.78rem;
  color: ${theme.colors.ink.secondary};
  font-family: ${theme.typography.fontFamily.sans};
`

const ToolbarLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
`

const LegendContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.72rem;
`

const LegendItem = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 4px;
`

const ResetButton = styled.button`
  display: flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: none;
  color: ${theme.colors.ink.primary};
  font-size: 0.78rem;
  font-weight: ${theme.typography.weights.medium};
  cursor: pointer;
  padding: 4px 8px;
  border-radius: ${theme.radius.sm};
  outline: none;
  transition: color 0.15s ease;

  &:hover {
    color: ${theme.colors.brand.coral};
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 1px;
  }
`

const SvgCanvasContainer = styled.div`
  background-color: ${theme.colors.canvas.subtle};
  border-radius: ${theme.radius.card};
  border: ${theme.colors.border.subtle};
  overflow: hidden;
  user-select: none;
  box-shadow: ${theme.shadows.sm};
`

const StyledSvg = styled.svg<{ $isDragging: boolean }>`
  width: 100%;
  height: 420px;
  cursor: ${({ $isDragging }) => ($isDragging ? 'grabbing' : 'grab')};
  display: block;
`

const ConnectionsCard = styled.div`
  display: flex;
  flex-direction: column;
  gap: 10px;
  background-color: ${theme.colors.canvas.primary};
  border-radius: ${theme.radius.card};
  padding: 16px 20px;
  border: ${theme.colors.border.subtle};
  box-shadow: ${theme.shadows.sm};
  font-family: ${theme.typography.fontFamily.sans};
`

const ConnectionsTitle = styled.span`
  font-size: 0.8rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
  display: block;
`

const ConnectionsList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`

const ConnectionRow = styled.div`
  font-size: 0.82rem;
  color: ${theme.colors.ink.secondary};
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-weight: ${theme.typography.weights.light};
`

export const CitationGraphVisualizer = ({
  references,
  edges,
  activeRefId,
  onSelectRef,
}: CitationGraphVisualizerProps) => {
  const [customPositions, setCustomPositions] = useState<
    Record<string, { x: number; y: number } | undefined>
  >({})
  const [draggingId, setDraggingId] = useState<string | null>(null)
  const svgRef = useRef<SVGSVGElement | null>(null)

  const inspectedRef: ReferenceItem | null =
    references.find((r) => r.id === activeRefId) || references[0]

  const getNodeCoords = (ref: ReferenceItem) => {
    return customPositions[ref.id] || { x: ref.x, y: ref.y }
  }

  const handleMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
    if (!draggingId || !svgRef.current) return
    const rect = svgRef.current.getBoundingClientRect()
    const scaleX = 500 / rect.width
    const scaleY = 340 / rect.height
    const newX = Math.max(45, Math.min(455, (e.clientX - rect.left) * scaleX))
    const newY = Math.max(40, Math.min(300, (e.clientY - rect.top) * scaleY))
    setCustomPositions((prev) => ({
      ...prev,
      [draggingId]: { x: newX, y: newY },
    }))
  }

  return (
    <GraphContainer>
      {/* Graph Toolbar */}
      <GraphToolbar>
        <ToolbarLeft>
          <span>Click a node to select • Drag a node to tidy up</span>
          <LegendContainer>
            <LegendItem>
              <StatusDot variant="warning" size={8} />
              Anchor
            </LegendItem>
            <LegendItem>
              <StatusDot variant="purple" size={8} />
              Deep Study
            </LegendItem>
            <LegendItem>
              <StatusDot variant="emerald" size={8} />
              Foundational
            </LegendItem>
            <LegendItem>
              <StatusDot color="#3B82F6" size={8} />
              Application
            </LegendItem>
          </LegendContainer>
        </ToolbarLeft>
        <ResetButton type="button" onClick={() => setCustomPositions({})}>
          <RotateCcw size={13} />
          <span>Reset Layout</span>
        </ResetButton>
      </GraphToolbar>

      {/* SVG Canvas */}
      <SvgCanvasContainer>
        <StyledSvg
          ref={svgRef}
          viewBox="0 0 500 340"
          $isDragging={Boolean(draggingId)}
          onMouseMove={handleMouseMove}
          onMouseUp={() => setDraggingId(null)}
          onMouseLeave={() => setDraggingId(null)}
        >
          <defs>
            <marker
              id="arrow-active"
              viewBox="0 0 10 10"
              refX="25"
              refY="5"
              markerWidth="6"
              markerHeight="6"
              orient="auto-start-reverse"
            >
              <path
                d="M 0 1 L 10 5 L 0 9 z"
                fill={theme.colors.accent.purple}
              />
            </marker>
            <marker
              id="arrow-muted"
              viewBox="0 0 10 10"
              refX="25"
              refY="5"
              markerWidth="5"
              markerHeight="5"
              orient="auto-start-reverse"
            >
              <path d="M 0 1 L 10 5 L 0 9 z" fill="#CBD5E1" />
            </marker>
          </defs>

          {/* Edges */}
          {edges.map((edge, idx) => {
            const sRef = references.find((r) => r.id === edge.from)
            const tRef = references.find((r) => r.id === edge.to)
            if (!sRef || !tRef) return null
            const s = getNodeCoords(sRef)
            const t = getNodeCoords(tRef)
            const isActive =
              edge.from === inspectedRef.id || edge.to === inspectedRef.id

            return (
              <line
                key={idx}
                x1={s.x}
                y1={s.y}
                x2={t.x}
                y2={t.y}
                stroke={
                  isActive
                    ? theme.colors.accent.purple
                    : 'rgba(100, 111, 121, 0.25)'
                }
                strokeWidth={isActive ? '2' : '1.2'}
                strokeDasharray={isActive ? 'none' : '4 4'}
                markerEnd={
                  isActive ? 'url(#arrow-active)' : 'url(#arrow-muted)'
                }
              />
            )
          })}

          {/* Nodes */}
          {references.map((ref) => {
            const coord = getNodeCoords(ref)
            const isSelected = inspectedRef.id === ref.id

            return (
              <g
                key={ref.id}
                transform={`translate(${coord.x}, ${coord.y})`}
                onMouseDown={(e) => {
                  e.stopPropagation()
                  onSelectRef(ref.id)
                  setDraggingId(ref.id)
                }}
                style={{ cursor: 'pointer' }}
              >
                {isSelected && (
                  <circle
                    r={ref.radius + 6}
                    fill={theme.colors.accent.purple}
                    opacity="0.18"
                  />
                )}

                {/* Node Circle */}
                <circle
                  r={ref.radius}
                  fill={
                    isSelected
                      ? theme.colors.accent.purple
                      : ref.color
                        ? `${ref.color}18`
                        : '#FFFFFF'
                  }
                  stroke={
                    isSelected
                      ? theme.colors.accent.purpleDark
                      : ref.color || theme.colors.ink.primary
                  }
                  strokeWidth={isSelected ? '2.5' : '2'}
                />

                {/* Number Badge Text */}
                <text
                  textAnchor="middle"
                  dy="4"
                  fill={isSelected ? '#FFFFFF' : theme.colors.ink.primary}
                  fontSize="11"
                  fontWeight="600"
                  fontFamily={theme.typography.fontFamily.mono}
                >
                  [{ref.num}]
                </text>

                {/* Author & Year Subtitle */}
                <text
                  textAnchor="middle"
                  dy={ref.radius + 15}
                  fill={
                    isSelected
                      ? theme.colors.accent.purple
                      : theme.colors.ink.secondary
                  }
                  fontSize="9.5"
                  fontWeight={
                    isSelected
                      ? theme.typography.weights.medium
                      : theme.typography.weights.regular
                  }
                  fontFamily={theme.typography.fontFamily.sans}
                >
                  {ref.authors.split(',')[0]} ({ref.year})
                </text>
              </g>
            )
          })}
        </StyledSvg>
      </SvgCanvasContainer>

      {/* Active Node Relations Breakdown */}
      <ConnectionsCard>
        <ConnectionsTitle>
          Citation Connections [{inspectedRef.num}]{' '}
          {inspectedRef.authors.split(',')[0]}:
        </ConnectionsTitle>

        <ConnectionsList>
          {edges
            .filter(
              (e) => e.from === inspectedRef.id || e.to === inspectedRef.id,
            )
            .map((e, i) => {
              const isOut = e.from === inspectedRef.id
              const other = references.find(
                (r) => r.id === (isOut ? e.to : e.from),
              )
              if (!other) return null
              return (
                <ConnectionRow key={i}>
                  <StatusDot variant="purple" size={6} />
                  <span>
                    {isOut ? 'Cites' : 'Cited by'}{' '}
                    <strong
                      style={{
                        color: theme.colors.ink.primary,
                        fontWeight: theme.typography.weights.medium,
                      }}
                    >
                      [{other.num}] {other.authors.split(',')[0]}
                    </strong>{' '}
                    — {e.relation}
                  </span>
                </ConnectionRow>
              )
            })}
        </ConnectionsList>
      </ConnectionsCard>
    </GraphContainer>
  )
}
