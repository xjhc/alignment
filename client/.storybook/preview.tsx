import type { Preview, Decorator } from '@storybook/react-vite'
import React from 'react';
import '../src/global.css';

// Create a decorator to wrap stories with your app's global context
const withGlobalStyles: Decorator = (Story) => {
  // You can set the theme here. Forcing dark for consistency.
  React.useEffect(() => {
    document.documentElement.setAttribute('data-theme', 'dark');
  }, []);

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
  // Apply the decorator to all stories
  decorators: [withGlobalStyles],
};

export default preview;