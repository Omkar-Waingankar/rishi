#' Get active tab/document context
#' 
#' Retrieves the current active document's filename and contents using rstudioapi
#' 
#' @get /context/active_tab
get_active_tab_context <- function() {
  tryCatch({
    # Get the active document context
    doc_context <- rstudioapi::getActiveDocumentContext()
    
    # Check if we have a valid document
    if (is.null(doc_context) || is.null(doc_context$path) || doc_context$path == "") {
      return(list(
        error = jsonlite::unbox("No active document found")
      ))
    }
    
    # Extract filename from path
    filename <- basename(doc_context$path)
    
    # Get document contents
    contents <- if (is.null(doc_context$contents)) "" else paste(doc_context$contents, collapse = "\n")
    
    return(list(
      filename = jsonlite::unbox(filename),
      content = jsonlite::unbox(contents)
    ))
    
  }, error = function(e) {
    return(list(
      error = jsonlite::unbox(paste("Failed to get active document:", e$message))
    ))
  })
}

#' Get plot context
#'
#' Captures the current plot as a base64-encoded image
#'
#' @get /context/plot
get_plot_context <- function() {  
  tryCatch({
    # Create a temporary file for the plot
    temp_file <- tempfile(fileext = ".png")

    # Try to save the current plot (whatever is displayed in the plot pane)
    rstudioapi::savePlotAsImage(
      file = temp_file,
      format = "png",
      width = 800,
      height = 600
    )

    # Check if the file was actually created and has meaningful content
    if (!file.exists(temp_file)) {
      return(list(
        error = jsonlite::unbox("No plot found - savePlotAsImage failed")
      ))
    }

    file_size <- file.info(temp_file)$size
    if (file_size == 0 || is.na(file_size) || file_size < 1000) { # Very small files are likely empty
      unlink(temp_file)
      return(list(
        error = jsonlite::unbox("Plot pane appears empty or contains no meaningful plot")
      ))
    }

    # Read the plot data
    plot_data <- readBin(temp_file, "raw", file_size)

    # Convert to base64
    plot_base64 <- base64enc::base64encode(plot_data)

    # Clean up temp file
    unlink(temp_file)

    return(list(
      imageBase64 = jsonlite::unbox(plot_base64),
      mediaType = jsonlite::unbox("image/png")
    ))

  }, error = function(e) {
    return(list(
      error = jsonlite::unbox(paste("Failed to get plot:", e$message))
    ))
  })
}
