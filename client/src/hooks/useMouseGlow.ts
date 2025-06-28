import { useEffect, useState, useRef, useCallback } from 'react';

export interface MouseGlowState {
  x: number;
  y: number;
  isHovering: boolean;
}

export const useMouseGlow = () => {
  const [mousePosition, setMousePosition] = useState<MouseGlowState>({
    x: 0,
    y: 0,
    isHovering: false
  });
  
  const elementRef = useRef<HTMLElement>(null);

  const handleMouseMove = useCallback((event: MouseEvent) => {
    if (elementRef.current) {
      const rect = elementRef.current.getBoundingClientRect();
      const x = ((event.clientX - rect.left) / rect.width) * 100;
      const y = ((event.clientY - rect.top) / rect.height) * 100;
      
      setMousePosition({
        x: Math.max(0, Math.min(100, x)),
        y: Math.max(0, Math.min(100, y)),
        isHovering: true
      });
    }
  }, []);

  const handleMouseEnter = useCallback(() => {
    setMousePosition(prev => ({ ...prev, isHovering: true }));
  }, []);

  const handleMouseLeave = useCallback(() => {
    setMousePosition(prev => ({ ...prev, isHovering: false }));
  }, []);

  useEffect(() => {
    const element = elementRef.current;
    if (element) {
      element.addEventListener('mousemove', handleMouseMove);
      element.addEventListener('mouseenter', handleMouseEnter);
      element.addEventListener('mouseleave', handleMouseLeave);

      return () => {
        element.removeEventListener('mousemove', handleMouseMove);
        element.removeEventListener('mouseenter', handleMouseEnter);
        element.removeEventListener('mouseleave', handleMouseLeave);
      };
    }
  }, [handleMouseMove, handleMouseEnter, handleMouseLeave]);

  const getGlowStyle = useCallback((
    color: string = 'rgba(59, 130, 246, 0.15)', 
    size: number = 200
  ) => {
    if (!mousePosition.isHovering) {
      return {};
    }

    return {
      background: `radial-gradient(${size}px circle at ${mousePosition.x}% ${mousePosition.y}%, ${color}, transparent 70%)`,
      transition: 'background 0.3s ease-out'
    };
  }, [mousePosition]);

  return {
    elementRef,
    mousePosition,
    getGlowStyle
  };
};