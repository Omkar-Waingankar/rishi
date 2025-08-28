import React, { useState } from 'react';
import MessageList from './MessageList';
import InputBox from './InputBox';
import { Message, ChatResponse } from './types';

const ChatApp: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([
    { 
      id: 1, 
      text: "👋 Hi, I'm Tibbl — your AI assistant for RStudio. Ask me anything about R, code, data, or your project. How can I assist you today?", 
      sender: 'assistant', 
      timestamp: new Date() 
    }
  ]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const handleSendMessage = async (messageText: string): Promise<void> => {
    if (!messageText.trim()) return;

    const userMessage: Message = {
      id: Date.now(),
      text: messageText,
      sender: 'user',
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setIsLoading(true);

    try {
      const response = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: messageText })
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
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessage.id 
                  ? { ...msg, text: assistantResponse }
                  : msg
              ));
            }
          } catch (e) {
            console.error('Error parsing JSON:', e);
          }
        }
      }
    } catch (error) {
      console.error('Error sending message:', error);
      const errorMessage: Message = {
        id: Date.now() + 1,
        text: "Could not connect to our server. Please wait or restart Tibbl and try again.",
        sender: 'assistant',
        timestamp: new Date(),
        type: 'error'
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="chat-app">
      <div className="chat-header">
        <h2>Tibbl</h2>
      </div>
      <MessageList messages={messages} isLoading={isLoading} />
      <InputBox onSendMessage={handleSendMessage} disabled={isLoading} />
    </div>
  );
};

export default ChatApp;