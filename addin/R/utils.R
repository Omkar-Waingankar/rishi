#' Utility functions shared across the rishiai package

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