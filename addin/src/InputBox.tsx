import React, { useState, useRef } from 'react';
import { InputBoxProps } from './types';

const InputBox: React.FC<InputBoxProps> = ({ onSendMessage, disabled }) => {
  const [message, setMessage] = useState<string>('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSendMessage(message);
      setMessage('');
      if (textareaRef.current) {
        textareaRef.current.style.height = '40px';
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
    textarea.style.height = '40px';
    const scrollHeight = textarea.scrollHeight;
    const maxHeight = 120;
    textarea.style.height = Math.min(scrollHeight, maxHeight) + 'px';
  };

  return (
    <div className="input-box">
      <form onSubmit={handleSubmit}>
        <div className="input-container">
          <textarea
            ref={textareaRef}
            value={message}
            onChange={handleTextareaChange}
            onKeyDown={handleKeyDown}
            placeholder="Type your message... (Shift+Enter for new line)"
            disabled={disabled}
            rows={1}
          />
          <button 
            type="submit" 
            disabled={disabled || !message.trim()}
            className="send-button"
          >
            {disabled ? '...' : 'Send'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default InputBox;