#' Constructor functions for structured data types

#' Create a tool result response
#' 
#' @param content Character string with result content
#' @param error Character string with error message (empty if no error)
#' @return List with content and error fields
text_editor_view_tool_result <- function(content = "", error = "") {
  stopifnot(is.character(content), is.character(error))
  list(content = content, error = error)
}

#' Create a str_replace tool result response
#' 
#' @param content Character string with result content
#' @param error Character string with error message (empty if no error)
#' @return List with content and error fields
text_editor_str_replace_tool_result <- function(content = "", error = "") {
  stopifnot(is.character(content), is.character(error))
  list(content = content, error = error)
}

#' Create a create tool result response
#' 
#' @param content Character string with result content
#' @param error Character string with error message (empty if no error)
#' @return List with content and error fields
text_editor_create_tool_result <- function(content = "", error = "") {
  stopifnot(is.character(content), is.character(error))
  list(content = content, error = error)
}

#' Create a insert tool result response
#' 
#' @param content Character string with result content
#' @param error Character string with error message (empty if no error)
#' @return List with content and error fields
text_editor_insert_tool_result <- function(content = "", error = "") {
  stopifnot(is.character(content), is.character(error))
  list(content = content, error = error)
}