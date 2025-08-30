#' Compute safe root directory for file operations
#' 
#' @return Character string of safe root directory or refusal message
compute_safe_root <- function() {
  # Try to get project root from RStudio API
  project_root <- tryCatch({
    rstudioapi::getActiveProject()
  }, error = function(e) {
    NULL
  })
  
  # Fallback to current working directory if no project
  if (is.null(project_root)) {
    project_root <- getwd()
  }
  
  # Normalize the path
  project_root <- normalizePath(project_root, winslash = "/", mustWork = FALSE)
  
  # Check if path resolves to home directory or system root
  home_dir <- normalizePath("~", winslash = "/", mustWork = FALSE)
  
  # Check for system root patterns
  is_root <- grepl("^/$", project_root) ||  # Unix root
             grepl("^[A-Za-z]:/$", project_root) ||  # Windows root (C:/, D:/, etc.)
             identical(project_root, home_dir)  # Home directory
  
  if (is_root) {
    return("You are not allowed to list files from root")
  }
  
  return(project_root)
}

# Global variables for the server
.tool_rpc_token <- "tibble-dev-local-please-change"

#' Authentication filter for tool RPC
#' @filter auth
auth_filter <- function(req, res) {
  auth_header <- req$HTTP_AUTHORIZATION
  
  if (is.null(auth_header)) {
    res$status <- 401
    return(list(error = jsonlite::unbox("Missing Authorization header")))
  }
  
  if (!startsWith(auth_header, "Bearer ")) {
    res$status <- 401
    return(list(error = jsonlite::unbox("Invalid Authorization header format")))
  }
  
  token <- substring(auth_header, 8)
  
  if (token != .tool_rpc_token) {
    res$status <- 401
    return(list(error = jsonlite::unbox("Invalid token")))
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
    plumber::pr_filter("auth", auth_filter) %>%
    plumber::pr_get("/healthz", healthz_endpoint) %>%
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