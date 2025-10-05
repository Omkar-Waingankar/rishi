import React, { useState, useRef, useEffect } from 'react';
import MessageList from './MessageList';
import InputBox from './InputBox';
import StatusBar from './StatusBar';
import ApiKeySetup from './ApiKeySetup';
import Settings from './Settings';
import { Message, ChatResponse, MessageContent } from './types';
import { Settings as SettingsIcon } from '@mui/icons-material';
import {
  ToolCommand,
  ToolCallStatus,
  ViewToolInput,
  StrReplaceToolInput,
  CreateToolInput,
  InsertToolInput,
  ConsoleExecToolInput
} from './tool_types';

const getToolCallText = (toolCall: { name: string; status: string; input?: object }) => {
  // Assume input is always an object - parse it based on the tool command
  const input = toolCall.input || {};

  switch (toolCall.name) {
    case ToolCommand.VIEW: {
      const viewInput = input as ViewToolInput;
      const displayPath = viewInput.path || 'current directory';

      if (toolCall.status === 'requesting') {
        return `Viewing ${displayPath}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to view ${displayPath}`;
      } else {
        return `Viewed ${displayPath}`;
      }
    }
    
    case ToolCommand.STR_REPLACE: {
      const replaceInput = input as StrReplaceToolInput;
      const displayPath = replaceInput.path || 'file';

      if (toolCall.status === 'requesting') {
        return `Editing ${displayPath}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to edit ${displayPath}`;
      } else {
        return `Edited ${displayPath}`;
      }
    }
    
    case ToolCommand.CREATE: {
      const createInput = input as CreateToolInput;
      const displayPath = createInput.path || 'file';

      if (toolCall.status === 'requesting') {
        return `Creating ${displayPath}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to create ${displayPath}`;
      } else {
        return `Created ${displayPath}`;
      }
    }
    
    case ToolCommand.INSERT: {
      const insertInput = input as InsertToolInput;
      const displayPath = insertInput.path || 'file';

      if (toolCall.status === 'requesting') {
        return `Inserting into ${displayPath}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to insert into ${displayPath}`;
      } else {
        return `Inserted into ${displayPath}`;
      }
    }

    case ToolCommand.CONSOLE_EXEC: {
      if (toolCall.status === 'requesting') {
        return `Writing to console`;
      } else if (toolCall.status === 'failed') {
        return `Failed to write to console`;
      } else {
        return `Wrote to console`;
      }
    }

    default:
      // Handle unknown tool commands gracefully
      if (toolCall.status === 'requesting') {
        return `Using ${toolCall.name}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to use ${toolCall.name}`;
      } else {
        return `Used ${toolCall.name}`;
      }
  }
};

