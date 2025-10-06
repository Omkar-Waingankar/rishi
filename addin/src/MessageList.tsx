import React, { useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { MessageListProps, Message } from './types';
import { ContentCopy, ThumbUp, ThumbDown, Close, Edit, Add, Terminal } from '@mui/icons-material';
import { RHelpToolInput } from './tool_types';

const MessageList: React.FC<MessageListProps> = ({ messages, isLoading }) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messageListRef = useRef<HTMLDivElement>(null);
  const [expandedErrors, setExpandedErrors] = useState<Set<string>>(new Set());

  // Helper function to generate tooltip text for R help tool calls
  const getRHelpTooltipText = (input: RHelpToolInput): string => {
    if (input.topic) {
      return `${input.package}::${input.topic}`;
    }
    return input.package;
  };

  // Helper function to get the appropriate icon for each tool type
  const getToolIcon = (toolName: string) => {
    switch (toolName) {
      case 'str_replace':
        return <Edit sx={{ fontSize: 14 }} />;
      case 'create':
        return <Add sx={{ fontSize: 14 }} />;
      case 'insert':
        return <Edit sx={{ fontSize: 14 }} />;
      case 'console_exec':
        return <Terminal sx={{ fontSize: 14 }} />;
      default:
        return null;
    }
  };


  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };



  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  const formatTime = (timestamp: Date): string => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const isInRStudioViewer = () => {
    return window.location.search.includes('viewer_pane') || 
           window.location.href.includes('viewer_pane');
  };

  const copyToClipboard = (text: string) => {
    // If we're in RStudio viewer or clipboard API is unavailable, use fallback immediately
    if (isInRStudioViewer() || !navigator.clipboard || !navigator.clipboard.writeText) {
      fallbackCopyToClipboard(text);
      return;
    }

    // Try clipboard API first, fallback on failure
    navigator.clipboard.writeText(text).catch(() => {
      fallbackCopyToClipboard(text);
    });
  };

  const buildCopyText = (content: any[]) => {
    return content.map((item, index) => {
      if (item.type === 'text') {
        return item.content;
      } else if (item.type === 'tool_call') {
        return item.content;
      } else if (item.type === 'error') {
        return item.content;
      } else if (item.type === 'image') {
        return '[Image]';
      }
      return item.content || '';
    }).join('\n\n');
  };

  const fallbackCopyToClipboard = (text: string) => {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    textArea.style.opacity = '0';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      const successful = document.execCommand('copy');
      if (!successful) {
        console.warn('Fallback copy method may not have worked');
      }
    } catch (err) {
      console.error('Copy failed:', err);
    }
    document.body.removeChild(textArea);
  };

  const handleReviewChanges = () => {
    // This would integrate with RStudio's diff viewer
    console.log('Review changes clicked');
  };

  const toggleErrorExpansion = (errorId: string) => {
    setExpandedErrors(prev => {
      const newSet = new Set(prev);
      if (newSet.has(errorId)) {
        newSet.delete(errorId);
      } else {
        newSet.add(errorId);
      }
      return newSet;
    });
  };

  const renderCollapsibleError = (content: string, errorId: string) => {
    const isExpanded = expandedErrors.has(errorId);
    const MAX_LENGTH = 200; // Characters to show when collapsed
    
    if (content.length <= MAX_LENGTH) {
      return <div className="error-content">{content}</div>;
    }

    const truncatedContent = content.substring(0, MAX_LENGTH) + '...';
    
    return (
      <div className="error-content">
        <div className="collapsible-error">
          {isExpanded ? content : truncatedContent}
          <button 
            className="expand-toggle" 
            onClick={() => toggleErrorExpansion(errorId)}
            aria-label={isExpanded ? "Show less" : "Show more"}
          >
            {isExpanded ? ' Show less' : ' See more'}
          </button>
        </div>
      </div>
    );
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
                      renderCollapsibleError(item.content, `${message.id}-${index}`)
                    ) : item.type === 'image' ? (
                      // Don't render images in user messages
                      null
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
                        // Use legacy styling for view and r_help commands, new box styling for edit commands
                        (item.toolCall?.name === 'view' || item.toolCall?.name === 'r_help') ? (
                          item.toolCall?.name === 'r_help' && item.toolCall?.input ? (
                            <div key={index} className="tooltip-container">
                              <div className={`inline-tool-call ${item.toolCall?.status}`}>
                                {item.content}
                                {item.toolCall?.status === 'failed' && (
                                  <span className="tool-call-error-indicator" aria-label="Failed">×</span>
                                )}
                              </div>
                              <div className="tooltip">
                                {getRHelpTooltipText(item.toolCall.input as RHelpToolInput)}
                              </div>
                            </div>
                          ) : (
                            <div key={index} className={`inline-tool-call ${item.toolCall?.status}`}>
                              {item.content}
                              {item.toolCall?.status === 'failed' && (
                                <span className="tool-call-error-indicator" aria-label="Failed">×</span>
                              )}
                            </div>
                          )
                        ) : (
                          <div key={index} className={`tool-call-box ${item.toolCall?.status}`}>
                            <div className="tool-call-icon">
                              {getToolIcon(item.toolCall?.name || '')}
                            </div>
                            <span className="tool-call-text">{item.content}</span>
                            <div className="tool-call-icon">
                              {item.toolCall?.status === 'requesting' && (
                                <div className="spinner"></div>
                              )}
                              {item.toolCall?.status === 'completed' && (
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="check-icon">
                                  <polyline points="20,6 9,17 4,12"></polyline>
                                </svg>
                              )}
                              {item.toolCall?.status === 'failed' && (
                                <Close className="error-icon" sx={{ fontSize: 14 }} />
                              )}
                            </div>
                          </div>
                        )
                      )
                    ))}
                  </div>
              </div>
            </div>
            {showFooter && (
              <div className="message-actions" role="toolbar" aria-label="Message actions">
                <button 
                  className="action-button copy-button"
                  onClick={() => copyToClipboard(buildCopyText(message.content))}
                  aria-label="Copy message"
                >
                  <ContentCopy className="action-icon" sx={{ fontSize: 14 }} />
                  Copy
                </button>
                <button 
                  className="action-button thumb-button"
                  aria-label="Like message"
                >
                  <ThumbUp className="action-icon" sx={{ fontSize: 14 }} />
                </button>
                <button 
                  className="action-button thumb-button"
                  aria-label="Dislike message"
                >
                  <ThumbDown className="action-icon" sx={{ fontSize: 14 }} />
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