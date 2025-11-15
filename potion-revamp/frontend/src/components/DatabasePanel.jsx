import { useEffect, useMemo, useState } from 'react';
import DatabaseViewRenderer from './DatabaseViewRenderer.jsx';
import { buildPropertyPayload } from '../utils.js';

const defaultProperty = { name: 'Name', type: 'title', optionsText: '' };

export default function DatabasePanel({ databases, selectedDatabase, viewPayload, onDatabaseCreated, onViewRefresh }) {
  const [dbTitle, setDbTitle] = useState('New Database');
  const [propertyDrafts, setPropertyDrafts] = useState([defaultProperty]);
  const [viewName, setViewName] = useState('Default view');
  const [viewType, setViewType] = useState('table');
  const [groupBy, setGroupBy] = useState('');
  const [coverProperty, setCoverProperty] = useState('');
  const [entryDraft, setEntryDraft] = useState({ title: '', properties: {} });

  const properties = selectedDatabase?.properties || [];
  const views = selectedDatabase?.views || [];
  const activeView = useMemo(() => viewPayload?.database?.views?.find((view) => view.id === viewPayload?.activeView?.id) || views?.[0], [viewPayload, views]);

  useEffect(() => {
    if (!selectedDatabase) return;
    const draft = { title: '', properties: {} };
    selectedDatabase.properties.forEach((property) => {
      if (property.type === 'checkbox') {
        draft.properties[property.id] = false;
      } else if (property.type === 'multi_select') {
        draft.properties[property.id] = [];
      } else {
        draft.properties[property.id] = '';
      }
    });
    setEntryDraft(draft);
  }, [selectedDatabase?.id]);

  const handleAddProperty = () => {
    setPropertyDrafts((prev) => [...prev, { name: '', type: 'text', optionsText: '' }]);
  };

  const handlePropertyChange = (index, updates) => {
    setPropertyDrafts((prev) => {
      const draft = [...prev];
      draft[index] = { ...draft[index], ...updates };
      return draft;
    });
  };

  const handleCreateDatabase = async (event) => {
    event.preventDefault();
    const propertiesPayload = propertyDrafts
      .filter((property) => property.name.trim())
      .map((property, index) => buildPropertyPayload(property, index));
    if (!propertiesPayload.length) return;
    const viewOptions = {};
    if (viewType === 'kanban' && groupBy) {
      viewOptions.groupBy = groupBy;
    }
    if (viewType === 'gallery' && coverProperty) {
      viewOptions.coverProperty = coverProperty;
    }
    const payload = {
      title: dbTitle,
      properties: propertiesPayload,
      views: [
        {
          name: viewName,
          type: viewType,
          options: viewOptions,
        },
      ],
    };
    try {
      const response = await fetch('/databases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (response.ok) {
        const created = await response.json();
        onDatabaseCreated(created.id);
        setDbTitle('New Database');
        setPropertyDrafts([defaultProperty]);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleEntryFieldChange = (propertyId, value) => {
    setEntryDraft((prev) => ({
      ...prev,
      properties: {
        ...prev.properties,
        [propertyId]: value,
      },
    }));
  };

  const handleCreateEntry = async (event) => {
    event.preventDefault();
    if (!selectedDatabase) return;
    const payload = {
      title: entryDraft.title,
      values: entryDraft.properties,
    };
    try {
      const response = await fetch(`/databases/${selectedDatabase.id}/entries`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (response.ok) {
        const created = await response.json();
        const nextDraft = { ...entryDraft, title: '' };
        setEntryDraft(nextDraft);
        const defaultView = activeView || views[0];
        if (defaultView) {
          onViewRefresh(selectedDatabase.id, defaultView.id);
        }
      }
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="database-panel">
      <h2>Database Builder</h2>
      <form onSubmit={handleCreateDatabase} className="database-form">
        <input value={dbTitle} onChange={(event) => setDbTitle(event.target.value)} placeholder="Database title" />
        <div className="property-list">
          {propertyDrafts.map((property, index) => (
            <div key={index} className="property-editor">
              <input
                placeholder="Property name"
                value={property.name}
                onChange={(event) => handlePropertyChange(index, { name: event.target.value })}
              />
              <select value={property.type} onChange={(event) => handlePropertyChange(index, { type: event.target.value })}>
                <option value="title">Title</option>
                <option value="text">Text</option>
                <option value="number">Number</option>
                <option value="select">Select</option>
                <option value="multi_select">Multi Select</option>
                <option value="date">Date</option>
                <option value="checkbox">Checkbox</option>
              </select>
              {(property.type === 'select' || property.type === 'multi_select') && (
                <input
                  placeholder="Option1, Option2"
                  value={property.optionsText || ''}
                  onChange={(event) => handlePropertyChange(index, { optionsText: event.target.value })}
                />
              )}
            </div>
          ))}
        </div>
        <button type="button" className="secondary" onClick={handleAddProperty}>
          Add Property
        </button>
        <div className="view-editor">
          <input value={viewName} onChange={(event) => setViewName(event.target.value)} placeholder="View name" />
          <select value={viewType} onChange={(event) => setViewType(event.target.value)}>
            <option value="table">Table</option>
            <option value="kanban">Kanban</option>
            <option value="gallery">Gallery</option>
          </select>
          {viewType === 'kanban' && (
            <input
              placeholder="Group by property name"
              value={groupBy}
              onChange={(event) => setGroupBy(event.target.value)}
            />
          )}
          {viewType === 'gallery' && (
            <input
              placeholder="Cover property name"
              value={coverProperty}
              onChange={(event) => setCoverProperty(event.target.value)}
            />
          )}
        </div>
        <button type="submit">Create Database</button>
      </form>

      {selectedDatabase ? (
        <div className="database-view">
          <header className="block-toolbar">
            <strong>{selectedDatabase.title}</strong>
            <select onChange={(event) => onViewRefresh(selectedDatabase.id, event.target.value)} value={activeView?.id || ''}>
              {views.map((view) => (
                <option key={view.id} value={view.id}>
                  {view.name}
                </option>
              ))}
            </select>
          </header>
          <DatabaseViewRenderer database={selectedDatabase} view={activeView} entries={viewPayload?.entries || []} />
          <form onSubmit={handleCreateEntry} className="entry-form">
            <h3>Create Entry</h3>
            <input
              value={entryDraft.title}
              onChange={(event) => setEntryDraft((prev) => ({ ...prev, title: event.target.value }))}
              placeholder="Entry title"
            />
            {properties.map((property) => (
              <PropertyInput
                key={property.id}
                property={property}
                value={entryDraft.properties?.[property.id]}
                onChange={(value) => handleEntryFieldChange(property.id, value)}
              />
            ))}
            <button type="submit">Add Entry</button>
          </form>
        </div>
      ) : (
        <p>Select a database to view and add entries.</p>
      )}
    </div>
  );
}

function PropertyInput({ property, value, onChange }) {
  switch (property.type) {
    case 'checkbox':
      return (
        <label>
          <input type="checkbox" checked={!!value} onChange={(event) => onChange(event.target.checked)} />
          {property.name}
        </label>
      );
    case 'select': {
      const options = property.options?.options || [];
      return (
        <select value={value?.id || ''} onChange={(event) => onChange(options.find((option) => option.id === event.target.value) || '')}>
          <option value="">Select {property.name}</option>
          {options.map((option) => (
            <option key={option.id} value={option.id}>
              {option.name}
            </option>
          ))}
        </select>
      );
    }
    case 'multi_select': {
      const options = property.options?.options || [];
      return (
        <select
          multiple
          value={Array.isArray(value) ? value.map((option) => option.id) : []}
          onChange={(event) => {
            const selected = Array.from(event.target.selectedOptions).map((option) => options.find((candidate) => candidate.id === option.value));
            onChange(selected);
          }}
        >
          {options.map((option) => (
            <option key={option.id} value={option.id}>
              {option.name}
            </option>
          ))}
        </select>
      );
    }
    case 'number':
      return (
        <input type="number" value={value || ''} onChange={(event) => onChange(Number(event.target.value))} placeholder={property.name} />
      );
    case 'date':
      return <input type="date" value={value || ''} onChange={(event) => onChange(event.target.value)} />;
    default:
      return <input value={value || ''} onChange={(event) => onChange(event.target.value)} placeholder={property.name} />;
  }
}
