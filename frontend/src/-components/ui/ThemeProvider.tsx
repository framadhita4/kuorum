import type { ReactNode } from 'react'
import { useState } from 'react'

import {
  CacheProvider,
  css,
  Global,
  ThemeProvider as EmotionThemeProvider,
} from '@emotion/react'

import createEmotionCache from './createEmotionCache'
import { theme } from './theme'

export interface ThemeProviderProps {
  children: ReactNode
}

const globalStyles = css`
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap');

  :root {
    /* Surface & Canvas */
    --color-bg-primary: ${theme.colors.canvas.primary};
    --color-bg-subtle: ${theme.colors.canvas.subtle};
    --color-bg-neutral: ${theme.colors.canvas.neutral};
    --color-bg-charcoal: ${theme.colors.canvas.charcoal};
    --color-bg-dark: ${theme.colors.canvas.dark};
    --color-bg-footer: ${theme.colors.canvas.footer};

    /* Ink / Text */
    --color-ink-primary: ${theme.colors.ink.primary};
    --color-ink-secondary: ${theme.colors.ink.secondary};
    --color-ink-muted: ${theme.colors.ink.muted};
    --color-ink-subtle: ${theme.colors.ink.subtle};
    --color-ink-inverse: ${theme.colors.ink.inverse};

    /* Brand & Accents */
    --color-brand-coral: ${theme.colors.brand.coral};
    --color-brand-vibrant: ${theme.colors.brand.vibrant};
    --color-brand-claret: ${theme.colors.brand.claret};
    --color-accent-purple: ${theme.colors.accent.purple};
    --color-accent-purple-dark: ${theme.colors.accent.purpleDark};
    --color-accent-lilac: ${theme.colors.accent.lilac};
    --color-accent-blue: ${theme.colors.accent.blue};
    --color-accent-midnight: ${theme.colors.accent.midnight};
    --color-accent-gold: ${theme.colors.accent.gold};
    --color-danger: ${theme.colors.accent.danger};

    /* Gradients */
    --gradient-surface-fade: ${theme.gradients.surfaceFade};
    --gradient-ai-purple: ${theme.gradients.aiPurple};
    --gradient-dark-scrim: ${theme.gradients.darkScrim};

    /* Typography */
    --font-sans: ${theme.typography.fontFamily.sans};
    --font-display: ${theme.typography.fontFamily.display};
    --font-mono: ${theme.typography.fontFamily.mono};

    /* Type Scale */
    --text-display: ${theme.typography.sizes.display};
    --text-h1: ${theme.typography.sizes.h1};
    --text-h2: ${theme.typography.sizes.h2};
    --text-h3: ${theme.typography.sizes.h3};
    --text-h4: ${theme.typography.sizes.h4};
    --text-body-lg: ${theme.typography.sizes.bodyLg};
    --text-body: ${theme.typography.sizes.body};
    --text-sm: ${theme.typography.sizes.sm};
    --text-xs: ${theme.typography.sizes.xs};

    /* Spacing (4px base) */
    --space-1: ${theme.spacing['1']};
    --space-2: ${theme.spacing['2']};
    --space-3: ${theme.spacing['3']};
    --space-4: ${theme.spacing['4']};
    --space-6: ${theme.spacing['6']};
    --space-8: ${theme.spacing['8']};
    --space-12: ${theme.spacing['12']};
    --space-16: ${theme.spacing['16']};
    --space-20: ${theme.spacing['20']};
    --space-30: ${theme.spacing['30']};

    /* Border Radius */
    --radius-sm: ${theme.radius.sm};
    --radius-button: ${theme.radius.button};
    --radius-card: ${theme.radius.card};
    --radius-hero: ${theme.radius.hero};
    --radius-pill: ${theme.radius.pill};

    /* Borders */
    --border-subtle-color: ${theme.colors.border.subtleColor};
    --border-subtle: ${theme.colors.border.subtle};

    /* Elevation & Motion */
    --shadow-sm: ${theme.shadows.sm};
    --shadow-md: ${theme.shadows.md};
    --shadow-lg: ${theme.shadows.lg};
    --ease-out-expo: ${theme.transitions.easeOutExpo};
    --duration-fast: ${theme.transitions.durationFast};
    --duration-normal: ${theme.transitions.durationNormal};

    /* Dark Mode Compatibility */
    --bg-primary: ${theme.colors.dark.base};
    --bg-secondary: ${theme.colors.dark.surface};
    --bg-surface: ${theme.colors.dark.elevated};
    --bg-elevated: ${theme.colors.dark.popover};
    --border-subtle: ${theme.colors.dark.borderSubtle};
    --border-medium: ${theme.colors.dark.borderMedium};
    --text-main: ${theme.colors.dark.textPrimary};
    --text-secondary: ${theme.colors.dark.textSecondary};
    --text-muted: ${theme.colors.dark.textMuted};
  }

  *,
  *::before,
  *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  html,
  body {
    min-height: 100vh;
    background-color: ${theme.colors.canvas.primary};
    color: ${theme.colors.ink.primary};
    font-family: ${theme.typography.fontFamily.sans};
    line-height: ${theme.typography.lineHeights.normal};
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    overflow-x: hidden;
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    font-family: ${theme.typography.fontFamily.sans};
    letter-spacing: ${theme.typography.letterSpacing.tight};
    color: ${theme.colors.ink.primary};
  }

  code,
  pre,
  kbd {
    font-family: ${theme.typography.fontFamily.mono};
  }

  ::selection {
    background-color: ${theme.colors.brand.coral};
    color: ${theme.colors.ink.primary};
  }
`

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const [cache] = useState(() => createEmotionCache())

  return (
    <CacheProvider value={cache}>
      <EmotionThemeProvider theme={theme}>
        <Global styles={globalStyles} />
        {children}
      </EmotionThemeProvider>
    </CacheProvider>
  )
}
