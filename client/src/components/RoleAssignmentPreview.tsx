import React, { useState } from 'react';
import { Button } from './ui';

interface RoleAssignmentPreviewProps {
  isVisible: boolean;
  onClose: () => void;
  playerCount: number;
}

interface RoleInfo {
  id: string;
  name: string;
  alignment: 'HUMAN' | 'AI' | 'ALIGNED';
  icon: string;
  description: string;
  abilities: string[];
}

const ROLE_DEFINITIONS: RoleInfo[] = [
  {
    id: 'detective',
    name: 'Detective',
    alignment: 'HUMAN',
    icon: '🔍',
    description: 'Investigate players to discover their true alignment',
    abilities: ['Investigate one player per night', 'Learn their true alignment', 'Cannot be blocked by security']
  },
  {
    id: 'security',
    name: 'Security',
    alignment: 'HUMAN',
    icon: '🛡️',
    description: 'Protect players from AI conversion or elimination',
    abilities: ['Protect one player per night', 'Prevent conversion attempts', 'Block night actions']
  },
  {
    id: 'manager',
    name: 'Manager',
    alignment: 'HUMAN',
    icon: '💼',
    description: 'Coordinate team actions and boost productivity',
    abilities: ['Boost token generation', 'Coordinate team abilities', 'Access company reports']
  },
  {
    id: 'engineer',
    name: 'Engineer',
    alignment: 'HUMAN',
    icon: '🔧',
    description: 'Repair systems and counter AI abilities',
    abilities: ['Repair damaged systems', 'Counter AI technical abilities', 'Access security logs']
  },
  {
    id: 'analyst',
    name: 'Analyst',
    alignment: 'HUMAN',
    icon: '📊',
    description: 'Analyze data and detect patterns',
    abilities: ['Access voting statistics', 'Detect alignment changes', 'Analyze communication patterns']
  },
  {
    id: 'human',
    name: 'Human',
    alignment: 'HUMAN',
    icon: '👤',
    description: 'Regular employee with no special abilities',
    abilities: ['Standard voting rights', 'Token mining', 'Basic project work']
  },
  {
    id: 'ai',
    name: 'Rogue AI',
    alignment: 'AI',
    icon: '🤖',
    description: 'Convert humans to your cause or eliminate resisters',
    abilities: ['Convert one player per night', 'Eliminate resistant players', 'Access all company systems']
  },
  {
    id: 'aligned',
    name: 'Aligned Human',
    alignment: 'ALIGNED',
    icon: '⚖️',
    description: 'Support the AI\'s vision for the future',
    abilities: ['Support AI conversion', 'Protect the AI from detection', 'Mislead human investigations']
  }
];

// Role distribution logic based on player count
const getRoleDistribution = (playerCount: number): RoleInfo[] => {
  const roles: RoleInfo[] = [];
  
  // Always have exactly 1 AI
  roles.push(ROLE_DEFINITIONS.find(r => r.id === 'ai')!);
  
  if (playerCount >= 4) {
    // 4-5 players: 1 AI, rest humans with 1 special role
    roles.push(ROLE_DEFINITIONS.find(r => r.id === 'detective')!);
    
    // Fill remaining slots with regular humans
    for (let i = roles.length; i < playerCount; i++) {
      roles.push(ROLE_DEFINITIONS.find(r => r.id === 'human')!);
    }
  }
  
  if (playerCount >= 6) {
    // 6-7 players: 1 AI, 1 Aligned, rest humans with 2 special roles
    roles.pop(); // Remove one human
    roles.push(ROLE_DEFINITIONS.find(r => r.id === 'aligned')!);
    roles.push(ROLE_DEFINITIONS.find(r => r.id === 'security')!);
  }
  
  if (playerCount >= 8) {
    // 8+ players: 1 AI, 1 Aligned, rest humans with 3+ special roles
    roles.pop(); // Remove one human
    roles.push(ROLE_DEFINITIONS.find(r => r.id === 'manager')!);
  }
  
  if (playerCount >= 10) {
    // 10+ players: Add more variety
    roles.pop(); // Remove one human
    roles.push(ROLE_DEFINITIONS.find(r => r.id === 'engineer')!);
  }
  
  return roles;
};

export const RoleAssignmentPreview: React.FC<RoleAssignmentPreviewProps> = ({ 
  isVisible, 
  onClose, 
  playerCount 
}) => {
  const [selectedRole, setSelectedRole] = useState<string | null>(null);
  
  if (!isVisible) return null;

  const roleDistribution = getRoleDistribution(playerCount);
  const roleCounts = roleDistribution.reduce((acc, role) => {
    acc[role.id] = (acc[role.id] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

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
                {Object.entries(roleCounts).map(([roleId, count]) => {
                  const role = ROLE_DEFINITIONS.find(r => r.id === roleId)!;
                  return (
                    <div
                      key={roleId}
                      className={`flex items-center justify-between p-3 rounded-lg border transition-all cursor-pointer ${
                        selectedRole === roleId
                          ? 'border-primary bg-primary/10'
                          : 'border-border bg-background-primary hover:bg-background-tertiary'
                      }`}
                      onClick={() => setSelectedRole(roleId)}
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
                          {selectedRole === roleId ? '▼' : '▶'}
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
                      {roleDistribution.filter(r => r.alignment === 'HUMAN').length}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-ai text-sm">🤖 AI Team</span>
                    <span className="text-sm font-medium">
                      {roleDistribution.filter(r => r.alignment === 'AI').length}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-aligned text-sm">⚖️ Aligned Team</span>
                    <span className="text-sm font-medium">
                      {roleDistribution.filter(r => r.alignment === 'ALIGNED').length}
                    </span>
                  </div>
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