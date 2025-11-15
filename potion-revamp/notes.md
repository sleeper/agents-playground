# Notes

## Day 1
- Created workspace for Potion revamp study.
- Reviewed Notion help guide on components to understand blocks, databases, views, and linking behaviors.
- Drafted detailed plan for backend, frontend, and testing strategy in plan.md.
- Implemented Go backend storage layer with schema to support pages, blocks, databases, and views.
- Added HTTP server with chi routing to expose CRUD endpoints for pages and databases.
- Wrote storage unit tests covering page backlinks and database lifecycle; resolved SQLite single-connection deadlock by deferring property fetch until after row iteration.
- Verified Go unit tests succeed with `go test ./...`.
- Crafted React frontend with block-level editing, page linking, database embedding, and database management panel.
- Added shared utility helpers and Vitest unit tests for client-side behaviors.
