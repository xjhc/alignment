import React, { createContext, useContext, useReducer, useCallback } from 'react';
import { PrivateNotification } from '../types';

// Notification types for different tiers
export type NotificationTier = 'toast' | 'system' | 'loebmate' | 'subtle';

export interface ToastNotification {
  id: string;
  type: 'toast';
  category: 'action_blocked' | 'conversion' | 'ability_unlocked' | 'kpi_complete' | 'generic';
  icon: string;
  color: 'danger' | 'magenta' | 'info' | 'success' | 'warning';
  title: string;
  message: string;
  priority: 'low' | 'medium' | 'high';
  timestamp: number;
  duration?: number; // Auto-dismiss time in ms, undefined for manual dismissal
}

export interface LoebmateNotification {
  id: string;
  type: 'loebmate';
  trigger: 'nomination_phase' | 'night_phase' | 'first_tokens' | 'role_unlocked' | 'system_shock';
  content: string;
  timestamp: number;
  isPrivate: boolean;
}

export interface SubtleIndicator {
  id: string;
  type: 'subtle';
  element: 'channel' | 'vote_button' | 'typing' | 'browser_tab';
  state: 'unread' | 'voted' | 'typing' | 'new_message';
  value?: string | number;
  tooltip?: string;
}

export type UnifiedNotification = ToastNotification | LoebmateNotification | SubtleIndicator;

interface NotificationState {
  toasts: ToastNotification[];
  loebmateMessages: LoebmateNotification[];
  subtleIndicators: SubtleIndicator[];
  settings: {
    loebmateEnabled: boolean;
    soundEnabled: boolean;
    toastLimit: number;
  };
}

type NotificationAction = 
  | { type: 'ADD_TOAST'; payload: ToastNotification }
  | { type: 'DISMISS_TOAST'; payload: string }
  | { type: 'ADD_LOEBMATE'; payload: LoebmateNotification }
  | { type: 'DISMISS_LOEBMATE'; payload: string }
  | { type: 'SET_SUBTLE_INDICATOR'; payload: SubtleIndicator }
  | { type: 'CLEAR_SUBTLE_INDICATOR'; payload: string }
  | { type: 'UPDATE_SETTINGS'; payload: Partial<NotificationState['settings']> }
  | { type: 'CLEAR_ALL_TOASTS' };

const initialState: NotificationState = {
  toasts: [],
  loebmateMessages: [],
  subtleIndicators: [],
  settings: {
    loebmateEnabled: true,
    soundEnabled: true,
    toastLimit: 3,
  },
};

function notificationReducer(state: NotificationState, action: NotificationAction): NotificationState {
  switch (action.type) {
    case 'ADD_TOAST':
      // Limit toast count and prioritize by priority
      const existingToasts = state.toasts;
      const newToast = action.payload;
      
      // If at limit, remove lowest priority toast
      if (existingToasts.length >= state.settings.toastLimit) {
        const sortedToasts = [...existingToasts].sort((a, b) => {
          const priorityOrder = { high: 3, medium: 2, low: 1 };
          return priorityOrder[b.priority] - priorityOrder[a.priority];
        });
        
        // Remove the lowest priority toast
        sortedToasts.pop();
        
        return {
          ...state,
          toasts: [...sortedToasts, newToast],
        };
      }
      
      return {
        ...state,
        toasts: [...existingToasts, newToast],
      };
      
    case 'DISMISS_TOAST':
      return {
        ...state,
        toasts: state.toasts.filter(toast => toast.id !== action.payload),
      };
      
    case 'ADD_LOEBMATE':
      if (!state.settings.loebmateEnabled) return state;
      
      return {
        ...state,
        loebmateMessages: [...state.loebmateMessages, action.payload],
      };
      
    case 'DISMISS_LOEBMATE':
      return {
        ...state,
        loebmateMessages: state.loebmateMessages.filter(msg => msg.id !== action.payload),
      };
      
    case 'SET_SUBTLE_INDICATOR':
      const existingIndicators = state.subtleIndicators.filter(
        indicator => indicator.id !== action.payload.id
      );
      
      return {
        ...state,
        subtleIndicators: [...existingIndicators, action.payload],
      };
      
    case 'CLEAR_SUBTLE_INDICATOR':
      return {
        ...state,
        subtleIndicators: state.subtleIndicators.filter(
          indicator => indicator.id !== action.payload
        ),
      };
      
    case 'UPDATE_SETTINGS':
      return {
        ...state,
        settings: { ...state.settings, ...action.payload },
      };
      
    case 'CLEAR_ALL_TOASTS':
      return {
        ...state,
        toasts: [],
      };
      
    default:
      return state;
  }
}

interface NotificationContextValue {
  state: NotificationState;
  
  // Toast notifications (Tier 1)
  showToast: (notification: Omit<ToastNotification, 'id' | 'timestamp'>) => void;
  dismissToast: (id: string) => void;
  
  // Loebmate messages (Tier 3)
  showLoebmateMessage: (notification: Omit<LoebmateNotification, 'id' | 'timestamp'>) => void;
  dismissLoebmateMessage: (id: string) => void;
  
  // Subtle indicators (Tier 4)
  setSubtleIndicator: (indicator: SubtleIndicator) => void;
  clearSubtleIndicator: (id: string) => void;
  
