import { useState, useEffect } from 'react';

export type Theme = 'light' | 'dark';

export const useTheme = () => {
  const [theme, setTheme] = useState<Theme>(() => {
    // Check for saved theme preference or default to 'dark'
    const saved = localStorage.getItem('theme') as Theme;
    const initialTheme = saved || 'dark';
    
    // Apply theme immediately to prevent FOUC
    document.documentElement.setAttribute('data-theme', initialTheme);
    
    return initialTheme;
  });

  useEffect(() => {
    // Apply theme to document
    document.documentElement.setAttribute('data-theme', theme);
    // Save theme preference
    localStorage.setItem('theme', theme);
    
    // Add smooth transition class temporarily
    document.body.style.transition = 'all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94)';
    
    // Remove transition after animation completes
    const timer = setTimeout(() => {
      document.body.style.transition = '';
    }, 300);
    
    return () => clearTimeout(timer);
  }, [theme]);

  const toggleTheme = () => {
    setTheme(prev => prev === 'dark' ? 'light' : 'dark');
  };

  const setThemeMode = (newTheme: Theme) => {
    setTheme(newTheme);
  };

  return {
    theme,
    toggleTheme,
    setThemeMode,
    isDark: theme === 'dark',
    isLight: theme === 'light'
  };
};