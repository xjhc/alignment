import React from 'react';

interface SkeletonProps {
  className?: string;
  width?: string | number;
  height?: string | number;
  rounded?: boolean;
  children?: React.ReactNode;
}

export function Skeleton({ 
  className = '', 
  width, 
  height, 
  rounded = false,
  children 
}: SkeletonProps) {
  const style: React.CSSProperties = {};
  if (width) style.width = typeof width === 'number' ? `${width}px` : width;
  if (height) style.height = typeof height === 'number' ? `${height}px` : height;

  return (
    <div 
      className={`
        bg-background-tertiary animate-pulse
        ${rounded ? 'rounded-full' : 'rounded'}
        ${className}
      `}
      style={style}
    >
      {children}
    </div>
  );
}

// Pre-built skeleton components for common use cases
export function SkeletonText({ lines = 1, className = '' }: { lines?: number; className?: string }) {
  return (
    <div className={`space-y-2 ${className}`}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton 
          key={i} 
          height="1rem" 
          className={i === lines - 1 ? 'w-3/4' : 'w-full'} 
        />
      ))}
    </div>
  );
}

export function SkeletonCard({ className = '' }: { className?: string }) {
  return (
    <div className={`p-4 border border-border rounded-lg ${className}`}>
      <div className="flex items-center space-x-3 mb-3">
        <Skeleton width={40} height={40} rounded />
        <div className="flex-1">
          <Skeleton height="1rem" className="mb-2" />
          <Skeleton height="0.75rem" width="60%" />
        </div>
      </div>
      <SkeletonText lines={2} />
    </div>
  );
}

export function SkeletonPlayerCard({ className = '' }: { className?: string }) {
  return (
    <div className={`p-3 border border-border rounded-lg ${className}`}>
      <div className="flex items-center space-x-2 mb-2">
        <Skeleton width={32} height={32} rounded />
        <Skeleton height="1rem" width="40%" />
      </div>
      <Skeleton height="0.75rem" width="60%" className="mb-1" />
      <Skeleton height="0.75rem" width="80%" />
    </div>
  );
}

export function SkeletonChatMessage({ className = '' }: { className?: string }) {
  return (
    <div className={`flex gap-3 p-3 ${className}`}>
      <Skeleton width={32} height={32} rounded />
      <div className="flex-1">
        <div className="flex items-center gap-2 mb-2">
          <Skeleton height="0.875rem" width="80px" />
          <Skeleton height="0.75rem" width="60px" />
        </div>
        <SkeletonText lines={Math.floor(Math.random() * 3) + 1} />
      </div>
    </div>
  );
}