import React, { useState, useRef } from 'react';
import MessageList from './MessageList';
import InputBox from './InputBox';
import { Message, ChatResponse } from './types';

const getToolCallText = (toolCall: { name: string; status: string; input?: string }) => {
  let filename = '';
  try {
    const input = JSON.parse(toolCall.input || '{}');
    filename = input.path || input.Path || '';
  } catch {
    // If parsing fails, just use the raw input
    filename = toolCall.input || '';
  }

  switch (toolCall.name) {
    case 'read_file':
      return toolCall.status === 'requesting' 
        ? `Claude asking to read ${filename}`
        : `Claude has read ${filename}`;
    case 'list_files':
      return toolCall.status === 'requesting'
        ? `Claude asking to list files ${filename ? `in ${filename}` : ''}`
        : `Claude has listed files ${filename ? `in ${filename}` : ''}`;
    default:
      return toolCall.status === 'requesting'
        ? `Claude asking to use ${toolCall.name}`
        : `Claude has used ${toolCall.name}`;
  }
};

const ChatApp: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([
    { 
      id: 1, 
      text: "👋 Hi, I'm Tibbl — your AI assistant for RStudio. Ask me anything about R, code, data, or your project. How can I assist you today?", 
      sender: 'assistant', 
      timestamp: new Date() 
    }
  ]);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const handleSendMessage = async (messageText: string): Promise<void> => {
    if (!messageText.trim()) return;

    const userMessage: Message = {
      id: Date.now(),
      text: messageText,
      sender: 'user',
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setIsStreaming(true);

    // Create new AbortController for this request
    abortControllerRef.current = new AbortController();

    try {
      // Convert messages to history format (exclude the initial greeting message)
      const conversationHistory = messages
        .slice(1) // Skip the initial greeting message
        .map(msg => ({
          role: msg.sender === 'user' ? 'user' : 'assistant',
          content: msg.text
        }));

      const response = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          message: messageText,
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
      let assistantResponse = '';
      let assistantContent: Array<{type: 'text' | 'tool_call', content: string, toolCall?: any}> = [];

      const assistantMessage: Message = {
        id: Date.now() + 1,
        text: '',
        sender: 'assistant',
        timestamp: new Date()
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
            if (data.text) {
              assistantResponse += data.text;
              assistantContent.push({type: 'text', content: data.text});
              
              // Render mixed content as combined text with tool calls inline
              const combinedText = assistantContent.map(item => {
                if (item.type === 'text') {
                  return item.content;
                } else {
                  return `\n\n<div class="inline-tool-call ${item.toolCall?.status}">${getToolCallText(item.toolCall)}</div>\n\n`;
                }
              }).join('');
              
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id 
                  ? { ...msg, text: combinedText }
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
                // Update the last tool call in content
                for (let i = assistantContent.length - 1; i >= 0; i--) {
                  if (assistantContent[i].type === 'tool_call' && 
                      assistantContent[i].toolCall?.name === data.tool_call.name &&
                      assistantContent[i].toolCall?.status === 'requesting') {
                    assistantContent[i] = {
                      type: 'tool_call',
                      content: getToolCallText(data.tool_call),
                      toolCall: data.tool_call
                    };
                    break;
                  }
                }
              }
              
              // Re-render combined content
              const combinedText = assistantContent.map(item => {
                if (item.type === 'text') {
                  return item.content;
                } else {
                  return `\n\n<div class="inline-tool-call ${item.toolCall?.status}">${getToolCallText(item.toolCall)}</div>\n\n`;
                }
              }).join('');
              
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id 
                  ? { ...msg, text: combinedText }
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
          text: "Could not connect to our server. Please wait or restart Tibbl and try again.",
          sender: 'assistant',
          timestamp: new Date(),
          type: 'error'
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
      <InputBox 
        onSendMessage={handleSendMessage} 
        disabled={isStreaming}
        onStopStreaming={handleStopStreaming}
      />
    </div>
  );
};

export default ChatApp;