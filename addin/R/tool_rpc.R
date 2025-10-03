

#' CORS filter for tool RPC
#' @filter cors
cors_filter <- function(req, res) {
  res$setHeader("Access-Control-Allow-Origin", "*")
  res$setHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
  res$setHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
  
  if (req$REQUEST_METHOD == "OPTIONS") {
    res$status <- 200
    return(list())
  }
  
  plumber::forward()
}


#' Health check endpoint
#' @get /healthz
healthz_endpoint <- function() {
  list(ok = TRUE)
}

#' Compute safe root endpoint
#' @get /safe_root
safe_root_endpoint <- function(req, res) {
  result <- compute_safe_root(is_startup = FALSE)

  # Check if we got an error
  if (result$source == "none") {
    res$status <- 400
    return(list(error = jsonlite::unbox(result$path)))
  }

  # Detect if working directory changed
  changed <- FALSE
  if (!is.null(.wd_state$last_known_wd) && .wd_state$last_known_wd != result$path) {
    changed <- TRUE
  }

  # Update cached state
  .wd_state$last_known_wd <- result$path

  list(
    safe_root = jsonlite::unbox(result$path),
    source = jsonlite::unbox(result$source),
    changed = jsonlite::unbox(changed)
  )
}

#' List files endpoint
#' @post /list
list_files_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  path <- if (is.null(body$path)) "" else body$path
  pattern <- if (is.null(body$pattern)) NULL else body$pattern
  recursive <- if (is.null(body$recursive)) FALSE else body$recursive
  max_items <- if (is.null(body$max_items)) 50 else body$max_items

  result <- compute_safe_root()

  if (result$source == "none") {
    res$status <- 400
    return(list(error = jsonlite::unbox(result$path)))
  }

  safe_root <- result$path
  
  full_path <- if (path == "") {
    safe_root
  } else {
    file.path(safe_root, path)
  }
  
  full_path <- normalizePath(full_path, winslash = "/", mustWork = FALSE)
  
  if (!startsWith(full_path, safe_root)) {
    res$status <- 400
    return(list(error = jsonlite::unbox("Path outside safe root")))
  }
  
  if (!dir.exists(full_path)) {
    res$status <- 400
    return(list(error = jsonlite::unbox("Directory does not exist")))
  }
  
  tryCatch({
    files <- list.files(
      path = full_path,
      pattern = pattern,
      recursive = recursive,
      all.files = FALSE,
      no.. = TRUE,
      full.names = FALSE
    )
    
    # Handle NULL or empty files result
    if (is.null(files)) {
      files <- character(0)
    }
    
    if (length(files) > max_items) {
      files <- files[1:max_items]
    }
    
    # Only modify paths if we have files and path is not empty
    if (path != "" && length(files) > 0) {
      # Ensure path is not NULL and files are valid
      if (!is.null(path) && !is.null(files) && all(!is.na(files))) {
        files <- file.path(path, files, fsep = "/")
      }
    }
    
    # Ensure we always return a character vector
    if (is.null(files)) {
      files <- character(0)
    }
    
    return(list(files = files))
  }, error = function(e) {
    res$status <- 500
    return(list(error = jsonlite::unbox(paste("Failed to list files:", e$message))))
  })
}

#' Text editor view endpoint
#' @post /text_editor/view
text_editor_view_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  # Check if path is missing
  if (is.null(body$path) || body$path == "") {
    return(text_editor_view_tool_result(error = "Path is required"))
  }

  relative_path <- body$path
  view_range <- body$view_range

  # Get safe root
  result <- compute_safe_root()
  if (result$source == "none") {
    return(text_editor_view_tool_result(error = result$path))
  }
  safe_root <- result$path

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

