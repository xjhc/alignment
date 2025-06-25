import { useState, useRef } from 'react';
import type { Meta } from '@storybook/react';
import { 
  STAGGER_CHILD,
  applyStaggeredAnimation
} from '../../utils/animations';

const meta: Meta = {
  title: 'Design System/Motion System',
  parameters: {
    layout: 'fullscreen',
    docs: {
      page: () => <MotionSystemPage />,
    },
  },
};

export default meta;

const MotionSystemPage = () => {
  const [animationKey, setAnimationKey] = useState(0);
  const staggerRef = useRef<HTMLDivElement>(null);

  const triggerStaggerAnimation = () => {
    if (staggerRef.current) {
      const children = staggerRef.current.querySelectorAll('.demo-stagger-item');
      // Reset first
      children.forEach(child => {
        (child as HTMLElement).classList.remove(STAGGER_CHILD);
        (child as HTMLElement).style.animationDelay = '';
      });
      
      // Apply stagger
      setTimeout(() => {
        applyStaggeredAnimation(children, 100);
      }, 50);
    }
  };

  const resetAnimation = () => {
    setAnimationKey(prev => prev + 1);
  };

  const cardStyle = {
    padding: '1rem',
    border: '1px solid var(--border)',
    borderRadius: 'var(--radius-md)',
    backgroundColor: 'var(--bg-secondary)',
  };

  const demoBoxStyle = {
    width: '80px',
    height: '80px',
    backgroundColor: 'var(--accent-primary)',
    borderRadius: 'var(--radius-md)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'white',
    fontWeight: 'var(--font-weight-semibold)',
    fontSize: 'var(--font-size-sm)',
  };

  const gridStyle = {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
    gap: '1rem',
    marginBottom: '2rem',
  };

  return (
    <div style={{ padding: '32px', fontFamily: 'Inter, sans-serif', maxWidth: '1200px' }}>
      <h1 style={{ fontSize: '32px', fontWeight: '700', marginBottom: '16px' }}>Motion System</h1>
      
      <p style={{ fontSize: '16px', color: 'var(--text-secondary)', marginBottom: '48px', lineHeight: '1.6' }}>
        The Motion System defines consistent, purposeful animations for the Alignment game. 
        All animations are designed to be crisp, efficient, and informative.
      </p>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Motion Principles</h2>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem', marginBottom: '32px' }}>
          <div style={cardStyle}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '12px' }}>Purposeful & Informative</h3>
            <p style={{ color: 'var(--text-secondary)', lineHeight: '1.5' }}>
              Motion must have a reason. It should guide the user's eye, show relationships between UI elements, and provide feedback on interactions.
            </p>
          </div>
          <div style={cardStyle}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '12px' }}>Crisp & Efficient</h3>
            <p style={{ color: 'var(--text-secondary)', lineHeight: '1.5' }}>
              Animations should be quick and precise, reflecting the game's high-stakes, data-driven theme. No slow or whimsical animations.
            </p>
          </div>
          <div style={cardStyle}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '12px' }}>Consistent</h3>
            <p style={{ color: 'var(--text-secondary)', lineHeight: '1.5' }}>
              The same interaction should always produce the same motion. This creates a predictable and intuitive rhythm for the user.
            </p>
          </div>
          <div style={cardStyle}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '12px' }}>Performant</h3>
            <p style={{ color: 'var(--text-secondary)', lineHeight: '1.5' }}>
              All animations must only animate transform and opacity to ensure they are smooth and do not cause layout shifts.
            </p>
          </div>
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Duration Tokens</h2>
        <div style={gridStyle}>
          {[
            { name: 'Fast', value: '150ms', var: '--duration-fast', usage: 'Immediate feedback on user interaction (button press, hover effect)' },
            { name: 'Medium', value: '300ms', var: '--duration-medium', usage: 'Standard UI element transitions (element appearing/disappearing)' },
            { name: 'Slow', value: '500ms', var: '--duration-slow', usage: 'Large-scale screen or panel transitions' },
          ].map(({ name, value, var: varName, usage }) => (
            <div key={varName} style={cardStyle}>
              <h3 style={{ fontSize: '16px', fontWeight: '600', marginBottom: '8px' }}>{name}</h3>
              <div style={{ 
                ...demoBoxStyle, 
                animation: `fadeIn ${value} var(--ease-out) infinite alternate`,
                marginBottom: '12px'
              }}>
                {value}
              </div>
              <code style={{ display: 'block', marginBottom: '8px' }}>{varName}</code>
              <p style={{ fontSize: '12px', color: 'var(--text-muted)', lineHeight: '1.4' }}>{usage}</p>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Easing Curves</h2>
        <div style={gridStyle}>
          {[
            { name: 'Ease Out', var: '--ease-out', value: 'cubic-bezier(0.25, 0.46, 0.45, 0.94)', usage: 'Standard curve for elements entering the screen' },
            { name: 'Ease In', var: '--ease-in', value: 'cubic-bezier(0.55, 0.085, 0.68, 0.53)', usage: 'Standard curve for elements leaving the screen' },
            { name: 'Ease In-Out', var: '--ease-in-out', value: 'cubic-bezier(0.445, 0.05, 0.55, 0.95)', usage: 'For elements that transform in place (color change)' },
            { name: 'Feedback', var: '--ease-feedback', value: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)', usage: '"Overshoot" curve for attention-grabbing feedback' },
          ].map(({ name, var: varName, value, usage }) => (
            <div key={varName} style={cardStyle}>
              <h3 style={{ fontSize: '16px', fontWeight: '600', marginBottom: '8px' }}>{name}</h3>
              <div style={{ 
                ...demoBoxStyle, 
                animation: `scaleIn var(--duration-medium) ${value} infinite alternate`,
                marginBottom: '12px'
              }}>
                Curve
              </div>
              <code style={{ display: 'block', marginBottom: '8px', fontSize: '10px' }}>{varName}</code>
              <p style={{ fontSize: '12px', color: 'var(--text-muted)', lineHeight: '1.4' }}>{usage}</p>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Animation Patterns</h2>
        
        <div style={{ marginBottom: '32px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
            <h3 style={{ fontSize: '18px', fontWeight: '600' }}>Fade In / Fade Out</h3>
            <button 
              onClick={resetAnimation}
              style={{ 
                padding: '8px 16px', 
                backgroundColor: 'var(--accent-primary)', 
                color: 'white', 
                border: 'none', 
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer'
              }}
            >
              Reset Animation
            </button>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
            For elements that appear or disappear in place (toasts, error messages)
          </p>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            <div key={`fade-${animationKey}`} style={{ ...demoBoxStyle, animation: `fadeIn var(--duration-medium) var(--ease-out) forwards` }}>
              Fade In
            </div>
          </div>
        </div>

        <div style={{ marginBottom: '32px' }}>
          <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '8px' }}>Slide In</h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
            For elements entering the viewport, like a new screen or side panel
          </p>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            <div key={`slide-up-${animationKey}`} style={{ ...demoBoxStyle, animation: `slideInUp var(--duration-medium) var(--ease-out) forwards` }}>
              Up
            </div>
            <div key={`slide-down-${animationKey}`} style={{ ...demoBoxStyle, animation: `slideInDown var(--duration-medium) var(--ease-out) forwards` }}>
              Down
            </div>
            <div key={`slide-left-${animationKey}`} style={{ ...demoBoxStyle, animation: `slideInLeft var(--duration-medium) var(--ease-out) forwards` }}>
              Left
            </div>
            <div key={`slide-right-${animationKey}`} style={{ ...demoBoxStyle, animation: `slideInRight var(--duration-medium) var(--ease-out) forwards` }}>
              Right
            </div>
          </div>
        </div>

        <div style={{ marginBottom: '32px' }}>
          <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '8px' }}>Scale In</h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
            For elements that need attention or emphasis
          </p>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            <div key={`scale-${animationKey}`} style={{ ...demoBoxStyle, animation: `scaleIn var(--duration-medium) var(--ease-out) forwards` }}>
              Scale
            </div>
            <div key={`scale-feedback-${animationKey}`} style={{ ...demoBoxStyle, animation: `scaleIn var(--duration-medium) var(--ease-feedback) forwards` }}>
              Feedback
            </div>
          </div>
        </div>

        <div style={{ marginBottom: '32px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
            <h3 style={{ fontSize: '18px', fontWeight: '600' }}>Staggered List Reveal</h3>
            <button 
              onClick={triggerStaggerAnimation}
              style={{ 
                padding: '8px 16px', 
                backgroundColor: 'var(--accent-primary)', 
                color: 'white', 
                border: 'none', 
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer'
              }}
            >
              Trigger Stagger
            </button>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
            To animate lists of items so they appear sequentially (e.g., player roster)
          </p>
          <div ref={staggerRef} style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            {Array.from({ length: 6 }, (_, i) => (
              <div 
                key={i}
                className="demo-stagger-item"
                style={{ 
                  ...demoBoxStyle, 
                  width: '60px', 
                  height: '60px',
                  fontSize: '12px'
                }}
              >
                {i + 1}
              </div>
            ))}
          </div>
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Usage Examples</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>CSS Classes</h3>
        <pre style={{ 
          backgroundColor: 'var(--bg-tertiary)', 
          padding: '16px', 
          borderRadius: '8px', 
          border: '1px solid var(--border)',
          fontFamily: 'var(--font-mono)',
          fontSize: '12px',
          marginBottom: '24px',
          overflow: 'auto'
        }}>
{`/* Basic animations */
.my-element {
  animation: fadeIn var(--duration-medium) var(--ease-out) forwards;
}

/* Using utility classes */
<div className="animation-fade-in">Content</div>
<div className="animation-slide-in-up">Content</div>
<div className="animation-scale-in-feedback">Button</div>`}
        </pre>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>JavaScript/React</h3>
        <pre style={{ 
          backgroundColor: 'var(--bg-tertiary)', 
          padding: '16px', 
          borderRadius: '8px', 
          border: '1px solid var(--border)',
          fontFamily: 'var(--font-mono)',
          fontSize: '12px',
          marginBottom: '24px',
          overflow: 'auto'
        }}>
{`import { 
  FADE_IN, 
  SLIDE_IN_UP, 
  applyStaggeredAnimation,
  triggerAnimation 
} from '../utils/animations';

// Using constants
<div className={FADE_IN}>Content</div>

// Programmatic animation
await triggerAnimation(element, SLIDE_IN_UP);

// Staggered list animation
const listItems = container.querySelectorAll('.list-item');
applyStaggeredAnimation(listItems, 100);`}
        </pre>
      </section>
    </div>
  );
};

export const Default = () => <MotionSystemPage />;