import React, { useState, useEffect } from 'react';

interface SettingsProps {
  onClose: () => void;
  apiKeyStatus: {
    anthropic: { has_key: boolean; api_key: string };
    openai: { has_key: boolean; api_key: string };
  } | null;
  onApiKeyUpdate: (provider: string, apiKey: string) => Promise<void>;
}

const Settings: React.FC<SettingsProps> = ({ onClose, apiKeyStatus, onApiKeyUpdate }) => {
  const [anthropicKey, setAnthropicKey] = useState<string>('');
  const [openaiKey, setOpenaiKey] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<{ [key: string]: boolean }>({});
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  // Initialize form with existing keys
  useEffect(() => {
    if (apiKeyStatus) {
      setAnthropicKey(apiKeyStatus.anthropic.has_key ? apiKeyStatus.anthropic.api_key : '');
      setOpenaiKey(apiKeyStatus.openai.has_key ? apiKeyStatus.openai.api_key : '');
    }
  }, [apiKeyStatus]);

  const handleSubmit = async (provider: string, apiKey: string) => {
    if (!apiKey.trim()) return;

    setIsSubmitting(prev => ({ ...prev, [provider]: true }));
    setErrors(prev => ({ ...prev, [provider]: '' }));

    try {
      await onApiKeyUpdate(provider, apiKey.trim());
    } catch (error) {
      setErrors(prev => ({ 
        ...prev, 
        [provider]: error instanceof Error ? error.message : 'Failed to save API key' 
      }));
    } finally {
      setIsSubmitting(prev => ({ ...prev, [provider]: false }));
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent, provider: string, apiKey: string) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSubmit(provider, apiKey);
    }
  };

  return (
    <div className="settings-fullscreen">
      <div className="settings-header">
        <h2>Settings</h2>
        <button className="settings-close-button" onClick={onClose}>
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M12 4L4 12M4 4L12 12"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
      </div>

      <div className="settings-content">
          {/* Anthropic API Key Section */}
          <div className="settings-section">
            <div className="settings-section-header">
              <h3>Anthropic API Key</h3>
            </div>
            <p className="settings-description">
              You can put in your Anthropic key to use Claude at cost. When enabled, this key will be used for all models beginning with 'claude-'.
            </p>
            <div className="api-key-input-wrapper">
              <input
                type="password"
                value={anthropicKey}
                onChange={(e) => setAnthropicKey(e.target.value)}
                onKeyPress={(e) => handleKeyPress(e, 'anthropic', anthropicKey)}
                placeholder="Enter your Anthropic API Key"
                className="api-key-input"
                disabled={isSubmitting.anthropic}
              />
              <button
                onClick={() => handleSubmit('anthropic', anthropicKey)}
                disabled={!anthropicKey.trim() || isSubmitting.anthropic}
                className={`api-key-submit-button ${anthropicKey.trim() && !isSubmitting.anthropic ? 'active' : ''}`}
              >
                {isSubmitting.anthropic ? (
                  <div className="spinner-small" />
                ) : (
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 16 16"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M8 3L13 8L8 13M13 8H3"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                )}
              </button>
            </div>
            {errors.anthropic && (
              <div className="api-key-error">
                {errors.anthropic}
              </div>
            )}
          </div>

          {/* OpenAI API Key Section */}
          <div className="settings-section">
            <div className="settings-section-header">
              <h3>OpenAI API Key</h3>
            </div>
            <p className="settings-description">
              You can put in your OpenAI key to use OpenAI models at cost.
            </p>
            <div className="api-key-input-wrapper">
              <input
                type="password"
                value={openaiKey}
                onChange={(e) => setOpenaiKey(e.target.value)}
                onKeyPress={(e) => handleKeyPress(e, 'openai', openaiKey)}
                placeholder="Enter your OpenAI API Key"
                className="api-key-input"
                disabled={isSubmitting.openai}
              />
              <button
                onClick={() => handleSubmit('openai', openaiKey)}
                disabled={!openaiKey.trim() || isSubmitting.openai}
                className={`api-key-submit-button ${openaiKey.trim() && !isSubmitting.openai ? 'active' : ''}`}
              >
                {isSubmitting.openai ? (
                  <div className="spinner-small" />
                ) : (
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 16 16"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M8 3L13 8L8 13M13 8H3"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                )}
              </button>
            </div>
            {errors.openai && (
              <div className="api-key-error">
                {errors.openai}
              </div>
            )}
          </div>
      </div>
    </div>
  );
};

export default Settings;
