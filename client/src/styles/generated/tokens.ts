/**
 * Do not edit directly, this file was auto-generated.
 */

export const tokens = {
  border: {
    radius: {
      sm: '3px',
      base: '4px',
      md: '6px',
      lg: '8px',
      xl: '12px',
      '2xl': '16px',
      full: '50%',
    },
  },
  brand: {
    accent: {
      primary: '#f59e0b',
      amber: '#f59e0b',
      amberLight: '#fbbf24',
      amberDark: '#d97706',
      cyan: '#06b6d4',
      cyanLight: '#22d3ee',
      cyanDark: '#0891b2',
      red: '#ef4444',
      redLight: '#f87171',
      redDark: '#dc2626',
      magenta: '#ec4899',
      magentaLight: '#f472b6',
      magentaDark: '#db2777',
      green: '#10b981',
      greenLight: '#34d399',
      greenDark: '#059669',
      blue: '#3b82f6',
      blueLight: '#60a5fa',
      blueDark: '#2563eb',
    },
  },
  semantic: {
    game: {
      human: '#f59e0b',
      aligned: '#06b6d4',
      ai: '#ec4899',
      success: '#10b981',
      danger: '#ef4444',
      warning: '#f59e0b',
      info: '#3b82f6',
    },
    component: {
      focus: '#3b82f6',
      active: '#f59e0b',
    },
  },
  theme: {
    light: {
      bg: {
        primary: '#ffffff',
        secondary: '#f8fafc',
        tertiary: '#f1f5f9',
        quaternary: '#e2e8f0',
        hover: '#f1f5f9',
        mention: 'rgba(245, 158, 11, 0.2)',
      },
      text: {
        primary: '#0f172a',
        secondary: '#475569',
        muted: '#64748b',
      },
      border: '#e2e8f0',
      shadow: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
    },
    dark: {
      bg: {
        primary: '#0f172a',
        secondary: '#1e293b',
        tertiary: '#334155',
        quaternary: '#475569',
        hover: '#334155',
        mention: 'rgba(251, 191, 36, 0.4)',
      },
      text: {
        primary: '#f8fafc',
        secondary: '#cbd5e1',
        muted: '#94a3b8',
      },
      border: '#334155',
      shadow: '0 1px 3px 0 rgb(0 0 0 / 0.3), 0 1px 2px -1px rgb(0 0 0 / 0.3)',
    },
  },
  font: {
    family: {
      sans: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      mono: "'JetBrains Mono', 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace",
    },
    weight: {
      light: '300',
      normal: '400',
      medium: '500',
      semibold: '600',
      bold: '700',
      extrabold: '800',
    },
    size: {
      xs: '10px',
      sm: '11px',
      base: '12px',
      md: '13px',
      lg: '14px',
      xl: '16px',
      '2xl': '18px',
      '3xl': '20px',
    },
    lineHeight: {
      tight: '1.2',
      normal: '1.4',
      relaxed: '1.6',
    },
  },
  spacing: {
    1: '4px',
    2: '8px',
    3: '12px',
    4: '16px',
    5: '20px',
    6: '24px',
    8: '32px',
    10: '40px',
    12: '48px',
    16: '64px',
  },
} as const;

export type TokensType = typeof tokens;