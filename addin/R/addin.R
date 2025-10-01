#' Launch Rishi
#'
#' This function launches the Rishi chat interface as an RStudio add-in.
#' The interface provides a chat-based UI for interacting with AI assistance.
#' Uses rstudioapi::viewer() to keep the R console free for code execution.
#'
#' @export
rishiAddin <- function() {
  # Clean up any previous Rishi processes and servers
  cleanupRishi()
  Sys.sleep(0.5)  # Give processes time to clean up

  # Get the path to the www directory
  www_dir <- system.file("www", package = "rishi")

  if (!dir.exists(www_dir)) {
    stop("Web assets not found. Make sure to run 'make build-addin' first.")
  }

  # Start the Tool RPC server first (HTTP for frontend) AND WebSocket connection
  # This needs to be running before the frontend loads so it can check for API key
  startToolRPC()
  startToolRPCWebSocket()

  # Start a simple HTTP server to serve the React app (always on port 8081)
  server_port <- startLocalServer(www_dir)

  # Try to start the daemon if API key is available
  # If not available, frontend will show setup UI
  tryCatch({
    startDaemon()
  }, error = function(e) {
    # Daemon failed to start - this is okay, user may need to set up API key
  })

  # Open in RStudio viewer pane
  viewer_url <- paste0("http://127.0.0.1:", server_port, "/index.html")
  rstudioapi::viewer(viewer_url, height = "maximize")

  # Display ASCII art and welcome message
  cat("\n")
  cat("  ██████╗ ██╗███████╗██╗  ██╗██╗\n")
  cat("  ██╔══██╗██║██╔════╝██║  ██║██║\n")
  cat("  ██████╔╝██║███████╗███████║██║\n")
  cat("  ██╔══██╗██║╚════██║██╔══██║██║\n")
  cat("  ██║  ██║██║███████║██║  ██║██║\n")
  cat("  ╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝\n")
  cat("\n")
  cat("🚀 Ready to transform your R workflow! Rishi is here to help.\n")
}

#' Start a simple HTTP server to serve static files
#' @param www_dir Directory containing web assets
#' @return Port number of the started server
startLocalServer <- function(www_dir) {
  # Always use port 8081 for HTTP server (fixed port for predictability)
  port <- 8081

  # Start the HTTP server to serve the React frontend
  server <- httpuv::startServer("127.0.0.1", port,
    list(
      call = function(req) {
        # Simple static file server
        path <- req$PATH_INFO
        if (path == "/" || path == "") path <- "/index.html"
        
        file_path <- file.path(www_dir, substring(path, 2))
        
        if (file.exists(file_path)) {
          # Determine content type
          ext <- tools::file_ext(file_path)
          content_type <- switch(ext,
            "html" = "text/html",
            "js" = "application/javascript", 
            "css" = "text/css",
            "json" = "application/json",
            "text/plain"
          )
          
          # Read and return file
          content <- readBin(file_path, "raw", file.info(file_path)$size)
          list(
            status = 200L,
            headers = list("Content-Type" = content_type),
            body = content
          )
        } else {
          list(
            status = 404L,
            headers = list("Content-Type" = "text/plain"),
            body = "File not found"
          )
        }
      }
    )
  )

  return(port)
}

#' Start the Rishi daemon if not already running
#'
#' This function detects the user's platform and starts the appropriate daemon binary
#' that was bundled with the R package during installation.
startDaemon <- function() {
  # Check if daemon is already running
  if (isDaemonRunning()) {
    return(invisible(TRUE))
  }

  # Get the daemon binary path for current platform
  daemon_path <- getDaemonPath()

  if (is.null(daemon_path) || !file.exists(daemon_path)) {
    stop("Daemon binary not found for your platform. Please reinstall the package.")
  }

  # Start daemon in background
  tryCatch({
    # Make binary executable on Unix systems (required for security)
    if (.Platform$OS.type == "unix") {
      Sys.chmod(daemon_path, mode = "0755")
    }

    # Start daemon as background process (works on both Windows and Unix)
    # No environment variables needed - daemon accepts API key via header
    result <- system2(daemon_path, wait = FALSE, stdout = FALSE, stderr = FALSE)

    # Give daemon time to start up (typically takes 1-3 seconds)
    Sys.sleep(3)

    # Verify daemon started successfully by checking health endpoint
    if (isDaemonRunning()) {
      return(invisible(TRUE))
    } else {
      stop("Failed to start daemon - daemon not responding on port 8080")
    }

  }, error = function(e) {
    stop(paste("Failed to start daemon:", e$message))
  })
}

