import { useNavigate } from 'react-router-dom';
import { useCallback } from 'react';

export function useAppNavigation() {
  const navigate = useNavigate();

  const navigateToLogin = useCallback(() => {
    navigate('/login');
  }, [navigate]);

  const navigateToLobbyList = useCallback(() => {
    navigate('/lobby-list');
  }, [navigate]);

  const navigateToWaiting = useCallback(() => {
    navigate('/waiting');
  }, [navigate]);

  const navigateToRoleReveal = useCallback(() => {
    navigate('/role-reveal');
  }, [navigate]);

  const navigateToGame = useCallback(() => {
    navigate('/game');
  }, [navigate]);

  const navigateToGameOver = useCallback(() => {
    navigate('/game-over');
  }, [navigate]);

  const navigateToAnalysis = useCallback(() => {
    navigate('/analysis');
  }, [navigate]);

  return {
    navigateToLogin,
    navigateToLobbyList,
    navigateToWaiting,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
    navigateToAnalysis
  };
}