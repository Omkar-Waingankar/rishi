import React, { useState, useRef, useEffect } from 'react';
import MessageList from './MessageList';
import InputBox from './InputBox';
import { Message, ChatResponse } from './types';
import { 
  ToolCommand, 
  LegacyToolCommand, 
  ToolCallStatus, 
  ViewToolInput,
  StrReplaceToolInput,
  CreateToolInput,
  InsertToolInput,
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
        content: "👋 Hi, I'm Rishi — your AI assistant for RStudio. Ask me anything about R, code, data, or your project. How can I assist you today?"
      }]
    }
  ]);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [safeRoot, setSafeRoot] = useState<string | null>(null);
  const [safeRootError, setSafeRootError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

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
        
        // Remove any existing safe root error messages from chat
        setMessages(prev => prev.filter(msg => 
          !(msg.sender === 'assistant' && 
            msg.content.some(c => c.type === 'safe_root_error'))
        ));
      } else {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to get safe root';
        setSafeRootError(errorMessage);
        setSafeRoot(null);
        
        // Add error message to chat with refresh link
        const safeRootErrorMessage: Message = {
          id: Date.now() + 1000,
          sender: 'assistant',
          timestamp: new Date(),
          content: [{
            type: 'safe_root_error',
            content: `⚠️ Rishi requires an active directory to work.\n\nPlease set your active directory by:\n• Opening an RStudio project (.Rproj file), or\n• Using \`setwd("/path/to/your/project")\` in the R console\n\nThen click refresh here to continue.`,
            refreshAction: checkSafeRoot
          }]
        };
        setMessages(prev => {
          // Remove any existing safe root error messages first
          const filtered = prev.filter(msg => 
            !(msg.sender === 'assistant' && 
              msg.content.some(c => c.type === 'safe_root_error'))
          );
          return [...filtered, safeRootErrorMessage];
        });
      }
    } catch (error) {
      console.error('Error checking safe root:', error);
      const errorMessage = 'Failed to connect to tool server';
      setSafeRootError(errorMessage);
      setSafeRoot(null);
      
      // Add error message to chat with refresh link
      const safeRootErrorMessage: Message = {
        id: Date.now() + 1000,
        sender: 'assistant',
        timestamp: new Date(),
        content: [{
          type: 'safe_root_error',
          content: `⚠️ Rishi requires an active directory to work.\n\nPlease set your active directory by:\n• Opening an RStudio project (.Rproj file), or\n• Using \`setwd("/path/to/your/project")\` in the R console\n\nThen click refresh here to continue.`,
          refreshAction: checkSafeRoot
        }]
      };
      setMessages(prev => {
        // Remove any existing safe root error messages first
        const filtered = prev.filter(msg => 
          !(msg.sender === 'assistant' && 
            msg.content.some(c => c.type === 'safe_root_error'))
        );
        return [...filtered, safeRootErrorMessage];
      });
    }
  };

  // Check safe root on app startup
  useEffect(() => {
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

  return (
    <div className="chat-app">
      <div className="chat-header">
        <h2>Rishi</h2>
      </div>
      <MessageList messages={messages} isLoading={isStreaming} />
      <InputBox 
        onSendMessage={handleSendMessage} 
        disabled={isStreaming || !safeRoot}
        isStreaming={isStreaming}
        onStopStreaming={handleStopStreaming}
        safeRoot={safeRoot}
      />
    </div>
  );
};

export default ChatApp;