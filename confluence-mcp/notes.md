## Work log
- Created `confluence-mcp` workspace folder per root instructions.
- Initialized Go module `confluence-mcp`.
- Added configuration loader, JWT authenticator, policy store, Confluence client, and HTTP handlers scaffolding.
- Wired main server entry point with graceful shutdown and policy-aware routing.
- Added example YAML config and README describing architecture and usage.
- Ran `go test ./...` (initial run compiled toolchain; tests pass).
- Captured current repo diff into `diff.patch` per instructions.
- Regenerated `diff.patch` after staging to capture the full change set.
