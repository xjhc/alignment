import { useSessionContext } from "../contexts/SessionContext";
import { Button } from "./ui";
import { useState, useMemo, useEffect } from "react";
import { InviteFriendsModal } from "./InviteFriendsModal";
import { PlayerProfile } from "./PlayerProfile";
import { GameRulesSummaryEmbedded } from "./GameRulesSummaryEmbedded";
import { RoleAssignmentPreview } from "./RoleAssignmentPreview";

export function WaitingScreen() {
  const {
    appState,
    lobbyState,
    isConnected,
    onStartGame,
    onLeaveLobby,
    onBackToLogin,
  } = useSessionContext();

  const {
    playerInfos,
    isHost,
    canStart,
    hostId,
    maxPlayers,
    connectionError,
    countdown,
    lobbyName,
    gameSettings,
  } = lobbyState;

  const [inviteCopied, setInviteCopied] = useState(false);
  const [showInviteFriends, setShowInviteFriends] = useState(false);
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);
  const [showRolePreview, setShowRolePreview] = useState(false);
  const [lobbySettings, setLobbySettings] = useState(() => ({
    name: lobbyName,
    maxPlayers: maxPlayers,
    isPrivate: false,
    playAsAI: gameSettings?.playAsAI || false,
    initialAlignedCount: gameSettings?.initialAlignedHumanCount || 0,
  }));

  const formatGameId = (id: string) => {
    return id ? id.substring(0, 6) : "unknown";
  };

  const copyInviteLink = async () => {
    if (!appState.gameId) return;

    const inviteUrl = `${window.location.origin}/join/${appState.gameId}`;

    try {
      await navigator.clipboard.writeText(inviteUrl);
      setInviteCopied(true);
      setTimeout(() => setInviteCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy invite link:", err);
      const textArea = document.createElement("textarea");
      textArea.value = inviteUrl;
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand("copy");
        setInviteCopied(true);
        setTimeout(() => setInviteCopied(false), 2000);
      } catch (fallbackErr) {
        console.error("Fallback copy failed:", fallbackErr);
      }
      document.body.removeChild(textArea);
    }
  };

  const togglePrivacy = async () => {
    if (!appState.gameId || !isHost) return;

    try {
      const response = await fetch(`/api/games/${appState.gameId}/privacy`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_private: !isPrivate }),
      });

      if (response.ok) {
        setIsPrivate(!isPrivate);
      } else {
        console.error("Failed to update lobby privacy");
      }
    } catch (err) {
      console.error("Error updating lobby privacy:", err);
    }
  };

  const gameSettingsDisplay = useMemo(() => {
    const settings = [];
    if (gameSettings?.playAsAI) {
      settings.push({ icon: "🤖", label: "Play as AI", color: "ai" });
    }
    if (
      gameSettings?.initialAlignedHumanCount &&
      gameSettings.initialAlignedHumanCount > 0
    ) {
      settings.push({
        icon: "🕵️",
        label: `+${gameSettings.initialAlignedHumanCount} Aligned`,
        color: "aligned",
      });
    }
    if (
      !gameSettings?.playAsAI &&
      (!gameSettings?.initialAlignedHumanCount ||
        gameSettings.initialAlignedHumanCount === 0)
    ) {
      settings.push({ icon: "⚖️", label: "Classic Mode", color: "human" });
    }
    return settings;
  }, [gameSettings]);

  if (connectionError) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background-primary to-background-secondary flex items-center justify-center px-4">
        <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-8 max-w-md w-full shadow-2xl animation-scale-in">
          <div className="text-center">
            <div className="w-16 h-16 mx-auto mb-6 bg-danger/20 rounded-full flex items-center justify-center">
              <span className="text-2xl">⚠️</span>
            </div>
            <h1 className="text-xl font-bold text-text-primary mb-2">
              Connection Error
            </h1>
            <p className="text-danger text-sm mb-6 leading-relaxed">
              {connectionError}
            </p>
            <div className="flex flex-col gap-3">
              <Button
                onClick={onLeaveLobby}
                variant="secondary"
                className="w-full"
              >
                ← Go Back
              </Button>
              <Button
                onClick={onBackToLogin}
                variant="ghost"
                className="w-full text-text-muted"
              >
                Logout
              </Button>
            </div>
            <p className="text-xs text-text-muted mt-4 text-center">
              If the lobby has ended or expired, try logging out and starting
              fresh
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!isConnected) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background-primary to-background-secondary flex items-center justify-center px-4">
        <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-8 max-w-md w-full shadow-2xl animation-scale-in">
          <div className="text-center">
            <div className="w-16 h-16 mx-auto mb-6 bg-primary/20 rounded-full flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent"></div>
            </div>
            <h1 className="text-xl font-bold text-text-primary mb-2">
              Connecting to Lobby
            </h1>
            <p className="text-text-secondary text-sm mb-6">
              Establishing secure connection...
            </p>
            <div className="flex flex-col gap-3">
              <Button
                onClick={onLeaveLobby}
                variant="secondary"
                className="w-full"
              >
                Cancel
              </Button>
              <Button
                onClick={onBackToLogin}
                variant="ghost"
                className="w-full text-text-muted"
              >
                Logout
              </Button>
            </div>
            <p className="text-xs text-text-muted mt-4 text-center">
              If connection takes too long, try canceling and rejoining
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Render the lobby UI directly. The player list will be populated reactively.
  // The ReconnectionOverlay will handle the "syncing" state for session restores.
  return (
    <>
      <div className="min-h-screen bg-gradient-to-br from-background-primary to-background-secondary">
        {/* Header */}
        <div className="bg-background-secondary/50 backdrop-blur-sm border-b border-border">
          <div className="max-w-4xl mx-auto px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-lg font-mono font-bold text-text-primary tracking-wide">
                  LOEBIAN INC. // EMERGENCY BRIDGE
                </h1>
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-xs text-text-muted">STATUS:</span>
                  <span className="text-xs font-semibold text-amber animate-pulse">
                    CONTAINMENT STANDBY
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="max-w-4xl mx-auto px-6 py-8">
          {/* Lobby Info Card */}
          <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-6 mb-8 shadow-lg">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-bold text-text-primary mb-1">
                  {lobbyName || (
                    <div className="flex items-center gap-3">
                      <div className="flex items-center gap-2">
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent"></div>
                        <span>Loading lobby info...</span>
                      </div>
                      <Button
                        onClick={onLeaveLobby}
                        variant="ghost"
                        size="sm"
                        className="text-xs text-danger hover:bg-danger/10"
                        title="Exit this lobby and return to lobby list"
                      >
                        Leave Lobby
                      </Button>
                    </div>
                  )}
                </h2>
                <div className="flex items-center gap-4 text-sm text-text-secondary">
                  <span className="flex items-center gap-1">
                    <span className="text-base">🏢</span>
                    <code className="bg-background-tertiary px-2 py-1 rounded font-mono">
                      {formatGameId(appState.gameId || "unknown")}
                    </code>
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="text-base">👥</span>
                    {playerInfos?.length || 0} / {maxPlayers} Personnel
                  </span>
                </div>
                {(!lobbyName ||
                  (lobbyName &&
                    isConnected &&
                    (playerInfos?.length || 0) === 0)) && (
                  <div className="mt-3 p-3 bg-amber/10 border border-amber/30 rounded-lg">
                    <div className="text-xs text-amber space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="animate-spin rounded-full h-3 w-3 border border-amber border-t-transparent"></div>
                        <span>
                          {!lobbyName
                            ? "Connecting to lobby..."
                            : "Synchronizing player list..."}
                        </span>
                      </div>
                      <div className="text-amber/80">
                        {!isConnected
                          ? "⏳ Establishing connection to server..."
                          : !lobbyName
                            ? "✓ Connected to server, waiting for lobby information..."
                            : "✓ Connected to lobby, waiting for player information..."}
                      </div>
                    </div>
                  </div>
                )}
              </div>
              <div className="text-right">
                <div className="flex flex-wrap gap-2 justify-end mb-2">
                  {gameSettingsDisplay.map((setting, index) => (
                    <span
                      key={index}
                      className={`text-xs font-semibold px-3 py-1 rounded-full bg-${setting.color}/20 text-${setting.color} border border-${setting.color}/30`}
                    >
                      {setting.icon} {setting.label}
                    </span>
                  ))}
                </div>
                <div className="flex gap-2">
                  <Button
                    onClick={copyInviteLink}
                    variant="ghost"
                    size="sm"
                    className="text-xs"
                  >
                    {inviteCopied ? "✓ Copied!" : "📋 Copy Invite"}
                  </Button>
                  <Button
                    onClick={() => setShowInviteFriends(true)}
                    variant="ghost"
                    size="sm"
                    className="text-xs"
                  >
                    👥 Invite
                  </Button>
                </div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="flex flex-wrap gap-3 mb-6">
              <Button
                onClick={() => setShowRolePreview(true)}
                variant="secondary"
                size="sm"
                className="text-xs"
              >
                🎭 Role Preview
              </Button>
              {isHost && (
                <label className="flex items-center gap-2 text-xs text-text-secondary cursor-pointer">
                  <input
                    type="checkbox"
                    checked={isPrivate}
                    onChange={togglePrivacy}
                    className="rounded"
                  />
                  🔒 Private Lobby
                </label>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Player List */}
            <div className="lg:col-span-2 space-y-6">
              {/* Game Rules Summary */}
              <GameRulesSummaryEmbedded />
              
              {/* Player List */}
              <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-6 shadow-lg">
                <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
                  <span className="text-base">👥</span>
                  Personnel Connected
                </h3>
                <div className="space-y-3">
                  {[...(playerInfos || [])]
                    .sort(
                      (a, b) =>
                        new Date(a.joinedAt).getTime() -
                        new Date(b.joinedAt).getTime()
                    )
                    .map((playerInfo, index) => (
                      <div
                        key={playerInfo.id}
                        className="flex items-center gap-3 p-3 rounded-xl bg-background-tertiary/50 border border-border/50 hover:bg-background-tertiary transition-all duration-200 group animation-slide-in-left"
                        style={{ animationDelay: `${index * 100}ms` }}
                      >
                        <div
                          className={`w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-lg border border-primary/30 cursor-pointer hover:border-primary transition-colors relative ${
                            playerInfo.connectionStatus === "DISCONNECTED"
                              ? "opacity-50 grayscale"
                              : ""
                          }`}
                          onClick={() => setSelectedPlayerId(playerInfo.id)}
                        >
                          {playerInfo.connectionStatus === "DISCONNECTED"
                            ? "🔌"
                            : playerInfo.avatar || "👤"}
                          {playerInfo.id === hostId && (
                            <div className="absolute -top-1 -right-1 w-5 h-5 bg-amber rounded-full flex items-center justify-center border-2 border-background-primary">
                              <span className="text-xs">👑</span>
                            </div>
                          )}
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <span
                              className={`font-semibold cursor-pointer hover:text-primary transition-colors ${
                                playerInfo.connectionStatus === "DISCONNECTED"
                                  ? "text-text-muted"
                                  : "text-text-primary"
                              }`}
                              onClick={() => setSelectedPlayerId(playerInfo.id)}
                            >
                              {playerInfo.name}
                            </span>
                            {playerInfo.id === hostId && (
                              <span className="text-xs px-2 py-0.5 bg-amber/20 text-amber rounded-full">
                                Host
                              </span>
                            )}
                            {playerInfo.id === appState.playerId && (
                              <span className="text-xs px-2 py-0.5 bg-primary/20 text-primary rounded-full">
                                You
                              </span>
                            )}
                            {playerInfo.connectionStatus === "DISCONNECTED" && (
                              <span className="text-xs px-2 py-0.5 bg-gray-500/20 text-gray-500 rounded-full">
                                🔌 Disconnected
                              </span>
                            )}
                          </div>
                          <div
                            className={`text-xs ${
                              playerInfo.connectionStatus === "DISCONNECTED"
                                ? "text-gray-500"
                                : "text-text-muted"
                            }`}
                          >
                            {playerInfo.connectionStatus === "DISCONNECTED"
                              ? "🔌 DISCONNECTED"
                              : "Emergency Response Agent"}
                          </div>
                        </div>
                        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          {playerInfo.id !== appState.playerId && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-xs px-2 py-1 h-auto"
                                onClick={async () => {
                                  try {
                                    const response = await fetch(
                                      "/api/friends/requests",
                                      {
                                        method: "POST",
                                        headers: {
                                          "Content-Type": "application/json",
                                        },
                                        body: JSON.stringify({
                                          recipient_username: playerInfo.name,
                                        }),
                                      }
                                    );
                                    if (response.ok) {
                                      alert("Friend request sent!");
                                    } else {
                                      const errorData = await response.json();
                                      alert(
                                        errorData.error ||
                                          "Failed to send friend request"
                                      );
                                    }
                                  } catch (err) {
                                    alert("Error sending friend request");
                                  }
                                }}
                                title="Add friend"
                              >
                                👥
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-xs px-2 py-1 h-auto"
                                onClick={async () => {
                                  const reason = prompt(
                                    "Why are you giving kudos to this player?"
                                  );
                                  if (reason) {
                                    try {
                                      const response = await fetch(
                                        "/api/users/me/kudos",
                                        {
                                          method: "POST",
                                          headers: {
                                            "Content-Type": "application/json",
                                          },
                                          body: JSON.stringify({
                                            target_player_id: playerInfo.id,
                                            reason: reason,
                                          }),
                                        }
                                      );
                                      if (response.ok) {
                                        alert("Kudos sent!");
                                      } else {
                                        const errorData = await response.json();
                                        alert(
                                          errorData.error ||
                                            "Failed to give kudos"
                                        );
                                      }
                                    } catch (err) {
                                      alert("Error giving kudos");
                                    }
                                  }
                                }}
                                title="Give kudos"
                              >
                                👏
                              </Button>
                              {isHost && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="text-xs px-2 py-1 h-auto text-danger hover:bg-danger/10"
                                  onClick={async () => {
                                    if (
                                      confirm(
                                        `Are you sure you want to remove ${playerInfo.name} from the lobby?`
                                      )
                                    ) {
                                      try {
                                        const response = await fetch(
                                          `/api/games/${appState.gameId}/kick`,
                                          {
                                            method: "POST",
                                            headers: {
                                              "Content-Type":
                                                "application/json",
                                            },
                                            body: JSON.stringify({
                                              player_id: playerInfo.id,
                                            }),
                                          }
                                        );
                                        if (response.ok) {
                                          alert(
                                            `${playerInfo.name} has been removed from the lobby`
                                          );
                                        } else {
                                          alert(
                                            "Failed to remove player - this feature may not be implemented yet"
                                          );
                                        }
                                      } catch (err) {
                                        alert("Error removing player");
                                      }
                                    }
                                  }}
                                  title="Remove player (Host only)"
                                >
                                  ❌
                                </Button>
                              )}
                            </>
                          )}
                        </div>
                      </div>
                    ))}
                  {/* Empty Slots */}
                  {Array.from({
                    length: Math.max(
                      0,
                      maxPlayers - (playerInfos?.length || 0)
                    ),
                  }).map((_, index) => (
                    <div
                      key={`empty-${index}`}
                      className="flex items-center gap-3 p-3 rounded-xl bg-background-tertiary/30 border border-border/30 border-dashed"
                    >
                      <div className="w-10 h-10 rounded-full bg-background-tertiary/50 flex items-center justify-center text-lg border border-border/50">
                        <span className="animate-pulse">⏳</span>
                      </div>
                      <div className="flex-1">
                        <div className="text-sm text-text-muted">
                          Waiting for personnel...
                        </div>
                        <div className="text-xs text-text-muted/70">
                          Open slot
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Right Column */}
            <div className="space-y-6">
              {/* Game Control */}
              <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-6 shadow-lg">
                <h3 className="text-lg font-bold text-text-primary mb-4">
                  {isHost ? "Host Control" : "Game Status"}
                </h3>

                {isHost ? (
                  <div className="space-y-4">
                    <div className="p-4 bg-amber/10 border border-amber/30 rounded-xl">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-amber">👑</span>
                        <span className="text-sm font-semibold text-amber">
                          Host Privileges
                        </span>
                      </div>
                      <p className="text-xs text-text-muted">
                        You control when the game starts and can manage players
                      </p>
                    </div>

                    <Button
                      variant="primary"
                      size="lg"
                      onClick={onStartGame}
                      disabled={
                        !canStart || !isConnected || countdown?.isActive
                      }
                      className="w-full text-sm font-semibold bg-amber hover:enabled:bg-amber-light text-black"
                    >
                      {countdown?.isActive
                        ? "INITIATING PROTOCOL..."
                        : canStart
                          ? "🚀 INITIATE CONTAINMENT PROTOCOL"
                          : `NEED ${Math.max(0, 4 - (playerInfos?.length || 0))} MORE PLAYERS`}
                    </Button>

                    {!canStart && (
                      <div className="text-xs text-text-muted text-center">
                        Minimum 4 players required to start
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="p-4 bg-primary/10 border border-primary/30 rounded-xl text-center">
                      <div className="text-2xl mb-2">⏳</div>
                      <div className="text-sm text-text-primary mb-1">
                        {countdown?.isActive
                          ? "Protocol Initiating..."
                          : "Awaiting Host Command"}
                      </div>
                      <div className="text-xs text-text-muted">
                        {countdown?.isActive
                          ? "Game starting soon..."
                          : "The host will start the game when ready"}
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Host Settings (if host) */}
              {isHost && (
                <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-6 shadow-lg">
                  <h3 className="text-lg font-bold text-text-primary mb-4">
                    Lobby Settings
                  </h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-xs text-text-muted uppercase font-medium mb-2">
                        Lobby Name
                      </label>
                      <input
                        type="text"
                        value={lobbySettings.name || ""}
                        onChange={(e) =>
                          setLobbySettings((prev) => ({
                            ...prev,
                            name: e.target.value,
                          }))
                        }
                        className="w-full px-3 py-2 bg-background-tertiary border border-border rounded-lg text-sm text-text-primary focus:border-primary focus:outline-none"
                        placeholder="Enter lobby name..."
                      />
                    </div>

                    <div>
                      <label className="block text-xs text-text-muted uppercase font-medium mb-2">
                        Max Players
                      </label>
                      <select
                        value={lobbySettings.maxPlayers || 8}
                        onChange={(e) =>
                          setLobbySettings((prev) => ({
                            ...prev,
                            maxPlayers: parseInt(e.target.value),
                          }))
                        }
                        className="w-full px-3 py-2 bg-background-tertiary border border-border rounded-lg text-sm text-text-primary focus:border-primary focus:outline-none"
                      >
                        <option value={4}>4 Players</option>
                        <option value={5}>5 Players</option>
                        <option value={6}>6 Players</option>
                        <option value={7}>7 Players</option>
                        <option value={8}>8 Players</option>
                        <option value={9}>9 Players</option>
                        <option value={10}>10 Players</option>
                      </select>
                    </div>

                    <div className="pt-2 border-t border-border">
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (
                              confirm(
                                "Are you sure you want to reset the lobby? This will remove all players except you."
                              )
                            )
                              alert(
                                "Lobby reset would be implemented with backend API"
                              );
                          }}
                          className="text-xs flex-1 text-amber hover:bg-amber/10"
                        >
                          🔄 Reset
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (
                              confirm(
                                "Are you sure you want to close this lobby? This will end the session for all players."
                              )
                            )
                              alert(
                                "Lobby closure would be implemented with backend API"
                              );
                          }}
                          className="text-xs flex-1 text-danger hover:bg-danger/10"
                        >
                          🚫 Close
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer Actions */}
          <div className="mt-8 flex justify-center gap-4">
            <Button
              onClick={onLeaveLobby}
              variant="ghost"
              className="text-text-muted hover:text-text-primary"
            >
              ← Leave Lobby
            </Button>
            <Button
              onClick={onBackToLogin}
              variant="ghost"
              className="text-text-muted hover:text-text-primary"
            >
              Logout
            </Button>
          </div>
        </div>

        {/* Countdown Overlay */}
        {countdown?.isActive && (
          <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-background-secondary border border-border rounded-2xl p-12 text-center max-w-md mx-4 animation-scale-in shadow-2xl">
              <div className="text-amber text-8xl font-mono font-bold mb-6 animation-pulse">
                {countdown.remaining > 0 ? countdown.remaining : "GO!"}
              </div>
              <div className="text-text-primary text-xl mb-4 font-semibold">
                INITIATING CONTAINMENT PROTOCOL
              </div>
              <div className="text-text-secondary mb-6">
                {countdown.remaining > 0
                  ? "Final preparations in progress..."
                  : "Protocol activated! Preparing game environment..."}
              </div>
              {countdown.remaining > 0 && (
                <Button
                  onClick={onLeaveLobby}
                  variant="ghost"
                  size="sm"
                  className="text-xs text-danger hover:bg-danger/10"
                >
                  Leave Lobby
                </Button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Modals */}
      <InviteFriendsModal
        isVisible={showInviteFriends}
        onClose={() => setShowInviteFriends(false)}
        lobbyId={appState.gameId}
      />
      <PlayerProfile
        playerId={selectedPlayerId || ""}
        isVisible={!!selectedPlayerId}
        onClose={() => setSelectedPlayerId(null)}
        currentPlayerId={appState.playerId}
      />
      <RoleAssignmentPreview
        isVisible={showRolePreview}
        onClose={() => setShowRolePreview(false)}
        playerCount={playerInfos?.length || 0}
      />
    </>
  );
}
