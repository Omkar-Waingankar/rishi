import React, { useState, useRef } from 'react';
import { InputBoxProps } from './types';
import ModelDropdown from './ModelDropdown';

interface DropdownOption {
  value: string;
  label: string;
}

const InputBox: React.FC<InputBoxProps> = ({ onSendMessage, disabled }) => {
  const [message, setMessage] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState<string>('claude-3.5-sonnet');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSendMessage(message);
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
    { value: 'claude-3.5-sonnet', label: 'Claude 3.5 Sonnet' },
    { value: 'claude-3-haiku', label: 'Claude 3 Haiku' },
    { value: 'gpt-4', label: 'GPT-4' }
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
            
            <button 
              type="submit" 
              disabled={disabled || !message.trim()}
              className="send-button"
              aria-label={disabled ? 'Sending...' : 'Send message'}
            >
              {disabled ? (
                <svg className="send-icon" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="3"/>
                </svg>
              ) : (
                <svg className="send-icon" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M2,21L23,12L2,3V10L17,12L2,14V21Z"/>
                </svg>
              )}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

export default InputBox;