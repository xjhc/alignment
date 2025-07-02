import { useEffect, useState, useCallback } from 'react';

interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  action: () => void;
  preventDefault?: boolean;
}

export function useKeyboardShortcuts() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  const openCommandPalette = useCallback(() => {
    setCommandPaletteOpen(true);
  }, []);

  const closeCommandPalette = useCallback(() => {
    setCommandPaletteOpen(false);
  }, []);

  // Define global keyboard shortcuts
  const shortcuts: KeyboardShortcut[] = [
    // Command Palette (Cmd/Ctrl + K)
    {
      key: 'k',
      ctrlKey: true,
      metaKey: true, // This will match either Ctrl or Cmd
      action: openCommandPalette,
      preventDefault: true,
    },
    // Alternative command palette trigger (Cmd/Ctrl + P)
    {
      key: 'p',
      ctrlKey: true,
      metaKey: true,
      action: openCommandPalette,
      preventDefault: true,
    },
    // Quick chat focus (Ctrl + /)
    {
      key: '/',
      ctrlKey: true,
      action: () => {
        const chatInput = document.querySelector('input[placeholder*="chat" i], textarea[placeholder*="chat" i]') as HTMLInputElement;
        if (chatInput) {
          chatInput.focus();
        }
      },
      preventDefault: true,
    },
    // Quick help (?)
    {
      key: '?',
      shiftKey: true,
      action: openCommandPalette,
      preventDefault: true,
    },
    // Escape to close modals/overlays
    {
      key: 'Escape',
      action: () => {
        // Close command palette if open
        if (commandPaletteOpen) {
          closeCommandPalette();
          return;
        }
        
        // Clear chat input if focused
        const chatInput = document.querySelector('input[placeholder*="chat" i], textarea[placeholder*="chat" i]') as HTMLInputElement;
        if (chatInput && document.activeElement === chatInput && chatInput.value) {
          chatInput.value = '';
          chatInput.dispatchEvent(new Event('input', { bubbles: true }));
          return;
        }
        
        // Close any open modals/popovers
        const closeButtons = document.querySelectorAll('[data-close-on-escape]');
        closeButtons.forEach(button => {
          if (button instanceof HTMLButtonElement) {
            button.click();
          }
        });
      },
      preventDefault: false, // Don't prevent default for Escape - let components handle it
    }
  ];

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in input fields (except for Escape)
      const isInputField = event.target instanceof HTMLInputElement || 
                          event.target instanceof HTMLTextAreaElement ||
                          event.target instanceof HTMLSelectElement ||
                          (event.target as HTMLElement)?.contentEditable === 'true';
      
      if (isInputField && event.key !== 'Escape') {
        return;
      }

      for (const shortcut of shortcuts) {
        const keyMatches = shortcut.key.toLowerCase() === event.key.toLowerCase();
        const ctrlMatches = shortcut.ctrlKey ? event.ctrlKey || event.metaKey : !event.ctrlKey && !event.metaKey;
        const metaMatches = shortcut.metaKey ? event.metaKey || event.ctrlKey : !event.metaKey && !event.ctrlKey;
        const shiftMatches = shortcut.shiftKey ? event.shiftKey : !event.shiftKey;
        const altMatches = shortcut.altKey ? event.altKey : !event.altKey;

        if (keyMatches && ctrlMatches && metaMatches && shiftMatches && altMatches) {
          if (shortcut.preventDefault) {
            event.preventDefault();
            event.stopPropagation();
          }
          shortcut.action();
          break; // Only trigger the first matching shortcut
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [shortcuts, commandPaletteOpen, closeCommandPalette]);

  return {
    commandPaletteOpen,
    openCommandPalette,
    closeCommandPalette,
  };
}

// Hook for components that need to register their own keyboard shortcuts
export function useComponentKeyboardShortcuts(shortcuts: KeyboardShortcut[], enabled = true) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        const keyMatches = shortcut.key.toLowerCase() === event.key.toLowerCase();
        const ctrlMatches = shortcut.ctrlKey ? event.ctrlKey || event.metaKey : !event.ctrlKey && !event.metaKey;
        const metaMatches = shortcut.metaKey ? event.metaKey || event.ctrlKey : !event.metaKey && !event.ctrlKey;
        const shiftMatches = shortcut.shiftKey ? event.shiftKey : !event.shiftKey;
        const altMatches = shortcut.altKey ? event.altKey : !event.altKey;

        if (keyMatches && ctrlMatches && metaMatches && shiftMatches && altMatches) {
          if (shortcut.preventDefault) {
            event.preventDefault();
            event.stopPropagation();
          }
          shortcut.action();
          break;
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [shortcuts, enabled]);
}