#' Check if the Rishi daemon is running
#'
#' @return Logical indicating if daemon is responding on port 8080
isDaemonRunning <- function() {
  tryCatch({
    # Try to connect to daemon health endpoint
    response <- httr::GET("http://localhost:8080/health", httr::timeout(2))
    return(httr::status_code(response) == 200)
  }, error = function(e) {
    return(FALSE)
  })
}

#' Get the daemon binary path for the current platform
#'
#' @return Character path to daemon binary, or NULL if not found
getDaemonPath <- function() {
  # Detect platform
  os <- tolower(Sys.info()[["sysname"]])
  arch <- tolower(Sys.info()[["machine"]])

  # Map R platform names to our binary names
  platform_map <- list(
    "darwin" = list("x86_64" = "darwin-amd64", "arm64" = "darwin-arm64"),
    "linux" = list("x86_64" = "linux-amd64"),
    "windows" = list("x86_64" = "windows-amd64")
  )

  # Get binary suffix
  binary_suffix <- ""
  if (os == "windows") {
    binary_suffix <- ".exe"
  }

  # Map architecture names
  if (arch %in% c("amd64", "x64")) arch <- "x86_64"

  # Get platform string
  if (os %in% names(platform_map) && arch %in% names(platform_map[[os]])) {
    platform_string <- platform_map[[os]][[arch]]
  } else {
    warning(paste("Unsupported platform:", os, arch))
    return(NULL)
  }

  # Build binary path
  binary_name <- paste0("rishi-daemon-", platform_string, binary_suffix)
  daemon_path <- system.file("bin", binary_name, package = "rishi")

  if (daemon_path == "") {
    return(NULL)
  }

  return(daemon_path)
}


#' Clean up all Rishi processes and servers
#'
#' This function stops all HTTP servers and kills any running daemon processes
#' to ensure a clean slate before starting Rishi.
#'
#' @export
cleanupRishi <- function() {
  # Clean up any existing HTTP servers
  tryCatch({
    httpuv::stopAllServers()
  }, error = function(e) {
    cat("Warning: Failed to stop HTTP servers:", e$message, "\n")
  })

  # Kill any processes using Rishi ports (8080 for daemon, 8081 for HTTP server)
  tryCatch({
    if (.Platform$OS.type == "unix") {
      # Use lsof to find and kill processes on our ports
      system("lsof -ti:8080 | xargs kill -9 2>/dev/null || true", ignore.stdout = TRUE, ignore.stderr = TRUE)
      system("lsof -ti:8081 | xargs kill -9 2>/dev/null || true", ignore.stdout = TRUE, ignore.stderr = TRUE)
      # Also kill by process name for good measure
      system("pkill -f 'rishi-daemon' 2>/dev/null || true", ignore.stdout = TRUE, ignore.stderr = TRUE)
    } else {
      # Windows: kill by port using netstat and taskkill
      system("for /f \"tokens=5\" %a in ('netstat -aon ^| find \":8080\"') do taskkill /F /PID %a 2>nul", ignore.stdout = TRUE, ignore.stderr = TRUE)
      system("for /f \"tokens=5\" %a in ('netstat -aon ^| find \":8081\"') do taskkill /F /PID %a 2>nul", ignore.stdout = TRUE, ignore.stderr = TRUE)
      # Also kill by process name
      system("taskkill /F /IM rishi-daemon*.exe 2>nul", ignore.stdout = TRUE, ignore.stderr = TRUE)
    }
  }, error = function(e) {
    cat("Warning: Port cleanup failed:", e$message, "\n")
  })
}