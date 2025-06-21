import { useState } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';

export function WasmTestScreen() {
  const [testResults, setTestResults] = useState<string[]>([]);
  const { 
    isLoaded, 
    isLoading, 
    error, 
    createGame, 
    gameState,
    canPlayerVote,
    checkWinCondition,
    calculateMiningSuccess
  } = useGameEngine();

  const addResult = (message: string) => {
    setTestResults(prev => [...prev, `${new Date().toLocaleTimeString()}: ${message}`]);
  };

  const runTests = async () => {
    setTestResults([]);
    addResult('Starting WASM integration tests...');

    try {
      // Test 1: Create a game
      addResult('Test 1: Creating game...');
      await createGame('test-game-123');
      addResult('✅ Game created successfully');

      // Test 2: Check game state
      addResult('Test 2: Checking game state...');
      if (gameState) {
        addResult(`✅ Game state loaded: ID=${gameState.id}, Day=${gameState.dayNumber}`);
      } else {
        addResult('❌ No game state available');
      }

      // Test 3: Test rule functions
      addResult('Test 3: Testing rule functions...');
      const canVote = canPlayerVote('test-player', 'NOMINATION');
      addResult(`✅ canPlayerVote: ${canVote}`);

      const winCondition = checkWinCondition();
      addResult(`✅ checkWinCondition: ${winCondition ? 'Has condition' : 'No condition'}`);

      const miningSuccess = calculateMiningSuccess('test-player', 0.3);
      addResult(`✅ calculateMiningSuccess: ${miningSuccess}`);

      addResult('🎉 All tests completed successfully!');
    } catch (error) {
      addResult(`❌ Test failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  if (isLoading) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Loading WASM Engine...</h2>
          <div className="loading-spinner">⏳</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>WASM Engine Error</h2>
          <p style={{ color: 'var(--color-danger)' }}>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="launch-screen">
      <div className="launch-form" style={{ width: '600px' }}>
        <h2>WASM Integration Test</h2>
        
        <div style={{ marginBottom: '16px' }}>
          <p>Engine Status: {isLoaded ? '✅ Loaded' : '❌ Not Loaded'}</p>
          {gameState && (
            <p>Game State: {gameState.id} (Day {gameState.dayNumber})</p>
          )}
        </div>

        <button 
          className="btn-primary" 
          onClick={runTests}
          disabled={!isLoaded}
        >
          Run WASM Tests
        </button>

        {testResults.length > 0 && (
          <div style={{
            marginTop: '24px',
            padding: '16px',
            background: 'var(--bg-secondary)',
            borderRadius: '8px',
            textAlign: 'left',
            maxHeight: '300px',
            overflowY: 'auto'
          }}>
            <h3>Test Results:</h3>
            {testResults.map((result, index) => (
              <div key={index} style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '11px',
                margin: '4px 0',
                color: result.includes('❌') ? 'var(--color-danger)' : 
                       result.includes('✅') ? 'var(--color-success)' : 
                       'var(--text-primary)'
              }}>
                {result}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}