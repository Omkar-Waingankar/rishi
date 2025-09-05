#' WebSocket-based Tool RPC Server for Cloud Communication
#' 
#' This server establishes a WebSocket connection to the cloud Go backend
#' and handles file operation requests via WebSocket messages.

# WebSocket connection environment to avoid locked binding issues
.ws_env <- new.env()
.ws_env$connection <- NULL
.ws_env$token <- if (Sys.getenv("RISHI_TOKEN") != "") Sys.getenv("RISHI_TOKEN") else "rishi-dev-local-please-change"

#' Text Editor Str Replace Implementation
#' 
#' Implements the str_replace command for both WebSocket and local usage
#' @param input List containing path, old_str, and new_str
#' @return List with content or error
text_editor_str_replace <- function(input) {
  # Check if required fields are missing
  if (is.null(input$path) || input$path == "") {
    return(text_editor_str_replace_tool_result(error = "Path is required"))
  }
  if (is.null(input$old_str)) {
    return(text_editor_str_replace_tool_result(error = "old_str is required"))
  }
  if (is.null(input$new_str)) {
    return(text_editor_str_replace_tool_result(error = "new_str is required"))
  }
  
  relative_path <- input$path
  old_str <- input$old_str
  new_str <- input$new_str
  
  # Get safe root
  safe_root <- compute_safe_root()
  if (safe_root == "You are not allowed to list files from root") {
    return(text_editor_str_replace_tool_result(error = safe_root))
  }
  
  # Build absolute path
  absolute_path <- file.path(safe_root, relative_path)
  absolute_path <- normalizePath(absolute_path, winslash = "/", mustWork = FALSE)
  
  # Check if path is within safe root
  if (!startsWith(absolute_path, safe_root)) {
    return(text_editor_str_replace_tool_result(error = "Path outside safe root"))
  }
  
  # Check if file exists
  if (!file.exists(absolute_path)) {
    return(text_editor_str_replace_tool_result(error = "File not found"))
  }
  
  # Check if path is a directory
  if (file.info(absolute_path)$isdir) {
    return(text_editor_str_replace_tool_result(error = "Cannot edit directory"))
  }
  
  tryCatch({
    # Read file contents
    content <- readChar(absolute_path, file.info(absolute_path)$size)
    
    # Count occurrences of old_str
    matches <- length(gregexpr(old_str, content, fixed = TRUE)[[1]])
    if (matches == 0 || (matches == 1 && gregexpr(old_str, content, fixed = TRUE)[[1]][1] == -1)) {
      return(text_editor_str_replace_tool_result(error = "No match found for replacement"))
    }
    if (matches > 1) {
      return(text_editor_str_replace_tool_result(error = paste("Found", matches, "matches for replacement text. Please provide more specific text to make a unique match.")))
    }
    
    # Perform replacement
    new_content <- gsub(old_str, new_str, content, fixed = TRUE)
    
    # Write back to file
    writeChar(new_content, absolute_path, eos = NULL)
    
    return(text_editor_str_replace_tool_result(content = "Successfully replaced text at exactly one location."))
    
  }, error = function(e) {
    return(text_editor_str_replace_tool_result(error = paste("Failed to replace text:", e$message)))
  })
}

#' Text Editor Create Implementation
#' 
#' Implements the create command for both WebSocket and local usage
#' @param input List containing path and file_text
#' @return List with content or error
text_editor_create <- function(input) {
  # Check if required fields are missing
  if (is.null(input$path) || input$path == "") {
    return(text_editor_create_tool_result(error = "Path is required"))
  }
  if (is.null(input$file_text)) {
    return(text_editor_create_tool_result(error = "file_text is required"))
  }
  
  relative_path <- input$path
  file_text <- input$file_text
  
  # Get safe root
  safe_root <- compute_safe_root()
  if (safe_root == "You are not allowed to list files from root") {
    return(text_editor_create_tool_result(error = safe_root))
  }
  
  # Build absolute path
  absolute_path <- file.path(safe_root, relative_path)
  absolute_path <- normalizePath(absolute_path, winslash = "/", mustWork = FALSE)
  
  # Check if path is within safe root
  if (!startsWith(absolute_path, safe_root)) {
    return(text_editor_create_tool_result(error = "Path outside safe root"))
  }
  
  # Check if file already exists
  if (file.exists(absolute_path)) {
    return(text_editor_create_tool_result(error = "File already exists"))
  }
  
  tryCatch({
    # Create directory if it doesn't exist
    dir_path <- dirname(absolute_path)
    if (!dir.exists(dir_path)) {
      dir.create(dir_path, recursive = TRUE)
    }
    
    # Write content to file
    writeChar(file_text, absolute_path, eos = NULL)
    
    return(text_editor_create_tool_result(content = paste("Successfully created file:", relative_path)))
    
  }, error = function(e) {
    return(text_editor_create_tool_result(error = paste("Failed to create file:", e$message)))
  })
}

