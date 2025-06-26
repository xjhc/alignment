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

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      ref={pickerRef}
      className="fixed z-50 bg-background-primary border border-border rounded-lg shadow-lg p-2 animation-scale-in"
      style={{
        top: position.top,
        left: position.left,
      }}
    >
      <div className="grid grid-cols-3 gap-1">
        {ALLOWED_EMOJIS.map(({ emoji, name, label }) => (
          <button
            key={name}
            onClick={() => {
              onEmojiSelect(emoji);
              onClose();
            }}
            className="w-10 h-10 flex items-center justify-center text-lg hover:bg-background-secondary rounded transition-colors"
            title={label}
          >
            {emoji}
          </button>
        ))}
      </div>
    </div>
  );
};