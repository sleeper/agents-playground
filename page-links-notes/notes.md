# Notes

## 2024-05-06
- Initialized notes for adding page linking support to local-notion.
- Added SQLite migration for storing page-to-page link relationships.
- Extended domain.Page to surface outbound and inbound linked page identifiers.
- Extended SQLite store to persist and load linked page identifiers; added helper for listing pages.
- Documented the new `/api/pages` listing endpoint and UI link exploration capabilities.
- Verified backend and handler updates with `go test ./...` and rebuilt the frontend via `npm run build`.
- Introduced a page directory widget to fetch IDs from `/api/pages` and wired supporting styles.
