import { useState } from 'react';

const createInitialState = () => ({
  slug: '',
  title: '',
  description: '',
  icon: '',
  coverImage: '',
  properties:
    '[\n  {\n    "name": "Status",\n    "slug": "status",\n    "type": "select",\n    "config": {"options": ["Idea", "Doing", "Done"]},\n    "is_required": false,\n    "default": null,\n    "order_index": 0\n  }\n]',
  views:
    '[\n  {\n    "name": "Table",\n    "type": "table",\n    "display_properties": ["status"],\n    "filters": {},\n    "sorts": [],\n    "grouping": null,\n    "layout_options": {}\n  }\n]',
});

export default function DatabaseCreator() {
  const [form, setForm] = useState(createInitialState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdDatabase, setCreatedDatabase] = useState(null);

  function handleChange(event) {
    const { name, value } = event.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  }

  function resetForm() {
    setForm(createInitialState());
  }

  async function handleSubmit(event) {
    event.preventDefault();
    setLoading(true);
    setError('');

    try {
      const properties = form.properties.trim() ? JSON.parse(form.properties) : [];
      const views = form.views.trim() ? JSON.parse(form.views) : [];

      const payload = {
        slug: form.slug || undefined,
        title: form.title || undefined,
        description: form.description || undefined,
        icon: form.icon || undefined,
        cover_image_id: form.coverImage || undefined,
        properties,
        views,
      };

      const res = await fetch('/api/databases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.errors?.[0]?.message ?? 'Request failed');
      }
      setCreatedDatabase(json.data ?? json);
      resetForm();
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError('Properties or views JSON is invalid.');
      } else {
        setError(err.message);
      }
      setCreatedDatabase(null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h2>Create Database</h2>
      <form className="panel" onSubmit={handleSubmit}>
        <label htmlFor="database-slug">Slug</label>
        <input
          id="database-slug"
          name="slug"
          type="text"
          value={form.slug}
          onChange={handleChange}
          placeholder="Optional slug"
        />

        <label htmlFor="database-title">Title</label>
        <input
          id="database-title"
          name="title"
          type="text"
          value={form.title}
          onChange={handleChange}
          placeholder="Name of the database"
          required
        />

        <label htmlFor="database-description">Description</label>
        <textarea
          id="database-description"
          name="description"
          value={form.description}
          onChange={handleChange}
          placeholder="Explain what this database stores"
          rows={3}
        />

        <label htmlFor="database-icon">Icon</label>
        <input
          id="database-icon"
          name="icon"
          type="text"
          value={form.icon}
          onChange={handleChange}
          placeholder="Optional emoji/icon"
        />

        <label htmlFor="database-cover">Cover Image ID</label>
        <input
          id="database-cover"
          name="coverImage"
          type="text"
          value={form.coverImage}
          onChange={handleChange}
          placeholder="Optional cover image asset ID"
        />

        <label htmlFor="database-properties">Properties JSON</label>
        <textarea
          id="database-properties"
          name="properties"
          value={form.properties}
          onChange={handleChange}
          rows={10}
        />
        <small className="help-text">
          Provide an array of property definitions. See README for full schema.
        </small>

        <label htmlFor="database-views">Views JSON</label>
        <textarea
          id="database-views"
          name="views"
          value={form.views}
          onChange={handleChange}
          rows={10}
        />
        <small className="help-text">
          Provide an array of view definitions that reference property slugs.
        </small>

        <button type="submit" disabled={loading}>
          {loading ? 'Creating…' : 'Create Database'}
        </button>
      </form>

      {error && <p className="error">{error}</p>}
      {createdDatabase && (
        <article className="card">
          <header>
            <h3>Database Created</h3>
            <small>ID: {createdDatabase.id}</small>
          </header>
          <pre>{JSON.stringify(createdDatabase, null, 2)}</pre>
        </article>
      )}
    </section>
  );
}
