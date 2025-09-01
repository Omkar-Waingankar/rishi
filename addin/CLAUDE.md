# Context

This is the directory that holds both the frontend for the Tibbl addin as well as critical http and websocket servers written in R. 

# Directories

- `inst`: You don't really need to worry about this one. It's basically where our final Tibbl frontend gets built. Specifically, `inst/www/chat-app.js`.
- `R`: This is where we define the entrypoint to our addin, HTTP/websocket servers to provide tool call capabilities to the LLM in our backend, and other R utility files.
- `src`: This is where the React frontend is defined for our Tibbl chat interface.

# Tips for navigating R directory

- `addin.R`: Short and sweet launcher for our addin. It is the entrypoint that spins up our HTTP and websocket servers, serves the frontend at port 8081, and also automatically opens the viewer in RStudio for the end-user.
- `constructors.R`: Use this file to define "types" for our HTTP and Websocket servers. It's much better than having websocket endpoints return freeform list().
- `utils.R`: Use this file to define common utility functions. Right now it just has `compute_safe_root` which uses the rstudioapi and getwd() to determine if the end user has set a working directory. This is essential to ensure LLMs do not have arbitrary root access to our end user's computers.
- `tool_rpc.R`: This is the HTTP server spun up at port 8082 by `addin.R`. It's very small because right now it only serves a very simple purpose: allowing the frontend to check that the end user has set a working directory via the endpoint /safe_root. For now, we want to keep this as minimal as possible. There are other endpoints like /list_files but these are legacy and were part of an older tools implementation that we don't need to expand upon. We've kept them around because they're decent code examples.
- `tool_rpc_ws.R`: This is the most important file in the `R` directory. It defines a websocket server that our backend pushes tool call uses/commands to and in turn this server processes them and return back tool call results. For now, we've just implemented the "view" command of Anthropic's text_editor tool, but will be expanding this functionality soon.

# Tips for navigating src directory

- `index.tsx`: This is the entrypoint for our React frontend. It's very simple - it just finds the DOM element with id "chat-root" and renders our main ChatApp component. This file gets built into `inst/www/chat-app.js` by webpack.
- `ChatApp.tsx`: This is the main React component that orchestrates everything. It manages the chat state, handles websocket connections to our R backend, and coordinates between the MessageList and InputBox components. This is where most of the frontend logic lives.
- `MessageList.tsx`: Renders the list of chat messages. It handles both user and assistant messages, including special rendering for tool calls and errors. Each message can have multiple content pieces (text, tool calls, etc.) and supports expandable tool call results.
- `InputBox.tsx`: The input component at the bottom of the chat. It handles text input, the send button, and shows the currently set LLM model via `ModelDropdown.tsx`. It also has a stop button that appears during streaming responses.
- `ModelDropdown.tsx`: A dropdown component for selecting different AI models. Currently supports Claude 3.5 Sonnet and Claude 3 Opus. This is where you'd add new models in the future, for now it's just a UI placeholder.
- `types.ts`: Defines the core TypeScript interfaces for our chat system. This includes Message, MessageContent, and various props interfaces. The Message interface is particularly important as it defines how we structure chat data.
- `tool_types.ts`: Defines the TypeScript interfaces for our tool system. This includes ToolCommand enum, ToolCallStatus, and various input/output types for tools like the text editor tool. This file ensures type safety when communicating with our R backend.
- `styles.css`: Contains all the CSS styling for our chat interface. It's designed to match the RStudio aesthetic with tight spacing, off-white backgrounds for user messages, and a fixed input box at the bottom.

# R programming best practices

- When implementing new functions/endpoints for the websocket server in R, don't forget to define new output "types" in constructors.R
- Attempt to leverage existing packages installed to accomplish all coding tasks. If you absolutely need to import a new package, go for one that is popular on CRAN. 
- Keep things as short and concise as possible.
- If code needs to be repeated twice, create a helper function. 

# React programming best practices 

- For now, keep all CSS styling in the single file styles.css. We may consider improving or refactoring this as the codebase grows, but it seems fine for now.
- It is very important to update types in `types.ts` as Message definitions expand/shrink, and similarly to add/update types in `tool_types.ts` as we enable LLMs
- I'm a big fan of componentization. If you're adding a significantly new element to the Chat interface, define a new component in a completely separate file and import it from there