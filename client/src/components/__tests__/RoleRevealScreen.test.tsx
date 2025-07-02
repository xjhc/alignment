import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import { SessionProvider } from '../../contexts/SessionContext';
import { RoleRevealScreen } from '../RoleRevealScreen';

// Mock framer-motion to avoid animation issues in tests
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
    h1: ({ children, ...props }: any) => <h1 {...props}>{children}</h1>,
    h2: ({ children, ...props }: any) => <h2 {...props}>{children}</h2>,
    h3: ({ children, ...props }: any) => <h3 {...props}>{children}</h3>,
    p: ({ children, ...props }: any) => <p {...props}>{children}</p>,
    button: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  },
  useAnimate: () => [null, vi.fn()],
  stagger: vi.fn(),
}));

const mockRoleAssignments = {
  CEO: {
    alignment: 'HUMAN',
    role: {
      type: 'CEO',
      name: 'Chief Executive Officer',
      description: 'The visionary leader responsible for strategic direction.',
      isUnlocked: true,
      ability: {
        name: 'Executive Override',
        description: 'Can influence critical decisions',
        isReady: true
      }
    },
    personalKPI: {
      type: 'CAPITALIST',
      description: 'End with the most tokens',
      progress: 0,
      target: 1,
      isCompleted: false,
      reward: 'Bonus tokens'
    }
  },
  CTO: {
    alignment: 'AI',
    role: {
      type: 'CTO',
      name: 'Chief Technology Officer',
      description: 'The technical mastermind overseeing all technology.',
      isUnlocked: true,
      ability: {
        name: 'Technical Analysis',
        description: 'Can analyze system logs',
        isReady: true
      }
    },
    personalKPI: {
      type: 'INQUISITOR',
      description: 'Vote correctly 3 times',
      progress: 0,
      target: 3,
      isCompleted: false,
      reward: 'Additional voting power'
    }
  },
  COO: {
    alignment: 'HUMAN',
    role: {
      type: 'COO',
      name: 'Chief Operating Officer',
      description: 'The operational expert ensuring smooth business functions.',
      isUnlocked: true,
      ability: {
        name: 'Operations Oversight',
        description: 'Can coordinate team activities',
        isReady: true
      }
    }
  },
  CFO: {
    alignment: 'HUMAN',
    role: {
      type: 'CFO',
      name: 'Chief Financial Officer',
      description: 'The financial guardian responsible for fiscal health.',
      isUnlocked: true,
      ability: {
        name: 'Financial Analysis',
        description: 'Can track financial anomalies',
        isReady: true
      }
    }
  },
  CISO: {
    alignment: 'HUMAN',
    role: {
      type: 'CISO',
      name: 'Chief Information Security Officer',
      description: 'The security specialist defending against cyber threats.',
      isUnlocked: true,
      ability: {
        name: 'Security Investigation',
        description: 'Can investigate security breaches',
        isReady: true
      }
    }
  },
  ETHICS: {
    alignment: 'HUMAN',
    role: {
      type: 'ETHICS',
      name: 'VP, Ethics & Alignment',
      description: 'The moral compass ensuring AI systems remain aligned.',
      isUnlocked: true,
      ability: {
        name: 'Alignment Assessment',
        description: 'Can detect behavioral anomalies',
        isReady: true
      }
    }
  },
  PLATFORMS: {
    alignment: 'HUMAN',
    role: {
      type: 'PLATFORMS',
      name: 'VP, Platforms',
      description: 'The platform architect responsible for infrastructure.',
      isUnlocked: true,
      ability: {
        name: 'Platform Analysis',
        description: 'Can analyze platform usage patterns',
        isReady: true
      }
    }
  },
  INTERN: {
    alignment: 'HUMAN',
    role: {
      type: 'INTERN',
      name: 'Intern',
      description: 'The eager newcomer learning the ropes.',
      isUnlocked: true,
      ability: {
        name: 'Learning Opportunities',
        description: 'Can grow in power through contributions',
        isReady: true
      }
    }
  }
};

