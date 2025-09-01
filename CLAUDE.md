# Context

This repository holds the code for Tibbl, an AI coding agent for R. It is currently in development, and the intention is that it will be packaged as a RStudio Addin for use by any existing RStudio user.

# Directories

- `/addin`: This is where the frontend for Tibbl lives (src). It's also where we define both a HTTP server (tool_rpc.R) and a websocket server (tool_rpc_ws.R) to service tool calls made by LLMs to explore files, write code, interact with RStudio, etc. 
- `daemon`: This is where the backend API server for Tibbl lives. It primarily serves two endpoints: /chat (used by the frontend to have conversations with LLMs) and /ws/tools (a websocket connection used to enable tool calling by the LLMs).

# Make commands

There are a bunch of Make commands we've defined, but there are only 4 important ones.
- `make build-server`: This builds the Go server and is used to verify there are no compile time issues.
- `make run-server`: This is used by me to actually run the Go server locally
- `make install-addin`: This builds the R addin and is used to verify there are no package install or compile time issues.
- `make install-and-launch-addin`: This builds the R addin and additionally launches open a new RStudio window.

# Local development / testing

If I want to use Tibbl in my local dev setup, I first run `make run-server` in one terminal window and then open another terminal window and run `make install-and-launch-addin`. From within the RStudio console, I then run `setwd("~/projects/r-testbed/penguin-analysis")` (Tibbl needs a working directory or .Rproj) followed `by tibblai:::tibblAddin()` to actually launch the addin.

# Maintenance

Please keep all CLAUDE.md files up to date as development progresses, especially if they are out of date relative to what is actually happening in the codebase. Do not change opinionated subjects like "Programming best practices" but do change factual information such as if particular packages, databases, or models we use change.