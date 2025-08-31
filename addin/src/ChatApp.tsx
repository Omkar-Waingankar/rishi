import React, { useState, useRef, useEffect } from 'react';
import MessageList from './MessageList';
import InputBox from './InputBox';
import { Message, ChatResponse } from './types';
import { 
  ToolCommand, 
  LegacyToolCommand, 
  ToolCallStatus, 
  ViewToolInput, 
  LegacyReadFileInput, 
  LegacyListFilesInput 
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
    
    // Legacy support for old tool names
    case LegacyToolCommand.READ_FILE: {
      const readInput = input as LegacyReadFileInput;
      const filename = readInput.path || readInput.Path || 'file';
      
      if (toolCall.status === 'requesting') {
        return `Reading ${filename}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to read ${filename}`;
      } else {
        return `Read ${filename}`;
      }
    }
    case LegacyToolCommand.LIST_FILES: {
      const listInput = input as LegacyListFilesInput;
      const dirPath = listInput.path || listInput.Path || 'current directory';
      
      if (toolCall.status === 'requesting') {
        return `Listing files in ${dirPath}`;
      } else if (toolCall.status === 'failed') {
        return `Failed to list files in ${dirPath}`;
      } else {
        return `Listed files in ${dirPath}`;
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
  const [messages, setMessages] = useState<Message[]>([
    { 
      id: 1, 
      sender: 'assistant', 
      timestamp: new Date(),
      content: [{
        type: 'text',
        content: "👋 Hi, I'm Tibbl — your AI assistant for RStudio. Ask me anything about R, code, data, or your project. How can I assist you today?"
      }]
    }
  ]);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [safeRoot, setSafeRoot] = useState<string | null>(null);
  const [safeRootError, setSafeRootError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Check safe root on app startup
  useEffect(() => {
    const checkSafeRoot = async () => {
      try {
        const response = await fetch('http://localhost:8082/safe_root', {
          method: 'GET',
          headers: {
            'Authorization': 'Bearer tibble-dev-local-please-change',
          },
        });

        if (response.ok) {
          const data = await response.json();
          setSafeRoot(data.safe_root);
          setSafeRootError(null);
        } else {
          const errorData = await response.json();
          setSafeRootError(errorData.error || 'Failed to get safe root');
          setSafeRoot(null);
        }
      } catch (error) {
        console.error('Error checking safe root:', error);
        setSafeRootError('Failed to connect to tool server');
        setSafeRoot(null);
      }
    };

    checkSafeRoot();
  }, []);

  const handleSendMessage = async (messageText: string): Promise<void> => {
    if (!messageText.trim()) return;

    const userMessage: Message = {
      id: Date.now(),
      sender: 'user',
      timestamp: new Date(),
      content: [{
        type: 'text',
        content: messageText
      }]
    };

    setMessages(prev => [...prev, userMessage]);
    setIsStreaming(true);

    // Create new AbortController for this request
    abortControllerRef.current = new AbortController();

    try {
      // Convert messages to history format (exclude the initial greeting message)
      const conversationHistory: Array<{role: string, content: string}> = [];
      
      messages.slice(1).forEach(msg => { // Skip the initial greeting message
        if (msg.sender === 'user') {
          conversationHistory.push({
            role: 'user',
            content: msg.content.map(c => c.content).join('')
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
                  content: currentTextContent
                });
                currentTextContent = ''; // Reset
              }
              
              // Add tool use request as assistant message (even if still requesting)
              if (contentItem.toolCall) {
                const inputStr = JSON.stringify(contentItem.toolCall.input || {});
                conversationHistory.push({
                  role: 'assistant', 
                  content: `[Using tool: ${contentItem.toolCall.name} with input: ${inputStr}]`
                });
              }
              
              // Add tool result as user message (for both completed and failed)
              if (
                (contentItem.toolCall?.status === 'completed' || contentItem.toolCall?.status === 'failed') 
                && contentItem.toolCall?.result
              ) {
                conversationHistory.push({
                  role: 'user',
                  content: `[Result for tool ${contentItem.toolCall.name}: ${contentItem.toolCall.result}]`
                });
              }
            }
          });
          // Add any remaining text content as final assistant message
          if (currentTextContent.trim()) {
            conversationHistory.push({
              role: 'assistant',
              content: currentTextContent
            });
          }
        }
      });

      const response = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          message: messageText,
          history: conversationHistory,
          safe_root: safeRoot
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
              // Handle error from backend
              assistantContent.push({type: 'error', content: data.error});
              
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id && 'content' in msg
                  ? { ...msg, content: [...assistantContent] }
                  : msg
              ));
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
            content: "Could not connect to our server. Please wait or restart Tibbl and try again."
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

  return (
    <div className="chat-app">
      <div className="chat-header">
        <h2>Tibbl</h2>
      </div>
      <MessageList messages={messages} isLoading={isStreaming} />
      {safeRootError && (
        <div className="safe-root-dialog">
          <div className="safe-root-message">
            <h3>⚠️ Active Directory Required</h3>
            <p>{safeRootError}</p>
            <p>Please set your active directory by:</p>
            <ul>
              <li>Opening an RStudio project (.Rproj file), or</li>
              <li>Using <code>setwd("/path/to/your/project")</code> in the R console</li>
            </ul>
          </div>
        </div>
      )}
      <InputBox 
        onSendMessage={handleSendMessage} 
        disabled={isStreaming || !safeRoot}
        onStopStreaming={handleStopStreaming}
        safeRoot={safeRoot}
      />
    </div>
  );
};

export default ChatApp;