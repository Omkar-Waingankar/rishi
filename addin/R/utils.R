#' Utility functions shared across the rishi package

# Package-level environment to cache working directory state
.wd_state <- new.env(parent = emptyenv())
.wd_state$last_known_wd <- NULL

#' Get config directory path
#' @return Character string of config directory path
get_config_dir <- function() {
  if (.Platform$OS.type == "windows") {
    config_base <- Sys.getenv("APPDATA")
    if (config_base == "") {
      config_base <- Sys.getenv("LOCALAPPDATA")
    }
    if (config_base == "") {
      config_base <- Sys.getenv("HOME")
    }
    return(file.path(config_base, "rishi"))
  } else {
    config_base <- Sys.getenv("XDG_CONFIG_HOME")
    if (config_base == "") {
      config_base <- file.path(Sys.getenv("HOME"), ".config")
    }
    return(file.path(config_base, "rishi"))
  }
}

#' Get config file path
#' @return Character string of config.json path
get_config_path <- function() {
  file.path(get_config_dir(), "config.json")
}

#' Load working directory from config
#' @return Character string of stored working directory or NULL
load_working_directory_config <- function() {
  config_path <- get_config_path()

  if (!file.exists(config_path)) {
    return(NULL)
  }

  tryCatch({
    config_data <- jsonlite::fromJSON(config_path, simplifyVector = FALSE)
    if (!is.null(config_data$last_working_directory)) {
      return(config_data$last_working_directory)
    }
    return(NULL)
  }, error = function(e) {
    return(NULL)
  })
}

#' Save working directory to config
#' @param path Character string of working directory path to save
save_working_directory_config <- function(path) {
  config_dir <- get_config_dir()
  config_path <- get_config_path()

  # Create config directory if it doesn't exist
  if (!dir.exists(config_dir)) {
    tryCatch({
      dir.create(config_dir, recursive = TRUE, mode = "0700")
    }, error = function(e) {
      warning(paste("Failed to create config directory:", e$message))
      return(invisible(FALSE))
    })
  }

  # Load existing config or create new one
  config_data <- list()
  if (file.exists(config_path)) {
    tryCatch({
      config_data <- jsonlite::fromJSON(config_path, simplifyVector = FALSE)
    }, error = function(e) {
      # If config is corrupted, start fresh
      config_data <- list()
    })
  }

  # Update working directory
  config_data$last_working_directory <- path

  # Write atomically using temp file
  tryCatch({
    temp_file <- tempfile(pattern = "config.json.", tmpdir = config_dir, fileext = ".tmp")
    writeLines(jsonlite::toJSON(config_data, pretty = TRUE, auto_unbox = TRUE), temp_file)

    # Move temp file to final location (atomic on most systems)
    file.rename(temp_file, config_path)

    return(invisible(TRUE))
  }, error = function(e) {
    warning(paste("Failed to save working directory config:", e$message))
    return(invisible(FALSE))
  })
}

#' Compute safe root directory for file operations
#'
#' @param is_startup Logical indicating if this is being called at startup
#' @return List with 'path' (character string of safe root directory or refusal message) and 'source' (character string: "rproj", "setwd", or "stored")
compute_safe_root <- function(is_startup = FALSE) {
  # Try to get project root from RStudio API
  project_root <- tryCatch({
    rstudioapi::getActiveProject()
  }, error = function(e) {
    NULL
  })

  source <- "rproj"

  # Fallback to current working directory if no project
  if (is.null(project_root)) {
    project_root <- getwd()
    source <- "setwd"
  }

  # Normalize the path
  project_root <- normalizePath(project_root, winslash = "/", mustWork = FALSE)

  # Check if path resolves to home directory or system root
  home_dir <- normalizePath("~", winslash = "/", mustWork = FALSE)

  # Check for system root patterns
  is_root <- grepl("^/$", project_root) ||  # Unix root
             grepl("^[A-Za-z]:/$", project_root) ||  # Windows root (C:/, D:/, etc.)
             identical(project_root, home_dir)  # Home directory

  # If we're at root/home, try to load from config
  if (is_root) {
    stored_wd <- load_working_directory_config()
    if (!is.null(stored_wd) && dir.exists(stored_wd)) {
      # Only auto-setwd on startup
      if (is_startup) {
        tryCatch({
          setwd(stored_wd)
        }, error = function(e) {
          # Failed to setwd, continue with error state
        })
      }
      project_root <- stored_wd
      source <- "stored"
      is_root <- FALSE
    }
  }

  if (is_root) {
    return(list(
      path = "You are not allowed to list files from root",
      source = "none"
    ))
  }

  # Save valid working directory to config
  save_working_directory_config(project_root)

  # Update cached state
  .wd_state$last_known_wd <- project_root

  return(list(
    path = project_root,
    source = source
  ))
}