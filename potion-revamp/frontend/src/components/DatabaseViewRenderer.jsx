import { useMemo } from 'react';
import { resolvePropertyId, resolvePropertyLabel, resolvePropertyValue, groupEntriesBySelect } from '../utils.js';

export default function DatabaseViewRenderer({ database, view, entries }) {
  const viewType = view?.type || 'table';
  const properties = database?.properties || [];
  const propertyIds = useMemo(() => properties.map((property) => property.id), [properties]);

  if (!database || !view) {
    return <p>Select a view to preview database content.</p>;
  }

  if (!entries?.length) {
    return <p>No entries yet. Use the form to add items.</p>;
  }

  if (viewType === 'kanban') {
    const groupKey = resolvePropertyId(properties, view.options?.groupBy);
    const groups = groupEntriesBySelect(entries, groupKey);
    return (
      <div className="kanban-board">
        {groups.map((column) => (
          <div key={column.key} className="kanban-column">
            <strong>{column.label}</strong>
            {column.items.map((entry) => (
              <div key={entry.id} className="block-card">
                <span>{entry.title}</span>
                <small>{resolvePropertyValue(entry, groupKey)}</small>
              </div>
            ))}
          </div>
        ))}
      </div>
    );
  }

  if (viewType === 'gallery') {
    const coverKey = resolvePropertyId(properties, view.options?.coverProperty);
    return (
      <div className="gallery-grid">
        {entries.map((entry) => (
          <div key={entry.id} className="gallery-card">
            <span className="gallery-title">{entry.title}</span>
            {coverKey && <small>{resolvePropertyValue(entry, coverKey)}</small>}
          </div>
        ))}
      </div>
    );
  }

  return (
    <table className="table-view">
      <thead>
        <tr>
          <th>Title</th>
          {propertyIds.map((propertyId) => (
            <th key={propertyId}>{resolvePropertyLabel(properties, propertyId)}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {entries.map((entry) => (
          <tr key={entry.id}>
            <td>{entry.title}</td>
            {propertyIds.map((propertyId) => (
              <td key={propertyId}>{resolvePropertyValue(entry, propertyId)}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
