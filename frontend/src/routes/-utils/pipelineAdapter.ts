import type {
  CitationEdge,
  ReferenceItem,
  ReportSection,
  ResearchTopic,
} from '#/routes/-types/research.type'
import type { ResearchPipeline } from '#/types/pipeline.type'

export function mapPipelineToTopic(
  pipeline: ResearchPipeline | null | undefined,
): ResearchTopic | null {
  if (!pipeline) return null

  const data = pipeline.stageData
  const synth = data.synthesis
  const graph = data.graph
  const anchor = data.anchorPaper

  // 1. Map Grounded Synthesis to Report Sections
  const sections: Array<ReportSection> = []

  if (synth?.definitions && synth.definitions.length > 0) {
    sections.push({
      id: 'sec-definitions',
      heading: '1. Formal Concepts & Domain Definitions',
      paragraphs: synth.definitions.map((d) => ({
        text: `**${d.term}**: ${d.definition} (Evidence: "${d.citation.verbatimText}")`,
        citations: [d.citation.paperId || 'anchor'],
      })),
    })
  }

  if (synth?.mechanisms && synth.mechanisms.length > 0) {
    sections.push({
      id: 'sec-mechanisms',
      heading: '2. Technical Architecture & Step-by-Step Mechanism',
      paragraphs: synth.mechanisms.map((m) => ({
        text: `**Step ${m.stepNumber}**: ${m.description} (Evidence: "${m.citation.verbatimText}")`,
        citations: [m.citation.paperId || 'anchor'],
      })),
    })
  }

  if (synth?.caseStudies && synth.caseStudies.length > 0) {
    sections.push({
      id: 'sec-cases',
      heading: '3. Empirical Outcomes & Industry Applications',
      paragraphs: synth.caseStudies.map((c) => ({
        text: `**${c.industryDomain}** — Solved: ${c.problemSolved}. Method: ${c.methodApplied}. Outcome: ${c.outcome}`,
        citations: [c.citation.paperId || 'anchor'],
      })),
    })
  }

  // 2. Map Explored Citation Network & Selected Literature to ReferenceItems
  type PaperMeta = NonNullable<typeof data.candidatePapers>[0]
  const paperMetadataMap = new Map<string, PaperMeta>()
  if (data.candidatePapers) {
    data.candidatePapers.forEach((p) => paperMetadataMap.set(p.id, p))
  }
  if (data.selectedPapers) {
    data.selectedPapers.forEach((p) => paperMetadataMap.set(p.id, p))
  }
  if (data.anchorPaper) {
    paperMetadataMap.set(data.anchorPaper.id, data.anchorPaper)
  }

  const references: Array<ReferenceItem> = []
  const addedRefIDs = new Set<string>()

  // A. When citation graph has been traversed, map the connected citation topology
  if (graph?.nodes && graph.nodes.length > 0) {
    const anchorNode = graph.nodes.find((n) => n.role === 'ANCHOR')
    const foundationalNodes = graph.nodes.filter(
      (n) => n.role === 'FOUNDATIONAL_REFERENCE',
    )
    const cohortNodes = graph.nodes.filter((n) => n.role === 'SELECTED_STUDY')
    const applicationNodes = graph.nodes.filter(
      (n) => n.role === 'PRACTICAL_APPLICATION',
    )

    let itemNum = 1

    // 1. Primary Anchor Node (Center)
    if (anchorNode) {
      const meta = paperMetadataMap.get(anchorNode.paperId)
      references.push({
        id: anchorNode.paperId,
        num: itemNum++,
        title: meta?.title || anchorNode.title,
        authors: meta?.authors.length
          ? meta.authors.map((a) => a.name).join(', ')
          : 'Primary Authors',
        year: meta?.year || anchorNode.year || 2023,
        journal: meta?.venue || 'Top-tier Journal / Venue',
        quartile: 'Q1 • Anchor Paper',
        citationsCount: meta?.citationCount ?? anchorNode.citationCount,
        doi: meta?.doi || anchorNode.paperId,
        role: 'Anchor Paper',
        keySnippet: meta?.abstract
          ? meta.abstract.slice(0, 180) + '...'
          : 'Primary anchor paper grounding the research synthesis.',
        abstract: meta?.abstract || '',
        x: 250,
        y: 170,
        radius: 30,
        color: '#F59E0B',
      })
      addedRefIDs.add(anchorNode.paperId)
    }

    // 2. Selected Cohort Studies (Arranged around center)
    cohortNodes.forEach((node, idx) => {
      if (addedRefIDs.has(node.paperId)) return
      const meta = paperMetadataMap.get(node.paperId)
      const angle =
        (idx / Math.max(1, cohortNodes.length)) * 2 * Math.PI - Math.PI / 2
      const cx = 250 + Math.cos(angle) * 115
      const cy = 170 + Math.sin(angle) * 75

      references.push({
        id: node.paperId,
        num: itemNum++,
        title: meta?.title || node.title,
        authors: meta?.authors.length
          ? meta.authors.map((a) => a.name).join(', ')
          : 'Cohort Authors',
        year: meta?.year || node.year || 2023,
        journal:
          meta?.venue ||
          (meta?.openAccessPdfUrl
            ? 'arXiv OpenAccess'
            : 'Peer-Reviewed Conference'),
        quartile: 'Selected Deep Study',
        citationsCount: meta?.citationCount ?? node.citationCount,
        doi: meta?.doi || node.paperId,
        role: 'Selected Study',
        keySnippet: meta?.abstract
          ? meta.abstract.slice(0, 180) + '...'
          : 'Selected relevant study analyzed in depth.',
        abstract: meta?.abstract || '',
        x: Math.round(cx),
        y: Math.round(cy),
        radius: 24,
        color: '#796EFF',
      })
      addedRefIDs.add(node.paperId)
    })

    // 3. Foundational References (Left Column)
    foundationalNodes.forEach((node, idx) => {
      if (addedRefIDs.has(node.paperId)) return
      const meta = paperMetadataMap.get(node.paperId)
      const count = foundationalNodes.length
      const stepY = count > 1 ? 260 / (count + 1) : 130
      const ny = 40 + (idx + 1) * stepY

      references.push({
        id: node.paperId,
        num: itemNum++,
        title: meta?.title || node.title,
        authors: meta?.authors.length
          ? meta.authors.map((a) => a.name).join(', ')
          : 'Seminal Authors',
        year: meta?.year || node.year || 2020,
        journal: meta?.venue || 'Foundational Venue',
        quartile: 'Foundational Baseline',
        citationsCount: meta?.citationCount ?? node.citationCount,
        doi: meta?.doi || node.paperId,
        role: 'Foundational Reference',
        keySnippet: meta?.abstract
          ? meta.abstract.slice(0, 180) + '...'
          : 'Foundational theoretical baseline referenced by the anchor.',
        abstract: meta?.abstract || '',
        x: 80,
        y: Math.round(ny),
        radius: 20,
        color: '#0D9488',
      })
      addedRefIDs.add(node.paperId)
    })

    // 4. Practical Applications (Right Column)
    applicationNodes.forEach((node, idx) => {
      if (addedRefIDs.has(node.paperId)) return
      const meta = paperMetadataMap.get(node.paperId)
      const count = applicationNodes.length
      const stepY = count > 1 ? 260 / (count + 1) : 130
      const ny = 40 + (idx + 1) * stepY

      references.push({
        id: node.paperId,
        num: itemNum++,
        title: meta?.title || node.title,
        authors: meta?.authors.length
          ? meta.authors.map((a) => a.name).join(', ')
          : 'Applied Authors',
        year: meta?.year || node.year || 2024,
        journal: meta?.venue || 'Industrial Application Venue',
        quartile: 'Downstream Extension',
        citationsCount: meta?.citationCount ?? node.citationCount,
        doi: meta?.doi || node.paperId,
        role: 'Practical Application',
        keySnippet: meta?.abstract
          ? meta.abstract.slice(0, 180) + '...'
          : 'Recent empirical extension or practical deployment.',
        abstract: meta?.abstract || '',
        x: 420,
        y: Math.round(ny),
        radius: 20,
        color: '#3B82F6',
      })
      addedRefIDs.add(node.paperId)
    })
  } else {
    // B. Prior to graph resolution, only display the curated cohort (avoid dumping 150 search results)
    const papersToDisplay =
      data.selectedPapers && data.selectedPapers.length > 0
        ? data.selectedPapers
        : data.candidatePapers && data.candidatePapers.length > 0
          ? data.candidatePapers.slice(0, 5)
          : []

    papersToDisplay.forEach((p, idx) => {
      const isAnchor = p.id === pipeline.anchorPaperId
      references.push({
        id: p.id,
        num: idx + 1,
        title: p.title,
        authors:
          p.authors.length > 0
            ? p.authors.map((a) => a.name).join(', ')
            : 'Unknown Authors',
        year: p.year || 2023,
        journal:
          p.venue ||
          (p.openAccessPdfUrl ? 'OpenAccess arXiv' : 'Academic Venue'),
        quartile: isAnchor ? 'Q1 • Anchor' : 'Peer-Reviewed',
        citationsCount: p.citationCount || 0,
        doi: p.doi || p.id,
        role: isAnchor ? 'Anchor Paper' : 'Selected Candidate',
        keySnippet: p.abstract
          ? p.abstract.slice(0, 180) + '...'
          : 'No abstract preview available.',
        abstract: p.abstract || '',
        x: 120 + (idx % 3) * 130,
        y: 110 + Math.floor(idx / 3) * 110,
        radius: isAnchor ? 28 : 22,
        color: isAnchor ? '#F59E0B' : '#796EFF',
      })
    })
  }

  // 3. Map Citation Graph Edges
  const edges: Array<CitationEdge> = []
  if (graph?.edges && graph.edges.length > 0) {
    graph.edges.forEach((e) => {
      edges.push({
        from: e.sourcePaperId,
        to: e.targetPaperId,
        relation: e.intent || 'cites',
      })
    })
  }

  // Dynamic Pipeline Steps Status
  const stageLabels: Record<string, string> = {
    EXPLORATION: '1. Academic Concept Decomposition',
    RETRIEVAL: '2. Federated Literature Search & Ranking',
    INGESTION: '3. PDF Sectioning & Layout Recovery',
    GRAPH: '4. Citation Topology Traversal',
    EXTRACTION: '5. Verbatim Grounded Synthesis',
  }

  const steps = [
    `Stage: ${stageLabels[pipeline.currentStage] ?? pipeline.currentStage}`,
    `Status: ${pipeline.status}${pipeline.checkpoint !== 'NONE' ? ` (Waiting for: ${pipeline.checkpoint})` : ''}`,
    `Mode: ${pipeline.mode}`,
    anchor
      ? `Anchor Paper: ${anchor.title.slice(0, 50)}...`
      : 'Selecting Anchor Paper...',
  ]

  return {
    id: pipeline.id,
    shortName: pipeline.title.slice(0, 24),
    query: pipeline.initialPrompt,
    aiIntro: `Research pipeline session (${pipeline.mode} mode) at stage: ${pipeline.currentStage}.`,
    title: pipeline.title,
    subtitle: `Session ID: ${pipeline.id} • ${pipeline.status}`,
    steps,
    summary:
      synth?.definitions[0]?.definition ??
      (data.decomposition?.angles[0]?.description ||
        'Research pipeline is actively executing cognitive literature discovery.'),
    sections: sections.length > 0 ? sections : [],
    comparisonTable: [],
    references,
    edges,
  }
}
