# Docinator

## Project Overview
Docinator is a CLI tool for scraping and storing documentation from websites. It uses Colly for scraping, Viper for configuration, and BoltDB for storage.

## Getting Started

### Installation
1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Run `go run cmd/docinator/main.go` to run the app

## Usage
docinator [flags]
 
### Commands
- Run `docinator help` for usage information

## Project Structure
- cmd/docinator: CLI entry point
- pkg/scraper: Web scraping logic using Colly
- pkg/parser: Document parsing
- pkg/config: Configuration management with Viper
- pkg/storage: Data storage using BoltDB
- internal/models: Internal data models
- internal/utils: Utility functions
- templates: Template files for output

## MongoDB Integration (Caching/Persistence)

Docinator can optionally cache and persist scraped package docs in MongoDB. On each run:
- For every import path argument (e.g., `github.com/spf13/cobra`), Docinator first checks MongoDB for a stored document using the import path as `_id`.
- If found, it loads from the database and skips scraping.
- If not found, it scrapes and then upserts the document back into MongoDB for future runs.

### Environment Variables
- `MONGODB_URI` (required to enable): Connection string to MongoDB. If not set, MongoDB is disabled and Docinator operates as before (pure scrape-to-output/files).
- `MONGODB_DB` (optional, default: `docinator`): Database name.
- `MONGODB_COLLECTION` (optional, default: `packages`): Collection name.

### Example
```
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DB="docinator"
export MONGODB_COLLECTION="packages"

# Scrape or load from cache for cobra and viper:
docinator scrape github.com/spf13/cobra github.com/spf13/viper -o ./test_output
```

If `MONGODB_URI` is unset or invalid, Docinator logs a message and continues without DB usage.

### Run MongoDB locally (Docker)
```
docker run --name mongo -p 27017:27017 -d mongo:7
export MONGODB_URI="mongodb://localhost:27017"
```

## Development

### Building
go build cmd/docinator/main.go

### Running Tests
go test ./...

## Markdown Output Format

The Markdown renderer converts scraped Go package documentation into a structured Markdown format suitable for LLM consumption and MCP server integration. The output includes:

- **Package Header**: Name, description, synopsis, module, import path, license, and repository information.
- **Functions Section**: Lists all functions with their signatures, descriptions, and example code blocks (```go ... ```) with outputs.
- **Types Section**: Lists all types with their kind, definition, description, methods (if any), and examples.
- **Variables and Constants**: Listed with their types and descriptions.

Example output structure for a package like cobra would include sections for Functions and Types, with code examples wrapped in language-specific fences for syntax highlighting.

## Contributing

Contributions are welcome! Please follow the conventional commit standard.

---
title: README.md
task_id: GDOCS-003
date: 2025-09-08
last_updated: 2025-09-08
status: DRAFT
owner: Builder
---

## Objective
Provide project overview and setup instructions.

## Inputs
- None

## Process
- Basic CLI scaffolding set up.

## Outputs
- Project ready for implementation.

## Dependencies
- Go 1.24.3
- Colly, Cobra, Viper, BoltDB

## Next Actions
- Implement scraping logic
- Implement parsing logic
- Implement storage logic