import { tokens } from './src/styles/generated/tokens.js';

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    // Use our design tokens as the foundation
    colors: {
      // Semantic brand colors
      primary: tokens.brand.accent.primary,
      amber: {
        DEFAULT: tokens.brand.accent.amber,
        light: tokens.brand.accent.amberLight,
        dark: tokens.brand.accent.amberDark,
      },
      cyan: {
        DEFAULT: tokens.brand.accent.cyan,
        light: tokens.brand.accent.cyanLight,
        dark: tokens.brand.accent.cyanDark,
      },
      red: {
        DEFAULT: tokens.brand.accent.red,
        light: tokens.brand.accent.redLight,
        dark: tokens.brand.accent.redDark,
      },
      magenta: {
        DEFAULT: tokens.brand.accent.magenta,
        light: tokens.brand.accent.magentaLight,
        dark: tokens.brand.accent.magentaDark,
      },
      green: {
        DEFAULT: tokens.brand.accent.green,
        light: tokens.brand.accent.greenLight,
        dark: tokens.brand.accent.greenDark,
      },
      blue: {
        DEFAULT: tokens.brand.accent.blue,
        light: tokens.brand.accent.blueLight,
        dark: tokens.brand.accent.blueDark,
      },
      // Game-specific colors
      human: tokens.semantic.game.human,
      aligned: tokens.semantic.game.aligned,
      ai: tokens.semantic.game.ai,
      success: tokens.semantic.game.success,
      danger: tokens.semantic.game.danger,
      warning: tokens.semantic.game.warning,
      info: tokens.semantic.game.info,
      // Theme colors (will be handled by CSS variables for theme switching)
      background: {
        primary: 'var(--bg-primary)',
        secondary: 'var(--bg-secondary)',
        tertiary: 'var(--bg-tertiary)',
        quaternary: 'var(--bg-quaternary)',
        hover: 'var(--bg-hover)',
        mention: 'var(--bg-mention)',
      },
      text: {
        primary: 'var(--text-primary)',
        secondary: 'var(--text-secondary)',
        muted: 'var(--text-muted)',
      },
      border: 'var(--border)',
      // Component state colors
      focus: tokens.semantic.component.focus,
      active: tokens.semantic.component.active,
    },
    fontFamily: {
      sans: tokens.font.family.sans.split(',').map(font => font.trim().replace(/['"]/g, '')),
      mono: tokens.font.family.mono.split(',').map(font => font.trim().replace(/['"]/g, '')),
    },
    fontSize: {
      xs: tokens.font.size.xs,
      sm: tokens.font.size.sm,
      base: tokens.font.size.base,
      md: tokens.font.size.md,
      lg: tokens.font.size.lg,
      xl: tokens.font.size.xl,
      '2xl': tokens.font.size['2xl'],
      '3xl': tokens.font.size['3xl'],
    },
    fontWeight: {
      light: tokens.font.weight.light,
      normal: tokens.font.weight.normal,
      medium: tokens.font.weight.medium,
      semibold: tokens.font.weight.semibold,
      bold: tokens.font.weight.bold,
      extrabold: tokens.font.weight.extrabold,
    },
    lineHeight: {
      tight: tokens.font.lineHeight.tight,
      normal: tokens.font.lineHeight.normal,
      relaxed: tokens.font.lineHeight.relaxed,
    },
    spacing: {
      1: tokens.spacing[1],
      2: tokens.spacing[2],
      3: tokens.spacing[3],
      4: tokens.spacing[4],
      5: tokens.spacing[5],
      6: tokens.spacing[6],
      8: tokens.spacing[8],
      10: tokens.spacing[10],
      12: tokens.spacing[12],
      16: tokens.spacing[16],
    },
    borderRadius: {
      sm: tokens.border.radius.sm,
      DEFAULT: tokens.border.radius.base,
      md: tokens.border.radius.md,
      lg: tokens.border.radius.lg,
      xl: tokens.border.radius.xl,
      '2xl': tokens.border.radius['2xl'],
      full: tokens.border.radius.full,
    },
    boxShadow: {
      DEFAULT: 'var(--shadow)',
    },
    extend: {
      // Custom animation utilities that match our existing ones
      animation: {
        'fade-in': 'fadeIn 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'fade-out': 'fadeOut 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'slide-in-up': 'fadeInUp 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'slide-in-down': 'fadeInDown 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'slide-in-left': 'slideInLeft 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'slide-in-right': 'slideInRight 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'scale-in': 'scaleIn 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'pulse': 'pulse 2s infinite',
        'shake': 'shake 0.5s',
        'bounce': 'bounce 1s',
        'flip-card': 'flipCard 0.8s ease-in-out',
        'stagger-reveal': 'staggerReveal 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'button-press': 'buttonPress 0.15s ease-out',
        'card-flip-in': 'cardFlipIn 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'elimination-fade': 'eliminationFade 1.5s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards',
        'loading-spinner': 'loadingSpinner 1s linear infinite',
        'progress-bar-fill': 'progressBarFill 1s ease-out forwards',
      },
    },
  },
  plugins: [],
}