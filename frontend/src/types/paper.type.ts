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
  influentialCitationCount?: number
  tldr?: string
  fieldsOfStudy?: Array<string>
  openAccessPdfUrl?: string
  doi?: string
  venue?: string
  fetchedAt?: string
  sources?: Array<string>
}

export interface SearchPapersParams {
  query: string
  limit?: number
  providers?: Array<string>
}

export interface SearchPapersResult {
  query: string
  total: number
  providers?: Array<string>
  papers: Array<PaperMetadata>
}

export interface PaperProvidersResult {
  providers: Array<string>
}

export interface PaperReferencesResult {
  paperId: string
  total: number
  references: Array<PaperMetadata>
}

export interface PaperCitationsResult {
  paperId: string
  total: number
  citations: Array<PaperMetadata>
}