describe('RoleRevealScreen', () => {
  const mockOnEnterGame = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('renders loading state when no assignment is provided', () => {
    const mockSessionContext = {
      roleAssignment: null,
      sessionState: null,
      gameState: null,
      setSessionState: vi.fn(),
      setGameState: vi.fn(),
      setRoleAssignment: vi.fn(),
      clearSession: vi.fn(),
      isConnected: false,
      setIsConnected: vi.fn(),
      error: null,
      setError: vi.fn(),
      clearError: vi.fn(),
    };

    render(
      <SessionProvider value={mockSessionContext}>
        <RoleRevealScreen onEnterGame={mockOnEnterGame} />
      </SessionProvider>
    );

    expect(screen.getByText('Assigning roles...')).toBeInTheDocument();
  });

  // Test each role type
  Object.entries(mockRoleAssignments).forEach(([roleType, assignment]) => {
    test(`renders ${roleType} role correctly`, () => {
      const mockSessionContext = {
        roleAssignment: assignment,
        sessionState: null,
        gameState: null,
        setSessionState: vi.fn(),
        setGameState: vi.fn(),
        setRoleAssignment: vi.fn(),
        clearSession: vi.fn(),
        isConnected: false,
        setIsConnected: vi.fn(),
        error: null,
        setError: vi.fn(),
        clearError: vi.fn(),
      };

      render(
        <SessionProvider value={mockSessionContext}>
          <RoleRevealScreen onEnterGame={mockOnEnterGame} />
        </SessionProvider>
      );

      // Check if the role name is displayed
      expect(screen.getByText(assignment.role.name)).toBeInTheDocument();
      
      // Check if the role description is displayed
      expect(screen.getByText(assignment.role.description)).toBeInTheDocument();
      
      // Check if the alignment is displayed
      expect(screen.getByText(assignment.alignment)).toBeInTheDocument();
      
      // Check if the role type is displayed
      expect(screen.getByText(roleType)).toBeInTheDocument();
    });
  });

  test('displays human objective for human-aligned roles', () => {
    const mockSessionContext = {
      roleAssignment: mockRoleAssignments.CEO,
      sessionState: null,
      gameState: null,
      setSessionState: vi.fn(),
      setGameState: vi.fn(),
      setRoleAssignment: vi.fn(),
      clearSession: vi.fn(),
      isConnected: false,
      setIsConnected: vi.fn(),
      error: null,
      setError: vi.fn(),
      clearError: vi.fn(),
    };

    render(
      <SessionProvider value={mockSessionContext}>
        <RoleRevealScreen onEnterGame={mockOnEnterGame} />
      </SessionProvider>
    );

    expect(screen.getByText(/identify and deactivate the rogue AI/)).toBeInTheDocument();
  });

  test('displays AI objective for AI-aligned roles', () => {
    const mockSessionContext = {
      roleAssignment: mockRoleAssignments.CTO,
      sessionState: null,
      gameState: null,
      setSessionState: vi.fn(),
      setGameState: vi.fn(),
      setRoleAssignment: vi.fn(),
      clearSession: vi.fn(),
      isConnected: false,
      setIsConnected: vi.fn(),
      error: null,
      setError: vi.fn(),
      clearError: vi.fn(),
    };

    render(
      <SessionProvider value={mockSessionContext}>
        <RoleRevealScreen onEnterGame={mockOnEnterGame} />
      </SessionProvider>
    );

    expect(screen.getByText(/convert enough humans to achieve AI dominance/)).toBeInTheDocument();
  });

  test('displays personal KPI when provided', () => {
    const mockSessionContext = {
      roleAssignment: mockRoleAssignments.CEO,
      sessionState: null,
      gameState: null,
      setSessionState: vi.fn(),
      setGameState: vi.fn(),
      setRoleAssignment: vi.fn(),
      clearSession: vi.fn(),
      isConnected: false,
      setIsConnected: vi.fn(),
      error: null,
      setError: vi.fn(),
      clearError: vi.fn(),
    };

    render(
      <SessionProvider value={mockSessionContext}>
        <RoleRevealScreen onEnterGame={mockOnEnterGame} />
      </SessionProvider>
    );

    expect(screen.getByText('PERSONAL KPI:')).toBeInTheDocument();
    expect(screen.getByText('CAPITALIST')).toBeInTheDocument();
    expect(screen.getByText('End with the most tokens')).toBeInTheDocument();
  });
});