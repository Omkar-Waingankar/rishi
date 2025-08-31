import { ToolCallStatus } from './tool_types';

interface MessageContent {
  type: 'text' | 'tool_call' | 'error';
  content: string;
  toolCall?: {
    name: string;
    status: ToolCallStatus;
    input?: object;
    result?: string;
  };
}

export interface Message {
  id: number;
  sender: 'user' | 'assistant';
  timestamp: Date;
  content: MessageContent[];
}

export interface InputBoxProps {
  onSendMessage: (message: string) => void;
  disabled: boolean;
  isStreaming: boolean;
  onStopStreaming: () => void;
  safeRoot: string | null;
}

export interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
}

export interface ChatResponse {
  text?: string;
  tool_call?: {
    name: string;
    input: object;
    status: ToolCallStatus;
    result?: string;
  };
  is_final?: boolean;
  error?: string;
}