

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
  safe_root <- compute_safe_root()
  if (safe_root == "You are not allowed to list files from root") {
    res$status <- 400
    return(list(error = jsonlite::unbox(safe_root)))
  }
  list(safe_root = jsonlite::unbox(safe_root))
}

#' List files endpoint
#' @post /list
list_files_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)
  
  path <- if (is.null(body$path)) "" else body$path
  pattern <- if (is.null(body$pattern)) NULL else body$pattern
  recursive <- if (is.null(body$recursive)) FALSE else body$recursive
  max_items <- if (is.null(body$max_items)) 50 else body$max_items
  
  safe_root <- compute_safe_root()
  
  if (safe_root == "You are not allowed to list files from root") {
    res$status <- 400
    return(list(error = jsonlite::unbox(safe_root)))
  }
  
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

#' Get API key endpoint
#' @get /api_key
get_api_key_endpoint <- function(req, res) {
  api_key <- getApiKey()

  if (is.null(api_key)) {
    return(list(has_key = jsonlite::unbox(FALSE)))
  }

  return(list(has_key = jsonlite::unbox(TRUE)))
}

#' Validate API key endpoint
#' @post /validate_api_key
validate_api_key_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  api_key <- body$api_key

  if (is.null(api_key) || api_key == "") {
    res$status <- 400
    return(list(error = jsonlite::unbox("Missing api_key parameter")))
  }

  # Basic format validation
  if (!startsWith(api_key, "sk-ant-")) {
    return(list(valid = jsonlite::unbox(FALSE)))
  }

  if (nchar(api_key) < 20) {
    return(list(valid = jsonlite::unbox(FALSE)))
  }

  # Test the API key with Anthropic API using httr
  tryCatch({
    response <- httr::POST(
      url = "https://api.anthropic.com/v1/messages",
      httr::add_headers(
        `x-api-key` = api_key,
        `anthropic-version` = "2023-06-01",
        `content-type` = "application/json"
      ),
      body = jsonlite::toJSON(list(
        model = "claude-3-haiku-20240307",
        max_tokens = 1,
        messages = list(list(role = "user", content = "hi"))
      ), auto_unbox = TRUE),
      encode = "raw"
    )

    # 200 = success, 400 = validation error but auth worked
    is_valid <- httr::status_code(response) %in% c(200, 400)

    return(list(valid = jsonlite::unbox(is_valid)))
  }, error = function(e) {
    # Network error or API unreachable
    return(list(valid = jsonlite::unbox(FALSE)))
  })
}

#' Set API key endpoint
#' @post /api_key
set_api_key_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)

  api_key <- body$api_key

  if (is.null(api_key) || api_key == "") {
    res$status <- 400
    return(list(error = jsonlite::unbox("Missing api_key parameter")))
  }

  success <- setApiKey(api_key)

  if (!success) {
    res$status <- 500
    return(list(error = jsonlite::unbox("Failed to save API key")))
  }

  # Try to start the daemon now that we have an API key
  tryCatch({
    startDaemon()
    # Also establish WebSocket connection
    Sys.sleep(1)  # Give daemon a moment to start
    startToolRPCWebSocket()
  }, error = function(e) {
    # Daemon start failed, but API key was saved successfully
    cat("Warning: Failed to start daemon:", e$message, "\n")
  })

  return(list(success = jsonlite::unbox(TRUE)))
}

#' Read file endpoint
#' @post /read
read_file_endpoint <- function(req, res) {
  body <- jsonlite::fromJSON(req$postBody)
  
  relpath <- body$relpath
  max_bytes <- if (is.null(body$max_bytes)) 2000000 else body$max_bytes
  
  if (is.null(relpath) || relpath == "") {
    res$status <- 400
    return(list(error = jsonlite::unbox("Missing relpath parameter")))
  }
  
  safe_root <- compute_safe_root()
  
  if (safe_root == "You are not allowed to list files from root") {
    res$status <- 400
    return(list(error = jsonlite::unbox(safe_root)))
  }
  
  full_path <- file.path(safe_root, relpath)
  full_path <- normalizePath(full_path, winslash = "/", mustWork = FALSE)
  
  if (!startsWith(full_path, safe_root)) {
    res$status <- 400
    return(list(error = jsonlite::unbox("Path outside safe root")))
  }
  
  if (!file.exists(full_path)) {
    res$status <- 404
    return(list(error = jsonlite::unbox("File not found")))
  }
  
  if (dir.exists(full_path)) {
    res$status <- 400
    return(list(error = jsonlite::unbox("Path is a directory, not a file")))
  }
  
  file_info <- file.info(full_path)
  if (is.na(file_info$size) || file_info$size > max_bytes) {
    res$status <- 400
    return(list(error = jsonlite::unbox(paste("File too large or unreadable:", file_info$size, "bytes, max:", max_bytes))))
  }
  
  # Handle empty files
  if (file_info$size == 0) {
    return(list(content = ""))
  }
  
  tryCatch({
    # Read file content using readLines
    content <- paste(readLines(full_path, warn = FALSE), collapse = "\n")
    
    # Ensure content is a string and create proper response
    if (is.null(content)) {
      content <- ""
    }
    content <- as.character(content)[1]
    
    # Manually create JSON to bypass plumber's boxing behavior
    json_response <- jsonlite::toJSON(list(content = content), auto_unbox = TRUE)
    
    # Set content type and return raw JSON
    res$setHeader("Content-Type", "application/json")
    res$body <- as.character(json_response)
    return(res)
  }, error = function(e) {
    res$status <- 500
    return(list(error = jsonlite::unbox(paste("Failed to read file:", e$message))))
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
    plumber::pr_get("/api_key", get_api_key_endpoint) %>%
    plumber::pr_post("/validate_api_key", validate_api_key_endpoint) %>%
    plumber::pr_post("/api_key", set_api_key_endpoint) %>%
    plumber::pr_post("/list", list_files_endpoint) %>%
    plumber::pr_post("/read", read_file_endpoint) %>%
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