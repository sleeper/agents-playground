# Confluence MCP Gateway (Go)

This folder contains a minimal Model Context Protocol (MCP) server that proxies Confluence Cloud for trusted AI agents. It enforces per-subject space scoping so downstream tools can only retrieve content their identity is authorized to see.

## Highlights
- **Zero-trust friendly**: Requests must carry a signed JWT (e.g., issued by Microsoft Entra or Purview) and are filtered by per-subject policies.
- **Space-aware queries**: Search and page fetches are constrained to allowed Confluence spaces, preventing cross-tenant data drift.
- **Cloud-native HTTP server**: Small, dependency-light Go service with graceful shutdown and observability-friendly middleware.

## Architecture
- `cmd/server`: boots the HTTP server, loads YAML configuration, and wires dependencies.
- `internal/config`: parses YAML config for server, Confluence, and access control settings.
- `internal/auth`: validates bearer JWTs using a shared HMAC secret (swap with your Entra signing keys in production).
- `internal/policy`: merges default and subject-specific allowed space lists.
- `internal/confluence`: thin client for Confluence Cloud REST APIs (search + page retrieval with storage body).
- `internal/handlers`: chi-based HTTP router exposing `/spaces`, `/search`, and `/pages/{id}` endpoints gated by policy.

## Configuration
Create a config file (see `config.example.yaml`) and point the server to it with `-config`.

```yaml
server:
  addr: ":8080"

confluence:
  base_url: "https://your-domain.atlassian.net"
  auth_token: "${CONFLUENCE_OAUTH_TOKEN}" # OAuth or PAT injected via env

access_control:
  shared_secret: "super-secret-hmac-key"   # replace with your Entra/Purview signing secret
  default_spaces: [ENG]
  policies:
    - subject: "team-analyst"
      allowed_spaces: [KM, DS]
    - subject: "agent-support"
      allowed_spaces: [SUPPORT]
```

- **Subjects** come from the JWT `sub` claim. Align this with your SSO identity (user object ID, app ID, or group ID).
- The policy store merges `default_spaces` with any subject-specific entries, deduping and sorting for deterministic behavior.

## Running
```bash
go run ./cmd/server -config ./config.example.yaml
```

Requests must include `Authorization: Bearer <jwt>` where the JWT is signed with the configured shared secret. Swap `auth.Authenticator` to verify Entra-issued tokens or JWKS for production; the handler contract already surfaces the subject for policy checks.

### Example calls
```bash
# List spaces the caller may access
curl -H "Authorization: Bearer $JWT" http://localhost:8080/spaces

# Search within allowed spaces
curl -G -H "Authorization: Bearer $JWT" --data-urlencode "q=zero trust" http://localhost:8080/search

# Fetch a specific page (rejected if space not in policy)
curl -H "Authorization: Bearer $JWT" http://localhost:8080/pages/123456
```

## Extending
- Replace the HMAC validator with an Entra JWKS-backed validator to honor your Microsoft Purview SSO integration.
- Persist policy entries in your governance source (e.g., Purview collections) and refresh the in-memory store on a schedule.
- Add caching for Confluence responses to reduce latency and API usage under heavy concurrent loads.
- Wire Prometheus or OpenTelemetry middleware for request tracing and rate/latency visibility.

## Testing
Run unit tests and code formatting checks via:
```bash
go test ./...
```
