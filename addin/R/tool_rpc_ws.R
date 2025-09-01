#' WebSocket-based Tool RPC Server for Cloud Communication
#' 
#' This server establishes a WebSocket connection to the cloud Go backend
#' and handles file operation requests via WebSocket messages.

# WebSocket connection environment to avoid locked binding issues
.ws_env <- new.env()
.ws_env$connection <- NULL
.ws_env$token <- if (Sys.getenv("TIBBL_TOKEN") != "") Sys.getenv("TIBBL_TOKEN") else "tibble-dev-local-please-change"

#' Text Editor View Implementation
#' 
#' Implements the view command for both WebSocket and local usage
#' @param input List containing path and optional view_range
#' @return List with content or error
text_editor_view <- function(input) {
  # Check if path is missing
  if (is.null(input$path) || input$path == "") {
    return(text_editor_view_tool_result(error = "Path is required"))
  }
  
  relative_path <- input$path
  view_range <- input$view_range
  
  # Get safe root
  safe_root <- compute_safe_root()
  if (safe_root == "You are not allowed to list files from root") {
    return(text_editor_view_tool_result(error = safe_root))
  }
  
  # Build absolute path
  absolute_path <- file.path(safe_root, relative_path)
  absolute_path <- normalizePath(absolute_path, winslash = "/", mustWork = FALSE)
  
  # Check if path is within safe root
  if (!startsWith(absolute_path, safe_root)) {
    return(text_editor_view_tool_result(error = "Path outside safe root"))
  }
  
  # Stat the file/directory
  if (!file.exists(absolute_path)) {
    return(text_editor_view_tool_result(error = "Failed to stat file"))
  }
  
  file_info <- file.info(absolute_path)
  result <- ""
  
  if (file_info$isdir) {
    # List directory contents
    tryCatch({
      entries <- list.files(absolute_path, all.files = FALSE, no.. = TRUE, full.names = FALSE)
      
      result <- paste0("Directory listing for '", relative_path, "':\n")
      
      for (entry in entries) {
        entry_path <- file.path(absolute_path, entry)
        entry_info <- file.info(entry_path)
        
        if (entry_info$isdir) {
          result <- paste0(result, entry, "/\n")
        } else {
          result <- paste0(result, entry, "\n")
        }
      }
    }, error = function(e) {
      return(text_editor_view_tool_result(error = "Failed to read directory"))
    })
  } else {
    # Read file contents
    tryCatch({
      lines <- readLines(absolute_path, warn = FALSE)
      
      # Determine which lines to include based on view_range
      start_line <- 1
      end_line <- length(lines)
      
      if (!is.null(view_range) && length(view_range) == 2) {
        start_line <- view_range[1]
        end_line <- min(view_range[2], length(lines))
      }
      
      # Build result with line numbers
      result <- paste0("File contents for '", relative_path, "':\n")
      
      for (i in start_line:end_line) {
        if (i <= length(lines)) {
          result <- paste0(result, i, ": ", lines[i], "\n")
        }
      }
    }, error = function(e) {
      return(text_editor_view_tool_result(error = "Failed to read file"))
    })
  }
  
  return(text_editor_view_tool_result(content = result))
}

#' Handle incoming WebSocket messages
#' 
#' @param message JSON message from Go backend
handle_ws_message <- function(message) {
  tryCatch({
    # Parse the JSON message
    data <- jsonlite::fromJSON(message)
    
    # Check if it's a tool request (from Go backend to R)
    if (!is.null(data$type) && data$type == "tool_request") {
      # Handle both text_editor and str_replace_based_edit_tool
      tool_name <- data$tool
      if (!is.null(tool_name) && (tool_name == "text_editor" || tool_name == "str_replace_based_edit_tool")) {
        if (!is.null(data$command) && data$command == "view" && !is.null(data$input)) {
          # Handle view command with local implementation
          result <- text_editor_view(data$input)
          
          # Send response back with general structure
          response <- list(
            id = data$id,
            type = "tool_response",
            tool = tool_name,
            command = data$command,
            result = result
          )
          
          if (!is.null(.ws_env$connection)) {
            .ws_env$connection$send(jsonlite::toJSON(response, auto_unbox = TRUE))
          } else {
            cat("❌ WebSocket connection is NULL, cannot send response\n")
          }
        }
        # Future: Add other text_editor commands (create, insert, etc.) here
      }
      # Future: Add other tools here
    }
    
    # Note: R doesn't send tool requests to Go, so no need to handle tool_response
  }, error = function(e) {
    # Send error response
    if (!is.null(.ws_env$connection) && !is.null(data$id)) {
      error_response <- list(
        id = data$id,
        type = "tool_response",
        result = text_editor_view_tool_result(error = paste("Internal error:", e$message))
      )
      .ws_env$connection$send(jsonlite::toJSON(error_response, auto_unbox = TRUE))
    }
  })
}

#' Start WebSocket Tool RPC Connection
#' 
#' Establishes a WebSocket connection to the cloud Go backend
#' @param backend_url URL of the Go backend WebSocket endpoint
startToolRPCWebSocket <- function(backend_url = "ws://localhost:8080/ws/tools") {
  tryCatch({
    
    # Create WebSocket connection with authentication header
    .ws_env$connection <- websocket::WebSocket$new(backend_url, 
      headers = list(
        "Authorization" = paste("Bearer", .ws_env$token)
      ),
      autoConnect = FALSE
    )
    
    # Set up event handlers
    .ws_env$connection$onOpen(function(event) {
      # Send initial handshake
      handshake <- list(
        type = "handshake",
        client = "r_tool_rpc",
        version = "1.0.0"
      )
      .ws_env$connection$send(jsonlite::toJSON(handshake, auto_unbox = TRUE))
    })
    
    .ws_env$connection$onMessage(function(event) {
      handle_ws_message(event$data)
    })
    
    .ws_env$connection$onClose(function(event) {
      cat("❌ WebSocket connection closed with code", event$code, "reason:", event$reason, "\n")
    })
    
    .ws_env$connection$onError(function(event) {
      cat("❌ WebSocket error:", event$message, "\n")
    })
    
    # Connect to the backend
    .ws_env$connection$connect()
    
    return(TRUE)
  }, error = function(e) {
    warning(paste("Failed to start WebSocket Tool RPC:", e$message))
    return(FALSE)
  })
}

#' Stop WebSocket Tool RPC Connection
stopToolRPCWebSocket <- function() {
  if (!is.null(.ws_env$connection)) {
    tryCatch({
      # Close the WebSocket connection
      .ws_env$connection$close()
      cat("WebSocket Tool RPC connection stopped\n")
    }, error = function(e) {
      cat("Error stopping WebSocket connection:", e$message, "\n")
    })
    .ws_env$connection <- NULL
  }
}