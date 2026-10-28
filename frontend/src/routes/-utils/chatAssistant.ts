import type {
  ChatCitationRef,
  ChatMessage,
  ResearchTopic,
  ThinkingStep,
} from '#/routes/-types/research.type'
import type { ResearchPipeline } from '#/types/pipeline.type'

/**
 * Creates the initial chat exchange for a topic or pipeline session.
 */
export function createInitialMessages(
  topic: ResearchTopic,
  pipeline?: ResearchPipeline | null,
): Array<ChatMessage> {
  const timestamp = new Date().toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })

  const citationRefs: Array<ChatCitationRef> = topic.references.map((r) => ({
    id: r.id,
    num: r.num,
    title: r.title,
    authors: r.authors,
    year: r.year,
    journal: r.journal,
    doi: r.doi,
  }))

  const initialUserMessage: ChatMessage = {
    id: `msg-user-init-${topic.id}`,
    role: 'user',
    content: topic.query,
    timestamp,
  }

  // Derive initial quote if available
  const quote =
    topic.references.length > 0
      ? {
          text: topic.references[0].keySnippet,
          source: `${topic.references[0].authors.split(',')[0]} et al. (${topic.references[0].year})`,
          paperId: topic.references[0].id,
        }
      : undefined

  // Derive checkpoint if pipeline is in waiting state
  let checkpoint: ChatMessage['checkpoint']
  if (pipeline && pipeline.status === 'WAITING_INPUT') {
    if (pipeline.checkpoint === 'ANGLE_SELECTION') {
      checkpoint = {
        type: 'ANGLE_SELECTION',
        angles: pipeline.stageData.decomposition?.angles,
      }
    } else if (pipeline.checkpoint === 'ANCHOR_SELECTION') {
      checkpoint = {
        type: 'ANCHOR_SELECTION',
        candidatePapers: pipeline.stageData.candidatePapers,
      }
    }
  }

  // Dynamic suggested follow-ups
  const suggestedFollowUps: Array<string> = [
    'Explain the ion transport mechanism steps in depth',
    'Compare material ionic conductivity in the comparison table',
    'What are the interfacial stability challenges and research limitations?',
    'Show verbatim empirical evidence citations from the main paper',
  ]

  const initialAssistantMessage: ChatMessage = {
    id: `msg-assistant-init-${topic.id}`,
    role: 'assistant',
    content: topic.summary,
    timestamp,
    steps: topic.steps,
    thinkingSteps: topic.thinkingSteps,
    statsSummary: topic.statsSummary,
    citations: citationRefs,
    quote,
    checkpoint,
    status: 'done',
    suggestedFollowUps,
  }

  return [initialUserMessage, initialAssistantMessage]
}

/**
 * Simulates intelligent, grounded assistant answering for follow-up questions
 * with smooth token-by-token streaming, rich Consensus-style thinking steps,
 * and custom state animations.
 */
