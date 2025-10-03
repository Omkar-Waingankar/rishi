import { ToolCallStatus } from './tool_types';

// Define supported image MIME types
export type ImageMimeType = 'image/jpeg' | 'image/png' | 'image/webp' | 'image/gif';

// Text content
interface TextContent {
  type: 'text';
  content: string;
}

// Image content
interface ImageContent {
  type: 'image';
  mediaType: ImageMimeType;
  dataBase64: string;
}

// Tool call content
interface ToolCallContent {
  type: 'tool_call';
  content: string;
  toolCall?: {
    name: string;
    status: ToolCallStatus;
    input?: object;
    result?: string;
  };
}

// Error content
interface ErrorContent {
  type: 'error';
  content: string;
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