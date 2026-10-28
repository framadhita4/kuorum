import type {
  ChatMessage,
  ResearchStatsSummary,
  ThinkingStep,
} from '#/routes/-types/research.type'

export type PipelineStage =
  | 'EXPLORATION'
  | 'RETRIEVAL'
  | 'INGESTION'
  | 'GRAPH'
  | 'EXTRACTION'
  | 'COMPLETED'

export type PipelineStatus =
  'DRAFT' | 'RUNNING' | 'WAITING_INPUT' | 'COMPLETED' | 'FAILED' | 'CANCELLED'

export type ExecutionMode = 'AUTONOMOUS' | 'INTERACTIVE'

export type ResearchApproach = 'NORMAL' | 'DEEP_RESEARCH'

export type CheckpointType = 'NONE' | 'ANGLE_SELECTION' | 'ANCHOR_SELECTION'

export interface PaperScoreInfo {
  paperId: string
  title: string
  totalPoints: number
  relevanceScore: number
  authorityScore: number
  recencyScore: number
  openAccessScore: number
  clefRelevanceScore?: number
  clefLabel?: string
  rank: number
  isSelected: boolean
  reason?: string
}

export interface ChatPaperSelectionData {
  approach: ResearchApproach
  totalRetrieved: number
  selectedCount: number
  selectedPapers: Array<PaperMetadata>
  scores?: Array<PaperScoreInfo>
}

export interface ResearchAngle {
  id: string
  title: string
  description: string
  suggestedQueries: Array<string>
}

export interface ConceptDecomposition {
  originalPrompt: string
  normalizedTerm: string
  angles: Array<ResearchAngle>
  createdAt: string
}

export interface PaperAuthor {
  id?: string
  name: string
}

export interface PaperMetadata {
  id: string
  title: string
  abstract: string
  authors: Array<PaperAuthor>
  year: number
  citationCount: number
  influentialCitationCount: number
  tldr?: string
  fieldsOfStudy?: Array<string>
  openAccessPdfUrl?: string
  doi?: string
  venue?: string
  fetchedAt?: string
  sources?: Array<string>
}

export interface GraphNode {
  paperId: string
  title: string
  year: number
  citationCount: number
  role:
    | 'ANCHOR'
    | 'SELECTED_STUDY'
    | 'FOUNDATIONAL_REFERENCE'
    | 'PRACTICAL_APPLICATION'
  inDegree: number
  outDegree: number
}

export interface GraphEdge {
  sourcePaperId: string
  targetPaperId: string
  intent?: string
}

export interface CitationGraph {
  anchorPaperId: string
  nodes: Array<GraphNode>
  edges: Array<GraphEdge>
  traversedAt: string
}

export interface CitationProof {
  paperId: string
  verbatimText: string
  section?: string
  pageNumber?: number
}

export interface GroundedDefinition {
  term: string
  definition: string
  citation: CitationProof
}

export interface GroundedMechanism {
  stepNumber: number
  description: string
  citation: CitationProof
}

export interface GroundedCaseStudy {
  industryDomain: string
  problemSolved: string
  methodApplied: string
  outcome: string
  citation: CitationProof
}

export interface GroundedConceptSynthesis {
  anchorPaperId: string
  concept: string
  definitions: Array<GroundedDefinition>
  mechanisms: Array<GroundedMechanism>
  caseStudies: Array<GroundedCaseStudy>
  generatedAt: string
}

export interface PipelineStageData {
  decomposition?: ConceptDecomposition
  candidatePapers?: Array<PaperMetadata>
  selectedPapers?: Array<PaperMetadata>
  paperScores?: Array<PaperScoreInfo>
  anchorPaper?: PaperMetadata
  graph?: CitationGraph
  synthesis?: GroundedConceptSynthesis
}

export interface ResearchPipeline {
  id: string
  title: string
  initialPrompt: string
  model?: string
  currentStage: PipelineStage
  status: PipelineStatus
  mode: ExecutionMode
  approach?: ResearchApproach
  checkpoint: CheckpointType
  selectedAngleId?: string
  anchorPaperId?: string
  selectedPaperIds?: Array<string>
  stageData: PipelineStageData
  messages?: Array<ChatMessage>
  pinned?: boolean
  isArchived?: boolean
  errorMessage?: string
  createdAt: string
  updatedAt: string
}

export type ChatHistory = ResearchPipeline

export interface PipelineEvent {
  pipelineId: string
  type:
    | 'pipeline_progress'
    | 'stage_completed'
    | 'awaiting_input'
    | 'pipeline_completed'
    | 'pipeline_failed'
  stage: PipelineStage
  message: string
  progress: number
  checkpoint?: CheckpointType
  data?: unknown
  step?: ThinkingStep
  stats?: ResearchStatsSummary
  timestamp: string
}
