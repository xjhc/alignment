import { BrowserRouter } from "react-router-dom";
import { ErrorBoundary } from "../components/ErrorBoundary";
import {
  ThemeProvider,
  GameEngineProvider,
  WebSocketProvider,
} from ".";

interface AppProvidersProps {
  children: React.ReactNode;
}

export function AppProviders({ children }: AppProvidersProps) {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ThemeProvider>
          <GameEngineProvider>
            <WebSocketProvider>
              {children}
            </WebSocketProvider>
          </GameEngineProvider>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}