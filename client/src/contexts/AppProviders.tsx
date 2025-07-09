import { BrowserRouter } from "react-router-dom";
import { ErrorBoundary } from "../components/ErrorBoundary";
import {
  ThemeProvider,
  GameEngineProvider,
  WebSocketProvider,
} from ".";
import { NotificationProvider } from "./NotificationContext";

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
              <NotificationProvider>
                {children}
              </NotificationProvider>
            </WebSocketProvider>
          </GameEngineProvider>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}