export async function generateFollowUpAnswer(
  prompt: string,
  topic: ResearchTopic,
  onProgress?: (partial: Partial<ChatMessage>) => void,
): Promise<ChatMessage> {
  const timestamp = new Date().toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })

  const lower = prompt.toLowerCase()
  const citationRefs: Array<ChatCitationRef> = topic.references.map((r) => ({
    id: r.id,
    num: r.num,
    title: r.title,
    authors: r.authors,
    year: r.year,
    journal: r.journal,
    doi: r.doi,
  }))

  const stats = {
    retrieved: topic.references.length * 3 || 18,
    eligible: topic.references.length || 6,
    included: Math.min(3, topic.references.length),
  }

  // ─── STEP 1: INITIAL REASONING & THOUGHT STREAM ────────────────────────────
  const thinkingSteps: Array<ThinkingStep> = [
    {
      id: `step-thought-1-${Date.now()}`,
      type: 'thought',
      description: `I'll examine the inquiry regarding "${prompt}" against the active knowledge graph and verified journal corpus (${topic.title}).`,
      status: 'in_progress',
    },
  ]

  onProgress?.({
    thinkingSteps: [...thinkingSteps],
    statsSummary: stats,
    isThinkingActive: true,
    liveProgressMessage: 'Analyzing inquiry against ontology...',
    content: '',
    status: 'streaming',
  })

  await new Promise((r) => setTimeout(r, 260))

  // Complete thought 1, start federated search step
  thinkingSteps[0].status = 'completed'
  const searchStepId = `step-search-${Date.now()}`
  thinkingSteps.push({
    id: searchStepId,
    type: 'search',
    description: `Federated literature search for "${prompt.slice(0, 48)}..."`,
    count: `${topic.references.length * 2 || 12} ↗`,
    status: 'in_progress',
  })

  onProgress?.({
    thinkingSteps: [...thinkingSteps],
    liveProgressMessage: 'Querying semantic literature index...',
  })

  await new Promise((r) => setTimeout(r, 240))

  // Complete search step, add intermediate thought
  const searchStep = thinkingSteps.find((s) => s.id === searchStepId)
  if (searchStep) searchStep.status = 'completed'

  thinkingSteps.push({
    id: `step-thought-2-${Date.now()}`,
    type: 'thought',
    description: `Targeting empirical definitions and verbatim sections. Cross-referencing primary references for consistency.`,
    status: 'completed',
  })

  // Start read step
  const readStepId = `step-read-${Date.now()}`
  thinkingSteps.push({
    id: readStepId,
    type: 'read_fulltext',
    label: 'Read',
    target: 'Full Text & Verbatim Snippets',
    description: `Extracting section claims from ${topic.title.slice(0, 36)}...`,
    count: `${Math.min(3, topic.references.length)} >`,
    status: 'in_progress',
  })

  onProgress?.({
    thinkingSteps: [...thinkingSteps],
    liveProgressMessage: 'Reading full-text sections & quotes...',
  })

  await new Promise((r) => setTimeout(r, 220))

  // Complete read step, add references & audit step
  const readStep = thinkingSteps.find((s) => s.id === readStepId)
  if (readStep) readStep.status = 'completed'

  thinkingSteps.push({
    id: `step-ref-${Date.now()}`,
    type: 'references',
    label: 'References',
    description: `Backward citation lookup for theoretical foundations and baseline data`,
    count: `${topic.references.length * 3 || 15} ↗`,
    status: 'completed',
  })

  thinkingSteps.push({
    id: `step-audit-${Date.now()}`,
    type: 'audit',
    label: 'Grounding Audit',
    description: `Verifying extracted claims against verbatim text (zero hallucination)`,
    count: '100% Grounded',
    status: 'completed',
  })

  onProgress?.({
    thinkingSteps: [...thinkingSteps],
    isThinkingActive: false,
    liveProgressMessage: undefined,
  })

  await new Promise((r) => setTimeout(r, 120))

  // ─── STEP 2: COMPOSE FULL GROUNDED ANSWER ──────────────────────────────────
  let finalContent = ''
  let quoteText: string | undefined
  let quoteSource: string | undefined
  let quotePaperId: string | undefined
  const steps: Array<string> = [
    'Verifying inquiry against active research corpus ontology',
    'Tracing verbatim claims within selected document sections',
    'Verifying citation attribution and empirical evidence consistency',
  ]
  const suggestedFollowUps: Array<string> = []

  if (
    lower.includes('mekanisme') ||
    lower.includes('cara kerja') ||
    lower.includes('mechanism') ||
    lower.includes('ion')
  ) {
    const mechSec = topic.sections.find(
      (s) =>
        s.id === 'sec-mechanisms' ||
        s.heading.toLowerCase().includes('mechanism'),
    )

    finalContent = `Based on verbatim analysis of **${topic.title}**, material transport and interaction mechanisms proceed through several key stages:

1. **Interfacial Activation**: Ions undergo partial de-solvation at the diffusion boundary [1], where charge transfer occurs without liquid phase resistance.
2. **Percolation Tunneling**: Ion diffusion across planar crystal lattices leverages lattice vacancies with exceptionally low measured activation energies [2].
3. **Parasitic Reduction Prevention**: Interfacial kinetic barriers suppress dendrite formation and side-reaction degradation [3].`

    if (mechSec && mechSec.paragraphs.length > 0) {
      finalContent += `\n\n${mechSec.paragraphs.map((p) => p.text).join('\n\n')}`
    }

    if (topic.references.length > 0) {
      quoteText = topic.references[0].keySnippet
      quoteSource = `${topic.references[0].authors.split(',')[0]} et al. (${topic.references[0].year})`
      quotePaperId = topic.references[0].id
    }

    suggestedFollowUps.push(
      'What is the lowest reported activation energy value?',
      'How does temperature influence the degradation rate?',
      'Compare cycling performance with conventional liquid electrolytes',
    )
  } else if (
    lower.includes('banding') ||
    lower.includes('komparasi') ||
    lower.includes('tabel') ||
    lower.includes('perbedaan') ||
    lower.includes('compare')
  ) {
    finalContent = `Here is a summary comparing the performance and stability of validated materials from relevant Q1 literature:\n\n`
    if (topic.comparisonTable.length > 0) {
      finalContent += topic.comparisonTable
        .map(
          (row, idx) =>
            `- **${row.type}** [${(idx % topic.references.length) + 1}]: Conductivity performance of **${row.conductivity}** with electrochemical stability of **${row.stability}**.`,
        )
        .join('\n')
    } else {
      finalContent += `- **Oxide LLZO** [1]: Offers a wide electrochemical stability window (> 4.5 V vs Li/Li⁺) but exhibits high solid-solid contact resistance.\n- **Sulfide LGPS** [2]: Ultra-high ionic conductivity (> 10⁻² S/cm at room temperature) but sensitive to moisture and ambient degradation.`
    }

    finalContent += `\n\nOverall, hybrid composites (polymer-oxide blends) represent the latest research consensus to balance ionic conductivity with mechanical flexibility.`

    if (topic.references.length > 1) {
      quoteText = topic.references[1].keySnippet
      quoteSource = `${topic.references[1].authors.split(',')[0]} et al. (${topic.references[1].year})`
      quotePaperId = topic.references[1].id
    }

    suggestedFollowUps.push(
      'Why are sulfide electrolytes highly moisture sensitive?',
      'Which method is most effective at resolving grain-boundary resistance?',
      'Show electrochemical impedance spectroscopy (EIS) methodology details',
    )
  } else if (
    lower.includes('limitasi') ||
    lower.includes('kelemahan') ||
    lower.includes('tantangan') ||
    lower.includes('challenge') ||
    lower.includes('problem')
  ) {
    finalContent = `Despite showing significant advantages, critical analysis of core publications [1], [2], [3] highlights several fundamental technical limitations:

1. **Grain-Boundary Resistance**: Formation of resistive micro-phases at inter-particle contacts reduces effective conductivity by up to 40%.
2. **Critical Current Density (CCD)**: Under fast charging rates (> 3 mA/cm²), micro-dendrite penetration along grain boundaries can still trigger short circuits.
3. **Manufacturing Scalability Complexity**: High-temperature synthesis requirements (> 1000 °C) and moisture-free dry-room conditions increase mass production costs.`

    if (topic.references.length > 2) {
      quoteText = topic.references[2].keySnippet
      quoteSource = `${topic.references[2].authors.split(',')[0]} et al. (${topic.references[2].year})`
      quotePaperId = topic.references[2].id
    }

    suggestedFollowUps.push(
      'What strategies are used to increase critical current density?',
      'How does interface engineering (interlayers) suppress dendrite growth?',
      'Are there recent papers from 2024-2025 resolving the grain boundary issue?',
    )
  } else {
    const firstSection = topic.sections[0]
    finalContent = `Responding to research focus **"${prompt}"**:

Based on grounded evidence synthesis from Q1 journal corpus [1], [2]:

`
    if (topic.sections.length > 0 && firstSection.paragraphs.length > 0) {
      finalContent += firstSection.paragraphs
        .slice(0, 2)
        .map((p) => p.text)
        .join('\n\n')
    } else {
      finalContent += `This study confirms that microscopic architecture optimization and interface passivation play crucial roles in achieving long-term cycling stability. Key findings demonstrate a direct correlation between microstructural homogeneity and Coulombic efficiency exceeding 99.8%.`
    }

    finalContent += `\n\nFor a deeper review, you can click on reference citations or open the **Detail Panel** on the right to examine complete verbatim evidence.`

    if (topic.references.length > 0) {
      quoteText = topic.references[0].keySnippet
      quoteSource = `${topic.references[0].authors.split(',')[0]} et al. (${topic.references[0].year})`
      quotePaperId = topic.references[0].id
    }

    suggestedFollowUps.push(
      'Explain the technical mechanism steps in depth',
      'Compare material performance in the comparison table',
      'What are the reported methodological limitations?',
    )
  }

  // ─── STEP 3: STREAM MAIN CONTENT TOKEN-BY-TOKEN ────────────────────────────
  const quote = quoteText
    ? {
        text: quoteText,
        source: quoteSource ?? 'Academic Evidence',
        paperId: quotePaperId,
      }
    : undefined

  // Split into token / word chunks for smooth natural streaming
  const words = finalContent.split(' ')
  let streamedContent = ''
  const chunkSize = 2 // stream in 2-word bursts for smooth readable pace

  for (let i = 0; i < words.length; i += chunkSize) {
    const nextWords = words.slice(i, i + chunkSize).join(' ')
    streamedContent += (streamedContent ? ' ' : '') + nextWords

    onProgress?.({
      content: streamedContent,
      status: 'streaming',
      isThinkingActive: false,
    })

    // Natural streaming interval (approx 20ms)
    await new Promise((r) => setTimeout(r, 20))
  }

  // Return final completed ChatMessage
  return {
    id: `msg-ans-${Date.now()}`,
    role: 'assistant',
    content: finalContent,
    timestamp,
    steps,
    thinkingSteps,
    statsSummary: stats,
    citations: citationRefs,
    quote,
    status: 'done',
    isThinkingActive: false,
    suggestedFollowUps,
  }
}
