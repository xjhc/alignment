import type { Preview, Decorator } from '@storybook/react';
import React from 'react';
import '../src/index.css';
import { withProviders } from './decorators';
import { StorybookProviders } from '../src/stories/util/StorybookProviders';
import { useTheme } from '../src/hooks/useTheme';

// New decorator that wraps stories with our centralized mock provider
const withStorybookProviders: Decorator = (Story, context) => {
  // Get custom state from story args
  const { storyGameState, localPlayerId, storySessionContext } = context.args;

  return (
    <StorybookProviders 
      storyGameState={storyGameState} 
      localPlayerId={localPlayerId}
      storySessionContext={storySessionContext}
    >
      <Story />
    </StorybookProviders>
  );
};

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
  // withProviders gives us Theme, etc.
  // Then withStorybookProviders gives us the game-specific contexts.
  decorators: [withProviders, withStorybookProviders, withGlobalStyles],
};

export default preview;