import React, { useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { MessageListProps, Message } from './types';

const MessageList: React.FC<MessageListProps> = ({ messages, isLoading }) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messageListRef = useRef<HTMLDivElement>(null);


  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };



  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

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
      role="log"
      aria-label="Chat conversation"
      aria-live="polite"
    >
      {messages.map((message, index) => {
        // Only show footer for assistant messages when not currently loading
        // or if it's not the most recent assistant message
        const isLastAssistantMessage = message.sender === 'assistant' && 
          index === messages.length - 1;
        const hasErrorContent = message.content.some(item => item.type === 'error');
        const showFooter = message.sender === 'assistant' && 
          !hasErrorContent &&
          (!isLoading || !isLastAssistantMessage);
        
        return (
          <div key={message.id} className={`message ${message.sender}${hasErrorContent ? ' error' : ''}`} role="article">
            <div className="message-content">
              <div className="message-text" aria-label={`${message.sender} message`}>
                <div>
                  {message.content.map((item, index) => (
                    item.type === 'error' ? (
                      <div key={index} className="error-content">{item.content}</div>
                    ) : item.type === 'text' ? (
                        <ReactMarkdown
                          key={index}
                          remarkPlugins={[remarkGfm]}
                          components={{
                            code: ({ node, inline, className, children, ...props }: any) => {
                              if (inline) {
                                return (
                                  <code className="inline-code" {...props}>
                                    {children}
                                  </code>
                                );
                              }
                              
                              // Extract language from className (e.g., "language-javascript" -> "javascript")
                              let language = '';
                              if (className && className.includes('language-')) {
                                language = className.replace('language-', '');
                              }
                              
                              // Fallback: try to detect language from content patterns
                              const codeContent = String(children || '').toLowerCase();
                              if (!language) {
                                if (codeContent.includes('function(') || codeContent.includes('<-') || codeContent.includes('print(')) {
                                  language = 'r';
                                } else if (codeContent.includes('def ') || codeContent.includes('import ')) {
                                  language = 'python';
                                } else if (codeContent.includes('function ') || codeContent.includes('const ') || codeContent.includes('console.log')) {
                                  language = 'javascript';
                                }
                              }
                              
                              const displayLanguage = language || 'text';
                              
                              return (
                                <code className="code-block-inner" data-language={displayLanguage} {...props}>
                                  {children}
                                </code>
                              );
                            },
                            pre: ({ children }: any) => {
                              // Extract language from the code element inside
                              const codeElement = React.Children.toArray(children)[0] as any;
                              
                              // Try multiple ways to get the language
                              let language = codeElement?.props?.['data-language'] || '';
                              
                              // If data-language isn't set, try to extract from className
                              if (!language && codeElement?.props?.className) {
                                const className = codeElement.props.className;
                                if (className.includes('language-')) {
                                  language = className.replace('language-', '');
                                }
                              }
                              
                              // Fallback to content-based detection
                              if (!language) {
                                const codeContent = String(codeElement?.props?.children || '').toLowerCase();
                                if (codeContent.includes('function(') || codeContent.includes('<-') || codeContent.includes('print(')) {
                                  language = 'r';
                                } else if (codeContent.includes('def ') || codeContent.includes('import ')) {
                                  language = 'python';
                                } else if (codeContent.includes('function ') || codeContent.includes('const ') || codeContent.includes('console.log')) {
                                  language = 'javascript';
                                }
                              }
                              
                              const displayLanguage = language || 'plaintext';
                              
                              return (
                                <div className="code-block-container">
                                  <div className="code-block-header">
                                    <span className="code-language">{displayLanguage}</span>
                                  </div>
                                  <pre className="code-block">
                                    {children}
                                  </pre>
                                </div>
                              );
                            },
                          }}
                        >
                          {item.content}
                        </ReactMarkdown>
                      ) : (
                        <div key={index} className={`inline-tool-call ${item.toolCall?.status}`}>
                          {item.content}
                          {item.toolCall?.status === 'failed' && (
                            <span className="tool-call-error-indicator" aria-label="Failed">×</span>
                          )}
                        </div>
                      )
                    ))}
                  </div>
              </div>
            </div>
            {showFooter && (
              <div className="message-actions" role="toolbar" aria-label="Message actions">
                <button 
                  className="action-button copy-button"
                  onClick={() => copyToClipboard(message.content.map(c => c.content).join(''))}
                  aria-label="Copy message"
                >
                  <svg className="action-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <rect width="14" height="14" x="8" y="8" rx="2" ry="2"/>
                    <path d="m4 16c-1.1 0-2-.9-2-2v-10c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>
                  </svg>
                  Copy
                </button>
                <button 
                  className="action-button thumb-button"
                  aria-label="Like message"
                >
                  <svg className="action-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M7 10v12"/>
                    <path d="M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2h0a3.13 3.13 0 0 1 3 3.88Z"/>
                  </svg>
                </button>
                <button 
                  className="action-button thumb-button"
                  aria-label="Dislike message"
                >
                  <svg className="action-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M17 14V2"/>
                    <path d="M9 18.12 10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79 1.11L12 22h0a3.13 3.13 0 0 1-3-3.88Z"/>
                  </svg>
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

      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageList;