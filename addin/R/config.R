#' Get the Rishi config directory path
#'
#' Returns the platform-appropriate config directory path for Rishi
#' @return Character path to config directory
getConfigDir <- function() {
  # Determine platform-specific config directory
  if (.Platform$OS.type == "windows") {
    # Windows: use APPDATA or LOCALAPPDATA
    config_base <- Sys.getenv("APPDATA", Sys.getenv("LOCALAPPDATA", Sys.getenv("HOME")))
    config_dir <- file.path(config_base, "rishi")
  } else {
    # Unix/Mac: use XDG_CONFIG_HOME or ~/.config
    config_base <- Sys.getenv("XDG_CONFIG_HOME", file.path(Sys.getenv("HOME"), ".config"))
    config_dir <- file.path(config_base, "rishi")
  }

  return(config_dir)
}

#' Get the Rishi config file path
#'
#' @return Character path to config.R file
getConfigPath <- function() {
  file.path(getConfigDir(), "config.R")
}

#' Load Rishi configuration
#'
#' Loads the configuration from the config file if it exists
#' @return List with configuration values, or NULL if not found
loadConfig <- function() {
  config_path <- getConfigPath()

  if (!file.exists(config_path)) {
    return(NULL)
  }

  tryCatch({
    # Create isolated environment for config
    # Use baseenv() as parent so we have access to assignment operator
    config_env <- new.env(parent = baseenv())
    source(config_path, local = config_env)

    # Return as list
    return(as.list(config_env))
  }, error = function(e) {
    warning(paste("Failed to load config:", e$message))
    return(NULL)
  })
}

#' Save Rishi configuration
#'
#' Saves configuration to the config file
#' @param config List of configuration values
#' @return Logical indicating success
saveConfig <- function(config) {
  config_dir <- getConfigDir()
  config_path <- getConfigPath()

  # Create config directory if it doesn't exist
  if (!dir.exists(config_dir)) {
    dir.create(config_dir, recursive = TRUE, mode = "0700")
  }

  tryCatch({
    # Write config file
    con <- file(config_path, "w")
    on.exit(close(con))

    # Write each config value
    for (name in names(config)) {
      value <- config[[name]]
      if (is.character(value)) {
        # Quote character values
        writeLines(sprintf('%s <- "%s"', name, value), con)
      } else {
        writeLines(sprintf('%s <- %s', name, as.character(value)), con)
      }
    }

    # Set restrictive permissions on Unix systems
    if (.Platform$OS.type == "unix") {
      Sys.chmod(config_path, mode = "0600")
    }

    return(TRUE)
  }, error = function(e) {
    warning(paste("Failed to save config:", e$message))
    return(FALSE)
  })
}

#' Get the API key from configuration
#'
#' Retrieves the ANTHROPIC_API_KEY from the config file
#' @return Character API key, or NULL if not found
getApiKey <- function() {
  config <- loadConfig()

  if (is.null(config)) {
    return(NULL)
  }

  return(config$ANTHROPIC_API_KEY)
}

#' Set the API key in configuration
#'
#' Saves the ANTHROPIC_API_KEY to the config file
#' @param api_key Character API key to save
#' @return Logical indicating success
setApiKey <- function(api_key) {
  # Load existing config or create new one
  config <- loadConfig()
  if (is.null(config)) {
    config <- list()
  }

  # Set API key
  config$ANTHROPIC_API_KEY <- api_key

  # Save config
  return(saveConfig(config))
}