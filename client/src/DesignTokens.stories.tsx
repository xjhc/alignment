import type { Meta } from '@storybook/react';
import tokens from './styles/generated/tokens.json';

const meta: Meta = {
  title: 'Design System/Design Tokens',
  parameters: {
    layout: 'fullscreen',
    docs: {
      page: () => <DesignTokensPage />,
    },
  },
};

export default meta;

const DesignTokensPage = () => {
  return (
    <div style={{ padding: '32px', fontFamily: 'Inter, sans-serif' }}>
      <h1 style={{ fontSize: '32px', fontWeight: '700', marginBottom: '16px' }}>Design Tokens</h1>
      
      <p style={{ fontSize: '16px', color: '#64748b', marginBottom: '48px', lineHeight: '1.6' }}>
        Our design system is built on a foundation of design tokens that ensure consistency across all components and interfaces. 
        These tokens are automatically generated from JSON definitions using Style Dictionary.
      </p>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Colors</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Brand Colors</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '32px' }}>
          {Object.entries(tokens.brand.accent).map(([name, value]) => (
            <div key={name} style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <div style={{ width: '40px', height: '40px', backgroundColor: value, borderRadius: '8px', border: '1px solid #e2e8f0' }} />
              <div>
                <div style={{ fontWeight: '600', fontSize: '14px' }}>{name}</div>
                <div style={{ fontSize: '12px', color: '#64748b', fontFamily: 'JetBrains Mono, monospace' }}>{value}</div>
              </div>
            </div>
          ))}
        </div>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Semantic Colors</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '32px' }}>
          {Object.entries(tokens.semantic.game).map(([name, value]) => (
            <div key={name} style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <div style={{ width: '40px', height: '40px', backgroundColor: value, borderRadius: '8px', border: '1px solid #e2e8f0' }} />
              <div>
                <div style={{ fontWeight: '600', fontSize: '14px' }}>{name}</div>
                <div style={{ fontSize: '12px', color: '#64748b', fontFamily: 'JetBrains Mono, monospace' }}>{value}</div>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Typography</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Font Families</h3>
        <div style={{ marginBottom: '32px' }}>
          {Object.entries(tokens.font.family).map(([name, value]) => (
            <div key={name} style={{ marginBottom: '16px' }}>
              <div style={{ fontWeight: '600', fontSize: '14px', marginBottom: '8px' }}>{name}</div>
              <div style={{ fontFamily: value, fontSize: '16px', marginBottom: '4px' }}>
                The quick brown fox jumps over the lazy dog
              </div>
              <div style={{ fontSize: '12px', color: '#64748b', fontFamily: 'JetBrains Mono, monospace' }}>{value}</div>
            </div>
          ))}
        </div>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Font Sizes</h3>
        <div style={{ marginBottom: '32px' }}>
          {Object.entries(tokens.font.size).map(([name, value]) => (
            <div key={name} style={{ marginBottom: '12px', display: 'flex', alignItems: 'center', gap: '16px' }}>
              <div style={{ fontSize: value, lineHeight: '1.2' }}>Typography Sample</div>
              <div style={{ fontSize: '12px', color: '#64748b', fontFamily: 'JetBrains Mono, monospace' }}>
                {name}: {value}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Spacing</h2>
        <div style={{ marginBottom: '32px' }}>
          {Object.entries(tokens.spacing).map(([name, value]) => (
            <div key={name} style={{ marginBottom: '12px', display: 'flex', alignItems: 'center', gap: '16px' }}>
              <div style={{ width: value, height: '24px', backgroundColor: '#3b82f6', borderRadius: '4px' }} />
              <div style={{ fontSize: '14px', fontFamily: 'JetBrains Mono, monospace' }}>
                {name}: {value}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Usage</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>CSS Variables</h3>
        <p style={{ marginBottom: '16px', color: '#64748b' }}>All tokens are available as CSS custom properties:</p>
        <pre style={{ 
          backgroundColor: '#f8fafc', 
          padding: '16px', 
          borderRadius: '8px', 
          border: '1px solid #e2e8f0',
          fontFamily: 'JetBrains Mono, monospace',
          fontSize: '14px',
          marginBottom: '24px'
        }}>
{`.my-component {
  background-color: var(--bg-primary);
  color: var(--text-primary);
  padding: var(--space-4);
  border-radius: var(--radius-md);
  font-family: var(--font-sans);
}`}
        </pre>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>TypeScript/JavaScript</h3>
        <p style={{ marginBottom: '16px', color: '#64748b' }}>Tokens can also be imported and used in JavaScript:</p>
        <pre style={{ 
          backgroundColor: '#f8fafc', 
          padding: '16px', 
          borderRadius: '8px', 
          border: '1px solid #e2e8f0',
          fontFamily: 'JetBrains Mono, monospace',
          fontSize: '14px'
        }}>
{`import tokens from './styles/generated/tokens.json';

const styles = {
  backgroundColor: tokens.theme.dark.bg.primary,
  color: tokens.theme.dark.text.primary,
  padding: tokens.spacing[4],
  borderRadius: tokens.border.radius.md,
};`}
        </pre>
      </section>
    </div>
  );
};

export const Default = () => <DesignTokensPage />;