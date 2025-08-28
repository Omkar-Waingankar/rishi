import React, { useEffect, useRef, useState } from 'react';
import { MessageListProps } from './types';

const MessageList: React.FC<MessageListProps> = ({ messages, isLoading }) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [showJumpToLatest, setShowJumpToLatest] = useState(false);
  const messageListRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const handleScroll = () => {
    if (messageListRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = messageListRef.current;
      const isScrolledUp = scrollHeight - scrollTop - clientHeight > 100;
      setShowJumpToLatest(isScrolledUp);
    }
  };

  useEffect(() => {
    if (!showJumpToLatest) {
      scrollToBottom();
    }
  }, [messages, isLoading, showJumpToLatest]);

  const formatTime = (timestamp: Date): string => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const handleReviewChanges = () => {
    // This would integrate with RStudio's diff viewer
    console.log('Review changes clicked');
  };

  return (
    <div 
      className="message-list" 
      ref={messageListRef} 
      onScroll={handleScroll}
      role="log"
      aria-label="Chat conversation"
      aria-live="polite"
    >
      {messages.map((message, index) => {
        // Only show footer for assistant messages when not currently loading
        // or if it's not the most recent assistant message
        const isLastAssistantMessage = message.sender === 'assistant' && 
          index === messages.length - 1;
        const showFooter = message.sender === 'assistant' && 
          (!isLoading || !isLastAssistantMessage);
        
        return (
          <div key={message.id} className={`message ${message.sender}${message.type ? ` ${message.type}` : ''}`} role="article">
            <div className="message-content">
              <div className="message-text" aria-label={`${message.sender} message`}>{message.text}</div>
            </div>
            {showFooter && (
              <div className="message-actions" role="toolbar" aria-label="Message actions">
                <button 
                  className="action-button primary"
                  onClick={handleReviewChanges}
                  aria-label="Review changes"
                >
                  Review changes
                </button>
                <button 
                  className="action-button"
                  onClick={() => copyToClipboard(message.text)}
                  aria-label="Copy message"
                >
                  <span className="action-button-icon">📋</span>
                  Copy
                </button>
                <button 
                  className="action-button"
                  aria-label="Like message"
                >
                  <span className="action-button-icon">👍</span>
                </button>
                <button 
                  className="action-button"
                  aria-label="Dislike message"
                >
                  <span className="action-button-icon">👎</span>
                </button>
              </div>
            )}
          </div>
        );
      })}
      {isLoading && (
        <div className="message assistant">
          <div className="message-content">
            <div className="message-text typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        </div>
      )}
      {showJumpToLatest && (
        <button 
          className="jump-to-latest"
          onClick={() => {
            setShowJumpToLatest(false);
            scrollToBottom();
          }}
          aria-label="Jump to latest message"
        >
          Jump to latest
        </button>
      )}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageList;