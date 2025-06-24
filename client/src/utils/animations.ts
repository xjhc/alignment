// Animation utility constants and helpers for the Alignment game UI

export const ANIMATION_DURATIONS = {
  FAST: 150,
  MEDIUM: 300,
  SLOW: 500,
  EXTRA_SLOW: 800,
} as const;

export const EASING = {
  EASE_OUT: 'cubic-bezier(0.25, 0.46, 0.45, 0.94)',
  EASE_IN_OUT: 'cubic-bezier(0.645, 0.045, 0.355, 1)',
  BOUNCE: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
} as const;

// Animation class names that correspond to CSS animations
export const ANIMATION_CLASSES = {
  FADE_IN: 'animate-fade-in',
  FADE_OUT: 'animate-fade-out',
  SLIDE_IN_UP: 'animate-slide-in-up',
  SLIDE_IN_DOWN: 'animate-slide-in-down',
  SLIDE_IN_LEFT: 'animate-slide-in-left',
  SLIDE_IN_RIGHT: 'animate-slide-in-right',
  SCALE_IN: 'animate-scale-in',
  PULSE: 'animate-pulse',
  SHAKE: 'animate-shake',
  BOUNCE: 'animate-bounce',
  FLIP_CARD: 'animate-flip-card',
  STAGGER_REVEAL: 'animate-stagger-reveal',
  BUTTON_PRESS: 'animate-button-press',
} as const;

// Utility function to add staggered animation delays to elements
export const applyStaggeredAnimation = (
  elements: NodeListOf<Element> | Element[],
  baseDelay: number = 100
) => {
  Array.from(elements).forEach((element, index) => {
    const delay = index * baseDelay;
    (element as HTMLElement).style.animationDelay = `${delay}ms`;
  });
};

// Utility function to trigger a CSS animation
export const triggerAnimation = (
  element: HTMLElement,
  animationClass: string,
  duration?: number
): Promise<void> => {
  return new Promise((resolve) => {
    const cleanup = () => {
      element.classList.remove(animationClass);
      element.removeEventListener('animationend', cleanup);
    };

    element.addEventListener('animationend', cleanup);
    element.classList.add(animationClass);

    if (duration) {
      element.style.animationDuration = `${duration}ms`;
    }

    // Fallback timeout
    setTimeout(resolve, duration || ANIMATION_DURATIONS.MEDIUM);
  });
};

// Screen transition helper
export const transitionScreen = async (
  outgoingElement: HTMLElement | null,
  incomingElement: HTMLElement | null,
  options: {
    outAnimation?: string;
    inAnimation?: string;
    duration?: number;
  } = {}
) => {
  const {
    outAnimation = ANIMATION_CLASSES.FADE_OUT,
    inAnimation = ANIMATION_CLASSES.FADE_IN,
    duration = ANIMATION_DURATIONS.MEDIUM,
  } = options;

  // Fade out current screen
  if (outgoingElement) {
    await triggerAnimation(outgoingElement, outAnimation, duration);
  }

  // Small delay for smoothness
  await new Promise(resolve => setTimeout(resolve, 50));

  // Fade in new screen
  if (incomingElement) {
    await triggerAnimation(incomingElement, inAnimation, duration);
  }
};