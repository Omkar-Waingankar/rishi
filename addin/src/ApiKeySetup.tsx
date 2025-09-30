import React, { useState } from 'react';

interface ApiKeySetupProps {
  onApiKeySubmit: (apiKey: string) => Promise<void>;
}

const ApiKeySetup: React.FC<ApiKeySetupProps> = ({ onApiKeySubmit }) => {
  const [apiKey, setApiKey] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!apiKey.trim()) return;

    setIsSubmitting(true);
    setError(null);

    try {
      await onApiKeySubmit(apiKey.trim());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save API key');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="api-key-setup">
      <div className="api-key-setup-content">
        <h3>Welcome to Rishi</h3>
        <p>To get started, please enter your Anthropic API key.</p>

        <form onSubmit={handleSubmit}>
          <div className="api-key-input-group">
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="sk-ant-..."
              disabled={isSubmitting}
              autoFocus
            />
          </div>

          {error && (
            <div className="api-key-error">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={isSubmitting || !apiKey.trim()}
            className="api-key-submit"
          >
            {isSubmitting ? 'Saving...' : 'Save and Connect'}
          </button>
        </form>

        <div className="api-key-help">
          <p>Don't have an API key? <a href="https://console.anthropic.com/" target="_blank" rel="noopener noreferrer">Get one here</a></p>
        </div>
      </div>
    </div>
  );
};

export default ApiKeySetup;