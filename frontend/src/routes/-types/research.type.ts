import type { ChatPaperSelectionData } from '#/types/pipeline.type'

export interface ReferenceItem {
  id: string
  num: number
  title: string
  authors: string
  year: number
  journal: string
  quartile: string
  citationsCount: number
  doi: string
  role: string
  keySnippet: string
  abstract: string
  x: number
  y: number
  radius: number
  color: string
}

export interface CitationEdge {
  from: string
  to: string
  relation: string
}

export interface ReportParagraph {
  text: string
  citations: Array<string>
}

export interface ReportSection {
  id: string
  heading: string
  paragraphs: Array<ReportParagraph>
}

export interface ComparisonRow {
  type: string
  conductivity: string
  stability: string
  primaryRef: string
}

export interface ChatCitationRef {
  id: string
  num: number
  title: string
  authors: string
  year: number
  journal: string
  doi?: string
}

export type ThinkingStepType =
  | 'thought'
  | 'search'
  | 'read_abstract'
  | 'read_fulltext'
  | 'references'
  | 'citations'
  | 'deconstruct'
  | 'audit'

export interface ThinkingStep {
  id: string
  type: ThinkingStepType
  label?: string
  description: string
  target?: string
  count?: number | string
  status: 'pending' | 'in_progress' | 'completed' | 'error'
  isExternalLink?: boolean
  timestamp?: number
}

export interface ResearchStatsSummary {
  retrieved: number | string
  eligible: number | string
  included: number | string
}

export interface ChatCheckpoint {
  type: 'ANGLE_SELECTION' | 'ANCHOR_SELECTION'
  angles?: Array<{
    id: string
    title: string
    description: string
    suggestedQueries?: Array<string>
  }>
  candidatePapers?: Array<{
    id: string
    title: string
    abstract?: string
    authors?: Array<{ name: string }>
    year?: number
    citationCount?: number
    doi?: string
  }>
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
  thinkingSteps?: Array<ThinkingStep>
  statsSummary?: ResearchStatsSummary
  isThinkingActive?: boolean
  thinkingDuration?: string
  steps?: Array<string>
  isLiveProgress?: boolean
  liveProgressMessage?: string
  checkpoint?: ChatCheckpoint
  paperSelection?: ChatPaperSelectionData
  citations?: Array<ChatCitationRef>
  status?: 'sending' | 'streaming' | 'done' | 'error'
  errorMessage?: string
  suggestedFollowUps?: Array<string>
  quote?: {
    text: string
    source: string
    paperId?: string
  }
}

export interface ResearchTopic {
  id: string
  shortName: string
  query: string
  aiIntro: string
  title: string
  subtitle: string
  steps: Array<string>
  thinkingSteps?: Array<ThinkingStep>
  statsSummary?: ResearchStatsSummary
  summary: string
  sections: Array<ReportSection>
  comparisonTable: Array<ComparisonRow>
  references: Array<ReferenceItem>
  edges: Array<CitationEdge>
}
