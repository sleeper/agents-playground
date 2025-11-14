import { useCallback, useEffect, useState } from 'react';

export default function PageDirectory({ onSelect, refreshKey = 0 }) {
  const [pages, setPages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const loadPages = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await fetch('/api/pages');
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.errors?.[0]?.message ?? 'Failed to load pages');
      }
      const data = Array.isArray(json.data) ? json.data : [];
      setPages(data);
    } catch (err) {
      setError(err.message);
      setPages([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPages();
  }, [loadPages]);

  useEffect(() => {
    if (refreshKey > 0) {
      loadPages();
    }
  }, [refreshKey, loadPages]);

  return (
    <section className="panel card directory-panel">
      <header className="directory-header">
        <div>
          <h3>Page Directory</h3>
          <p>Browse stored pages and copy their IDs when creating new links.</p>
        </div>
        <button type="button" onClick={loadPages} disabled={loading}>
          {loading ? 'Loading…' : 'Refresh'}
        </button>
      </header>
      {error && <p className="error">{error}</p>}
      {!error && pages.length === 0 && !loading && <p>No pages loaded yet.</p>}
      {pages.length > 0 && (
        <ul className="directory-list">
          {pages.map((page) => (
            <li key={page.id}>
              <div className="directory-info">
                <strong>{page.title || 'Untitled page'}</strong>
                <small>{page.slug || '—'}</small>
              </div>
              <div className="directory-actions">
                <button type="button" onClick={() => onSelect?.(page.id)}>
                  Load
                </button>
                <code>{page.id}</code>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
