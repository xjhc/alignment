import React, { useState, useRef, useEffect, forwardRef } from 'react';

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
    'flex flex-col gap-1',
    fullWidth ? 'w-full' : '',
  ].filter(Boolean).join(' ');

  const sizeClasses = {
    sm: 'px-3 py-2 text-xs rounded-sm min-h-[32px]',
    md: 'px-3 py-2 text-sm rounded-md min-h-[40px]',
    lg: 'px-4 py-3 text-base rounded-md min-h-[48px]',
  };

  const selectClasses = [
    'w-full flex items-center justify-between gap-2 font-inherit font-normal text-text-primary bg-background-primary border border-border transition-all duration-150 cursor-pointer text-left',
    'focus:outline-none focus:border-primary focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]',
    'hover:enabled:border-primary',
    sizeClasses[size],
    hasError ? '!border-danger !shadow-[0_0_0_3px_rgba(239,68,68,0.1)]' : '',
    disabled ? 'opacity-50 cursor-not-allowed bg-background-secondary' : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className={wrapperClasses}>
      {label && (
        <label htmlFor={selectId} className="text-sm font-medium text-text-primary mb-1">
          {label}
        </label>
      )}
      
      <div className="relative">
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
          <span className={`flex-1 overflow-hidden text-ellipsis whitespace-nowrap ${selectedOption ? 'text-text-primary' : 'text-text-muted'}`}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
          <div className={`flex items-center justify-center text-text-muted transition-transform duration-150 flex-shrink-0 ${isOpen ? 'rotate-180' : ''}`}>
            <svg width="12" height="8" viewBox="0 0 12 8" fill="none">
              <path d="M1 1.5L6 6.5L11 1.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
        </button>
        
        {isOpen && (
          <ul
            ref={listRef}
            className="absolute top-full left-0 right-0 z-[100] bg-background-primary border border-border rounded-md shadow-lg mt-0.5 max-h-[200px] overflow-y-auto list-none py-1"
            role="listbox"
          >
            {options.map((option, index) => (
              <li key={option.value}>
                <button
                  type="button"
                  className={`w-full px-3 py-2 text-inherit font-inherit text-text-primary bg-none border-none cursor-pointer text-left transition-colors duration-150 hover:enabled:bg-background-secondary ${
                    index === focusedIndex ? '!bg-background-secondary' : ''
                  } ${
                    option.value === value ? '!bg-primary !text-background-primary hover:!bg-primary' : ''
                  } ${
                    option.disabled ? 'opacity-50 cursor-not-allowed' : ''
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
        <div className={`text-xs mt-1 ${hasError ? 'text-danger' : 'text-text-secondary'}`}>
          {error || helperText}
        </div>
      )}
    </div>
  );
});