  // Settings
  updateSettings: (settings: Partial<NotificationState['settings']>) => void;
  
  // Utility functions
  clearAllToasts: () => void;
}

const NotificationContext = createContext<NotificationContextValue | undefined>(undefined);

export const NotificationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(notificationReducer, initialState);
  
  const showToast = useCallback((notification: Omit<ToastNotification, 'id' | 'timestamp'>) => {
    const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const toast: ToastNotification = {
      ...notification,
      id,
      timestamp: Date.now(),
    };
    
    dispatch({ type: 'ADD_TOAST', payload: toast });
    
    // Auto-dismiss based on duration or default timing
    if (notification.duration !== undefined) {
      setTimeout(() => {
        dispatch({ type: 'DISMISS_TOAST', payload: id });
      }, notification.duration);
    } else if (notification.priority !== 'high') {
      // Auto-dismiss low/medium priority toasts after 8 seconds
      setTimeout(() => {
        dispatch({ type: 'DISMISS_TOAST', payload: id });
      }, 8000);
    }
  }, []);
  
  const dismissToast = useCallback((id: string) => {
    dispatch({ type: 'DISMISS_TOAST', payload: id });
  }, []);
  
  const showLoebmateMessage = useCallback((notification: Omit<LoebmateNotification, 'id' | 'timestamp'>) => {
    const id = `loebmate-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const message: LoebmateNotification = {
      ...notification,
      id,
      timestamp: Date.now(),
    };
    
    dispatch({ type: 'ADD_LOEBMATE', payload: message });
  }, []);
  
  const dismissLoebmateMessage = useCallback((id: string) => {
    dispatch({ type: 'DISMISS_LOEBMATE', payload: id });
  }, []);
  
  const setSubtleIndicator = useCallback((indicator: SubtleIndicator) => {
    dispatch({ type: 'SET_SUBTLE_INDICATOR', payload: indicator });
  }, []);
  
  const clearSubtleIndicator = useCallback((id: string) => {
    dispatch({ type: 'CLEAR_SUBTLE_INDICATOR', payload: id });
  }, []);
  
  const updateSettings = useCallback((settings: Partial<NotificationState['settings']>) => {
    dispatch({ type: 'UPDATE_SETTINGS', payload: settings });
  }, []);
  
  const clearAllToasts = useCallback(() => {
    dispatch({ type: 'CLEAR_ALL_TOASTS' });
  }, []);
  
  const value: NotificationContextValue = {
    state,
    showToast,
    dismissToast,
    showLoebmateMessage,
    dismissLoebmateMessage,
    setSubtleIndicator,
    clearSubtleIndicator,
    updateSettings,
    clearAllToasts,
  };
  
  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
};

export const useNotifications = () => {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
};

// Utility functions for creating specific notification types
export const createToastNotification = {
  actionBlocked: (message: string): Omit<ToastNotification, 'id' | 'timestamp'> => ({
    type: 'toast',
    category: 'action_blocked',
    icon: '🚫',
    color: 'danger',
    title: 'Action Blocked',
    message,
    priority: 'high',
  }),
  
  conversion: (message: string): Omit<ToastNotification, 'id' | 'timestamp'> => ({
    type: 'toast',
    category: 'conversion',
    icon: '⚡',
    color: 'magenta',
    title: 'System Shock',
    message,
    priority: 'high',
  }),
  
  abilityUnlocked: (abilityName: string): Omit<ToastNotification, 'id' | 'timestamp'> => ({
    type: 'toast',
    category: 'ability_unlocked',
    icon: '✨',
    color: 'info',
    title: 'Ability Unlocked',
    message: `You have unlocked **${abilityName}**. You can now use it during the Night Phase.`,
    priority: 'medium',
  }),
  
  kpiComplete: (objective: string): Omit<ToastNotification, 'id' | 'timestamp'> => ({
    type: 'toast',
    category: 'kpi_complete',
    icon: '🏆',
    color: 'success',
    title: 'KPI Complete',
    message: `You have completed your "${objective}" objective. Reward unlocked.`,
    priority: 'medium',
  }),
};

export const createLoebmateMessage = {
  nominationPhase: (): Omit<LoebmateNotification, 'id' | 'timestamp'> => ({
    type: 'loebmate',
    trigger: 'nomination_phase',
    content: '**New Action Unlocked: Nominate!** The discussion is over. It\'s time to choose who to put on trial. Click the "Nominate" button next to a player\'s name in the roster to cast your vote.',
    isPrivate: true,
  }),
  
  nightPhase: (): Omit<LoebmateNotification, 'id' | 'timestamp'> => ({
    type: 'loebmate',
    trigger: 'night_phase',
    content: '**Welcome to the Night Phase!** The main channel is locked. You have 30 seconds to secretly choose an action from the menu below. "Mine for Tokens" helps your team, while "Project Milestones" helps you unlock your ability. Choose wisely!',
    isPrivate: true,
  }),
  
  firstTokens: (): Omit<LoebmateNotification, 'id' | 'timestamp'> => ({
    type: 'loebmate',
    trigger: 'first_tokens',
    content: '**You\'ve received new Tokens!** Tokens increase your voting power. The more you have, the more your vote counts. They are also your defense against AI conversion.',
    isPrivate: true,
  }),
};