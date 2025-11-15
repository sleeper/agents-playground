import { useCallback, useEffect, useMemo, useState } from 'react';
import { marked } from 'marked';
import DatabasePanel from './components/DatabasePanel.jsx';
import PageBacklinks from './components/PageBacklinks.jsx';
import { BlockEditor, blockDefaults } from './components/blocks.jsx';
import { normalizeViewOptions, prepareBlocksForSave } from './utils.js';

const API_ROOT = '';

async function jsonFetch(url, options = {}) {
  const response = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed: ${response.status}`);
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

function App() {
  const [pages, setPages] = useState([]);
  const [databases, setDatabases] = useState([]);
  const [selectedPageId, setSelectedPageId] = useState(null);
  const [pageData, setPageData] = useState(null);
  const [blocks, setBlocks] = useState([]);
  const [isSavingBlocks, setIsSavingBlocks] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [selectedDatabaseId, setSelectedDatabaseId] = useState(null);
  const [viewPayload, setViewPayload] = useState(null);

  const loadPages = useCallback(async () => {
    try {
      const payload = await jsonFetch(`${API_ROOT}/pages`);
      setPages(payload);
      if (!selectedPageId && payload.length) {
        setSelectedPageId(payload[0].id);
      }
    } catch (err) {
      console.error(err);
      setErrorMessage(err.message);
    }
  }, [selectedPageId]);

  const loadDatabases = useCallback(async () => {
    try {
      const payload = await jsonFetch(`${API_ROOT}/databases`);
      setDatabases(payload);
      if (!selectedDatabaseId && payload.length) {
        setSelectedDatabaseId(payload[0].id);
      }
    } catch (err) {
      console.error(err);
      setErrorMessage(err.message);
    }
  }, [selectedDatabaseId]);

  useEffect(() => {
    loadPages();
    loadDatabases();
  }, [loadPages, loadDatabases]);

  useEffect(() => {
    if (!selectedPageId) return;
    (async () => {
      try {
        const payload = await jsonFetch(`${API_ROOT}/pages/${selectedPageId}`);
        setPageData(payload);
        setBlocks(payload.blocks?.length ? payload.blocks : []);
      } catch (err) {
        console.error(err);
        setErrorMessage(err.message);
      }
    })();
  }, [selectedPageId]);

  useEffect(() => {
    if (!selectedDatabaseId) {
      setViewPayload(null);
      return;
    }
    (async () => {
      try {
        const metadata = await jsonFetch(`${API_ROOT}/databases/${selectedDatabaseId}`);
        if (!metadata.views.length) {
          setViewPayload({ database: metadata, entries: [] });
          return;
        }
        const defaultView = metadata.views[0];
        const view = await jsonFetch(`${API_ROOT}/databases/${selectedDatabaseId}/views/${defaultView.id}`);
        view.database = {
          ...metadata,
          views: metadata.views,
        };
        setViewPayload(view);
      } catch (err) {
        console.error(err);
        setErrorMessage(err.message);
      }
    })();
  }, [selectedDatabaseId]);

  const pageTitle = pageData?.page?.title ?? '';

  const handleCreatePage = async () => {
    try {
      const draftName = `Untitled ${pages.length + 1}`;
      const created = await jsonFetch(`${API_ROOT}/pages`, {
        method: 'POST',
        body: JSON.stringify({ title: draftName }),
      });
      await loadPages();
      setSelectedPageId(created.id);
    } catch (err) {
      setErrorMessage(err.message);
    }
  };

  const handleUpdatePageMeta = async (updates) => {
    if (!pageData) return;
    try {
      const payload = {
        title: updates.title ?? pageData.page.title,
        icon: updates.icon ?? pageData.page.icon,
        parentId: updates.parentId ?? pageData.page.parentId ?? null,
      };
      const updated = await jsonFetch(`${API_ROOT}/pages/${pageData.page.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
      setPageData(updated);
      await loadPages();
    } catch (err) {
      setErrorMessage(err.message);
    }
  };

  const handleAddBlock = (type) => {
    setBlocks((prev) => [...prev, blockDefaults(type)]);
  };

  const handleBlockChange = (index, nextBlock) => {
    setBlocks((prev) => {
      const draft = [...prev];
      draft[index] = { ...draft[index], ...nextBlock };
      return draft;
    });
  };

  const handleRemoveBlock = (index) => {
    setBlocks((prev) => prev.filter((_, idx) => idx !== index));
  };

  const handleSaveBlocks = async () => {
    if (!pageData) return;
    setIsSavingBlocks(true);
    try {
      const prepared = prepareBlocksForSave(blocks);
      await jsonFetch(`${API_ROOT}/pages/${pageData.page.id}/blocks`, {
        method: 'PUT',
        body: JSON.stringify({ blocks: prepared }),
      });
      const refreshed = await jsonFetch(`${API_ROOT}/pages/${pageData.page.id}`);
      setPageData(refreshed);
      setBlocks(refreshed.blocks);
    } catch (err) {
      setErrorMessage(err.message);
    } finally {
      setIsSavingBlocks(false);
    }
  };

  const handleDatabaseCreated = async (databaseId) => {
    await loadDatabases();
    setSelectedDatabaseId(databaseId);
  };

  const handleViewRefresh = async (databaseId, viewId) => {
    try {
      const view = await jsonFetch(`${API_ROOT}/databases/${databaseId}/views/${viewId}`);
      const metadata = await jsonFetch(`${API_ROOT}/databases/${databaseId}`);
      const normalized = normalizeViewOptions(view.database?.views?.find((v) => v.id === viewId) || view.database.views[0], metadata.properties);
      setViewPayload({ ...view, database: metadata, activeView: normalized });
    } catch (err) {
      setErrorMessage(err.message);
    }
  };

  const backlinks = pageData?.backlinks ?? [];

  const selectedDatabase = viewPayload?.database ?? null;
  const selectedView = useMemo(() => {
    if (!selectedDatabase) return null;
    const resolved = viewPayload?.database?.views?.find((view) => view.id === viewPayload?.activeView?.id);
    return resolved || viewPayload?.database?.views?.[0] || null;
  }, [viewPayload]);

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="nav-section">
          <h1>Potion</h1>
          <button onClick={handleCreatePage}>New Page</button>
        </div>
        <div className="nav-section">
          <h2>Pages</h2>
          <div className="nav-list">
            {pages.map((page) => (
              <div
                key={page.id}
                className={`nav-item ${page.id === selectedPageId ? 'active' : ''}`}
                onClick={() => setSelectedPageId(page.id)}
              >
                {page.icon ? `${page.icon} ` : ''}{page.title}
              </div>
            ))}
          </div>
        </div>
        <div className="nav-section">
          <h2>Databases</h2>
          <div className="nav-list">
            {databases.map((database) => (
              <div
                key={database.id}
                className={`nav-item ${database.id === selectedDatabaseId ? 'active' : ''}`}
                onClick={() => setSelectedDatabaseId(database.id)}
              >
                {database.icon ? `${database.icon} ` : ''}{database.title}
              </div>
            ))}
          </div>
        </div>
        {errorMessage && <div className="error-banner">{errorMessage}</div>}
      </aside>
      <main className="content">
        <section className="page-editor">
          {pageData ? (
            <>
              <input
                className="page-title"
                value={pageTitle}
                onChange={(event) => handleUpdatePageMeta({ title: event.target.value })}
              />
              <div className="block-toolbar">
                <button onClick={() => handleAddBlock('markdown')}>Markdown</button>
                <button onClick={() => handleAddBlock('heading')} className="secondary">Heading</button>
                <button onClick={() => handleAddBlock('pageLink')} className="secondary">Page Link</button>
                <button onClick={() => handleAddBlock('databaseView')} className="secondary">Database View</button>
                <button onClick={handleSaveBlocks} disabled={isSavingBlocks}>
                  {isSavingBlocks ? 'Saving…' : 'Save Blocks'}
                </button>
              </div>
              <div className="block-list">
                {blocks.map((block, index) => (
                  <BlockEditor
                    key={block.id || index}
                    block={block}
                    pages={pages}
                    databases={databases}
                    onChange={(next) => handleBlockChange(index, next)}
                    onRemove={() => handleRemoveBlock(index)}
                    onPreviewRefresh={handleViewRefresh}
                  />
                ))}
                {!blocks.length && <p>No blocks yet. Use the toolbar to add content.</p>}
              </div>
            </>
          ) : (
            <p>Select a page to begin authoring.</p>
          )}
        </section>
        <section className="database-panel">
          <DatabasePanel
            databases={databases}
            selectedDatabase={selectedDatabase}
            viewPayload={viewPayload}
            onDatabaseCreated={handleDatabaseCreated}
            onViewRefresh={handleViewRefresh}
          />
          <PageBacklinks backlinks={backlinks} onSelectPage={setSelectedPageId} />
        </section>
      </main>
    </div>
  );
}

export function MarkdownPreview({ markdown }) {
  const html = useMemo(() => marked.parse(markdown || ''), [markdown]);
  return <div className="markdown-preview" dangerouslySetInnerHTML={{ __html: html }} />;
}

export default App;
