# Potion Experiment

Potion is an experimental, Raspberry Pi-friendly workspace inspired by Notion. It combines a Go + SQLite backend with a React + Vite frontend to provide a lightweight way to author markdown pages, interlink them, and manage simple databases.

## Project layout

```
potion/
├── backend/      # Go HTTP API for pages, links, and databases
├── frontend/     # React/Vite single-page app consuming the API
├── notes.md      # Running log of investigation steps
└── README.md     # This report
```

## Backend

- **Language:** Go 1.24+
- **Frameworks:** chi router, modernc SQLite driver
- **Persistence:** SQLite, schema created automatically on first run
- **Key features:**
  - CRUD endpoints for markdown pages
  - Page-to-page linking
  - Database definitions with JSON schemas
  - Database entry management
  - Health endpoint for monitoring

### Running the server

```
cd backend
go run ./cmd/server --addr :8080 --db ./potion.db
```

Configuration flags:

- `--addr` — listening address (defaults to `:8080`)
- `--db` — SQLite database path (defaults to `./potion.db`)

The API automatically creates tables for pages, links, databases, and database entries. All timestamps are stored in UTC.

### API overview

| Method | Path                           | Description                               |
| ------ | ------------------------------ | ----------------------------------------- |
| GET    | `/healthz`                     | Health check                              |
| GET    | `/pages`                       | List pages                                |
| POST   | `/pages`                       | Create a page                             |
| GET    | `/pages/{id}`                  | Fetch a page                              |
| PUT    | `/pages/{id}`                  | Update a page                             |
| GET    | `/pages/{id}/links`            | List outgoing links                       |
| POST   | `/links`                       | Create a page-to-page link                |
| GET    | `/databases`                   | List databases                            |
| POST   | `/databases`                   | Create a database                         |
| GET    | `/databases/{id}`              | Fetch a database                          |
| GET    | `/databases/{id}/entries`      | List entries in a database                |
| POST   | `/databases/{id}/entries`      | Create an entry for a database            |

All request and response payloads are JSON encoded. Page content is stored in markdown and can be rendered on the frontend.

## Frontend

- **Framework:** React 18 with Vite bundler
- **Styling:** CSS modules in `src/styles.css`
- **Features:**
  - Page list with markdown preview using `marked`
  - Page creation form with optional parent selection
  - Link creation between pages
  - Database list with simple table view of entries
  - Forms for creating databases and entries (JSON properties)

### Running the UI

```
cd frontend
npm install
npm run dev
```

By default the app expects the backend at `http://localhost:8080`. Set `VITE_API_BASE` to point elsewhere in development, or rely on Vite's proxy configuration to forward `/pages`, `/databases`, and `/links` to another host.

### Building for production

```
npm run build
```

Bundles are emitted into `frontend/dist`. Serve the static files with your preferred HTTP server and configure it to reverse-proxy API calls to the Go service.

## Next steps

- Implement authentication for multi-user environments
- Add richer database schemas and views (kanban, gallery rendering)
- Support page version history and collaborative editing
- Package backend + frontend into a single deployment artifact for the Raspberry Pi
