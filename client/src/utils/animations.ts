/**
 * Animation System Utilities
 * 
 * This file centralizes all animation class names and utility functions for the Motion System.
 * Import animation classes from here to avoid magic strings throughout the codebase.
 */

// Core Animation Classes
export const FADE_IN = 'animation-fade-in';
export const FADE_OUT = 'animation-fade-out';

export const SLIDE_IN_UP = 'animation-slide-in-up';
export const SLIDE_IN_DOWN = 'animation-slide-in-down';
export const SLIDE_IN_LEFT = 'animation-slide-in-left';
export const SLIDE_IN_RIGHT = 'animation-slide-in-right';

export const SLIDE_OUT_UP = 'animation-slide-out-up';
export const SLIDE_OUT_DOWN = 'animation-slide-out-down';
export const SLIDE_OUT_LEFT = 'animation-slide-out-left';
export const SLIDE_OUT_RIGHT = 'animation-slide-out-right';

export const SCALE_IN = 'animation-scale-in';
export const SCALE_OUT = 'animation-scale-out';

// Duration Variants
export const FADE_IN_FAST = 'animation-fade-in-fast';
export const SLIDE_IN_UP_FAST = 'animation-slide-in-up-fast';
export const SCALE_IN_FAST = 'animation-scale-in-fast';

export const FADE_IN_SLOW = 'animation-fade-in-slow';
export const SLIDE_IN_UP_SLOW = 'animation-slide-in-up-slow';
export const SCALE_IN_SLOW = 'animation-scale-in-slow';

// Feedback Variants
export const SCALE_IN_FEEDBACK = 'animation-scale-in-feedback';

// Stagger Animation
export const STAGGER_CHILD = 'stagger-child';

// Game-specific Animations
export const GLITCH = 'animation-glitch';
export const PULSE = 'animation-pulse';
export const SHAKE = 'animation-shake';
export const FLIP_CARD = 'animation-flip-card';
export const CARD_FLIP_IN = 'animation-card-flip-in';
export const ELIMINATION_FADE = 'animation-elimination-fade';

/**
 * Legacy Animation Classes (for backwards compatibility during migration)
 * @deprecated Use the new animation constants above instead
 */
export const ANIMATION_CLASSES = {
  FADE_IN: FADE_IN,
  FADE_OUT: FADE_OUT,
  SLIDE_IN_UP: SLIDE_IN_UP,
  SLIDE_IN_DOWN: SLIDE_IN_DOWN,
  SLIDE_IN_LEFT: SLIDE_IN_LEFT,
  SLIDE_IN_RIGHT: SLIDE_IN_RIGHT,
  SCALE_IN: SCALE_IN,
  PULSE: PULSE,
  SHAKE: SHAKE,
  FLIP_CARD: FLIP_CARD,
  STAGGER_REVEAL: STAGGER_CHILD,
} as const;

/**
 * Applies staggered animation to a list of elements
 * @param elements - NodeList or Array of HTML elements to animate
 * @param staggerDelay - Delay between each element's animation in milliseconds (default: 50ms)
 * @param animationClass - CSS class to apply to each element (default: STAGGER_CHILD)
 */
export function applyStaggeredAnimation(
  elements: NodeListOf<Element> | Element[],
  staggerDelay: number = 50,
  animationClass: string = STAGGER_CHILD
): void {
  const elementsArray = Array.from(elements);
  
  elementsArray.forEach((element, index) => {
    const htmlElement = element as HTMLElement;
    
    // Add the animation class
    htmlElement.classList.add(animationClass);
    
    // Apply stagger delay
    htmlElement.style.animationDelay = `${index * staggerDelay}ms`;
  });
}

/**
 * Removes animation classes and resets animation delay
 * @param elements - NodeList or Array of HTML elements to reset
 * @param animationClasses - Array of animation class names to remove
 */
export function resetAnimation(
  elements: NodeListOf<Element> | Element[],
  animationClasses: string[] = [STAGGER_CHILD]
): void {
  const elementsArray = Array.from(elements);
  
  elementsArray.forEach((element) => {
    const htmlElement = element as HTMLElement;
    
    // Remove animation classes
    animationClasses.forEach(className => {
      htmlElement.classList.remove(className);
    });
    
    // Reset animation delay
    htmlElement.style.animationDelay = '';
  });
}

/**
 * Creates a promise that resolves when an animation completes
 * @param element - The element with the animation
 * @returns Promise that resolves when animation ends
 */
export function waitForAnimation(element: HTMLElement): Promise<void> {
  return new Promise((resolve) => {
    const handleAnimationEnd = () => {
      element.removeEventListener('animationend', handleAnimationEnd);
      resolve();
    };
    
    element.addEventListener('animationend', handleAnimationEnd);
  });
}

/**
 * Motion System Durations (in milliseconds)
 * These match the CSS custom properties for JavaScript usage
 */
export const DURATIONS = {
  FAST: 150,
  MEDIUM: 300,
  SLOW: 500,
} as const;

/**
 * Legacy duration constants (for backwards compatibility)
 * @deprecated Use DURATIONS instead
 */
export const ANIMATION_DURATIONS = DURATIONS;

/**
 * Motion System Easing Curves
 * These match the CSS custom properties for JavaScript usage
 */
export const EASING = {
  OUT: 'cubic-bezier(0.25, 0.46, 0.45, 0.94)',
  IN: 'cubic-bezier(0.55, 0.085, 0.68, 0.53)',
  IN_OUT: 'cubic-bezier(0.445, 0.05, 0.55, 0.95)',
  FEEDBACK: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
} as const;

/**
 * Utility function to trigger a CSS animation
 * @param element - The HTML element to animate
 * @param animationClass - The animation class to apply
 * @param duration - Optional duration override in milliseconds
 * @returns Promise that resolves when animation completes
 */
export const triggerAnimation = (
  element: HTMLElement,
  animationClass: string,
  duration?: number
): Promise<void> => {
  return new Promise((resolve) => {
    const cleanup = () => {
      element.classList.remove(animationClass);
      element.removeEventListener('animationend', cleanup);
      resolve();
    };

    element.addEventListener('animationend', cleanup);
    element.classList.add(animationClass);

    if (duration) {
      element.style.animationDuration = `${duration}ms`;
    }

    // Fallback timeout
    setTimeout(() => {
      cleanup();
    }, duration || DURATIONS.MEDIUM);
  });
};

/**
 * Screen transition helper
 * @param outgoingElement - Element to animate out
 * @param incomingElement - Element to animate in
 * @param options - Transition options
 */
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
    outAnimation = FADE_OUT,
    inAnimation = FADE_IN,
    duration = DURATIONS.MEDIUM,
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