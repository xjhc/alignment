import { useState, useEffect, useRef } from "react";
import { Button, Modal } from "./ui";
import { FriendsPanel } from "./FriendsPanel";
import { PartyPanel } from "./PartyPanel";
import { getUserIdForApi } from "../services/guestIdentity";
import { websocketClient } from "../services/websocket";
import { GameSettings, LobbyInfo } from "../types";
import { applyStaggeredAnimation, SLIDE_IN_UP } from "../utils/animations";
import { CreateGameModal } from "./CreateGameModal";

interface LobbyListScreenProps {
  playerName: string;
  playerAvatar?: string;
  onJoinLobby: (gameId: string, playerId: string, sessionToken: string, lobbyName?: string) => void;
  onCreateGame: (
    gameId: string,
    playerId: string,
    sessionToken: string
  ) => void;
  onSpectateGame: (gameId: string, playerId: string, sessionToken: string, lobbyName?: string) => void;
  onBack: () => void;
}

function LobbyCard({
  lobby,
  onJoin,
  onSpectate,
  joinCooldowns,
}: {
  lobby: LobbyInfo;
  onJoin: (id: string) => void;
  onSpectate: (id: string) => void;
  joinCooldowns: Record<string, number>;
}) {
  const now = Date.now();
  const cooldownEnd = joinCooldowns[lobby.id];
  const inCooldown = cooldownEnd && now < cooldownEnd;
  const remainingSeconds = inCooldown
    ? Math.ceil((cooldownEnd - now) / 1000)
    : 0;

  const getStatusChip = () => {
    switch (lobby.status) {
      case "WAITING":
        return (
          <div className="px-2 py-1 text-xs font-semibold uppercase rounded-full bg-success/20 text-success">
            Waiting
          </div>
        );
      case "IN_PROGRESS":
        return (
          <div className="px-2 py-1 text-xs font-semibold uppercase rounded-full bg-amber/20 text-amber">
            In Progress
          </div>
        );
      default:
        return (
          <div className="px-2 py-1 text-xs font-semibold uppercase rounded-full bg-text-muted/20 text-text-muted">
            {lobby.status}
          </div>
        );
    }
  };

  return (
    <div
      className={`p-4 bg-background-secondary rounded-lg border border-border transition-all duration-200 hover:border-primary hover:shadow-lg hover:bg-background-tertiary group ${SLIDE_IN_UP}`}
    >
      <div className="flex justify-between items-start">
        <div>
          <h3 className="font-semibold text-text-primary mb-1">{lobby.name}</h3>
          <p className="text-xs text-text-muted font-mono">
            ID: {lobby.id.substring(0, 8)}
          </p>
        </div>
        {getStatusChip()}
      </div>
      <div className="mt-4 flex items-center justify-between">
        <div className="flex items-center gap-4 text-sm">
          <div className="flex items-center gap-1.5 text-text-secondary">
            <span className="text-base">👥</span>
            <span>
              {lobby.player_count} / {lobby.max_players}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {lobby.game_settings?.playAsAI && (
              <span
                className="text-xs font-semibold bg-ai/20 text-ai px-2 py-1 rounded"
                title="Play as AI Mode"
              >
                🤖 AI
              </span>
            )}
            {lobby.game_settings?.initialAlignedHumanCount && lobby.game_settings.initialAlignedHumanCount > 0 && (
              <span
                className="text-xs font-semibold bg-aligned/20 text-aligned px-2 py-1 rounded"
                title={`${lobby.game_settings.initialAlignedHumanCount} starting Aligned player(s)`}
              >
                🕵️ +{lobby.game_settings.initialAlignedHumanCount}
              </span>
            )}
          </div>
        </div>
        {lobby.status === "IN_PROGRESS" ? (
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onSpectate(lobby.id)}
            disabled={inCooldown}
            className={`text-sm font-medium transition-all duration-200 ${
              inCooldown 
                ? 'cursor-not-allowed opacity-50' 
                : 'hover:shadow-md group-hover:bg-secondary-dark'
            }`}
            title={
              inCooldown
                ? `Please wait ${remainingSeconds}s`
                : `Spectate ${lobby.name}`
            }
          >
            {inCooldown ? (
              <span className="flex items-center gap-1">
                <span className="animate-spin">⏳</span>
                {remainingSeconds}s
              </span>
            ) : (
              <span className="flex items-center gap-1">
                👁️ Spectate
              </span>
            )}
          </Button>
        ) : (
          <Button
            variant="primary"
            size="sm"
            onClick={() => onJoin(lobby.id)}
            disabled={!lobby.can_join || inCooldown}
            className={`text-sm font-medium transition-all duration-200 ${
              inCooldown 
                ? 'cursor-not-allowed opacity-50' 
                : lobby.can_join 
                  ? 'hover:shadow-md group-hover:bg-primary-dark' 
                  : 'cursor-not-allowed opacity-50'
            }`}
            title={
              inCooldown
                ? `Please wait ${remainingSeconds}s`
                : !lobby.can_join
                  ? "Lobby is full or in progress"
                  : `Join ${lobby.name}`
            }
          >
            {inCooldown ? (
              <span className="flex items-center gap-1">
                <span className="animate-spin">⏳</span>
                {remainingSeconds}s
              </span>
            ) : (
              "Join"
            )}
          </Button>
        )}
      </div>
    </div>
  );
}

