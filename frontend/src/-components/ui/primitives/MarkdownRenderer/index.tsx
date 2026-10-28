import type { ReactNode } from 'react'
import {
  Children,
  cloneElement,
  isValidElement,
  memo,
  useCallback,
  useMemo,
  useState,
} from 'react'

import { keyframes } from '@emotion/react'
import styled from '@emotion/styled'
import { Check, Copy, ExternalLink } from 'lucide-react'
import remarkGfm from 'remark-gfm'
import Markdown from 'react-markdown'

import { theme } from '../../theme'
import { CitationPill } from '../CitationPill'

export interface MarkdownReference {
  id: string
  num: number
  title?: string
  authors?: string
  year?: string | number
}

export interface MarkdownRendererProps {
  content: string
  references?: Array<MarkdownReference>
  onSelectCitation?: (citationId: string) => void
  isStreaming?: boolean
  className?: string
}

// ─── ANIMATIONS ─────────────────────────────────────────────────────────────

const blinkKeyframe = keyframes`
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.2;
  }
`

// ─── ROOT CONTAINER ─────────────────────────────────────────────────────────

const RootProse = styled.div`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 1.02rem;
  font-weight: ${theme.typography.weights.light};
  color: ${theme.colors.ink.primary};
  line-height: 1.75;
  max-width: 68ch;
  word-break: break-word;

  /* Smooth font rendering */
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
`

// ─── TYPOGRAPHY ELEMENTS ───────────────────────────────────────────────────

const MarkdownH1 = styled.h1`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h2}; /* 36px / 2.25rem */
  font-weight: ${theme.typography.weights.light}; /* 300 Light per design rule */
  letter-spacing: ${theme.typography.letterSpacing.tightest}; /* -0.03em */
  line-height: ${theme.typography.lineHeights.tight};
  color: ${theme.colors.ink.primary};
  margin: 32px 0 14px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownH2 = styled.h2`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h3}; /* 30px / 1.875rem */
  font-weight: ${theme.typography.weights.light}; /* 300 Light per design rule */
  letter-spacing: ${theme.typography.letterSpacing.tight}; /* -0.02em */
  line-height: ${theme.typography.lineHeights.snug};
  color: ${theme.colors.ink.primary};
  margin: 28px 0 12px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownH3 = styled.h3`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h4}; /* 24px / 1.5rem */
  font-weight: ${theme.typography.weights.regular}; /* 400 Regular */
  letter-spacing: -0.015em;
  line-height: 1.3;
  color: ${theme.colors.ink.primary};
  margin: 22px 0 10px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownH4 = styled.h4`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h5}; /* 20px / 1.25rem */
  font-weight: ${theme.typography.weights.medium}; /* 500 Medium */
  letter-spacing: -0.01em;
  line-height: 1.35;
  color: ${theme.colors.ink.primary};
  margin: 18px 0 8px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownH5 = styled.h5`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 1.05rem;
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
  margin: 16px 0 6px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownH6 = styled.h6`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.xs}; /* 12px */
  font-weight: ${theme.typography.weights.medium};
  text-transform: uppercase;
  letter-spacing: ${theme.typography.letterSpacing.widest}; /* 0.08em */
  color: ${theme.colors.ink.muted};
  margin: 14px 0 4px;

  &:first-of-type {
    margin-top: 0;
  }
`

const MarkdownParagraph = styled.p`
  margin: 0 0 14px 0;
  line-height: 1.75;

  &:last-child {
    margin-bottom: 0;
  }
`

const MarkdownStrong = styled.strong`
  font-weight: ${theme.typography.weights.semibold};
  color: ${theme.colors.ink.primary};
`

const MarkdownEm = styled.em`
  font-style: italic;
  color: ${theme.colors.ink.secondary};
`

const MarkdownDel = styled.del`
  text-decoration: line-through;
  color: ${theme.colors.ink.muted};
`

// ─── LISTS ──────────────────────────────────────────────────────────────────

const MarkdownUl = styled.ul`
  margin: 8px 0 16px;
  padding-left: 22px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  list-style-type: disc;

  & & {
    margin: 4px 0 4px;
  }

  &::marker {
    color: ${theme.colors.accent.purple};
  }
`

const MarkdownOl = styled.ol`
  margin: 8px 0 16px;
  padding-left: 22px;
  display: flex;
  flex-direction: column;
  gap: 6px;

  & & {
    margin: 4px 0 4px;
  }

  &::marker {
    color: ${theme.colors.ink.muted};
    font-weight: ${theme.typography.weights.medium};
    font-family: ${theme.typography.fontFamily.mono};
    font-size: 0.9em;
  }
