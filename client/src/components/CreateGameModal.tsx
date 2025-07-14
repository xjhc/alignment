import React, { useState } from "react";
import { Button, Input, Modal } from "./ui";
import { GameSettings } from "../types";

interface CreateGameModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (settings: Partial<GameSettings> & { lobbyName: string }) => void;
  playerName: string;
}

export const CreateGameModal: React.FC<CreateGameModalProps> = ({
  isOpen,
  onClose,
  onCreate,
  playerName,
}) => {
  const [lobbyName, setLobbyName] = useState(`${playerName} Game`);
  const [playAsAI, setPlayAsAI] = useState(false);
  const [hasInitialAligned, setHasInitialAligned] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCreate({
      lobbyName,
      playAsAI,
      initialAlignedHumanCount: hasInitialAligned ? 1 : 0,
    });
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="md"
    >
      <form onSubmit={handleSubmit}>
        <Modal.Header>
          🚨 Create Emergency Session
        </Modal.Header>
        <Modal.Body>
          <div className="space-y-6">
            <div>
              <Input
                label="Lobby Name"
                value={lobbyName}
                onChange={(e) => setLobbyName(e.target.value)}
                maxLength={100}
                required
                placeholder="Enter a lobby name (avoid special characters like quotes and brackets)"
              />
            </div>

            <div className="border-t border-border pt-4">
              <h4 className="text-sm font-bold text-text-primary mb-3 flex items-center gap-2">
                <span>⚙️</span>
                Custom Game Rules
              </h4>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-3 bg-background-secondary rounded-lg hover:bg-background-tertiary transition-colors duration-200">
                  <div>
                    <label className="font-medium text-text-primary text-sm">
                      Play as AI
                    </label>
                    <p className="text-xs text-text-secondary">
                      One random player will be assigned the AI role at game
                      start.
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={playAsAI}
                    onChange={(e) => setPlayAsAI(e.target.checked)}
                    className="h-5 w-5 rounded-md border-border text-primary focus:ring-primary"
                  />
                </div>

                <div className="flex items-center justify-between p-3 bg-background-secondary rounded-lg hover:bg-background-tertiary transition-colors duration-200">
                  <div>
                    <label className="font-medium text-text-primary text-sm">
                      Start with 1 Aligned Human
                    </label>
                    <p className="text-xs text-text-secondary">
                      One human player will secretly start on the AI's team.
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={hasInitialAligned}
                    onChange={(e) => setHasInitialAligned(e.target.checked)}
                    className="h-5 w-5 rounded-md border-border text-primary focus:ring-primary"
                  />
                </div>
              </div>
            </div>
          </div>
        </Modal.Body>
        <Modal.Footer>
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" className="font-semibold">
            🚀 Create Game
          </Button>
        </Modal.Footer>
      </form>
    </Modal>
  );
};