#' Text editor str_replace endpoint
#' @post /text_editor/str_replace
text_editor_str_replace_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  # Check if required fields are missing
  if (is.null(body$path) || body$path == "") {
    return(text_editor_str_replace_tool_result(error = "Path is required"))
  }
  if (is.null(body$old_str)) {
    return(text_editor_str_replace_tool_result(error = "old_str is required"))
  }
  if (is.null(body$new_str)) {
    return(text_editor_str_replace_tool_result(error = "new_str is required"))
  }

  relative_path <- body$path
  old_str <- body$old_str
  new_str <- body$new_str

  # Get safe root
  result <- compute_safe_root()
  if (result$source == "none") {
    return(text_editor_str_replace_tool_result(error = result$path))
  }
  safe_root <- result$path

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

#' Text editor create endpoint
#' @post /text_editor/create
text_editor_create_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  # Check if required fields are missing
  if (is.null(body$path) || body$path == "") {
    return(text_editor_create_tool_result(error = "Path is required"))
  }
  if (is.null(body$file_text)) {
    return(text_editor_create_tool_result(error = "file_text is required"))
  }

  relative_path <- body$path
  file_text <- body$file_text

  # Get safe root
  result <- compute_safe_root()
  if (result$source == "none") {
    return(text_editor_create_tool_result(error = result$path))
  }
  safe_root <- result$path

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

#' Text editor insert endpoint
#' @post /text_editor/insert
text_editor_insert_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  # Check if required fields are missing
  if (is.null(body$path) || body$path == "") {
    return(text_editor_insert_tool_result(error = "Path is required"))
  }
  if (is.null(body$insert_line) || !is.numeric(body$insert_line)) {
    return(text_editor_insert_tool_result(error = "insert_line is required and must be a number"))
  }
  if (is.null(body$new_str)) {
    return(text_editor_insert_tool_result(error = "new_str is required"))
  }

  relative_path <- body$path
  insert_line <- as.integer(body$insert_line)
  new_str <- body$new_str

  # Get safe root
  result <- compute_safe_root()
  if (result$source == "none") {
    return(text_editor_insert_tool_result(error = result$path))
  }
  safe_root <- result$path

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


#' Console exec endpoint
#' @post /console/exec
console_exec_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  # Check if code is missing
  if (is.null(body$code) || body$code == "") {
    return(list(
      content = jsonlite::unbox(""),
      error = jsonlite::unbox("Code is required")
    ))
  }

  code <- body$code

  tryCatch({
    # Send code to console and execute it
    rstudioapi::sendToConsole(code, execute = TRUE, focus = TRUE)

    return(list(
      content = jsonlite::unbox("Code executed successfully."),
      error = jsonlite::unbox("")
    ))
  }, error = function(e) {
    return(list(
      content = jsonlite::unbox(""),
      error = jsonlite::unbox(paste("Failed to execute code:", e$message))
    ))
  })
}

#' Start Tool RPC Server
#'
#' Starts a plumber server on port 8082 for tool operations
startToolRPC <- function() {
  library(plumber)
  library(jsonlite)

  # Create plumber API programmatically
  pr <- plumber::pr() %>%
    plumber::pr_filter("cors", cors_filter) %>%
    plumber::pr_get("/healthz", healthz_endpoint) %>%
    plumber::pr_get("/safe_root", safe_root_endpoint) %>%
    plumber::pr_post("/list", list_files_endpoint) %>%
    plumber::pr_post("/text_editor/view", text_editor_view_endpoint) %>%
    plumber::pr_post("/text_editor/str_replace", text_editor_str_replace_endpoint) %>%
    plumber::pr_post("/text_editor/create", text_editor_create_endpoint) %>%
    plumber::pr_post("/text_editor/insert", text_editor_insert_endpoint) %>%
    plumber::pr_post("/console/exec", console_exec_endpoint) %>%
    plumber::pr_set_serializer(plumber::serializer_unboxed_json())

  # Start server
  tryCatch({
    server <- httpuv::startServer(host = "127.0.0.1", port = 8082, pr)
    return(invisible(server))
  }, error = function(e) {
    warning(paste("Failed to start Tool RPC server on port 8082:", e$message))
    return(NULL)
  })
}