`

const MarkdownLi = styled.li`
  line-height: 1.7;
  font-weight: ${theme.typography.weights.light};
  color: ${theme.colors.ink.primary};

  & > p {
    margin: 0;
  }
`

// ─── BLOCKQUOTE ─────────────────────────────────────────────────────────────

const MarkdownBlockquote = styled.blockquote`
  margin: 16px 0;
  padding: 12px 18px;
  background-color: ${theme.colors.canvas.subtle};
  border-left: 3px solid ${theme.colors.brand.coral};
  border-radius: 0 ${theme.radius.md} ${theme.radius.md} 0;
  color: ${theme.colors.ink.secondary};
  font-size: 0.95rem;
  font-style: italic;
  line-height: 1.65;

  & > p {
    margin: 0;
    font-style: italic;
  }
`

// ─── LINKS ──────────────────────────────────────────────────────────────────

const MarkdownLink = styled.a`
  color: ${theme.colors.accent.purple};
  text-decoration: underline;
  text-underline-offset: 3px;
  text-decoration-color: rgba(121, 110, 255, 0.4);
  font-weight: ${theme.typography.weights.medium};
  display: inline-inline;
  align-items: center;
  gap: 3px;
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.brand.coral};
    text-decoration-color: ${theme.colors.brand.coral};
  }
`

const LinkIcon = styled(ExternalLink)`
  display: inline-block;
  margin-left: 3px;
  vertical-align: middle;
  opacity: 0.7;
`

// ─── CODE ELEMENTS ──────────────────────────────────────────────────────────

const InlineCode = styled.code`
  font-family: ${theme.typography.fontFamily.mono};
  font-size: 0.86em;
  font-weight: 500;
  padding: 2px 6px;
  border-radius: ${theme.radius.sm};
  border: 1px solid rgba(100, 111, 121, 0.2);
  background-color: ${theme.colors.canvas.neutral};
  color: ${theme.colors.brand.claret};
  white-space: pre-wrap;
  word-break: break-all;
`

const CodeBlockContainer = styled.div`
  margin: 18px 0;
  background-color: ${theme.colors.canvas.charcoal};
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: ${theme.radius.card};
  overflow: hidden;
  box-shadow: ${theme.shadows.md};
`

const CodeBlockHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: rgba(0, 0, 0, 0.25);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
`

const CodeLanguageLabel = styled.span`
  font-family: ${theme.typography.fontFamily.mono};
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: ${theme.colors.accent.lilac};
  user-select: none;
`

const CopyButton = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.75);
  font-family: ${theme.typography.fontFamily.sans};
  font-size: 0.75rem;
  font-weight: ${theme.typography.weights.medium};
  cursor: pointer;
  padding: 4px 8px;
  border-radius: ${theme.radius.sm};
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.brand.coral};
    background: rgba(255, 255, 255, 0.06);
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 1px;
  }
`

const CodePre = styled.pre`
  margin: 0;
  padding: 14px 18px;
  overflow-x: auto;
  font-family: ${theme.typography.fontFamily.mono};
  font-size: 0.85rem;
  line-height: 1.6;
  color: #f5f5f7;
  background-color: transparent;

  & > code {
    background: none;
    border: none;
    padding: 0;
    font-size: inherit;
    color: inherit;
    font-family: inherit;
    white-space: pre;
    word-break: normal;
  }

  /* Sleek scrollbar */
  &::-webkit-scrollbar {
    height: 6px;
  }
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: ${theme.radius.pill};
  }
  &::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.35);
  }
`

// ─── TABLES ─────────────────────────────────────────────────────────────────

const TableWrapper = styled.div`
  margin: 18px 0;
  width: 100%;
  overflow-x: auto;
  border: ${theme.colors.border.subtle};
  border-radius: ${theme.radius.card};
  box-shadow: ${theme.shadows.sm};
  background-color: ${theme.colors.canvas.primary};
`

const StyledTable = styled.table`
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  text-align: left;
`

const StyledTh = styled.th`
  background-color: ${theme.colors.canvas.subtle};
  color: ${theme.colors.ink.secondary};
  font-size: 0.78rem;
  font-weight: ${theme.typography.weights.semibold};
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 12px 16px;
  border-bottom: ${theme.colors.border.subtle};
  white-space: nowrap;
