# Rishi

Rishi is an AI coding assistant for R, built as a RStudio add-in. It provides an agentic chat interface powered by Claude to help you write, debug, and understand R code.

## Installation

### Install from GitHub (Recommended)

```r
# Install remotes if you don't have it
install.packages("remotes")

# Install Rishi
remotes::install_github("Omkar-Waingankar/rishi", subdir = "addin")
```

### Set up your API key

Add your Anthropic API key to your `.Renviron` file:

```r
# Edit your .Renviron file
usethis::edit_r_environ()

# Add this line (replace with your actual key):
ANTHROPIC_API_KEY=your-api-key-here
```

Restart R after editing `.Renviron`.

### Launch Rishi

```r
rishi::rishiAddin()
```

Or access it from the RStudio Addins menu.

---

## Development Setup

For contributors and developers:

1. Clone this repository.
2. Set up your Anthropic API key as described above.
3. Uninstall any existing version:
   ```bash
   make uninstall-local
   ```
4. Build and install for local development:
   ```bash
   make up
   ```
5. In RStudio, run:
   ```r
   setwd("~/your-project-directory")
   rishi::rishiAddin()
   ```

---

## Creating a New Release

For maintainers publishing a new version:

1. **Update the version** in `addin/DESCRIPTION`:
   ```
   Version: 1.0.4
   ```

2. **Build everything** (frontend + daemon binaries for all platforms):
   ```bash
   make package-all
   ```

3. **Commit and push the built assets**:
   ```bash
   git add addin/DESCRIPTION addin/inst/www/ addin/inst/bin/
   git commit -m "Release v1.0.4"
   git push
   ```

4. **Create and publish the GitHub release**:
   ```bash
   make release
   ```

This will:
- Create a git tag (e.g., `v1.0.4`)
- Push the tag to GitHub
- Create a GitHub release with standalone daemon binaries attached

**Requirements**: You need the [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated.

---

## Repository Structure

- **addin/**  
  Contains the Rishi add-in source code, including the React frontend and R integration.

- **daemon/**
  Contains an HTTP backend server that provides AI chat functionality via Anthropic's Claude API. The backend is automatically launched by the add-in.

---
