import { useEffect, useMemo, useState } from 'react';
import { marked } from 'marked';

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080';

const emptyPageForm = { title: '', content: '', parentId: '' };
const emptyDatabaseForm = { title: '', description: '', view: 'table' };
const emptyEntryForm = { title: '', properties: '{}' };

export default function App() {
  const [pages, setPages] = useState([]);
  const [pageForm, setPageForm] = useState(emptyPageForm);
  const [selectedPageId, setSelectedPageId] = useState(null);
  const [databases, setDatabases] = useState([]);
  const [databaseForm, setDatabaseForm] = useState(emptyDatabaseForm);
  const [selectedDatabaseId, setSelectedDatabaseId] = useState(null);
  const [entryForm, setEntryForm] = useState(emptyEntryForm);
  const [entries, setEntries] = useState([]);
  const [linkTarget, setLinkTarget] = useState('');
  const [status, setStatus] = useState(null);
  const [error, setError] = useState(null);

  const selectedPage = useMemo(() => pages.find((page) => page.id === selectedPageId) ?? null, [pages, selectedPageId]);
  const selectedDatabase = useMemo(
    () => databases.find((database) => database.id === selectedDatabaseId) ?? null,
    [databases, selectedDatabaseId]
  );

  useEffect(() => {
    refreshPages();
    refreshDatabases();
  }, []);

  useEffect(() => {
    if (selectedDatabaseId) {
      refreshEntries(selectedDatabaseId);
    } else {
      setEntries([]);
    }
  }, [selectedDatabaseId]);

  async function refreshPages() {
    try {
      const response = await fetch(`${API_BASE}/pages`);
      if (!response.ok) throw new Error('Failed to load pages');
      const data = await response.json();
      const pageList = Array.isArray(data) ? data : [];
      setPages(pageList);
      if (pageList.length && !selectedPageId) {
        setSelectedPageId(pageList[0].id);
      }
    } catch (err) {
      setError(err.message);
    }
  }

  async function refreshDatabases() {
    try {
      const response = await fetch(`${API_BASE}/databases`);
      if (!response.ok) throw new Error('Failed to load databases');
      const data = await response.json();
      const databaseList = Array.isArray(data) ? data : [];
      setDatabases(databaseList);
      if (databaseList.length && !selectedDatabaseId) {
        setSelectedDatabaseId(databaseList[0].id);
      }
    } catch (err) {
      setError(err.message);
    }
  }

  async function refreshEntries(databaseId) {
    try {
      const response = await fetch(`${API_BASE}/databases/${databaseId}/entries`);
      if (!response.ok) throw new Error('Failed to load database entries');
      const data = await response.json();
      const entryList = Array.isArray(data) ? data : [];
      setEntries(entryList);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCreatePage(event) {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        title: pageForm.title.trim(),
        content: pageForm.content,
        parentId: pageForm.parentId ? pageForm.parentId : null,
      };
      const response = await fetch(`${API_BASE}/pages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error('Unable to create page');
      setPageForm(emptyPageForm);
      setStatus('Page created');
      await refreshPages();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCreateDatabase(event) {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        title: databaseForm.title.trim(),
        description: databaseForm.description,
        view: databaseForm.view,
        schema: {},
      };
      const response = await fetch(`${API_BASE}/databases`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error('Unable to create database');
      setDatabaseForm(emptyDatabaseForm);
      setStatus('Database created');
      await refreshDatabases();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCreateEntry(event) {
    event.preventDefault();
    if (!selectedDatabaseId) return;
    setError(null);
    try {
      const payload = {
        title: entryForm.title.trim(),
        properties: JSON.parse(entryForm.properties || '{}'),
      };
      const response = await fetch(`${API_BASE}/databases/${selectedDatabaseId}/entries`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error('Unable to create entry');
      setEntryForm(emptyEntryForm);
      setStatus('Entry created');
      await refreshEntries(selectedDatabaseId);
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError('Properties must be valid JSON');
      } else {
        setError(err.message);
      }
    }
  }

  async function handleCreateLink(event) {
    event.preventDefault();
    if (!selectedPageId || !linkTarget) return;
    setError(null);
    try {
      const response = await fetch(`${API_BASE}/links`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sourceId: selectedPageId, targetId: linkTarget }),
      });
      if (!response.ok) throw new Error('Unable to create link');
      setLinkTarget('');
      setStatus('Link created');
    } catch (err) {
      setError(err.message);
    }
  }

  function renderMarkdown(content) {
    return { __html: marked.parse(content || '') };
  }

  return (
    <div className="app">
      <header className="top-bar">
        <h1>Potion</h1>
        <p className="tagline">A lightweight Notion-style workspace for Raspberry Pi</p>
      </header>

      {status && <div className="status">{status}</div>}
      {error && <div className="error">{error}</div>}

      <main className="workspace">
        <section className="sidebar">
          <h2>Pages</h2>
          <ul className="page-list">
            {pages.map((page) => (
              <li
                key={page.id}
                className={selectedPageId === page.id ? 'active' : ''}
                onClick={() => setSelectedPageId(page.id)}
              >
                {page.title || 'Untitled'}
              </li>
            ))}
          </ul>

          <form className="panel" onSubmit={handleCreatePage}>
            <h3>New Page</h3>
            <label>
              Title
              <input
                value={pageForm.title}
                onChange={(event) => setPageForm({ ...pageForm, title: event.target.value })}
                placeholder="Page title"
                required
              />
            </label>
            <label>
              Parent page
              <select
                value={pageForm.parentId}
                onChange={(event) => setPageForm({ ...pageForm, parentId: event.target.value })}
              >
                <option value="">None</option>
                {pages.map((page) => (
                  <option key={page.id} value={page.id}>
                    {page.title}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Markdown content
              <textarea
                value={pageForm.content}
                onChange={(event) => setPageForm({ ...pageForm, content: event.target.value })}
                rows={6}
                placeholder="Write in markdown..."
              />
            </label>
            <button type="submit">Create page</button>
          </form>

          <form className="panel" onSubmit={handleCreateLink}>
            <h3>Link page</h3>
            <label>
              Target page
              <select value={linkTarget} onChange={(event) => setLinkTarget(event.target.value)}>
                <option value="">Select a page</option>
                {pages
                  .filter((page) => page.id !== selectedPageId)
                  .map((page) => (
                    <option key={page.id} value={page.id}>
                      {page.title}
                    </option>
                  ))}
              </select>
            </label>
            <button type="submit" disabled={!selectedPageId || !linkTarget}>
              Create link
            </button>
          </form>
        </section>

        <section className="content">
          {selectedPage ? (
            <article>
              <h2>{selectedPage.title || 'Untitled'}</h2>
              <div className="markdown" dangerouslySetInnerHTML={renderMarkdown(selectedPage.content)} />
            </article>
          ) : (
            <p className="placeholder">Select or create a page to start writing.</p>
          )}

          <div className="databases">
            <div className="database-header">
              <h2>Databases</h2>
              <div className="database-tabs">
                {databases.map((database) => (
                  <button
                    key={database.id}
                    className={selectedDatabaseId === database.id ? 'active' : ''}
                    onClick={() => setSelectedDatabaseId(database.id)}
                    type="button"
                  >
                    {database.title || 'Untitled database'}
                  </button>
                ))}
              </div>
            </div>

            {selectedDatabase ? (
              <div className="database-content">
                <h3>{selectedDatabase.title}</h3>
                <p className="description">{selectedDatabase.description}</p>
                <div className="entries">
                  {entries.length ? (
                    <table>
                      <thead>
                        <tr>
                          <th>Title</th>
                          <th>Properties</th>
                        </tr>
                      </thead>
                      <tbody>
                        {entries.map((entry) => (
                          <tr key={entry.id}>
                            <td>{entry.title}</td>
                            <td>
                              <pre>{JSON.stringify(entry.properties, null, 2)}</pre>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <p className="placeholder">No entries yet.</p>
                  )}
                </div>

                <form className="panel" onSubmit={handleCreateEntry}>
                  <h4>New entry</h4>
                  <label>
                    Title
                    <input
                      value={entryForm.title}
                      onChange={(event) => setEntryForm({ ...entryForm, title: event.target.value })}
                      placeholder="Entry title"
                      required
                    />
                  </label>
                  <label>
                    Properties (JSON)
                    <textarea
                      value={entryForm.properties}
                      onChange={(event) => setEntryForm({ ...entryForm, properties: event.target.value })}
                      rows={4}
                      placeholder='{"status": "Todo"}'
                    />
                  </label>
                  <button type="submit" disabled={!selectedDatabaseId}>
                    Create entry
                  </button>
                </form>
              </div>
            ) : (
              <p className="placeholder">Create a database to get started.</p>
            )}
          </div>
        </section>

        <aside className="sidebar">
          <form className="panel" onSubmit={handleCreateDatabase}>
            <h3>New database</h3>
            <label>
              Title
              <input
                value={databaseForm.title}
                onChange={(event) => setDatabaseForm({ ...databaseForm, title: event.target.value })}
                placeholder="Database name"
                required
              />
            </label>
            <label>
              Description
              <textarea
                value={databaseForm.description}
                onChange={(event) => setDatabaseForm({ ...databaseForm, description: event.target.value })}
                rows={4}
                placeholder="Describe the database"
              />
            </label>
            <label>
              Default view
              <select
                value={databaseForm.view}
                onChange={(event) => setDatabaseForm({ ...databaseForm, view: event.target.value })}
              >
                <option value="table">Table</option>
                <option value="kanban">Kanban</option>
                <option value="gallery">Gallery</option>
              </select>
            </label>
            <button type="submit">Create database</button>
          </form>
        </aside>
      </main>
    </div>
  );
}
