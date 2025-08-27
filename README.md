# Tibbl

## Getting Started

1. Clone this repository.
2. Set up your Anthropic API key:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```
3. Start the backend daemon:
   ```bash
   make run-server
   ```
   The daemon will run on port 8080.
4. Build and install the RStudio add-in:
   ```bash
   make install-addin
   ```
5. Open RStudio, then launch the Tibbl add-in from the Addins menu.
   The frontend will run on port 8081.

---

## Repository Structure

- **addin/**  
  Contains the RStudio add-in source code, including the React frontend and R integration.

- **daemon/**  
  Contains an HTTP backend server that provides AI chat functionality via Anthropic's Claude API. The backend runs on port 8080 and is integrated with the add-in.

---
