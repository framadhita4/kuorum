/**
 * Single Main Theme File for Research Helper
 * Powered by @emotion/react
 *
 * All visual design tokens (colors, gradients, typography, spacing, radius, shadows, transitions)
 * across the entire application are strictly controlled and adjusted from this file.
 */

export const theme = {
  // 1. Semantic Color Palette
  colors: {
    // Surface & Canvas
    canvas: {
      primary: '#ffffff',
      subtle: '#f6f8f9',
      neutral: '#f3f3f3',
      surface: '#ffffff',
      elevated: '#ffffff',
      charcoal: '#222b37',
      dark: '#151b26',
      footer: '#0d1b2f',
      glass: 'rgba(255, 255, 255, 0.85)',
      overlay: 'rgba(13, 14, 16, 0.6)',
    },

    // Ink (Text & Iconography)
    ink: {
      primary: '#0d0e10', //
      secondary: '#4a5464',
      muted: '#646f79',
      subtle: '#6e6e6e',
      inverse: '#ffffff',
    },

    // Aliased to ink for seamless backward compatibility
    text: {
      primary: '#0d0e10',
      secondary: '#4a5464',
      muted: '#646f79',
      subtle: '#6e6e6e',
      inverse: '#ffffff',
    },

    // Brand & Semantic Accents
    brand: {
      coral: '#f06a6a',
      vibrant: '#ff584a',
      claret: '#690031',
    },

    accent: {
      purple: '#796eff',
      purpleDark: '#635ac7',
      lilac: '#ffc9ff',
      blue: '#00b2ff',
      midnight: '#222875',
      gold: '#fcbd01',
      danger: '#cd4848',
    },

    // Status Feedback
    status: {
      success: '#34c759',
      warning: '#eab308',
      error: '#ef4444',
      info: '#00b2ff',
    },

    // Borders
    border: {
      subtle: '1px solid rgba(100, 111, 121, 0.2)',
      subtleColor: '#e2e6e9',
      medium: 'rgba(100, 111, 121, 0.35)',
      strong: 'rgba(100, 111, 121, 0.6)',
      focus: '#796eff',
    },

    // Neutral Grayscale Scale
    neutral: {
      '50': '#fafafa',
      '100': '#f5f5f5',
      '200': '#e5e5e5',
      '300': '#d4d4d4',
      '400': '#a3a3a3',
      '500': '#737373',
      '600': '#525252',
      '700': '#404040',
      '800': '#262626',
      '900': '#171717',
    },

    // Dark Mode Compatibility Tokens (For dark terminal/explore/graph widgets)
    dark: {
      base: '#121212',
      surface: '#181818',
      elevated: '#1e1e1e',
      popover: '#242424',
      borderSubtle: '#262626',
      borderMedium: '#333333',
      borderStrong: '#444444',
      textPrimary: '#f5f5f7',
      textSecondary: '#a1a1a6',
      textMuted: '#6e6e73',
      textInverse: '#121212',
    },
  },

  // 2. Gradients
  gradients: {
    surfaceFade: 'linear-gradient(180deg, #ffffff 0%, #f6f8f9 100%)',
    aiPurple:
      'linear-gradient(90deg, #796eff 0%, #d472c5 45%, #df70b8 65%, #ff6d92 100%)',
    darkScrim:
      'linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.75) 100%)',
    heroGaze:
      'radial-gradient(ellipse at 50% 0%, rgba(121, 110, 255, 0.15) 0%, rgba(255, 255, 255, 0) 70%)',
  },

  // 3. Typography Scale & Families
  typography: {
    fontFamily: {
      sans: "'TWK Lausanne', 'Inter', -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif",
      display:
        "'Ghost', 'TWK Lausanne', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
      mono: "'JetBrains Mono', SFMono-Regular, Menlo, Monaco, Consolas, monospace",
    },
    weights: {
      light: 300,
      regular: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
    sizes: {
      display: '6.375rem', // 102px
      h1: '3.375rem', // 54px
      h2: '2.25rem', // 36px
      h3: '1.875rem', // 30px
      h4: '1.5rem', // 24px
      h5: '1.25rem', // 20px
      title: '1.25rem', // 20px
      bodyLg: '1.25rem', // 20px
      subtitle: '1.0625rem', // 17px
      body: '1rem', // 16px
      sm: '0.875rem', // 14px
      xs: '0.75rem', // 12px
      caption: '0.75rem', // 12px
      micro: '0.6875rem', // 11px
    },
    lineHeights: {
      tight: 1.1,
      snug: 1.25,
      normal: 1.5,
      relaxed: 1.65,
    },
    letterSpacing: {
      tightest: '-0.035em',
      tight: '-0.02em',
      normal: '0em',
      wide: '0.04em',
      widest: '0.08em',
    },
  },

  // 4. Spacing Scale (4px base)
  spacing: {
    '1': '4px',
    '2': '8px',
    '3': '12px',
    '4': '16px',
    '6': '24px',
    '8': '32px',
    '12': '48px',
    '16': '64px',
    '20': '80px',
    '30': '120px',
    // Semantic aliases
    none: '0px',
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    xxl: '48px',
    '3xl': '64px',
    '4xl': '80px',
  },

  // 5. Border Radius
  radius: {
    none: '0px',
    xs: '4px',
    sm: '4px',
    button: '8px',
    md: '8px',
    lg: '12px',
    card: '16px',
    xl: '16px',
    hero: '24px',
    '2xl': '24px',
    pill: '9999px',
    full: '9999px',
  },

  // 6. Elevation & Shadows
  shadows: {
    none: 'none',
    sm: '0 2px 6px rgba(13, 14, 16, 0.04)',
    md: '0 8px 24px rgba(13, 14, 16, 0.08)',
    lg: '0 16px 40px rgba(13, 14, 16, 0.12)',
  },

  // 7. Transitions & Motion
  transitions: {
    fast: 'all 200ms cubic-bezier(0.22, 1, 0.36, 1)',
    normal: 'all 350ms cubic-bezier(0.22, 1, 0.36, 1)',
    easeOutExpo: 'cubic-bezier(0.22, 1, 0.36, 1)',
    durationFast: '200ms',
    durationNormal: '350ms',
  },
} as const

export type Theme = typeof theme
