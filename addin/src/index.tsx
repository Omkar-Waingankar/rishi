import React from 'react';
import { createRoot } from 'react-dom/client';
import ChatApp from './ChatApp';
import './styles.css';

const container = document.getElementById('chat-root');
if (!container) {
  throw new Error('Failed to find the root element');
}

const root = createRoot(container);
root.render(<ChatApp />);