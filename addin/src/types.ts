export interface Message {
  id: number;
  text: string;
  sender: 'user' | 'assistant';
  timestamp: Date;
}

export interface InputBoxProps {
  onSendMessage: (message: string) => void;
  disabled: boolean;
}

export interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
}

export interface ChatResponse {
  text: string;
}