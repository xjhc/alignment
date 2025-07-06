import { BrowserRouter } from "react-router-dom";
import { ErrorBoundary } from "react-error-boundary";
import {
  ThemeProvider,
  GameEngineProvider,
  WebSocketProvider,
} from ".";

function ErrorFallback({ error, resetErrorBoundary }: any) {
  return (
    <div role="alert" className="launch-screen">
      <div className="launch-form">
        <h2>Something went wrong:</h2>
        <pre style={{ color: "red" }}>{error.message}</pre>
        <button className="btn-primary" onClick={resetErrorBoundary}>
          Try again
        </button>
      </div>
    </div>
  );
}

interface AppProvidersProps {
  children: React.ReactNode;
}

export function AppProviders({ children }: AppProvidersProps) {
  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onReset={() => window.location.reload()}
    >
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