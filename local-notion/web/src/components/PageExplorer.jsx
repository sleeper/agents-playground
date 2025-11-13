import { useState } from 'react';

export default function PageExplorer() {
  const [pageId, setPageId] = useState('');
  const [page, setPage] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleFetch(event) {
    event.preventDefault();
    if (!pageId) {
      setError('Provide a page ID to load data.');
      setPage(null);
      return;
    }

    setLoading(true);
    setError('');
    try {
      const res = await fetch(`/api/pages/${pageId}`);
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
  }

  return (
    <section>
      <h2>Page Explorer</h2>
      <form onSubmit={handleFetch} className="panel">
        <label htmlFor="page-id">Page ID</label>
        <input
          id="page-id"
          type="text"
          value={pageId}
          onChange={(event) => setPageId(event.target.value)}
          placeholder="Enter a page ID"
        />
        <button type="submit" disabled={loading}>
          {loading ? 'Loading…' : 'Load Page'}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {page && (
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
          {Array.isArray(page.tags) && page.tags.length > 0 && (
            <footer>
              <strong>Tags:</strong> {page.tags.join(', ')}
            </footer>
          )}
        </article>
      )}
    </section>
  );
}
