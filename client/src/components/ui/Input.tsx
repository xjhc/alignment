import React, { forwardRef } from 'react';

interface InputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  helperText?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'filled';
  fullWidth?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  hasError?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(({
  label,
  error,
  helperText,
  size = 'md',
  variant = 'default',
  fullWidth = false,
  leftIcon,
  rightIcon,
  hasError = false,
  className = '',
  id,
  ...props
}, ref) => {
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
  
  // hasError prop takes precedence, but fallback to error prop for backward compatibility
  const showError = hasError || Boolean(error);

  const wrapperClasses = [
    'flex flex-col gap-1',
    fullWidth ? 'w-full' : '',
  ].filter(Boolean).join(' ');

  const sizeClasses = {
    sm: 'px-3 py-2 text-xs rounded-sm min-h-[32px]',
    md: 'px-3 py-2 text-sm rounded-md min-h-[40px]',
    lg: 'px-4 py-3 text-base rounded-md min-h-[48px]',
  };

  const variantClasses = {
    default: 'bg-background-primary border-border focus:bg-background-primary',
    filled: 'bg-background-secondary border-transparent focus:bg-background-primary focus:border-primary',
  };

  const iconPadding = {
    left: leftIcon ? {
      sm: 'pl-10',
      md: 'pl-10', 
      lg: 'pl-11'
    }[size] : '',
    right: rightIcon ? {
      sm: 'pr-10',
      md: 'pr-10',
      lg: 'pr-11'  
    }[size] : '',
  };

  const inputClasses = [
    'w-full font-inherit font-normal text-text-primary border transition-all duration-150',
    'focus:outline-none focus:border-primary focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]',
    'placeholder:text-text-muted disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-background-secondary',
    sizeClasses[size],
    variantClasses[variant],
    showError ? '!border-danger !shadow-[0_0_0_3px_rgba(239,68,68,0.1)]' : '',
    iconPadding.left,
    iconPadding.right,
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className={wrapperClasses}>
      {label && (
        <label htmlFor={inputId} className="text-sm font-medium text-text-primary mb-1">
          {label}
        </label>
      )}
      
      <div className="relative flex items-center">
        {leftIcon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 flex items-center justify-center text-text-muted z-10">
            {leftIcon}
          </div>
        )}
        
        <input
          {...props}
          ref={ref}
          id={inputId}
          className={inputClasses}
        />
        
        {rightIcon && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center justify-center text-text-muted z-10">
            {rightIcon}
          </div>
        )}
      </div>
      
      {(error || helperText) && (
        <div className={`text-xs mt-1 ${showError ? 'text-danger' : 'text-text-secondary'}`}>
          {error || helperText}
        </div>
      )}
    </div>
  );
});