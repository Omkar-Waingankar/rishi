import React from 'react';
import { createRoot } from 'react-dom/client';
import ChatApp from './ChatApp';
import './styles.css';

const container = document.getElementById('chat-root');
const root = createRoot(container);
root.render(<ChatApp />);