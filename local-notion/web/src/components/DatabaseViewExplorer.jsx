import { useState } from 'react';

export default function DatabaseViewExplorer() {
  const [databaseId, setDatabaseId] = useState('');
  const [viewId, setViewId] = useState('');
  const [items, setItems] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleFetch(event) {
    event.preventDefault();
    if (!databaseId || !viewId) {
      setError('Both database ID and view ID are required.');
      setItems([]);
      return;
    }

    setLoading(true);
    setError('');
    try {
      const res = await fetch(`/api/databases/${databaseId}/views/${viewId}/items`);
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.errors?.[0]?.message ?? 'Request failed');
      }
      setItems(Array.isArray(json.data) ? json.data : []);
    } catch (err) {
      setError(err.message);
      setItems([]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h2>Database View Explorer</h2>
      <form onSubmit={handleFetch} className="panel">
        <label htmlFor="database-id">Database ID</label>
        <input
          id="database-id"
          type="text"
          value={databaseId}
          onChange={(event) => setDatabaseId(event.target.value)}
          placeholder="Enter a database ID"
        />
        <label htmlFor="view-id">View ID</label>
        <input
          id="view-id"
          type="text"
          value={viewId}
          onChange={(event) => setViewId(event.target.value)}
          placeholder="Enter a view ID"
        />
        <button type="submit" disabled={loading}>
          {loading ? 'Loading…' : 'Load Items'}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {items.length > 0 && (
        <ul className="list">
          {items.map((item) => (
            <li key={item.id}>
              <strong>{item.page?.title ?? 'Untitled'}</strong>
              {item.page?.summary && <p>{item.page.summary}</p>}
              {item.properties && (
                <details>
                  <summary>Properties</summary>
                  <pre>{JSON.stringify(item.properties, null, 2)}</pre>
                </details>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
