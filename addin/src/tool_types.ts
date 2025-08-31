// Tool types and interfaces for the text editor tool system

export enum ToolCommand {
  VIEW = 'view',
  // Future commands can be added here
  // EDIT = 'edit',
  // CREATE = 'create',
}

export type ToolCallStatus = 'requesting' | 'completed' | 'failed';

// Input types for different tool commands
export interface ViewToolInput {
  path: string;
  view_range?: [number, number];
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
  input?: ViewToolInput; // Can be extended with union types for other commands
  result?: string;
}

// Legacy tool command support (for backward compatibility)
export enum LegacyToolCommand {
  READ_FILE = 'read_file',
  LIST_FILES = 'list_files',
}

export interface LegacyReadFileInput {
  path: string;
  Path?: string; // Alternative capitalization
}

export interface LegacyListFilesInput {
  path?: string;
  Path?: string; // Alternative capitalization
}