import React, { useState, useEffect } from 'react';

interface ApiKeySetupProps {
  onApiKeySubmit: (apiKey: string) => Promise<void>;
}

const ApiKeySetup: React.FC<ApiKeySetupProps> = ({ onApiKeySubmit }) => {
  const [apiKey, setApiKey] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [isValid, setIsValid] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Validate API key format and test with Anthropic API via backend
  useEffect(() => {
    const validateApiKey = async () => {
      const trimmedKey = apiKey.trim();

      // Basic format validation
      if (!trimmedKey) {
        setIsValid(false);
        return;
      }

      // Check for sk-ant prefix
      if (!trimmedKey.startsWith('sk-ant-')) {
        setIsValid(false);
        return;
      }

      // Check minimum length (Anthropic keys are typically longer)
      if (trimmedKey.length < 20) {
        setIsValid(false);
        return;
      }

      // Test with backend validation endpoint
      setIsValidating(true);
      try {
        const response = await fetch('http://localhost:8082/validate_api_key', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            api_key: trimmedKey,
          }),
        });

        if (response.ok) {
          const data = await response.json();
          setIsValid(data.valid);
        } else {
          setIsValid(false);
        }
      } catch (err) {
        // Network error or backend unreachable
        setIsValid(false);
      } finally {
        setIsValidating(false);
      }
    };

    // Debounce the validation to avoid too many API calls
    const timeoutId = setTimeout(validateApiKey, 500);
    return () => clearTimeout(timeoutId);
  }, [apiKey]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!isValid || isSubmitting) return;

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
      <h3>Welcome to Rishi</h3>
      <p>To get started, please enter your Anthropic API key. Don't have an API key? <a href="https://console.anthropic.com/" target="_blank" rel="noopener noreferrer">Get one here</a></p>

      <form onSubmit={handleSubmit}>
        <div className="api-key-input-wrapper">
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="sk-ant-..."
            disabled={isSubmitting}
            autoFocus
            className="api-key-input"
          />
          <button
            type="submit"
            disabled={!isValid || isSubmitting || isValidating}
            className={`api-key-submit-arrow ${isValid && !isSubmitting && !isValidating ? 'active' : ''} ${isValidating ? 'validating' : ''}`}
          >
            {isSubmitting || isValidating ? (
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

        {error && (
          <div className="api-key-error">
            {error}
          </div>
        )}
      </form>

    </div>
  );
};

export default ApiKeySetup;