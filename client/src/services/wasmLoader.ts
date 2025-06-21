// WebAssembly loader for the Alignment game core
declare global {
  interface Window {
    Go?: any;
    AlignmentCore?: AlignmentCore;
    wasmReady?: () => void;
  }
}

// Types for the WASM bridge
export interface AlignmentCore {
  createGame: (gameId: string) => WasmResult;
  applyEvent: (eventJson: string) => WasmResult;
  getGameState: () => string;
  canPlayerVote: (playerId: string, phaseType: string) => boolean;
  canPlayerAffordAbility: (playerId: string) => boolean;
  checkWinCondition: () => string | null;
  calculateMiningSuccess: (playerId: string, difficulty: number) => boolean;
  isValidNightActionTarget: (actorId: string, targetId: string, actionType: string) => boolean;
  getVoteWinner: (threshold: number) => VoteWinnerResult;
  isGamePhaseOver: () => boolean;
  setStateCallback: (callbackName: string, callback: (stateJson: string) => void) => WasmResult;
  serializeGameState: () => string;
  deserializeGameState: (stateJson: string) => WasmResult;
}

export interface WasmResult {
  success?: boolean;
  error?: string;
  gameId?: string;
}

export interface VoteWinnerResult {
  winner: string;
  hasWinner: boolean;
}

class WasmLoader {
  private wasmInstance: WebAssembly.Instance | null = null;
  private isLoaded = false;
  private loadPromise: Promise<void> | null = null;
  private stateChangeCallbacks: ((stateJson: string) => void)[] = [];

  async load(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }

    this.loadPromise = this.loadWasm();
    return this.loadPromise;
  }

  private async loadWasm(): Promise<void> {
    try {
      // Load the WASM support script
      if (!window.Go) {
        await this.loadScript('/wasm_exec.js');
      }

      // Create Go instance
      const go = new window.Go();

      // Set up the ready callback
      window.wasmReady = () => {
        this.isLoaded = true;
        this.setupCallbacks();
        console.log('Alignment Core WASM loaded successfully');
      };

      // Load and instantiate the WASM module
      const wasmResponse = await fetch('/core.wasm');
      const wasmBytes = await wasmResponse.arrayBuffer();
      const wasmModule = await WebAssembly.instantiate(wasmBytes, go.importObject);
      
      this.wasmInstance = wasmModule.instance;

      // Run the Go program
      go.run(this.wasmInstance);

      // Wait for the WASM to signal it's ready
      return new Promise((resolve, reject) => {
        const checkReady = () => {
          if (this.isLoaded) {
            resolve();
          } else {
            setTimeout(checkReady, 10);
          }
        };
        checkReady();
        
        // Timeout after 5 seconds
        setTimeout(() => {
          if (!this.isLoaded) {
            reject(new Error('WASM loading timeout'));
          }
        }, 5000);
      });
    } catch (error) {
      console.error('Failed to load WASM:', error);
      throw error;
    }
  }

  private loadScript(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = src;
      script.onload = () => resolve();
      script.onerror = () => reject(new Error(`Failed to load script: ${src}`));
      document.head.appendChild(script);
    });
  }

  private setupCallbacks(): void {
    if (window.AlignmentCore) {
      // Set up state change callback
      window.AlignmentCore.setStateCallback('stateChange', (stateJson: string) => {
        this.stateChangeCallbacks.forEach(callback => {
          try {
            callback(stateJson);
          } catch (error) {
            console.error('Error in state change callback:', error);
          }
        });
      });
    }
  }

  getCore(): AlignmentCore {
    if (!this.isLoaded || !window.AlignmentCore) {
      throw new Error('WASM not loaded. Call load() first.');
    }
    return window.AlignmentCore;
  }

  isReady(): boolean {
    return this.isLoaded && !!window.AlignmentCore;
  }

  onStateChange(callback: (stateJson: string) => void): () => void {
    this.stateChangeCallbacks.push(callback);
    
    // Return unsubscribe function
    return () => {
      const index = this.stateChangeCallbacks.indexOf(callback);
      if (index > -1) {
        this.stateChangeCallbacks.splice(index, 1);
      }
    };
  }
}

// Singleton instance
export const wasmLoader = new WasmLoader();