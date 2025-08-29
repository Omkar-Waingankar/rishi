export interface Message {
  id: number;
  text: string;
  sender: 'user' | 'assistant';
  timestamp: Date;
  type?: 'normal' | 'error' | 'warning' | 'tool_call';
  toolCall?: {
    name: string;
    status: 'requesting' | 'completed';
    input?: string;
  };
}

export interface InputBoxProps {
  onSendMessage: (message: string) => void;
  disabled: boolean;
  onStopStreaming: () => void;
}

export interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
}

export interface ChatResponse {
  text?: string;
  tool_call?: {
    name: string;
    input: string;
    status: 'requesting' | 'completed';
  };
  is_final?: boolean;
}