export interface ResearchAngle {
  angleId: string
  title: string
  description: string
  suggestedQueries: Array<string>
}

export interface DecomposeConceptResult {
  prompt: string
  normalizedTerm: string
  angles: Array<ResearchAngle>
}

export interface SystemHealth {
  service: string
  status: string
  timestamp: string
  uptime: string
}

export interface ReformulateQueryPayload {
  prompt: string
  model?: string
}

export interface ReformulateQueryResult {
  originalPrompt: string
  keywords: Array<string>
  synonyms: Array<string>
  booleanQuery: string
  denseQuery: string
}

export interface AngleSummaryItem {
  id: string
  title: string
  description: string
}

export interface GetAnglesResult {
  concept: string
  angles: Array<AngleSummaryItem>
}
