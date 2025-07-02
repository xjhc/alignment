import React, { useState, useEffect } from 'react';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FADE_IN, SLIDE_IN_UP } from '../../utils/animations';

interface SettingsState {
  masterVolume: number;
  sfxVolume: number;
  musicVolume: number;
  loebmateAssistantEnabled: boolean;
  keyboardShortcutsEnabled: boolean;
  animationsEnabled: boolean;
  highContrastMode: boolean;
  reducedMotion: boolean;
}

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_SETTINGS: SettingsState = {
  masterVolume: 70,
  sfxVolume: 80,
  musicVolume: 60,
  loebmateAssistantEnabled: true,
  keyboardShortcutsEnabled: true,
  animationsEnabled: true,
  highContrastMode: false,
  reducedMotion: false,
};

const SETTINGS_STORAGE_KEY = 'alignment-game-settings';

export const SettingsModal: React.FC<SettingsModalProps> = ({ isOpen, onClose }) => {
  const [settings, setSettings] = useState<SettingsState>(DEFAULT_SETTINGS);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  // Load settings from localStorage on mount
  useEffect(() => {
    const savedSettings = localStorage.getItem(SETTINGS_STORAGE_KEY);
    if (savedSettings) {
      try {
        const parsed = JSON.parse(savedSettings);
        setSettings({ ...DEFAULT_SETTINGS, ...parsed });
      } catch (error) {
        console.error('Failed to parse saved settings:', error);
      }
    }
  }, []);

  // Save settings to localStorage
  const saveSettings = async () => {
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(settings));
    setHasUnsavedChanges(false);
    
    // Apply settings to the document/application
    await applySettings(settings);
  };

  // Apply settings to the application
  const applySettings = async (settingsToApply: SettingsState) => {
    // Apply audio settings (when sound system is implemented)
    if (window.soundManager) {
      window.soundManager.setMasterVolume(settingsToApply.masterVolume / 100);
      window.soundManager.setSfxVolume(settingsToApply.sfxVolume / 100);
      window.soundManager.setMusicVolume(settingsToApply.musicVolume / 100);
    }

    // Apply visual settings
    const root = document.documentElement;
    
    if (settingsToApply.reducedMotion) {
      root.style.setProperty('--duration-fast', '0ms');
      root.style.setProperty('--duration-medium', '0ms');
      root.style.setProperty('--duration-slow', '0ms');
    } else {
      root.style.removeProperty('--duration-fast');
      root.style.removeProperty('--duration-medium');
      root.style.removeProperty('--duration-slow');
    }

    if (settingsToApply.highContrastMode) {
      root.classList.add('high-contrast');
    } else {
      root.classList.remove('high-contrast');
    }

    // Send gameplay settings to backend if in a game
    try {
      await fetch('/api/users/me/settings', {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          disableLoebmateHints: !settingsToApply.loebmateAssistantEnabled
        }),
      });
    } catch (error) {
      console.warn('Failed to update backend settings:', error);
      // Continue anyway - this is not critical for local settings
    }
  };

  const updateSetting = <K extends keyof SettingsState>(
    key: K,
    value: SettingsState[K]
  ) => {
    setSettings(prev => ({ ...prev, [key]: value }));
    setHasUnsavedChanges(true);
  };

  const resetToDefaults = () => {
    setSettings(DEFAULT_SETTINGS);
    setHasUnsavedChanges(true);
  };

  const handleClose = async () => {
    if (hasUnsavedChanges) {
      const shouldSave = window.confirm('You have unsaved changes. Save before closing?');
      if (shouldSave) {
        await saveSettings();
      } else {
        // Reload settings from localStorage to discard changes
        const savedSettings = localStorage.getItem(SETTINGS_STORAGE_KEY);
        if (savedSettings) {
          try {
            const parsed = JSON.parse(savedSettings);
            setSettings({ ...DEFAULT_SETTINGS, ...parsed });
          } catch (error) {
            setSettings(DEFAULT_SETTINGS);
          }
        } else {
          setSettings(DEFAULT_SETTINGS);
        }
        setHasUnsavedChanges(false);
      }
    }
    onClose();
  };

  const SliderControl: React.FC<{
    label: string;
    value: number;
    onChange: (value: number) => void;
    min?: number;
    max?: number;
    step?: number;
  }> = ({ label, value, onChange, min = 0, max = 100, step = 1 }) => (
    <div className="flex items-center justify-between">
      <label className="text-sm font-medium text-text-primary">{label}</label>
      <div className="flex items-center gap-3">
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          className="w-24 h-2 bg-background-tertiary rounded-lg appearance-none cursor-pointer slider"
        />
        <span className="w-10 text-xs font-mono text-text-secondary">{value}%</span>
      </div>
    </div>
  );

  const ToggleControl: React.FC<{
    label: string;
    description?: string;
    checked: boolean;
    onChange: (checked: boolean) => void;
  }> = ({ label, description, checked, onChange }) => (
    <div className="flex items-start justify-between">
      <div className="flex-1">
        <label className="text-sm font-medium text-text-primary cursor-pointer">
          {label}
        </label>
        {description && (
          <p className="text-xs text-text-muted mt-1">{description}</p>
        )}
      </div>
      <label className="relative inline-flex items-center cursor-pointer">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="sr-only"
        />
        <div className={`w-11 h-6 rounded-full transition-colors ${
          checked ? 'bg-primary' : 'bg-background-tertiary'
        }`}>
          <div className={`w-5 h-5 bg-white rounded-full shadow transform transition-transform ${
            checked ? 'translate-x-5' : 'translate-x-0.5'
          } mt-0.5`} />
        </div>
      </label>
    </div>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Settings"
      size="lg"
    >
      <div className={`space-y-6 ${FADE_IN}`}>
        {/* Audio Settings */}
        <section>
          <h3 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
            🔊 Audio Settings
          </h3>
          <div className="space-y-4 bg-background-secondary rounded-lg p-4">
            <SliderControl
              label="Master Volume"
              value={settings.masterVolume}
              onChange={(value) => updateSetting('masterVolume', value)}
            />
            <SliderControl
              label="Sound Effects"
              value={settings.sfxVolume}
              onChange={(value) => updateSetting('sfxVolume', value)}
            />
            <SliderControl
              label="Background Music"
              value={settings.musicVolume}
              onChange={(value) => updateSetting('musicVolume', value)}
            />
          </div>
        </section>

        {/* Gameplay Settings */}
        <section>
          <h3 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
            🎮 Gameplay Settings
          </h3>
          <div className="space-y-4 bg-background-secondary rounded-lg p-4">
            <ToggleControl
              label="Loebmate Assistant"
              description="Receive helpful phase-specific hints and guidance during your first games"
              checked={settings.loebmateAssistantEnabled}
              onChange={(checked) => updateSetting('loebmateAssistantEnabled', checked)}
            />
            <ToggleControl
              label="Keyboard Shortcuts"
              description="Enable keyboard shortcuts for quick actions (Cmd/Ctrl+K for command palette)"
              checked={settings.keyboardShortcutsEnabled}
              onChange={(checked) => updateSetting('keyboardShortcutsEnabled', checked)}
            />
          </div>
        </section>

        {/* Accessibility Settings */}
        <section>
          <h3 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
            ♿ Accessibility
          </h3>
          <div className="space-y-4 bg-background-secondary rounded-lg p-4">
            <ToggleControl
              label="Reduced Motion"
              description="Minimize animations and transitions for users sensitive to motion"
              checked={settings.reducedMotion}
              onChange={(checked) => updateSetting('reducedMotion', checked)}
            />
            <ToggleControl
              label="High Contrast Mode"
              description="Increase contrast for better visibility"
              checked={settings.highContrastMode}
              onChange={(checked) => updateSetting('highContrastMode', checked)}
            />
            <ToggleControl
              label="Animations Enabled"
              description="Enable smooth animations and transitions"
              checked={settings.animationsEnabled}
              onChange={(checked) => updateSetting('animationsEnabled', checked)}
            />
          </div>
        </section>

        {/* Keyboard Shortcuts Reference */}
        <section>
          <h3 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
            ⌨️ Keyboard Shortcuts
          </h3>
          <div className="bg-background-secondary rounded-lg p-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <div className="flex justify-between">
                <span className="text-text-secondary">Command Palette:</span>
                <code className="text-xs bg-background-tertiary px-2 py-1 rounded font-mono">Cmd/Ctrl + K</code>
              </div>
              <div className="flex justify-between">
                <span className="text-text-secondary">Focus Chat:</span>
                <code className="text-xs bg-background-tertiary px-2 py-1 rounded font-mono">Ctrl + /</code>
              </div>
              <div className="flex justify-between">
                <span className="text-text-secondary">Help:</span>
                <code className="text-xs bg-background-tertiary px-2 py-1 rounded font-mono">?</code>
              </div>
              <div className="flex justify-between">
                <span className="text-text-secondary">Close Modals:</span>
                <code className="text-xs bg-background-tertiary px-2 py-1 rounded font-mono">Esc</code>
              </div>
            </div>
          </div>
        </section>
      </div>

      {/* Action Buttons */}
      <div className="flex items-center justify-between pt-6 border-t border-border">
        <Button
          variant="secondary"
          onClick={resetToDefaults}
          disabled={JSON.stringify(settings) === JSON.stringify(DEFAULT_SETTINGS)}
        >
          Reset to Defaults
        </Button>
        
        <div className="flex items-center gap-3">
          <Button variant="secondary" onClick={handleClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={async () => {
              await saveSettings();
              onClose();
            }}
            disabled={!hasUnsavedChanges}
          >
            Save Changes
          </Button>
        </div>
      </div>
    </Modal>
  );
};