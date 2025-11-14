import { useCallback, useEffect, useState } from 'react';
import PageCreator from './components/PageCreator.jsx';
import PageExplorer from './components/PageExplorer.jsx';
import DatabaseCreator from './components/DatabaseCreator.jsx';
import DatabaseViewExplorer from './components/DatabaseViewExplorer.jsx';
import './App.css';

export default function App() {
  const [health, setHealth] = useState(null);
  const [config, setConfig] = useState(null);
  const [healthError, setHealthError] = useState('');
  const [selectedPageId, setSelectedPageId] = useState('');
  const [pageRefreshKey, setPageRefreshKey] = useState(0);

  useEffect(() => {
    async function loadHealth() {
      try {
        const res = await fetch('/api/health');
        if (!res.ok) {
          throw new Error('Health check failed');
        }
        const json = await res.json();
        setHealth(json.data ?? json);
      } catch (err) {
        setHealthError(err.message);
      }
    }

    async function loadConfig() {
      try {
        const res = await fetch('/api/config');
        if (!res.ok) {
          throw new Error('Config fetch failed');
        }
        const json = await res.json();
        setConfig(json.data ?? json);
      } catch (err) {
        console.warn('Failed to load config', err);
      }
    }

    loadHealth();
    loadConfig();
  }, []);

  const handlePageCreated = useCallback((page) => {
    if (page?.id) {
      setSelectedPageId(page.id);
      setPageRefreshKey((value) => value + 1);
    }
  }, []);

  const handleSelectPage = useCallback((id) => {
    setSelectedPageId(id);
  }, []);

  return (
    <div className="app-shell">
      <header>
        <h1>Local Notion Workspace</h1>
        <p>
          Use the panels below to create pages, browse everything stored locally, and inspect rich
          page details. The layout mirrors a minimal Notion-like workspace with controls on the left
          and content on the right.
        </p>
      </header>

      <section className="status">
        <h2>System Status</h2>
        {healthError ? (
          <p className="error">{healthError}</p>
        ) : health ? (
          <pre>{JSON.stringify(health, null, 2)}</pre>
        ) : (
          <p>Checking health…</p>
        )}
        {config && (
          <details>
            <summary>Configuration</summary>
            <pre>{JSON.stringify(config, null, 2)}</pre>
          </details>
        )}
      </section>

      <main className="workspace">
        <aside className="sidebar">
          <PageCreator onCreated={handlePageCreated} />
          <DatabaseCreator />
        </aside>
        <section className="content-column">
          <PageExplorer
            selectedPageId={selectedPageId}
            onSelectPage={handleSelectPage}
            refreshKey={pageRefreshKey}
          />
          <DatabaseViewExplorer />
        </section>
      </main>

      <footer>
        <small>Backend proxy target: <code>/api</code> → <code>http://localhost:8080</code></small>
      </footer>
    </div>
  );
}
