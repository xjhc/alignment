import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Button } from './ui';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log the error for debugging
    console.error('Error Boundary caught an error:', error, errorInfo);
    
    this.setState({
      error,
      errorInfo
    });
  }

  handleLogout = () => {
    try {
      // Clear all session data
      sessionStorage.clear();
      localStorage.clear();
      
      // Clear WebSocket credentials if they exist
      try {
        sessionStorage.removeItem('alignmentGameSession');
        sessionStorage.removeItem('wsConnectionCredentials');
      } catch (e) {
        console.warn('Failed to clear session storage:', e);
      }
      
      // Reload the page to start fresh
      window.location.href = '/login';
    } catch (e) {
      // If we can't even do that, just reload the page
      console.error('Failed to logout cleanly:', e);
      window.location.reload();
    }
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
          <div className="flex flex-col gap-6 items-center max-w-lg mx-4 text-center">
            {/* Error Icon */}
            <div className="text-6xl mb-4">⚠️</div>
            
            {/* Error Title */}
            <h1 className="text-2xl font-bold text-danger">
              System Error Detected
            </h1>
            
            {/* Error Description */}
            <div className="bg-danger-bg border border-danger-border rounded-lg p-4 text-left w-full">
              <h2 className="text-sm font-semibold text-danger-text mb-2">
                Critical Application Error
              </h2>
              <p className="text-sm text-danger-text mb-3">
                The application encountered an unexpected error and cannot continue. 
                This could be due to a temporary system issue or corrupted session data.
              </p>
              
              {/* Error Details (collapsed by default) */}
              {this.state.error && (
                <details className="mt-3">
                  <summary className="text-xs text-danger-text cursor-pointer hover:text-danger">
                    Technical Details
                  </summary>
                  <pre className="text-xs text-danger-text mt-2 p-2 bg-danger/20 rounded overflow-auto max-h-32">
                    {this.state.error.toString()}
                    {this.state.errorInfo?.componentStack}
                  </pre>
                </details>
              )}
            </div>
            
            {/* Recovery Options */}
            <div className="flex flex-col gap-3 w-full">
              <Button
                variant="primary"
                onClick={this.handleLogout}
                className="w-full"
              >
                🚪 Clear Session & Return to Login
              </Button>
              
              <Button
                variant="secondary"
                onClick={this.handleReload}
                className="w-full"
              >
                🔄 Reload Application
              </Button>
            </div>
            
            {/* Help Text */}
            <p className="text-xs text-text-secondary max-w-md">
              If this error persists, try clearing your browser cache or contact support.
              Your game progress may be recovered by rejoining with the same credentials.
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}