# Contributing to Rishi

## Prerequisites

- R (version 4.0 or higher recommended)
- RStudio
- Go (for building the daemon backend)
- Node.js and npm (for building the frontend)
- Make

## Repository Structure

- **addin/**
  Contains the Rishi add-in source code, including the React frontend and R integration.

- **daemon/**
  Contains an HTTP backend server that provides AI chat functionality via Anthropic's Claude API. The backend is automatically launched by the add-in.

## Development Setup

1. **Clone this repository:**
   ```bash
   git clone https://github.com/Omkar-Waingankar/rishi.git
   cd rishi
   ```

2. **Build and install for local development:**
   ```bash
   make up
   ```

   This command:
   - Builds both the Go servers and the R addin
   - Immediately launches RStudio so you can test your changes

5. **In RStudio, run:**
   ```r
   setwd("~/projects/penguin-analysis")  # Or your preferred test project
   rishi:::rishiAddin()
   ```

   Note: Rishi needs a set working directory or .Rproj to function properly.

## Development Workflow

The typical development workflow is:

1. Make your changes to the code
2. Run `make uninstall-addin` to ensure a clean slate
3. Run `make up` to rebuild and launch RStudio
4. Test your changes by running `rishi:::rishiAddin()` in the RStudio console
5. The daemon server will be launched automatically by the addin

## Make Commands

There are several Make commands available, but these are the most important for development:

- `make up`: Builds both the Go servers and the R addin, then launches RStudio for testing
- `make uninstall-addin`: Uninstalls the R addin to ensure we're starting from a clean slate

## Questions?

If you have questions or need help, feel free to:
- Open an issue on [GitHub Issues](https://github.com/Omkar-Waingankar/rishi/issues)
- Connect with me (Omkar Waingankar) on [LinkedIn](https://www.linkedin.com/in/omkar-waingankar/)
