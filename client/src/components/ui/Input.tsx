import React, { forwardRef } from 'react';
import styles from './Input.module.css';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'filled';
  fullWidth?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
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
  className = '',
  id,
  ...props
}, ref) => {
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
  const hasError = Boolean(error);
  const hasIcons = Boolean(leftIcon || rightIcon);

  const wrapperClasses = [
    styles.wrapper,
    fullWidth ? styles.fullWidth : '',
  ].filter(Boolean).join(' ');

  const inputClasses = [
    styles.input,
    styles[variant],
    styles[size],
    hasError ? styles.error : '',
    hasIcons ? styles.withIcons : '',
    leftIcon ? styles.withLeftIcon : '',
    rightIcon ? styles.withRightIcon : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className={wrapperClasses}>
      {label && (
        <label htmlFor={inputId} className={styles.label}>
          {label}
        </label>
      )}
      
      <div className={styles.inputContainer}>
        {leftIcon && (
          <div className={styles.leftIcon}>
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
          <div className={styles.rightIcon}>
            {rightIcon}
          </div>
        )}
      </div>
      
      {(error || helperText) && (
        <div className={`${styles.helpText} ${hasError ? styles.errorText : ''}`}>
          {error || helperText}
        </div>
      )}
    </div>
  );
});