import React from 'react';
import { useNotificationBridge } from '../hooks/useNotificationBridge';

/**
 * Component that initializes the notification bridge
 * Must be inside GameProvider context
 */
export const NotificationBridge: React.FC = () => {
  useNotificationBridge();
  return null; // This component doesn't render anything
};