import type { Preview, Decorator } from '@storybook/react';
import React from 'react';
import '../src/index.css';
import { withProviders } from './decorators';
import { useTheme } from '../src/hooks/useTheme';

// Create a decorator to wrap stories with your app's global context
const withGlobalStyles: Decorator = (Story, context) => {
  const { theme } = useTheme();

  // Set the theme on the document element for global styles
  React.useEffect(() => {
    // The theme is managed by the ThemeProvider/useTheme hook
    // This just ensures the storybook background matches
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  // This provides a root div like your actual app, which can be useful for layout styles
  return (
    <div id="story-root" style={{ minHeight: '100vh', background: 'var(--bg-primary)'}}>
      <Story />
    </div>
  );
};

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
       color: /(background|color)$/i,
       date: /Date$/i,
      },
    },
  },
  // Apply the decorators to all stories. The order matters.
  decorators: [withProviders, withGlobalStyles],
};

export default preview;