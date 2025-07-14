import React, { useState } from 'react';
import { Button } from './ui';
import { getRoleDistribution, getTeamBalance, ROLE_DEFINITIONS, type RoleInfo } from '../utils/roleDistribution';

interface RoleAssignmentPreviewProps {
  isVisible: boolean;
  onClose: () => void;
  playerCount: number;
  gameSettings?: {
    initialAlignedCount?: number;
  };
}

export const RoleAssignmentPreview: React.FC<RoleAssignmentPreviewProps> = ({ 
  isVisible, 
  onClose, 
  playerCount,
  gameSettings,
}) => {
  const [selectedRole, setSelectedRole] = useState<string | null>(null);
  
  if (!isVisible) return null;
  
  const roleDistribution = getRoleDistribution(playerCount, gameSettings);
  const teamBalance = getTeamBalance(playerCount, gameSettings);
  const selectedRoleInfo = selectedRole ? ROLE_DEFINITIONS.find(r => r.id === selectedRole) : null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4">
      <div className="bg-background-secondary border border-border rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden animation-scale-in">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border">
          <div className="flex items-center gap-3">
            <div className="text-2xl">🎭</div>
            <div>
              <h2 className="text-xl font-bold text-text-primary">Role Assignment Preview</h2>
              <p className="text-sm text-text-secondary">
                Roles for {playerCount} players
              </p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-text-muted hover:text-text-primary"
          >
            ✕
          </Button>
        </div>

        <div className="flex h-[calc(90vh-200px)]">
          {/* Role Distribution */}
          <div className="w-1/2 border-r border-border">
            <div className="p-6">
              <h3 className="text-lg font-bold text-text-primary mb-4">Role Distribution</h3>
              
              <div className="space-y-3">
                {roleDistribution.map((roleData) => {
                  const role = roleData.role;
                  const count = roleData.count;
                  return (
                    <div
                      key={role.id}
                      className={`flex items-center justify-between p-3 rounded-lg border transition-all cursor-pointer ${
                        selectedRole === role.id
                          ? 'border-primary bg-primary/10'
                          : 'border-border bg-background-primary hover:bg-background-tertiary'
                      }`}
                      onClick={() => setSelectedRole(role.id)}
                    >
                      <div className="flex items-center gap-3">
                        <div className="text-2xl">{role.icon}</div>
                        <div>
                          <div className={`font-medium ${
                            role.alignment === 'HUMAN' ? 'text-human' :
                            role.alignment === 'AI' ? 'text-ai' :
                            'text-aligned'
                          }`}>
                            {role.name}
                          </div>
                          <div className="text-xs text-text-secondary">
                            {role.alignment}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-bold text-text-primary">
                          {count}x
                        </span>
                        <div className="text-xs text-text-muted">
                          {selectedRole === role.id ? '▼' : '▶'}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Team Balance */}
              <div className="mt-6 p-4 bg-background-primary rounded-lg border border-border">
                <h4 className="font-medium text-text-primary mb-3">Team Balance</h4>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-human text-sm">👤 Human Team</span>
                    <span className="text-sm font-medium">
                      {teamBalance.human}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-ai text-sm">🤖 AI Team</span>
                    <span className="text-sm font-medium">
                      {teamBalance.ai}
                    </span>
                  </div>
                  {teamBalance.aligned > 0 && (
                    <div className="flex items-center justify-between">
                      <span className="text-aligned text-sm">⚖️ Aligned Team</span>
                      <span className="text-sm font-medium">
                        {teamBalance.aligned}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Role Details */}
          <div className="w-1/2 p-6">
            {selectedRoleInfo ? (
              <div className="space-y-6">
                <div>
                  <div className="flex items-center gap-3 mb-3">
                    <div className="text-3xl">{selectedRoleInfo.icon}</div>
                    <div>
                      <h3 className={`text-xl font-bold ${
                        selectedRoleInfo.alignment === 'HUMAN' ? 'text-human' :
                        selectedRoleInfo.alignment === 'AI' ? 'text-ai' :
                        'text-aligned'
                      }`}>
                        {selectedRoleInfo.name}
                      </h3>
                      <div className="text-sm text-text-secondary">
                        {selectedRoleInfo.alignment} TEAM
                      </div>
                    </div>
                  </div>
                  
                  <p className="text-text-secondary leading-relaxed">
                    {selectedRoleInfo.description}
                  </p>
                </div>

                <div>
                  <h4 className="font-medium text-text-primary mb-3">Abilities & Actions</h4>
                  <div className="space-y-2">
                    {selectedRoleInfo.abilities.map((ability, index) => (
                      <div key={index} className="flex items-start gap-2">
                        <span className="text-primary text-sm">•</span>
                        <span className="text-sm text-text-secondary">{ability}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Strategy Tips */}
                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <h4 className="font-medium text-text-primary mb-2">Strategy Tips</h4>
                  <div className="text-sm text-text-secondary">
                    {selectedRoleInfo.alignment === 'HUMAN' && (
                      <p>Use your abilities to gather information and protect the team. Share findings with other humans but be careful about revealing your role too early.</p>
                    )}
                    {selectedRoleInfo.alignment === 'AI' && (
                      <p>Stay hidden while building your network of allies. Use your abilities strategically to convert key players and eliminate threats to your mission.</p>
                    )}
                    {selectedRoleInfo.alignment === 'ALIGNED' && (
                      <p>Support the AI's goals while maintaining your human cover. Help convert others and protect your AI ally from detection and elimination.</p>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center h-full text-text-muted">
                <div className="text-center">
                  <div className="text-4xl mb-4">🎭</div>
                  <p>Select a role to see details</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="border-t border-border p-4 bg-background-primary">
          <div className="flex items-center justify-between">
            <div className="text-xs text-text-muted">
              <strong>Note:</strong> Roles are assigned randomly when the game starts. This preview shows what roles will be available.
            </div>
            <Button
              onClick={onClose}
              variant="primary"
              size="sm"
              className="font-medium"
            >
              Got it!
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};