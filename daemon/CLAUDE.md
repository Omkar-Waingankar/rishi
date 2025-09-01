# Context

This is the directory that holds the backend code for the Tibbl addin. In local development, it is served at port 8080 but in production it will be hosted in the cloud (not deployed yet, but assume at api.tibbl.ai on Render).

# Endpoints

There are two mission critical endpoints in the Tibbl backend
- `/chat`: Our frontend hits this endpoint to allow the user to chat with LLMs. These LLMs have agentic capabilities (they can access tools to interact with RStudio, the end-user's filesystem, web search, etc.) It currently supports only Anthropic Claude Sonnet 4, but our offerings will expand with time. 
- `/ws/tools`: This exposes a websocket endpoint that our R websocket server (served on the end-user's local machine) connects to in order to bidirectionally service tool calls made by our LLMs, and avoid repeated handshakes that would be required of a HTTP endpoint. For now, we're focused on implementing and supporting just the text_editor tool for Anthropic, but this will expand over time as well. 

# Directories

- `cmd/server`: There is a single main.go file in this directory, and it is the entrypoint to the HTTP server. This is where we initialize critical clients to communicate with external services. For now, that's just Anthropic, but in the future it will expand to other LLM providers such as OpenAI and also other development tools like Supabase.
- `internal/api`: This is the directory in which the meat of our business logic lives. This is where our handlers, middleware, tool calls, errors, utils etc. are all defined.

# Core files in internal/api

- `handlers.go`: Where the handleChat function is defined which powers the `/chat` endpoint. As this handler expands in complexity, it's imperative that the majority of its business logic has to do with only 1. assembling requests to anthropic, 2. serving as a high level orchestrator between the various tool calls and the agentic flow within a single turn of conversation, and 3. writing back to the frontend so the end-user can understand what the LLM is doing to solve the problem. Anything else (tool call definitions, system prompt, etc.) should be delegated to other files. 
- `text_editor_tool.go`: Where we define the textEditorController which is essentially our implementation/support of the Anthropic text_editor tool. So far, we've only supported the "view" command but will be filling in support for all the other commands soon. 
- `websocket.go`: Where we define critical boilerplate to power the websocket endpoint. This should hopefully not have to change much in the future, but if it does, make additions/updates keeping in mind future use and extensibility. 
- `websocket_tool_rpc.go`: This is where our business logic related to websockets is defined. It's where we send tool commands on behalf of the LLMs and also parse tool responses before passing them upstream to the textEditorController and ultimately the `/chat` handler.

# Go programming best practices

- When adding environment variables, manage them via the envconfig and godotenv packages
- Use zerolog for logging. No specific required fields for logging yet.
- When adding tools (in either tools.go or text_editor_tool.go), create explicit types for their inputs and outputs so it's easy for us to marshal/unmarshal. 

# Miscellaneous files in internal/api 
- `tools.go`: This is a legacy file for tool calls. I've kept it here for future reference in case we need to define custom tools outside of what Anthropic's native text_editor tool can provide. 
- `utils.go`: Utilities for other files in the api package.
- `prompts.go`: Where we define critical prompts for the LLMs.
- `server.go`: Where the ServerClient is defined which has supporting clients for external services like Anthropic and is also where our endpoints are defined.
- `middleware.go`: This is where our API middleware is defined.