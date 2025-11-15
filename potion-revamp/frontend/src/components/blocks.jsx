import { useEffect, useMemo, useState } from 'react';
import { marked } from 'marked';
import DatabaseViewRenderer from './DatabaseViewRenderer.jsx';

export const blockDefaults = (type) => {
  switch (type) {
    case 'heading':
      return { id: '', type, data: { text: 'New heading' } };
    case 'pageLink':
      return { id: '', type, data: { targetPageId: '', alias: '' } };
    case 'databaseView':
      return { id: '', type, data: { databaseId: '', viewId: '' } };
    default:
      return { id: '', type: 'markdown', data: { markdown: '' } };
  }
};

export function BlockEditor({ block, pages, databases, onChange, onRemove }) {
  if (!block) return null;
  const headerLabel = block.type === 'markdown' ? 'Markdown' : block.type === 'heading' ? 'Heading' : block.type === 'pageLink' ? 'Linked Page' : 'Database View';

  return (
    <div className="block-card">
      <header>
        <span>{headerLabel}</span>
        <div className="block-toolbar">
          <button className="secondary" onClick={onRemove}>Remove</button>
        </div>
      </header>
      {block.type === 'markdown' && <MarkdownBlock block={block} onChange={onChange} />}
      {block.type === 'heading' && <HeadingBlock block={block} onChange={onChange} />}
      {block.type === 'pageLink' && <PageLinkBlock block={block} pages={pages} onChange={onChange} />}
      {block.type === 'databaseView' && <DatabaseBlock block={block} databases={databases} onChange={onChange} />}
    </div>
  );
}

function MarkdownBlock({ block, onChange }) {
  const markdown = block.data?.markdown || '';
  const preview = useMemo(() => marked.parse(markdown || ''), [markdown]);
  return (
    <div className="markdown-editor">
      <textarea
        value={markdown}
        onChange={(event) => onChange({ data: { ...block.data, markdown: event.target.value } })}
      />
      <div className="markdown-preview" dangerouslySetInnerHTML={{ __html: preview }} />
    </div>
  );
}

function HeadingBlock({ block, onChange }) {
  return (
    <input
      value={block.data?.text || ''}
      onChange={(event) => onChange({ data: { ...block.data, text: event.target.value } })}
    />
  );
}

function PageLinkBlock({ block, pages, onChange }) {
  const targetPageId = block.data?.targetPageId || '';
  const alias = block.data?.alias || '';
  return (
    <div className="page-link-editor">
      <select
        value={targetPageId}
        onChange={(event) => onChange({ data: { ...block.data, targetPageId: event.target.value } })}
      >
        <option value="">Select page…</option>
        {pages.map((page) => (
          <option key={page.id} value={page.id}>
            {page.title}
          </option>
        ))}
      </select>
      <input
        placeholder="Link alias"
        value={alias}
        onChange={(event) => onChange({ data: { ...block.data, alias: event.target.value } })}
      />
    </div>
  );
}

function DatabaseBlock({ block, databases, onChange }) {
  const databaseId = block.data?.databaseId || '';
  const viewId = block.data?.viewId || '';
  const [metadata, setMetadata] = useState(null);
  const [viewPayload, setViewPayload] = useState(null);

  useEffect(() => {
    if (!databaseId) {
      setMetadata(null);
      setViewPayload(null);
      return;
    }
    (async () => {
      try {
        const dbRes = await fetch(`/databases/${databaseId}`);
        if (!dbRes.ok) return;
        const meta = await dbRes.json();
        setMetadata(meta);
        const resolvedView = viewId || meta.views?.[0]?.id;
        if (resolvedView) {
          const viewRes = await fetch(`/databases/${databaseId}/views/${resolvedView}`);
          if (viewRes.ok) {
            const payload = await viewRes.json();
            setViewPayload({ ...payload, view: meta.views.find((view) => view.id === resolvedView) });
          }
        }
      } catch (err) {
        console.error(err);
      }
    })();
  }, [databaseId, viewId]);

  const handleDatabaseChange = (event) => {
    const nextDatabaseId = event.target.value;
    const nextData = { ...block.data, databaseId: nextDatabaseId, viewId: '' };
    onChange({ data: nextData });
  };

  const handleViewChange = (event) => {
    const nextViewId = event.target.value;
    const nextData = { ...block.data, viewId: nextViewId };
    onChange({ data: nextData });
  };

  const viewOptions = metadata?.views || [];
  const resolvedView = viewOptions.find((view) => view.id === (viewId || viewPayload?.view?.id));

  return (
    <div className="database-embed">
      <div className="page-link-editor">
        <select value={databaseId} onChange={handleDatabaseChange}>
          <option value="">Select database…</option>
          {databases.map((database) => (
            <option key={database.id} value={database.id}>
              {database.title}
            </option>
          ))}
        </select>
        <select value={viewId} onChange={handleViewChange} disabled={!databaseId || !viewOptions.length}>
          <option value="">Select view…</option>
          {viewOptions.map((view) => (
            <option key={view.id} value={view.id}>
              {view.name}
            </option>
          ))}
        </select>
      </div>
      {viewPayload && metadata && resolvedView ? (
        <DatabaseViewRenderer database={metadata} view={resolvedView} entries={viewPayload.entries || []} />
      ) : (
        <p>Select a database and view to embed its content.</p>
      )}
      {resolvedView && metadata && (
        <small className="hint">Embedding view {resolvedView.name} from {metadata.title}</small>
      )}
    </div>
  );
}
