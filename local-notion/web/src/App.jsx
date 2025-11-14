import { useEffect, useState } from 'react';
import PageCreator from './components/PageCreator.jsx';
import PageExplorer from './components/PageExplorer.jsx';
import DatabaseCreator from './components/DatabaseCreator.jsx';
import DatabaseViewExplorer from './components/DatabaseViewExplorer.jsx';
import './App.css';

export default function App() {
  const [health, setHealth] = useState(null);
  const [config, setConfig] = useState(null);
  const [healthError, setHealthError] = useState('');

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

  return (
    <div className="app-shell">
      <header>
        <h1>Agents Playground Control Panel</h1>
        <p>
          Inspect API health, load pages, and browse database view items. This UI is intentionally
          lightweight to perform well on small devices.
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

      <main>
        <PageCreator />
        <PageExplorer />
        <DatabaseCreator />
        <DatabaseViewExplorer />
      </main>

      <footer>
        <small>Backend proxy target: <code>/api</code> → <code>http://localhost:8080</code></small>
      </footer>
    </div>
  );
}
