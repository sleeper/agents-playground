# local-notion

Self-hosted knowledge platform prototype combining free-form pages and structured databases.

## Backend

* Language: Go
* Storage: SQLite (via modernc.org/sqlite for pure Go builds)
* HTTP stack: chi router, zerolog logging

### Running the API server

From the repository root:

```bash
cd local-notion
go run ./cmd/server
```

Environment variables:

* `HTTP_ADDRESS` – HTTP listen address (default `:8080`).
* `DATABASE_DSN` – SQLite DSN (default `file:data/app.db?_fk=1`).

### Testing

From the repository root:

```bash
cd local-notion
go test ./...
```

## REST Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/pages` | List stored pages for quick lookup. |
| `POST` | `/api/pages` | Create a new page. |
| `GET` | `/api/pages/{id}` | Retrieve page details. |
| `POST` | `/api/databases` | Create a database with properties/views. |
| `GET` | `/api/databases/{id}` | Retrieve database metadata. |
| `POST` | `/api/databases/{id}/items` | Create a database item and its page. |
| `GET` | `/api/databases/{id}/views/{viewID}/items` | List items rendered for a view. |
| `GET` | `/api/health` | Health check including DB ping. |
| `GET` | `/api/metrics` | Prometheus-style placeholder metrics. |
| `GET` | `/api/config` | Runtime configuration snapshot. |

Responses follow the envelope structure `{ "data": ..., "errors": [...] }`.

## Frontend

The `web/` directory contains a lightweight React single-page app for interacting with the
platform API. It is intentionally minimal so it can run on constrained devices such as a
Raspberry Pi while still providing quick insight into stored pages and database view items.
The UI now includes forms for creating standalone pages and databases directly from the
browser in addition to the existing read-only explorers. Page creation supports defining
links to other pages, and the explorer reveals both outbound relations and backlinks to
mirror Notion-style navigation.

### Installing dependencies

From the repository root:

```bash
cd local-notion/web
npm install
```

### Developing locally

From the repository root:

```bash
cd local-notion/web
npm run dev
```

The Vite dev server proxies `/api/*` requests to `http://localhost:8080`, so start the Go API
server first.

### Building for production

From the repository root:

```bash
cd local-notion/web
npm run build
```

The compiled assets are emitted to `local-notion/web/dist/` and can be served by the Go backend or a static
file server.
