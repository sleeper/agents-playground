# Potion Revamp

This experiment rebuilds the Potion workspace to more closely mirror Notion's core experience while remaining deployable on a Raspberry Pi class device. It provides a Go + SQLite backend and a React frontend that support hierarchical pages, block authoring with Markdown, bi-directional page links, rich databases, and embedded database views.

## Research Notes
- Reviewed Notion's public help guides to capture their block model, database customization, and multi-view experience.
- Identified high-impact capabilities to ship first: page linking, markdown blocks, configurable databases with kanban/table/gallery views, and inline database embeds.
- Prioritized performance-aware choices (single SQLite file, precomputed backlink index, lightweight React state) to keep latency low on constrained hardware.

## Backend Overview
- **Language / Storage:** Go 1.24 with `modernc.org/sqlite` for a self-contained SQLite driver.
- **HTTP Stack:** `chi` router with permissive CORS middleware.
- **Domain Model:**
  - `pages` with hierarchical parent references and associated `blocks` (markdown, heading, page link, database embed).
  - `databases` with typed `properties`, saved `views`, and per-entry property values.
  - `page_links` table maintains backlinks automatically when blocks change.
- **Key Endpoints:**
  - `GET/POST /pages`, `GET/PUT/DELETE /pages/{id}`, `PUT /pages/{id}/blocks` for block replacement.
  - `GET/POST /databases`, `GET /databases/{id}`, `POST /databases/{id}/entries`, `PUT /databases/{id}/entries/{entryID}`.
  - `GET /databases/{id}/views/{viewID}` resolves a database view plus entries.
- **Testing:** `internal/storage` tests validate backlink indexing and database lifecycle behavior.

## Frontend Overview
- **Stack:** React 18 + Vite + Vitest.
- **Features:**
  - Sidebar navigation listing all pages and databases.
  - Page editor with toolbar to add markdown, heading, page link, and database embed blocks.
  - Markdown blocks render live preview while editing.
  - Page links update backlinks card for quick navigation.
  - Database embeds fetch live view data and render table/kanban/gallery previews.
  - Database panel allows creating databases (with select/multi-select options), toggling views, and adding entries with property-aware inputs.
- **Utilities:** Shared helpers normalize view options, format property values, and sanitize block payloads.
- **Testing:** Vitest suite covers block preparation, database grouping, and property payload parsing.

## Running Locally
1. **Backend**
   ```bash
   cd backend
   go run ./cmd/server -addr :8080 -db potion.db
   ```
2. **Frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
3. Visit `http://localhost:5173` to interact with the workspace.

## Testing
- Backend: `go test ./...`
- Frontend: `npm test`

## Repository Layout
```
potion-revamp/
├── README.md
├── notes.md
├── plan.md
├── backend/
│   ├── cmd/server/main.go
│   ├── go.mod
│   ├── go.sum
│   └── internal/
│       ├── models/
│       ├── server/
│       └── storage/
└── frontend/
    ├── index.html
    ├── package.json
    ├── src/
    └── vite.config.js
```

## Next Steps
- Enhance block editing with drag-and-drop reordering and inline slash commands.
- Add granular API filtering (e.g., query databases by property filters).
- Persist per-embed view configuration overrides to mirror Notion's linked database feature.
