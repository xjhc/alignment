import type { Meta } from '@storybook/react';

const meta: Meta = {
  title: 'Design System/CSS Variables',
  parameters: {
    layout: 'fullscreen',
    docs: {
      page: () => <DesignSystemPage />,
    },
  },
};

export default meta;

const DesignSystemPage = () => {
  const cardStyle = {
    padding: '1rem',
    border: '1px solid var(--border)',
    borderRadius: 'var(--radius-md)',
  };

  const gridStyle = {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '1rem',
    marginBottom: '2rem',
  };

  return (
    <div style={{ padding: '32px', fontFamily: 'Inter, sans-serif', maxWidth: '1200px' }}>
      <h1 style={{ fontSize: '32px', fontWeight: '700', marginBottom: '16px' }}>Design System</h1>
      
      <p style={{ fontSize: '16px', color: '#64748b', marginBottom: '48px', lineHeight: '1.6' }}>
        This document showcases the CSS variables and design tokens used throughout the Alignment game application.
      </p>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Typography</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Font Families</h3>
        <ul style={{ marginBottom: '32px', lineHeight: '1.6' }}>
          <li><strong>Sans Serif</strong>: <code>--font-sans</code> → Inter, system fonts</li>
          <li><strong>Monospace</strong>: <code>--font-mono</code> → JetBrains Mono, SF Mono, Monaco</li>
        </ul>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Font Weights</h3>
        <div style={gridStyle}>
          {[
            { name: 'Light (300)', var: '--font-weight-light' },
            { name: 'Normal (400)', var: '--font-weight-normal' },
            { name: 'Medium (500)', var: '--font-weight-medium' },
            { name: 'Semibold (600)', var: '--font-weight-semibold' },
            { name: 'Bold (700)', var: '--font-weight-bold' },
            { name: 'Extra Bold (800)', var: '--font-weight-extrabold' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ fontWeight: `var(${varName})`, marginBottom: '0.5rem' }}>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Font Sizes</h3>
        <div style={gridStyle}>
          {[
            { name: 'Extra Small (10px)', var: '--font-size-xs' },
            { name: 'Small (11px)', var: '--font-size-sm' },
            { name: 'Base (12px)', var: '--font-size-base' },
            { name: 'Medium (13px)', var: '--font-size-md' },
            { name: 'Large (14px)', var: '--font-size-lg' },
            { name: 'Extra Large (16px)', var: '--font-size-xl' },
            { name: '2X Large (18px)', var: '--font-size-2xl' },
            { name: '3X Large (20px)', var: '--font-size-3xl' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ fontSize: `var(${varName})`, marginBottom: '0.5rem' }}>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Spacing</h2>
        <div style={gridStyle}>
          {[
            { size: '4px', var: '--space-1' },
            { size: '8px', var: '--space-2' },
            { size: '12px', var: '--space-3' },
            { size: '16px', var: '--space-4' },
            { size: '24px', var: '--space-6' },
            { size: '32px', var: '--space-8' },
          ].map(({ size, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ 
                width: `var(${varName})`, 
                height: `var(${varName})`, 
                backgroundColor: 'var(--accent-primary)', 
                marginBottom: '0.5rem' 
              }}></div>
              <div>{size}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Colors</h2>
        
        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Game Semantic Colors</h3>
        <div style={gridStyle}>
          {[
            { name: 'Human', var: '--color-human' },
            { name: 'Aligned', var: '--color-aligned' },
            { name: 'AI', var: '--color-ai' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ 
                width: '100%', 
                height: '60px', 
                backgroundColor: `var(${varName})`, 
                borderRadius: 'var(--radius-md)', 
                marginBottom: '0.5rem' 
              }}></div>
              <div>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Status Colors</h3>
        <div style={gridStyle}>
          {[
            { name: 'Success', var: '--color-success' },
            { name: 'Danger', var: '--color-danger' },
            { name: 'Warning', var: '--color-warning' },
            { name: 'Info', var: '--color-info' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ 
                width: '100%', 
                height: '60px', 
                backgroundColor: `var(${varName})`, 
                borderRadius: 'var(--radius-md)', 
                marginBottom: '0.5rem' 
              }}></div>
              <div>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Accent Colors</h3>
        <div style={gridStyle}>
          {[
            { name: 'Primary', var: '--accent-primary' },
            { name: 'Amber', var: '--accent-amber' },
            { name: 'Cyan', var: '--accent-cyan' },
            { name: 'Magenta', var: '--accent-magenta' },
            { name: 'Green', var: '--accent-green' },
            { name: 'Blue', var: '--accent-blue' },
            { name: 'Red', var: '--accent-red' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ 
                width: '100%', 
                height: '60px', 
                backgroundColor: `var(${varName})`, 
                borderRadius: 'var(--radius-md)', 
                marginBottom: '0.5rem' 
              }}></div>
              <div>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Border Radius</h2>
        <div style={gridStyle}>
          {[
            { name: 'Small (3px)', var: '--radius-sm' },
            { name: 'Base (4px)', var: '--radius-base' },
            { name: 'Medium (6px)', var: '--radius-md' },
            { name: 'Large (8px)', var: '--radius-lg' },
            { name: 'Extra Large (12px)', var: '--radius-xl' },
            { name: 'Full (50%)', var: '--radius-full' },
          ].map(({ name, var: varName }) => (
            <div key={varName} style={cardStyle}>
              <div style={{ 
                width: '60px', 
                height: '60px', 
                backgroundColor: 'var(--accent-primary)', 
                borderRadius: `var(${varName})`, 
                marginBottom: '0.5rem' 
              }}></div>
              <div>{name}</div>
              <code>{varName}</code>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: '48px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: '600', marginBottom: '24px' }}>Usage</h2>
        
        <p style={{ marginBottom: '16px', color: '#64748b', lineHeight: '1.6' }}>
          All of these CSS variables are available globally and can be used in any component. 
          The design system supports both light and dark themes, which automatically switch the appropriate color variables.
        </p>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Example Usage</h3>
        <pre style={{ 
          backgroundColor: '#f8fafc', 
          padding: '16px', 
          borderRadius: '8px', 
          border: '1px solid #e2e8f0',
          fontFamily: 'JetBrains Mono, monospace',
          fontSize: '14px',
          marginBottom: '24px',
          overflow: 'auto'
        }}>
{`.my-component {
  background: var(--bg-primary);
  color: var(--text-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
}`}
        </pre>

        <h3 style={{ fontSize: '18px', fontWeight: '600', marginBottom: '16px' }}>Theme Switching</h3>
        <p style={{ color: '#64748b', lineHeight: '1.6' }}>
          The application automatically switches between light and dark themes based on the <code>data-theme</code> attribute on the <code>html</code> element:
        </p>
        <ul style={{ marginTop: '12px', color: '#64748b', lineHeight: '1.6' }}>
          <li><code>data-theme="light"</code> → Light theme</li>
          <li><code>data-theme="dark"</code> → Dark theme (default)</li>
        </ul>
      </section>
    </div>
  );
};

export const Default = () => <DesignSystemPage />;