import React, { useState, useRef, useEffect } from 'react';

interface EmojiPickerProps {
  isOpen: boolean;
  onClose: () => void;
  onEmojiSelect: (emoji: string) => void;
  anchorElement?: HTMLElement | null;
}

const ALLOWED_EMOJIS = [
  { emoji: '👍', name: 'thumbs_up', label: 'Thumbs Up' },
  { emoji: '👎', name: 'thumbs_down', label: 'Thumbs Down' },
  { emoji: '🤔', name: 'thinking_face', label: 'Thinking' },
  { emoji: '👀', name: 'eyes', label: 'Eyes' },
  { emoji: '😂', name: 'joy', label: 'Joy' },
  { emoji: '🔥', name: 'fire', label: 'Fire' },
];

export const EmojiPicker: React.FC<EmojiPickerProps> = ({
  isOpen,
  onClose,
  onEmojiSelect,
  anchorElement,
}) => {
  const [position, setPosition] = useState({ top: 0, left: 0 });
  const [selectedIndex, setSelectedIndex] = useState(0);
  const pickerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isOpen && anchorElement) {
      const rect = anchorElement.getBoundingClientRect();
      const pickerHeight = 120; // Approximate height of picker
      const pickerWidth = 200; // Approximate width of picker
      
      let top = rect.top - pickerHeight - 8; // Position above the button
      let left = rect.left;
      
      // Ensure picker stays within viewport
      if (top < 8) {
        top = rect.bottom + 8; // Position below if not enough space above
      }
      
      if (left + pickerWidth > window.innerWidth - 8) {
        left = window.innerWidth - pickerWidth - 8;
      }
      
      if (left < 8) {
        left = 8;
      }
      
      setPosition({ top, left });
    }
  }, [isOpen, anchorElement]);

  // Reset selection when opened
  useEffect(() => {
    if (isOpen) {
      setSelectedIndex(0);
    }
  }, [isOpen]);

  // Enhanced keyboard navigation
  useEffect(() => {
    const handleKeyboard = (event: KeyboardEvent) => {
      if (!isOpen) return;

      switch (event.key) {
        case 'Escape':
          event.preventDefault();
          onClose();
          break;
        case 'ArrowRight':
          event.preventDefault();
          setSelectedIndex(prev => (prev + 1) % ALLOWED_EMOJIS.length);
          break;
        case 'ArrowLeft':
          event.preventDefault();
          setSelectedIndex(prev => (prev - 1 + ALLOWED_EMOJIS.length) % ALLOWED_EMOJIS.length);
          break;
        case 'ArrowDown':
          event.preventDefault();
          // Move down one row (3 columns)
          setSelectedIndex(prev => (prev + 3) % ALLOWED_EMOJIS.length);
          break;
        case 'ArrowUp':
          event.preventDefault();
          // Move up one row (3 columns)
          setSelectedIndex(prev => (prev - 3 + ALLOWED_EMOJIS.length) % ALLOWED_EMOJIS.length);
          break;
        case 'Enter':
        case ' ':
          event.preventDefault();
          const selectedEmoji = ALLOWED_EMOJIS[selectedIndex];
          if (selectedEmoji) {
            onEmojiSelect(selectedEmoji.emoji);
            onClose();
          }
          break;
        case 'Tab':
          // Allow tab navigation within the picker
          const buttons = pickerRef.current?.querySelectorAll('button');
          if (buttons && event.shiftKey) {
            // Shift+Tab - go backwards
            event.preventDefault();
            setSelectedIndex(prev => (prev - 1 + ALLOWED_EMOJIS.length) % ALLOWED_EMOJIS.length);
          } else if (buttons) {
            // Tab - go forwards
            event.preventDefault();
            setSelectedIndex(prev => (prev + 1) % ALLOWED_EMOJIS.length);
          }
          break;
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyboard);
      document.addEventListener('mousedown', handleClickOutside);
      
      // Focus the picker when opened
      pickerRef.current?.focus();
    }

    return () => {
      document.removeEventListener('keydown', handleKeyboard);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, onClose, onEmojiSelect, selectedIndex]);

  if (!isOpen) return null;

  return (
    <div
      ref={pickerRef}
      className="fixed z-50 bg-background-primary border border-border rounded-lg shadow-lg p-2 animation-scale-in"
      style={{
        top: position.top,
        left: position.left,
      }}
      tabIndex={-1}
      role="dialog"
      aria-label="Emoji picker"
    >
      <div className="text-xs text-text-muted mb-2 px-1">
        Use arrow keys to navigate, Enter to select
      </div>
      <div className="grid grid-cols-3 gap-1" role="grid">
        {ALLOWED_EMOJIS.map(({ emoji, name, label }, index) => (
          <button
            key={name}
            onClick={() => {
              onEmojiSelect(emoji);
              onClose();
            }}
            onMouseEnter={() => setSelectedIndex(index)}
            className={`w-10 h-10 flex items-center justify-center text-lg rounded transition-colors ${
              index === selectedIndex 
                ? 'bg-primary text-white' 
                : 'hover:bg-background-secondary'
            }`}
            title={label}
            aria-label={`${label} emoji`}
            role="gridcell"
            tabIndex={index === selectedIndex ? 0 : -1}
          >
            {emoji}
          </button>
        ))}
      </div>
    </div>
  );
};