interface MessageContent {
  type: 'text' | 'tool_call' | 'error';
  content: string;
  toolCall?: {
    name: string;
    status: ToolCallStatus;
    input?: string;
    result?: string;
  };
}

export type ToolCallStatus = 'requesting' | 'completed' | 'failed';

export interface Message {
  id: number;
  sender: 'user' | 'assistant';
  timestamp: Date;
  content: MessageContent[];
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
    status: ToolCallStatus;
    result?: string;
  };
  is_final?: boolean;
  error?: string;
}