export function LobbyListScreen({
  playerName,
  playerAvatar,
  onJoinLobby,
  onCreateGame,
  onSpectateGame,
  onBack,
}: LobbyListScreenProps) {
  const [lobbies, setLobbies] = useState<LobbyInfo[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showFriendsPanel, setShowFriendsPanel] = useState(false);
  const [showPartyPanel, setShowPartyPanel] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [pollingInterval, setPollingInterval] = useState(10000); // Increased to 10 seconds to reduce server load
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const [showSessionConflict, setShowSessionConflict] = useState(false);
  const [conflictDetails, setConflictDetails] = useState<{gameId?: string, sessionState?: string} | null>(null);
  const [joinCooldowns, setJoinCooldowns] = useState<Record<string, number>>(
    {}
  );
  const lobbyListRef = useRef<HTMLDivElement>(null);

  const clearSession = async () => {
    try {
      console.log("Clearing existing session...");
      
      // First, try to abandon the session via WebSocket if we have connection details
      const savedSession = localStorage.getItem("alignmentGameSession");
      if (savedSession) {
        try {
          const sessionData = JSON.parse(savedSession);
          if (sessionData.gameId && sessionData.playerId && sessionData.sessionToken) {
            // Attempt to send abandon action via WebSocket to cleanly leave
            websocketClient.connect(sessionData.gameId, sessionData.playerId, sessionData.sessionToken)
              .then(() => {
                websocketClient.sendAction({
                  type: "ABANDON_GAME",
                  payload: {
                    game_id: sessionData.gameId,
                    player_id: sessionData.playerId,
                  },
                });
                // Give it a moment to process, then disconnect
                setTimeout(() => websocketClient.disconnect(), 1000);
              })
              .catch((err) => {
                console.log("Could not connect to send abandon action:", err);
                websocketClient.disconnect();
              });
          }
        } catch (parseError) {
          console.log("Could not parse session data for clean abandon:", parseError);
        }
      }
      
      // Clear local storage regardless
      localStorage.removeItem("alignmentGameSession");
      websocketClient.disconnect();
      setShowSessionConflict(false);
      setConflictDetails(null);
      setError(null);
      await fetchLobbies();
    } catch (error) {
      console.error("Error clearing session:", error);
      setError("Failed to clear session. Please try refreshing the page.");
    }
  };

  const rejoinSession = () => {
    try {
      const savedSession = localStorage.getItem("alignmentGameSession");
      if (savedSession) {
        const sessionData = JSON.parse(savedSession);
        if (
          sessionData.gameId &&
          sessionData.playerId &&
          sessionData.sessionToken
        ) {
          console.log("Rejoining existing session:", sessionData);
          onJoinLobby(
            sessionData.gameId,
            sessionData.playerId,
            sessionData.sessionToken
          );
          setShowSessionConflict(false);
          return;
        }
      }
      setError("Unable to rejoin: No valid session data found");
      setShowSessionConflict(false);
    } catch (error) {
      console.error("Error rejoining session:", error);
      setError("Failed to rejoin session");
      setShowSessionConflict(false);
    }
  };

  const fetchLobbies = async (isRetry = false) => {
    try {
      if (!isRetry) setError(null);
      const response = await fetch("/api/games");

      if (response.status === 429) {
        const newRetryCount = retryCount + 1;
        const backoffDelay = Math.min(1000 * Math.pow(2, newRetryCount), 60000);
        setRetryCount(newRetryCount);
        setPollingInterval(Math.max(backoffDelay, 30000));
        console.warn(`Rate limited. Backing off for ${backoffDelay}ms.`);
        restartPolling();
        return;
      }

      if (!response.ok)
        throw new Error(`Failed to fetch lobbies: ${response.statusText}`);

      const data = await response.json();
      const newLobbies = data.lobbies || [];

      setLobbies(newLobbies);

      setTimeout(() => {
        if (lobbyListRef.current) {
          applyStaggeredAnimation(
            Array.from(lobbyListRef.current.children),
            50,
            SLIDE_IN_UP
          );
        }
      }, 50);

      if (retryCount > 0) {
        setRetryCount(0);
        setPollingInterval(5000);
        restartPolling();
      }
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Failed to fetch lobbies"
      );
    } finally {
      setIsLoading(false);
    }
  };

  const restartPolling = () => {
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => fetchLobbies(), pollingInterval);
  };

  useEffect(() => {
    // Check for existing active session on component mount
    const savedSession = localStorage.getItem("alignmentGameSession");
    if (savedSession) {
      try {
        const sessionData = JSON.parse(savedSession);
        if (
          sessionData.gameId &&
          sessionData.playerId &&
          sessionData.sessionToken &&
          sessionData.sessionState
        ) {
          console.log("Found existing session, showing conflict modal:", sessionData);
          setConflictDetails({ 
            gameId: sessionData.gameId, 
            sessionState: sessionData.sessionState || "unknown" 
          });
          setShowSessionConflict(true);
        }
      } catch (error) {
        console.error("Error parsing saved session data:", error);
        localStorage.removeItem("alignmentGameSession");
      }
    }
    
    fetchLobbies();
    restartPolling();
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, []);

  const handleJoinLobby = async (gameId: string) => {
    try {
      setError(null);
      const now = Date.now();
      const cooldownEnd = joinCooldowns[gameId];
      if (cooldownEnd && now < cooldownEnd) {
        const remainingSeconds = Math.ceil((cooldownEnd - now) / 1000);
        setError(
          `Please wait ${remainingSeconds} seconds before trying to join this lobby again.`
        );
        return;
      }

      const userId = getUserIdForApi();
      const response = await fetch(`/api/games/${gameId}/join`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          user_id: userId,
          player_name: playerName,
          player_avatar: playerAvatar || "",
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        if (
          response.status === 409 &&
          errorText.includes("already in an active game session")
        ) {
          // Extract game ID from error message if possible
          const gameIdMatch = errorText.match(/game: ([a-f0-9-]+)/);
          const conflictGameId = gameIdMatch ? gameIdMatch[1] : undefined;
          setConflictDetails({ gameId: conflictGameId, sessionState: "join_conflict" });
          setShowSessionConflict(true);
        } else if (response.status === 429) {
          setJoinCooldowns((prev) => ({ ...prev, [gameId]: now + 5000 }));
          setError(
            "Too many join attempts. Please wait a moment before trying again."
          );
          return;
        }
        throw new Error(errorText || "Failed to join lobby");
      }

      setJoinCooldowns((prev) => {
        const updated = { ...prev };
        delete updated[gameId];
        return updated;
      });

      const data = await response.json();
      
      // Find the lobby name from the current lobby list
      const lobby = lobbies.find(l => l.id === gameId);
      const lobbyName = lobby?.name || "";
      
      onJoinLobby(data.game_id, data.player_id, data.session_token, lobbyName);
    } catch (error) {
      setError(error instanceof Error ? error.message : "Failed to join lobby");
    }
  };

  const handleSpectateGame = async (gameId: string) => {
    try {
      setError(null);
      const now = Date.now();
      const cooldownEnd = joinCooldowns[gameId];
      if (cooldownEnd && now < cooldownEnd) {
        const remainingSeconds = Math.ceil((cooldownEnd - now) / 1000);
        setError(
          `Please wait ${remainingSeconds} seconds before trying to spectate this game again.`
        );
        return;
      }

      const userId = getUserIdForApi();
      const response = await fetch(`/api/games/${gameId}/spectate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          user_id: userId,
          spectator_name: playerName,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        if (
          response.status === 409 &&
          errorText.includes("already in an active game session")
        ) {
          // Extract game ID from error message if possible
          const gameIdMatch = errorText.match(/game: ([a-f0-9-]+)/);
          const conflictGameId = gameIdMatch ? gameIdMatch[1] : undefined;
          setConflictDetails({ gameId: conflictGameId, sessionState: "spectate_conflict" });
          setShowSessionConflict(true);
        } else if (response.status === 429) {
          setJoinCooldowns((prev) => ({ ...prev, [gameId]: now + 5000 }));
          setError(
            "Too many spectate attempts. Please wait a moment before trying again."
          );
          return;
        }
        throw new Error(errorText || "Failed to join as spectator");
      }

      setJoinCooldowns((prev) => {
        const updated = { ...prev };
        delete updated[gameId];
        return updated;
      });

      const data = await response.json();
      
      // Find the lobby name from the current lobby list
      const lobby = lobbies.find(l => l.id === gameId);
      const lobbyName = lobby?.name || "";
      
      onSpectateGame(data.game_id, data.player_id, data.session_token, lobbyName);
    } catch (error) {
      setError(error instanceof Error ? error.message : "Failed to join as spectator");
    }
  };

  const handleCreateGame = async (
    settings: Partial<GameSettings> & { lobbyName: string }
  ) => {
    try {
      setError(null);
      const userId = getUserIdForApi();
      const response = await fetch("/api/games", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          user_id: userId,
          lobby_name: settings.lobbyName,
          player_name: playerName,
          player_avatar: playerAvatar || "",
          play_as_ai: settings.playAsAI,
          initial_aligned_human_count: settings.initialAlignedHumanCount,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        if (
          response.status === 409 &&
          errorText.includes("already in an active game session")
        ) {
          // Extract game ID from error message if possible
          const gameIdMatch = errorText.match(/game: ([a-f0-9-]+)/);
          const conflictGameId = gameIdMatch ? gameIdMatch[1] : undefined;
          setConflictDetails({ gameId: conflictGameId, sessionState: "create_conflict" });
          setShowSessionConflict(true);
        }
        
        // Provide user-friendly error messages for common validation issues
        if (errorText.includes("lobby_name contains invalid characters")) {
          throw new Error("Lobby name contains invalid characters. Please avoid using < > ' \" & symbols.");
        }
        if (errorText.includes("lobby_name too long")) {
          throw new Error("Lobby name is too long. Please keep it under 100 characters.");
        }
        
        throw new Error(errorText || "Failed to create game");
      }

      const data = await response.json();
      onCreateGame(data.game_id, data.player_id, data.session_token, settings.lobbyName);
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Failed to create game"
      );
    }
  };

  return (
    <>
      <div className="w-screen min-h-screen flex flex-col items-center bg-background-primary text-text-primary p-6">
        <header className="w-full max-w-5xl mx-auto flex justify-between items-center mb-8">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-primary flex items-center justify-center text-background-primary text-lg font-semibold border-2 border-primary/50 shadow-lg">
              {playerAvatar || playerName.charAt(0).toUpperCase()}
            </div>
            <div>
              <span className="font-semibold text-text-primary">
                {playerName}
              </span>
              <p className="text-xs text-text-muted">
                Emergency Response Agent
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setShowFriendsPanel(true)}
              className="text-sm"
            >
              👥 Friends
            </Button>
            <Button 
              variant="secondary" 
              size="sm" 
              onClick={() => setShowPartyPanel(true)}
              className="text-sm"
            >
              🎉 Party
            </Button>
            <Button variant="ghost" size="sm" onClick={onBack} className="text-sm">
              Logout
            </Button>
          </div>
        </header>

        <main className="w-full max-w-5xl mx-auto">
          <div className="flex justify-between items-center mb-6">
            <div>
              <h1 className="text-2xl font-bold text-text-primary mb-1">
                Active Lobbies
              </h1>
              <p className="text-sm text-text-muted">
                {lobbies.length} active session{lobbies.length !== 1 ? 's' : ''} available
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="primary"
                onClick={() => setShowCreateModal(true)}
                className="font-semibold"
              >
                + Create Game
              </Button>
            </div>
          </div>

          {error && (
            <div className="my-4 p-3 bg-danger/10 text-danger rounded-md border border-danger/20">
              {error}
            </div>
          )}

          {showSessionConflict && (
            <div className="bg-warning/10 border border-warning/20 rounded-lg p-4 my-4">
              <h3 className="text-warning font-semibold mb-2">
                🚨 Active Session Detected
              </h3>
              <div className="text-text-secondary text-sm mb-4">
                <p className="mb-2">
                  You're already in an active game session. This might be due to:
                </p>
                <ul className="list-disc ml-4 space-y-1">
                  <li>A previous session that didn't close properly</li>
                  <li>Another browser tab or device connected to the same account</li>
                  <li>Network issues that prevented clean disconnection</li>
                </ul>
                {conflictDetails?.gameId && (
                  <p className="mt-2 text-xs font-mono bg-background-secondary p-2 rounded">
                    Conflicting Game ID: {conflictDetails.gameId.substring(0, 12)}...
                  </p>
                )}
              </div>
              <div className="flex gap-2 flex-wrap">
                <Button variant="primary" size="sm" onClick={rejoinSession}>
                  🔗 Rejoin Previous Game
                </Button>
                <Button variant="danger" size="sm" onClick={clearSession}>
                  🧹 Force Clear & Start Fresh
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setShowSessionConflict(false);
                    setConflictDetails(null);
                  }}
                >
                  Cancel
                </Button>
              </div>
              <p className="text-xs text-text-muted mt-3">
                💡 "Force Clear" will attempt to properly abandon your previous session
              </p>
            </div>
          )}

          {isLoading ? (
            <div className="text-center py-20">
              <div className="animate-pulse">
                <div className="text-4xl mb-4">🔍</div>
                <p className="text-text-muted mb-4">Scanning for active sessions...</p>
                <div className="text-xs text-text-muted">Connecting to Loebian Inc. network</div>
              </div>
            </div>
          ) : lobbies.length === 0 ? (
            <div className="text-center py-20 bg-background-secondary rounded-lg border-2 border-dashed border-border transition-all duration-300 hover:border-primary/50">
              <div className="text-4xl mb-4 animate-pulse">🏢</div>
              <h3 className="text-lg font-medium text-text-primary mb-2">
                No Active Emergency Sessions
              </h3>
              <p className="text-text-secondary text-sm mb-6 max-w-md mx-auto">
                The corporate network is quiet. Be the first to respond to the crisis situation and establish a new containment protocol.
              </p>
              <Button
                variant="primary"
                onClick={() => setShowCreateModal(true)}
                className="font-semibold text-sm px-8 py-3 shadow-lg hover:shadow-xl transition-all duration-200"
              >
                🚀 Initiate New Session
              </Button>
            </div>
          ) : (
            <div
              ref={lobbyListRef}
              className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
            >
              {lobbies.map((lobby) => (
                <LobbyCard
                  key={lobby.id}
                  lobby={lobby}
                  onJoin={handleJoinLobby}
                  onSpectate={handleSpectateGame}
                  joinCooldowns={joinCooldowns}
                />
              ))}
            </div>
          )}
        </main>
      </div>

      <CreateGameModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreate={handleCreateGame}
        playerName={playerName}
      />
      <FriendsPanel
        isVisible={showFriendsPanel}
        onClose={() => setShowFriendsPanel(false)}
      />
      <PartyPanel
        isVisible={showPartyPanel}
        onClose={() => setShowPartyPanel(false)}
      />
    </>
  );
}
