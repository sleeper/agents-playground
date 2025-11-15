# Potion Revamp Plan

## Goals
- Support markdown-authored rich pages with hierarchical structure and bidirectional links.
- Allow databases with configurable properties and multiple rendered views (table, kanban, gallery).
- Enable embedding databases inside pages with selectable view per embed.
- Provide REST API and React frontend optimized for Raspberry Pi class devices (Go + SQLite backend).

## Steps
1. **Domain Modeling**
   - Introduce normalized schema for spaces, pages, blocks, databases, properties, and entries.
   - Support block types: markdown text, heading, page link, database embed.
   - Maintain backlink index for quick retrieval of page references.
2. **Backend API**
   - Implement CRUD endpoints for pages, blocks, and databases.
   - Add endpoints for managing database properties, entries, and views.
   - Provide markdown rendering assistance (server returns parsed link metadata).
3. **Frontend Application**
   - Build React layout with sidebar navigation, page editor, and database view renderer.
   - Implement markdown editor with live preview, slash command menu, and page linking helper.
   - Render database views (table, kanban by select property, gallery by cover property).
4. **Testing Strategy**
   - Write Go unit tests for repository layer and HTTP handlers.
   - Add React unit tests for state management and view rendering.
   - Provide integration test that creates page, links, database, embed, and validates retrieval.
5. **Performance Considerations**
   - Use efficient SQL queries and indexes.
   - Bundle frontend with code-splitting and local caching of static assets.
6. **Documentation**
   - Document API endpoints and frontend workflows in README.

## Deliverables
- Go backend module with unit tests.
- React frontend with tests and build configuration.
- Documentation summarizing implementation and usage.
