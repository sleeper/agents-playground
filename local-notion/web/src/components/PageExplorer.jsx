import { useCallback, useEffect, useState } from 'react';
import PageDirectory from './PageDirectory.jsx';

export default function PageExplorer({ selectedPageId = '', onSelectPage, refreshKey }) {
  const [pageId, setPageId] = useState(selectedPageId);
  const [page, setPage] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const fetchPage = useCallback(async (targetId) => {
    if (!targetId) {
      setError('Provide a page ID to load data.');
      setPage(null);
      return;
    }
    setLoading(true);
    setError('');
    try {
      const res = await fetch(`/api/pages/${targetId}`);
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.errors?.[0]?.message ?? 'Request failed');
      }
      setPage(json.data);
    } catch (err) {
      setError(err.message);
      setPage(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleQuickLoad = useCallback(
    (id) => {
      setPageId(id);
      onSelectPage?.(id);
      fetchPage(id);
    },
    [fetchPage, onSelectPage]
  );

  async function handleFetch(event) {
    event.preventDefault();
    const trimmed = pageId.trim();
    await fetchPage(trimmed);
    if (trimmed) {
      onSelectPage?.(trimmed);
    }
  }

  useEffect(() => {
    if (selectedPageId && selectedPageId !== pageId) {
      setPageId(selectedPageId);
      fetchPage(selectedPageId);
    }
  }, [selectedPageId, pageId, fetchPage]);

  return (
    <section className="stack page-explorer">
      <article className="card stack">
        <header>
          <h2>Page Explorer</h2>
          <p className="help-text">
            Enter a page ID or pick one from the directory to inspect the stored content, linked
            pages, and backlinks.
          </p>
        </header>
        <form onSubmit={handleFetch} className="panel">
          <label htmlFor="page-id">Page ID</label>
          <input
            id="page-id"
            type="text"
            value={pageId}
            onChange={(event) => setPageId(event.target.value)}
            placeholder="Enter or paste a page ID"
          />
          <button type="submit" disabled={loading}>
            {loading ? 'Loading…' : 'Load Page'}
          </button>
        </form>
        {error && <p className="error">{error}</p>}
        {page ? (
          <article className="card">
            <header>
              <h3>{page.title}</h3>
              <small>Slug: {page.slug}</small>
            </header>
            <p>{page.summary}</p>
            {page.content && (
              <details>
                <summary>Content</summary>
                <pre>{page.content}</pre>
              </details>
            )}
            {Array.isArray(page.linked_page_ids) && page.linked_page_ids.length > 0 && (
              <section className="link-list">
                <h4>Linked pages</h4>
                <ul>
                  {page.linked_page_ids.map((id) => (
                    <li key={id}>
                      <button type="button" className="link-button" onClick={() => handleQuickLoad(id)}>
                        {id}
                      </button>
                    </li>
                  ))}
                </ul>
              </section>
            )}
            {Array.isArray(page.backlinked_page_ids) && page.backlinked_page_ids.length > 0 && (
              <section className="link-list">
                <h4>Backlinks</h4>
                <ul>
                  {page.backlinked_page_ids.map((id) => (
                    <li key={id}>
                      <button type="button" className="link-button" onClick={() => handleQuickLoad(id)}>
                        {id}
                      </button>
                    </li>
                  ))}
                </ul>
              </section>
            )}
            {Array.isArray(page.tags) && page.tags.length > 0 && (
              <footer>
                <strong>Tags:</strong> {page.tags.join(', ')}
              </footer>
            )}
          </article>
        ) : (
          <p className="empty-state">Select a page from the directory or load one by ID to see its details.</p>
        )}
      </article>
      <PageDirectory onSelect={handleQuickLoad} refreshKey={refreshKey} />
    </section>
  );
}