`

const StyledTd = styled.td`
  padding: 12px 16px;
  font-size: 0.92rem;
  font-weight: ${theme.typography.weights.light};
  color: ${theme.colors.ink.primary};
  border-bottom: 1px solid rgba(100, 111, 121, 0.1);
  line-height: 1.55;

  tr:last-child & {
    border-bottom: none;
  }

  tr:hover & {
    background-color: rgba(246, 248, 249, 0.7);
  }
`

// ─── HORIZONTAL RULE ────────────────────────────────────────────────────────

const StyledHr = styled.hr`
  border: none;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(100, 111, 121, 0.22) 20%,
    rgba(100, 111, 121, 0.22) 80%,
    transparent 100%
  );
  margin: 24px 0;
`

// ─── STREAMING CURSOR ───────────────────────────────────────────────────────

const StreamingCaret = styled.span`
  display: inline-block;
  margin-left: 4px;
  color: ${theme.colors.accent.purple};
  font-weight: 700;
  animation: ${blinkKeyframe} 1s infinite ease-in-out;
  vertical-align: baseline;
`

// ─── CITATION REGEX & HELPER ────────────────────────────────────────────────

const CITATION_REGEX = /\[(\d+(?:[\s,\u2013-]+\d+)*)\]/g

function renderWithCitations(
  children: ReactNode,
  references: Array<MarkdownReference> = [],
  onSelectCitation?: (citationId: string) => void,
): ReactNode {
  return Children.map(children, (child) => {
    if (typeof child === 'string') {
      if (!CITATION_REGEX.test(child)) {
        return child
      }
      CITATION_REGEX.lastIndex = 0

      const result: Array<ReactNode> = []
      let lastIndex = 0
      let match: RegExpExecArray | null

      while ((match = CITATION_REGEX.exec(child)) !== null) {
        if (match.index > lastIndex) {
          result.push(child.slice(lastIndex, match.index))
        }

        const token = match[1]
        const segments = token.split(',')
        const nums: Array<number> = []

        for (const seg of segments) {
          const trimmed = seg.trim()
          const rangeMatch = /^(\d+)[\u2013-](\d+)$/.exec(trimmed)
          if (rangeMatch) {
            const start = parseInt(rangeMatch[1], 10)
            const end = parseInt(rangeMatch[2], 10)
            if (start <= end && end - start < 15) {
              for (let i = start; i <= end; i++) nums.push(i)
            } else {
              nums.push(start, end)
            }
          } else {
            const val = parseInt(trimmed, 10)
            if (!Number.isNaN(val)) nums.push(val)
          }
        }

        // Validate that these numbers are plausible citations
        const isCitation =
          nums.length > 0 &&
          nums.every((n) => n < 100 || references.some((ref) => ref.num === n))

        if (isCitation) {
          nums.forEach((num, nIdx) => {
            const refItem = references.find((r) => r.num === num)
            const title = refItem
              ? `[${refItem.num}] ${refItem.title || ''}${
                  refItem.authors
                    ? ` (${String(refItem.authors).split(',')[0]} et al., ${refItem.year || ''})`
                    : ''
                }`
              : `Citation [${num}]`

            result.push(
              <CitationPill
                key={`cite-${match?.index ?? 0}-${num}-${nIdx}`}
                num={num}
                onClick={() => {
                  if (refItem && onSelectCitation) {
                    onSelectCitation(refItem.id)
                  }
                }}
                title={title}
              />,
            )
          })
        } else {
          result.push(match[0])
        }

        lastIndex = CITATION_REGEX.lastIndex
      }

      if (lastIndex < child.length) {
        result.push(child.slice(lastIndex))
      }
      return result
    }

    if (isValidElement<{ children?: ReactNode }>(child)) {
      // Do not replace inside code blocks or pills
      if (
        child.type === 'code' ||
        child.type === 'pre' ||
        child.type === CitationPill
      ) {
        return child
      }
      if (child.props.children) {
        return cloneElement(child, {
          children: renderWithCitations(
            child.props.children,
            references,
            onSelectCitation,
          ),
        })
      }
    }

    return child
  })
}

// ─── CODE BLOCK SUBCOMPONENT ────────────────────────────────────────────────

interface CodeBlockProps {
  code: string
  language?: string
}

const CodeBlock = memo(({ code, language }: CodeBlockProps) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Fallback
    }
  }, [code])

  return (
    <CodeBlockContainer>
      <CodeBlockHeader>
        <CodeLanguageLabel>{language || 'Code'}</CodeLanguageLabel>
        <CopyButton type="button" onClick={handleCopy} aria-label="Copy code">
          {copied ? <Check size={13} /> : <Copy size={13} />}
          <span>{copied ? 'Tersalin' : 'Salin'}</span>
        </CopyButton>
      </CodeBlockHeader>
      <CodePre>
        <code>{code}</code>
      </CodePre>
    </CodeBlockContainer>
  )
})

CodeBlock.displayName = 'CodeBlock'

// ─── MAIN MARKDOWN RENDERER COMPONENT ───────────────────────────────────────

export const MarkdownRenderer = memo(
  ({
    content,
    references = [],
    onSelectCitation,
    isStreaming = false,
    className,
  }: MarkdownRendererProps) => {
    const components = useMemo(
      () => ({
        h1: ({ children }: { children?: ReactNode }) => (
          <MarkdownH1>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH1>
        ),
        h2: ({ children }: { children?: ReactNode }) => (
          <MarkdownH2>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH2>
        ),
        h3: ({ children }: { children?: ReactNode }) => (
          <MarkdownH3>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH3>
        ),
        h4: ({ children }: { children?: ReactNode }) => (
          <MarkdownH4>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH4>
        ),
        h5: ({ children }: { children?: ReactNode }) => (
          <MarkdownH5>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH5>
        ),
        h6: ({ children }: { children?: ReactNode }) => (
          <MarkdownH6>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownH6>
        ),
        p: ({ children }: { children?: ReactNode }) => (
          <MarkdownParagraph>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownParagraph>
        ),
        ul: ({ children }: { children?: ReactNode }) => (
          <MarkdownUl>{children}</MarkdownUl>
        ),
        ol: ({ children }: { children?: ReactNode }) => (
          <MarkdownOl>{children}</MarkdownOl>
        ),
        li: ({ children }: { children?: ReactNode }) => (
          <MarkdownLi>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownLi>
        ),
        blockquote: ({ children }: { children?: ReactNode }) => (
          <MarkdownBlockquote>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownBlockquote>
        ),
        strong: ({ children }: { children?: ReactNode }) => (
          <MarkdownStrong>
            {renderWithCitations(children, references, onSelectCitation)}
          </MarkdownStrong>
        ),
        em: ({ children }: { children?: ReactNode }) => (
          <MarkdownEm>{children}</MarkdownEm>
        ),
        del: ({ children }: { children?: ReactNode }) => (
          <MarkdownDel>{children}</MarkdownDel>
        ),
        hr: () => <StyledHr />,
        a: ({ href, children }: { href?: string; children?: ReactNode }) => {
          const isExternal =
            href?.startsWith('http://') || href?.startsWith('https://')
          return (
            <MarkdownLink
              href={href}
              target={isExternal ? '_blank' : undefined}
              rel={isExternal ? 'noopener noreferrer' : undefined}
            >
              {renderWithCitations(children, references, onSelectCitation)}
              {isExternal && <LinkIcon size={12} />}
            </MarkdownLink>
          )
        },
        pre: ({ children }: { children?: ReactNode }) => <>{children}</>,
        code: ({
          className: codeClassName,
          children,
          ...rest
        }: {
          className?: string
          children?: ReactNode
        }) => {
          const codeStr =
            typeof children === 'string' ? children : String(children ?? '')
          const hasLang = Boolean(codeClassName)
          const isBlock = hasLang || codeStr.endsWith('\n')

          if (isBlock) {
            const language = codeClassName
              ? codeClassName.replace(/^language-/, '')
              : ''
            return (
              <CodeBlock
                code={codeStr.replace(/\n$/, '')}
                language={language}
              />
            )
          }

          return <InlineCode {...rest}>{children}</InlineCode>
        },
        table: ({ children }: { children?: ReactNode }) => (
          <TableWrapper>
            <StyledTable>{children}</StyledTable>
          </TableWrapper>
        ),
        th: ({ children }: { children?: ReactNode }) => (
          <StyledTh>
            {renderWithCitations(children, references, onSelectCitation)}
          </StyledTh>
        ),
        td: ({ children }: { children?: ReactNode }) => (
          <StyledTd>
            {renderWithCitations(children, references, onSelectCitation)}
          </StyledTd>
        ),
      }),
      [references, onSelectCitation],
    )

    return (
      <RootProse className={className}>
        <Markdown remarkPlugins={[remarkGfm]} components={components}>
          {content}
        </Markdown>
        {isStreaming && <StreamingCaret>▌</StreamingCaret>}
      </RootProse>
    )
  },
)

MarkdownRenderer.displayName = 'MarkdownRenderer'
