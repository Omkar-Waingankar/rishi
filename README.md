# Rishi

## Getting Started

1. Clone this repository.
2. Set up your Anthropic API key:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```
3. Uninstall any existing version of the add-in:
   ```bash
   make uninstall-addin
   ```
4. Build and install the Rishi add-in and launch RStudio:
   ```bash
   make up
   ```
5. Open RStudio, then launch the Rishi add-in from the Addins menu.
   The backend daemon will be launched automatically by the add-in.

---

## Repository Structure

- **addin/**  
  Contains the Rishi add-in source code, including the React frontend and R integration.

- **daemon/**
  Contains an HTTP backend server that provides AI chat functionality via Anthropic's Claude API. The backend is automatically launched by the add-in.

---
