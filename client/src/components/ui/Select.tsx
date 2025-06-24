import React, { useState, useRef, useEffect, forwardRef } from 'react';
import styles from './Select.module.css';

interface Option {
  value: string;
  label: string;
  disabled?: boolean;
}

interface SelectProps {
  options: Option[];
  value?: string;
  onChange: (value: string) => void;
  placeholder?: string;
  label?: string;
  error?: string;
  helperText?: string;
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  fullWidth?: boolean;
  className?: string;
  id?: string;
}

export const Select = forwardRef<HTMLButtonElement, SelectProps>(({
  options,
  value,
  onChange,
  placeholder = 'Select an option...',
  label,
  error,
  helperText,
  size = 'md',
  disabled = false,
  fullWidth = false,
  className = '',
  id,
}, ref) => {
  const [isOpen, setIsOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const selectRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLUListElement>(null);
  
  const selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;
  const hasError = Boolean(error);
  
  const selectedOption = options.find(option => option.value === value);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (selectRef.current && !selectRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setFocusedIndex(-1);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (disabled) return;

    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
          setFocusedIndex(value ? options.findIndex(opt => opt.value === value) : 0);
        } else if (focusedIndex >= 0 && !options[focusedIndex]?.disabled) {
          onChange(options[focusedIndex].value);
          setIsOpen(false);
          setFocusedIndex(-1);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        setFocusedIndex(-1);
        break;
      case 'ArrowDown':
        event.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
          setFocusedIndex(0);
        } else {
          const nextIndex = Math.min(focusedIndex + 1, options.length - 1);
          setFocusedIndex(nextIndex);
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (isOpen) {
          const prevIndex = Math.max(focusedIndex - 1, 0);
          setFocusedIndex(prevIndex);
        }
        break;
    }
  };

  const handleOptionClick = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
    setFocusedIndex(-1);
    selectRef.current?.focus();
  };

  const wrapperClasses = [
    styles.wrapper,
    fullWidth ? styles.fullWidth : '',
  ].filter(Boolean).join(' ');

  const selectClasses = [
    styles.select,
    styles[size],
    hasError ? styles.error : '',
    isOpen ? styles.open : '',
    disabled ? styles.disabled : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className={wrapperClasses}>
      {label && (
        <label htmlFor={selectId} className={styles.label}>
          {label}
        </label>
      )}
      
      <div className={styles.selectContainer}>
        <button
          ref={ref || selectRef}
          id={selectId}
          type="button"
          className={selectClasses}
          onClick={() => !disabled && setIsOpen(!isOpen)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          aria-expanded={isOpen}
          aria-haspopup="listbox"
        >
          <span className={styles.selectValue}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
          <div className={`${styles.chevron} ${isOpen ? styles.chevronOpen : ''}`}>
            <svg width="12" height="8" viewBox="0 0 12 8" fill="none">
              <path d="M1 1.5L6 6.5L11 1.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
        </button>
        
        {isOpen && (
          <ul
            ref={listRef}
            className={styles.dropdown}
            role="listbox"
          >
            {options.map((option, index) => (
              <li key={option.value}>
                <button
                  type="button"
                  className={`${styles.option} ${
                    index === focusedIndex ? styles.focused : ''
                  } ${
                    option.value === value ? styles.selected : ''
                  } ${
                    option.disabled ? styles.optionDisabled : ''
                  }`}
                  onClick={() => !option.disabled && handleOptionClick(option.value)}
                  disabled={option.disabled}
                  role="option"
                  aria-selected={option.value === value}
                >
                  {option.label}
                </button>
              </li>
            ))}
          </ul>
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