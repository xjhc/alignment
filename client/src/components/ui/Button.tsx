import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  fullWidth?: boolean;
  isLoading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const LoadingSpinner: React.FC<{ size: 'sm' | 'md' | 'lg' }> = ({ size }) => {
  const sizeClasses = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  return (
    <div className={`animate-spin rounded-full border-2 border-current border-t-transparent ${sizeClasses[size]}`} />
  );
};

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  fullWidth = false,
  disabled = false,
  isLoading = false,
  leftIcon,
  rightIcon,
  className = '',
  ...props
}) => {
  const baseClasses = 'relative inline-flex items-center justify-center gap-2 font-medium transition-all duration-150 border border-transparent cursor-pointer select-none font-sans no-underline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none';
  
  const sizeClasses = {
    sm: 'px-3 py-1 text-xs rounded-sm min-h-[28px]',
    md: 'px-4 py-2 text-sm rounded-md min-h-[36px]',
    lg: 'px-6 py-3 text-base rounded-md min-h-[44px]',
  };

  const variantClasses = {
    primary: 'bg-primary text-background-primary hover:enabled:-translate-y-px hover:enabled:shadow-md active:enabled:translate-y-0 active:enabled:shadow-sm',
    secondary: 'bg-background-secondary border-border text-text-primary hover:enabled:bg-background-tertiary hover:enabled:border-primary hover:enabled:-translate-y-px active:enabled:translate-y-0 active:enabled:bg-background-quaternary',
    danger: 'bg-danger text-white hover:enabled:-translate-y-px hover:enabled:shadow-md active:enabled:translate-y-0 active:enabled:shadow-sm',
    ghost: 'bg-transparent text-text-primary hover:enabled:bg-background-secondary active:enabled:bg-background-tertiary',
    outline: 'bg-transparent border-primary text-primary hover:enabled:bg-primary hover:enabled:text-background-primary hover:enabled:-translate-y-px active:enabled:translate-y-0',
  };

  const widthClass = fullWidth ? 'w-full' : '';

  const classes = [
    baseClasses,
    sizeClasses[size],
    variantClasses[variant],
    widthClass,
    className,
  ].filter(Boolean).join(' ');

  const isDisabled = disabled || isLoading;

  return (
    <button
      {...props}
      disabled={isDisabled}
      className={classes}
    >
      {isLoading ? (
        <LoadingSpinner size={size} />
      ) : (
        <>
          {leftIcon && <span className="flex-shrink-0">{leftIcon}</span>}
          <span className={leftIcon || rightIcon ? 'flex-1' : ''}>{children}</span>
          {rightIcon && <span className="flex-shrink-0">{rightIcon}</span>}
        </>
      )}
    </button>
  );
};