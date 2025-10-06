import { ToolCallStatus } from './tool_types';

// Define supported image MIME types
export type ImageMimeType = 'image/jpeg' | 'image/png' | 'image/webp' | 'image/gif';

// Text content
interface TextContent {
  type: 'text';
  content: string;
  invisible?: boolean;
}

// Image content
interface ImageContent {
  type: 'image';
  mediaType: ImageMimeType;
  dataBase64: string;
  invisible?: boolean;
}

// Tool call content
interface ToolCallContent {
  type: 'tool_call';
  content: string;
  invisible?: boolean;
  toolCall?: {
    name: string;
    status: ToolCallStatus;
    input?: object;
    result?: any;
  };
}

// Error content
interface ErrorContent {
  type: 'error';
  content: string;
  invisible?: boolean;
}

// Union of all content types
export type MessageContent = TextContent | ImageContent | ToolCallContent | ErrorContent;

export interface Message {
  id: number;
  sender: 'user' | 'assistant';
  timestamp: Date;
  content: MessageContent[];
}

export interface InputBoxProps {
  onSendMessage: (content: MessageContent[], model: string) => void;
  disabled: boolean;
  isStreaming: boolean;
  onStopStreaming: () => void;
  safeRoot: string | null;
  triggerStatusBarError?: React.MutableRefObject<(() => void) | null>;
  apiKeyStatus?: {
    anthropic: { has_key: boolean };
    openai: { has_key: boolean };
  } | null;
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
    result?: any;
  };
  is_final?: boolean;
  error?: string;
}

// Context state types
export interface ContextState {
  activeTab: boolean;
  plot: boolean;
}

// API response types
export interface ActiveTabResponse {
  filename: string;
  content: string;
  error?: string;
}

export interface PlotResponse {
  imageBase64: string;
  mediaType: string;
  error?: string;
}