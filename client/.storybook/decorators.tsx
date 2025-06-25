import React from 'react';
import type { Decorator } from '@storybook/react';
import { GameEngineProvider } from '../src/contexts/GameEngineContext';
import { WebSocketProvider } from '../src/contexts/WebSocketContext';
import { ThemeProvider } from '../src/contexts/ThemeContext';

/**
 * A global decorator to wrap all stories with application-wide context providers.
 */
export const withProviders: Decorator = (Story) => {
  return (
    <ThemeProvider>
      <GameEngineProvider>
        <WebSocketProvider>
          <Story />
        </WebSocketProvider>
      </GameEngineProvider>
    </ThemeProvider>
  );
};