#' Text Editor Insert Implementation
#' 
#' Implements the insert command for both WebSocket and local usage
#' @param input List containing path, insert_line, and new_str
#' @return List with content or error
text_editor_insert <- function(input) {
  # Check if required fields are missing
  if (is.null(input$path) || input$path == "") {
    return(text_editor_insert_tool_result(error = "Path is required"))
  }
  if (is.null(input$insert_line) || !is.numeric(input$insert_line)) {
    return(text_editor_insert_tool_result(error = "insert_line is required and must be a number"))
  }
  if (is.null(input$new_str)) {
    return(text_editor_insert_tool_result(error = "new_str is required"))
  }
  
  relative_path <- input$path
  insert_line <- as.integer(input$insert_line)
  new_str <- input$new_str
  
  # Get safe root
  safe_root <- compute_safe_root()
  if (safe_root == "You are not allowed to list files from root") {
    return(text_editor_insert_tool_result(error = safe_root))
  }
  
  # Build absolute path
  absolute_path <- file.path(safe_root, relative_path)
  absolute_path <- normalizePath(absolute_path, winslash = "/", mustWork = FALSE)
  
  # Check if path is within safe root
  if (!startsWith(absolute_path, safe_root)) {
    return(text_editor_insert_tool_result(error = "Path outside safe root"))
  }
  
  # Check if file exists
  if (!file.exists(absolute_path)) {
    return(text_editor_insert_tool_result(error = "File not found"))
  }
  
  # Check if path is a directory
  if (file.info(absolute_path)$isdir) {
    return(text_editor_insert_tool_result(error = "Cannot edit directory"))
  }
  
  tryCatch({
    # Read file lines
    lines <- readLines(absolute_path, warn = FALSE)
    
    # Validate insert_line
    if (insert_line < 0) {
      return(text_editor_insert_tool_result(error = "insert_line must be >= 0"))
    }
    if (insert_line > length(lines)) {
      return(text_editor_insert_tool_result(error = paste("insert_line", insert_line, "is beyond file length", length(lines))))
    }
    
    # Split new_str by newlines for proper insertion
    new_lines <- strsplit(new_str, "\n")[[1]]
    
    # Insert text at the specified line
    if (insert_line == 0) {
      # Insert at the beginning of the file
      result_lines <- c(new_lines, lines)
    } else {
      # Insert after the specified line
      result_lines <- c(lines[1:insert_line], new_lines, lines[(insert_line + 1):length(lines)])
    }
    
    # Remove any NA lines that might result from empty parts
    result_lines <- result_lines[!is.na(result_lines)]
    
    # Write back to file
    writeLines(result_lines, absolute_path)
    
    return(text_editor_insert_tool_result(content = paste("Successfully inserted text after line", insert_line)))
    
  }, error = function(e) {
    return(text_editor_insert_tool_result(error = paste("Failed to insert text:", e$message)))
  })
}

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
        } else if (!is.null(data$command) && data$command == "str_replace" && !is.null(data$input)) {
          # Handle str_replace command with local implementation
          result <- text_editor_str_replace(data$input)
          
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
        } else if (!is.null(data$command) && data$command == "create" && !is.null(data$input)) {
          # Handle create command with local implementation
          result <- text_editor_create(data$input)
          
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
        } else if (!is.null(data$command) && data$command == "insert" && !is.null(data$input)) {
          # Handle insert command with local implementation
          result <- text_editor_insert(data$input)
          
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
        # All text_editor commands have been implemented
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