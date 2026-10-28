import type { ResearchTopic } from '#/routes/-types/research.type'

export const NEW_RESEARCH_TOPIC: ResearchTopic = {
  id: 'new',
  shortName: 'New Research',
  query: '',
  aiIntro: '',
  title: 'Start New Scientific Research',
  subtitle:
    'Literature Retrieval • Concept Decomposition • Citation Graph Traversal',
  steps: [],
  summary: '',
  sections: [],
  comparisonTable: [],
  references: [],
  edges: [],
}

export const RESEARCH_TOPICS: Record<string, ResearchTopic> = {}
