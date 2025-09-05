import React, { useState, useRef } from 'react';
import { InputBoxProps } from './types';
import ModelDropdown from './ModelDropdown';

interface DropdownOption {
  value: string;
  label: string;
}

const InputBox: React.FC<InputBoxProps> = ({ onSendMessage, disabled, isStreaming, onStopStreaming, safeRoot }) => {
  const [message, setMessage] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState<string>('claude-4-sonnet');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSendMessage(message, selectedModel);
      setMessage('');
      if (textareaRef.current) {
        textareaRef.current.style.height = '24px';
      }
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent<HTMLFormElement>);
    }
  };

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
    setMessage(e.target.value);
    
    const textarea = e.target;
    textarea.style.height = '24px';
    const scrollHeight = textarea.scrollHeight;
    const maxHeight = 120;
    textarea.style.height = Math.min(scrollHeight, maxHeight) + 'px';
  };

  const handleModelChange = (value: string): void => {
    setSelectedModel(value);
    localStorage.setItem('selectedModel', value);
  };

  const modelOptions: DropdownOption[] = [
    { value: 'claude-4-sonnet', label: 'Claude 4 Sonnet' },
    { value: 'claude-3.7-sonnet', label: 'Claude 3.7 Sonnet' }
  ];

  // Load saved model preference on component mount
  React.useEffect(() => {
    const savedModel = localStorage.getItem('selectedModel');
    if (savedModel) {
      setSelectedModel(savedModel);
    }
  }, []);

  return (
    <div className="input-box">
      <form onSubmit={handleSubmit}>
        <div className="input-container">
          <div className="input-text-section">
            <textarea
              ref={textareaRef}
              value={message}
              onChange={handleTextareaChange}
              onKeyDown={handleKeyDown}
              placeholder="Plan, build, analyze anything"
              disabled={disabled}
              rows={1}
            />
          </div>
          
          <div className="input-footer">
            <ModelDropdown
              value={selectedModel}
              onChange={handleModelChange}
              disabled={disabled}
              options={modelOptions}
            />
            
            {isStreaming ? (
              <button 
                type="button"
                onClick={onStopStreaming}
                className="send-button stop-button"
                aria-label="Stop streaming"
              >
                <svg className="send-icon" viewBox="0 0 24 24" fill="currentColor">
                  <rect x="6" y="6" width="12" height="12" rx="2"/>
                </svg>
              </button>
            ) : (
              <button 
                type="submit" 
                disabled={disabled || !message.trim() || !safeRoot}
                className="send-button"
                aria-label="Send message"
              >
                <svg className="send-icon" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M2,21L23,12L2,3V10L17,12L2,14V21Z"/>
                </svg>
              </button>
            )}
          </div>
        </div>
      </form>
    </div>
  );
};

export default InputBox;