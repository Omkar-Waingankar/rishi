// Tool types and interfaces for the text editor tool system

export enum ToolCommand {
  VIEW = 'view',
  STR_REPLACE = 'str_replace',
  CREATE = 'create',
  INSERT = 'insert',
}

export type ToolCallStatus = 'requesting' | 'completed' | 'failed';

// Input types for different tool commands
export interface ViewToolInput {
  path: string;
  view_range?: [number, number];
}

export interface StrReplaceToolInput {
  path: string;
  old_str: string;
  new_str: string;
}

export interface CreateToolInput {
  path: string;
  file_text: string;
}

export interface InsertToolInput {
  path: string;
  insert_line: number;
  new_str: string;
}

// Output types for different tool commands
export interface ViewToolOutput {
  content?: string;
  error?: string;
}

// Generic tool call structure
export interface ToolCall {
  name: ToolCommand;
  status: ToolCallStatus;
  input?: ViewToolInput | StrReplaceToolInput | CreateToolInput | InsertToolInput;
  result?: string;
}