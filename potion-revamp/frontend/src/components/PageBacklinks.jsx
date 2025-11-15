export default function PageBacklinks({ backlinks, onSelectPage }) {
  if (!backlinks?.length) {
    return (
      <div className="backlinks-card">
        <strong>Backlinks</strong>
        <p>No backlinks yet. Link this page elsewhere to build your knowledge graph.</p>
      </div>
    );
  }
  return (
    <div className="backlinks-card">
      <strong>Backlinks</strong>
      <ul>
        {backlinks.map((page) => (
          <li key={page.id}>
            <button className="secondary" onClick={() => onSelectPage(page.id)}>
              {page.title}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
