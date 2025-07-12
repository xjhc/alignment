import { BrowserRouter } from "react-router-dom";
import { ErrorBoundary } from "../components/ErrorBoundary";
import {
  ThemeProvider,
  GameEngineProvider,
  WebSocketProvider,
  SessionProvider,
} from ".";
import { NotificationProvider } from "./NotificationContext";
import { useSessionManager } from "../hooks/useSessionManager";

interface AppProvidersProps {
  children: React.ReactNode;
}

function SessionProviderWrapper({ children }: { children: React.ReactNode }) {
  const { state, dispatch, gameEngineLoading, gameEngineError, isConnected, sessionActions } = useSessionManager();

  const sessionContextValue = {
    appState: state.appState,
    sessionState: state.sessionState,
    lobbyState: state.lobbyState,
    gameState: state.gameState,
    roleAssignment: state.roleAssignment,
    gameAnalysis: state.gameAnalysis,
    gameUIState: state.gameUIState,
    isConnected,
    dispatch,
    gameEngineLoading,
    gameEngineError,
    ...sessionActions,
  };

  return (
    <SessionProvider value={sessionContextValue}>
      {children}
    </SessionProvider>
  );
}

export function AppProviders({ children }: AppProvidersProps) {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ThemeProvider>
          <GameEngineProvider>
            <WebSocketProvider>
              <NotificationProvider>
                <SessionProviderWrapper>
                  {children}
                </SessionProviderWrapper>
              </NotificationProvider>
            </WebSocketProvider>
          </GameEngineProvider>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}