const ChatApp: React.FC = () => {
  // Message state
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 1,
      sender: 'assistant',
      timestamp: new Date(),
      content: [{
        type: 'text',
        content: "👋 Hi, I'm Rishi — your AI assistant for RStudio. Ask me anything about R, code, data, or your project. How can I assist you today?"
      }]
    }
  ]);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const abortControllerRef = useRef<AbortController | null>(null);
  const triggerStatusBarErrorRef = useRef<(() => void) | null>(null);

  // Safe root state
  const [safeRoot, setSafeRoot] = useState<string | null>(null);
  const [safeRootError, setSafeRootError] = useState<string | null>(null);

  // Connection status state
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'failed'>('connecting');

  // API key state
  const [showApiKeySetup, setShowApiKeySetup] = useState<boolean>(false);
  const [selectedProvider, setSelectedProvider] = useState<string>('anthropic');
  const [apiKeyStatus, setApiKeyStatus] = useState<{
    anthropic: { has_key: boolean; api_key: string };
    openai: { has_key: boolean; api_key: string };
  } | null>(null);

  // Settings state
  const [showSettings, setShowSettings] = useState<boolean>(false);

  const checkSafeRoot = async () => {
    try {
      const response = await fetch('http://localhost:8082/safe_root', {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer rishi-dev-local-please-change',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSafeRoot(data.safe_root);
        setSafeRootError(null);
      } else {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to get safe root';
        setSafeRootError(errorMessage);
        setSafeRoot(null);
      }
    } catch (error) {
      console.error('Error checking safe root:', error);
      const errorMessage = 'Failed to connect to tool server';
      setSafeRootError(errorMessage);
      setSafeRoot(null);
    }
  };

  // Check API key status for both providers on app startup
  useEffect(() => {
    const checkApiKeys = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/keys', {
          method: 'GET',
        });

        if (response.ok) {
          const data = await response.json();
          setApiKeyStatus(data);
          
          // Check if we have any API keys
          const hasAnyKey = data.anthropic.has_key || data.openai.has_key;
          if (!hasAnyKey) {
            setShowApiKeySetup(true);
          }
        } else {
          setShowApiKeySetup(true);
        }
      } catch (error) {
        // Backend server not ready yet, will retry
        console.error('Error checking API keys:', error);
      }
    };

    checkApiKeys();
  }, []);

  // Check safe root on app startup
  useEffect(() => {
    checkSafeRoot();
  }, []);

  // Poll for working directory changes every 2.5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      checkSafeRoot();
    }, 2500);

    return () => clearInterval(interval);
  }, []);

  // Check daemon connection status
  useEffect(() => {
    const checkConnection = async () => {
      try {
        const response = await fetch('http://localhost:8080/health', {
          method: 'GET',
        });

        if (response.ok) {
          setConnectionStatus('connected');
        } else {
          setConnectionStatus('failed');
        }
      } catch (error) {
        setConnectionStatus('failed');
      }
    };

    // Initial check
    checkConnection();

    // Check every 2 seconds continuously
    const interval = setInterval(() => {
      checkConnection();
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  const handleSendMessage = async (content: MessageContent[], selectedModel: string): Promise<void> => {
    if (content.length === 0) return;

    // Filter out invisible content for display
    const displayContent = content.filter(item => !item.invisible);

    const userMessage: Message = {
      id: Date.now(),
      sender: 'user',
      timestamp: new Date(),
      content: displayContent
    };

    setMessages(prev => [...prev, userMessage]);
    setIsStreaming(true);

    // Create new AbortController for this request
    abortControllerRef.current = new AbortController();

    try {
      // Convert messages to history format (exclude the initial greeting message)
      // Backend expects content arrays for all messages
      const conversationHistory: Array<{role: string, content: MessageContent[]}> = [];
      
      messages.slice(1).forEach(msg => { // Skip the initial greeting message
        if (msg.sender === 'user') {
          // For user messages, we need to send the actual content structure to backend
          conversationHistory.push({
            role: 'user',
            content: msg.content
          });
        } else if (msg.sender === 'assistant') {
          // Process content chronologically, maintaining interleaved structure
          let currentTextContent = '';
          
          msg.content.forEach(contentItem => {
            if (contentItem.type === 'text') {
              // Accumulate text content
              currentTextContent += contentItem.content;
            } else if (contentItem.type === 'tool_call') {
              // If we have accumulated text, add it as an assistant message first
              if (currentTextContent.trim()) {
                conversationHistory.push({
                  role: 'assistant',
                  content: [{ type: 'text', content: currentTextContent }]
                });
                currentTextContent = ''; // Reset
              }
              
              // Add tool use request as assistant message (even if still requesting)
              if (contentItem.toolCall) {
                const inputStr = JSON.stringify(contentItem.toolCall.input || {});
                conversationHistory.push({
                  role: 'assistant', 
                  content: [{ type: 'text', content: `[Using tool: ${contentItem.toolCall.name} with input: ${inputStr}]` }]
                });
              }
              
              // Add tool result as user message (for both completed and failed)
              if (
                (contentItem.toolCall?.status === 'completed' || contentItem.toolCall?.status === 'failed') 
                && contentItem.toolCall?.result
              ) {
                conversationHistory.push({
                  role: 'user',
                  content: [{ type: 'text', content: `[Result for tool ${contentItem.toolCall.name}: ${contentItem.toolCall.result}]` }]
                });
              }
            }
          });
          // Add any remaining text content as final assistant message
          if (currentTextContent.trim()) {
            conversationHistory.push({
              role: 'assistant',
              content: [{ type: 'text', content: currentTextContent }]
            });
          }
        }
      });

      // Determine endpoint and headers based on model type
      let endpoint: string;
      let apiKey: string;
      
      if (selectedModel.substring(0, 6) === 'claude') {
        endpoint = 'http://localhost:8080/chat/anthropic';
        apiKey = apiKeyStatus?.anthropic.api_key || '';
      } else {
        endpoint = 'http://localhost:8080/chat/openai';
        apiKey = apiKeyStatus?.openai.api_key || '';
      }

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        'X-Provider-API-Key': apiKey,
        'X-Model': selectedModel,
      };

      const response = await fetch(endpoint, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          message: content,
          history: conversationHistory
        }),
        signal: abortControllerRef.current.signal
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Handle streaming response
      if (!response.body) {
        throw new Error('Response body is null');
      }
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let assistantContent: Array<{type: 'text' | 'tool_call' | 'error', content: string, toolCall?: any}> = [];

      const assistantMessage: Message = {
        id: Date.now() + 1,
        sender: 'assistant',
        timestamp: new Date(),
        content: []
      };
      
      setMessages(prev => [...prev, assistantMessage]);

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());
        
        for (const line of lines) {
          try {
            const data: ChatResponse = JSON.parse(line);
            if (data.error) {
              // Handle error from backend - create a separate error message
              const errorMessage: Message = {
                id: Date.now() + 2,
                sender: 'assistant',
                timestamp: new Date(),
                content: [{
                  type: 'error',
                  content: data.error
                }]
              };
              
              setMessages(prev => [...prev, errorMessage]);
            } else if (data.text) {
              // Accumulate text chunks - find the last text item or create new one
              const lastItem = assistantContent[assistantContent.length - 1];
              if (lastItem && lastItem.type === 'text') {
                // Append to existing text content
                lastItem.content += data.text;
              } else {
                // Create new text content item
                assistantContent.push({type: 'text', content: data.text});
              }
              
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id && 'content' in msg
                  ? { ...msg, content: [...assistantContent] }
                  : msg
              ));
            } else if (data.tool_call) {
              if (data.tool_call.status === 'requesting') {
                assistantContent.push({
                  type: 'tool_call', 
                  content: getToolCallText(data.tool_call),
                  toolCall: data.tool_call
                });
              } else if (data.tool_call.status === 'completed') {
                // Check if the tool call actually failed by parsing the result
                let actualStatus = 'completed';
                if (data.tool_call.result) {
                  try {
                    const resultObj = JSON.parse(data.tool_call.result);
                    if (resultObj.error) {
                      actualStatus = 'failed';
                    }
                  } catch {
                    // If result isn't JSON, assume success
                  }
                }

                // Update the last tool call in content
                for (let i = assistantContent.length - 1; i >= 0; i--) {
                  if (assistantContent[i].type === 'tool_call' && 
                      assistantContent[i].toolCall?.name === data.tool_call.name &&
                      assistantContent[i].toolCall?.status === 'requesting') {
                    const updatedToolCall = {
                      ...data.tool_call,
                      status: actualStatus as ToolCallStatus
                    };
                    assistantContent[i] = {
                      type: 'tool_call',
                      content: getToolCallText(updatedToolCall),
                      toolCall: updatedToolCall
                    };
                    break;
                  }
                }
              }
              
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id && 'content' in msg
                  ? { ...msg, content: [...assistantContent] }
                  : msg
              ));
            }
          } catch (e) {
            console.error('Error parsing JSON:', e);
          }
        }
      }
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        // Request was cancelled by user
        console.log('Request cancelled by user');
      } else {
        console.error('Error sending message:', error);
        const errorMessage: Message = {
          id: Date.now() + 1,
          sender: 'assistant',
          timestamp: new Date(),
          content: [{
            type: 'error',
            content: "Could not connect to our server. Please wait or restart Rishi and try again."
          }]
        };
        setMessages(prev => [...prev, errorMessage]);
      }
    } finally {
      setIsStreaming(false);
      abortControllerRef.current = null;
    }
  };

  const handleStopStreaming = (): void => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  };

  const handleApiKeySubmit = async (submittedApiKey: string): Promise<void> => {
    try {
      const response = await fetch(`http://localhost:8080/api/key/${selectedProvider}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ api_key: submittedApiKey }),
      });

      if (!response.ok) {
        throw new Error('Failed to save API key');
      }

      setShowApiKeySetup(false);
      
      // Update API key status
      if (apiKeyStatus) {
        const updatedStatus = { ...apiKeyStatus };
        updatedStatus[selectedProvider as keyof typeof updatedStatus] = {
          has_key: true,
          api_key: submittedApiKey
        };
        setApiKeyStatus(updatedStatus);
      }
    } catch (error) {
      throw new Error('Failed to save API key. Please try again.');
    }
  };

  const handleApiKeyUpdate = async (provider: string, apiKey: string): Promise<void> => {
    try {
      // First validate the API key
      const validateResponse = await fetch(`http://localhost:8080/api/key/${provider}/validate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ api_key: apiKey }),
      });

      if (!validateResponse.ok) {
        throw new Error('Invalid API key');
      }

      const validateData = await validateResponse.json();
      if (!validateData.valid) {
        throw new Error('Invalid API key');
      }

      // If validation passes, save the API key
      const response = await fetch(`http://localhost:8080/api/key/${provider}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ api_key: apiKey }),
      });

      if (!response.ok) {
        throw new Error('Failed to save API key');
      }

      // Update API key status
      if (apiKeyStatus) {
        const updatedStatus = { ...apiKeyStatus };
        updatedStatus[provider as keyof typeof updatedStatus] = {
          has_key: true,
          api_key: apiKey
        };
        setApiKeyStatus(updatedStatus);
      }
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to save API key. Please try again.');
    }
  };

  if (showApiKeySetup) {
    return (
      <div className="chat-app">
        <ApiKeySetup 
          onApiKeySubmit={handleApiKeySubmit} 
          selectedProvider={selectedProvider}
          onProviderChange={setSelectedProvider}
        />
      </div>
    );
  }

  if (showSettings) {
    return (
      <Settings
        onClose={() => setShowSettings(false)}
        apiKeyStatus={apiKeyStatus}
        onApiKeyUpdate={handleApiKeyUpdate}
      />
    );
  }

  return (
    <div className="chat-app">
      <div className="chat-header">
        <h2>Rishi</h2>
        <button 
          className="settings-button"
          onClick={() => setShowSettings(true)}
          title="Settings"
        >
          <SettingsIcon sx={{ fontSize: 16 }} />
        </button>
      </div>
      <StatusBar connectionStatus={connectionStatus} workingDirectory={safeRoot} triggerErrorRef={triggerStatusBarErrorRef} />
      <MessageList messages={messages} isLoading={isStreaming} />
        <InputBox
          onSendMessage={handleSendMessage}
          disabled={isStreaming}
          isStreaming={isStreaming}
          onStopStreaming={handleStopStreaming}
          safeRoot={safeRoot}
          triggerStatusBarError={triggerStatusBarErrorRef}
          apiKeyStatus={apiKeyStatus ? {
            anthropic: { has_key: apiKeyStatus.anthropic.has_key },
            openai: { has_key: apiKeyStatus.openai.has_key }
          } : null}
        />
    </div>
  );
};

export default ChatApp;