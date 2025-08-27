#' Launch Tibbl
#'
#' This function launches the Tibbl chat interface as an RStudio add-in.
#' The interface provides a chat-based UI for interacting with AI assistance.
#' Uses rstudioapi::viewer() to keep the R console free for code execution.
#'
#' @export
tibblAddin <- function() {
  # Get the path to the www directory
  www_dir <- system.file("www", package = "tibblai")
  
  if (!dir.exists(www_dir)) {
    stop("Web assets not found. Make sure to run 'make build-addin' first.")
  }
  
  # Start a simple HTTP server to serve the React app
  server_port <- startLocalServer(www_dir)
  
  # Open in RStudio viewer pane
  viewer_url <- paste0("http://127.0.0.1:", server_port, "/index.html")
  rstudioapi::viewer(viewer_url, height = "maximize")
  
  # Display ASCII art and welcome message
  cat("\n")
  cat("  ████████╗██╗██████╗ ██████╗ ██╗     \n")
  cat("  ╚══██╔══╝██║██╔══██╗██╔══██╗██║     \n")
  cat("     ██║   ██║██████╔╝██████╔╝██║     \n")
  cat("     ██║   ██║██╔══██╗██╔══██╗██║     \n")
  cat("     ██║   ██║██████╔╝██████╔╝███████╗\n")
  cat("     ╚═╝   ╚═╝╚═════╝ ╚═════╝ ╚══════╝\n")
  cat("\n")
  cat("🚀 Ready to transform your R workflow! Visit tibbl.ai to learn more.\n")
}

#' Start a simple HTTP server to serve static files
#' @param www_dir Directory containing web assets
#' @return Port number of the started server
startLocalServer <- function(www_dir) {
  # Start server on port 8081
  port <- 8081
  
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