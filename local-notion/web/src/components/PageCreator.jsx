import { useState } from 'react';

const createInitialState = () => ({
  slug: '',
  title: '',
  summary: '',
  content: '',
  parentPageId: '',
  tags: '',
  linkedPageIds: '',
});

export default function PageCreator({ onCreated }) {
  const [form, setForm] = useState(createInitialState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdPage, setCreatedPage] = useState(null);

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

    const payload = {
      slug: form.slug || undefined,
      title: form.title || undefined,
      summary: form.summary || undefined,
      content: form.content || undefined,
      parent_page_id: form.parentPageId.trim() ? form.parentPageId.trim() : undefined,
      tags: form.tags
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean),
      linked_page_ids: form.linkedPageIds
        .split(',')
        .map((id) => id.trim())
        .filter(Boolean),
    };

    try {
      const res = await fetch('/api/pages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.errors?.[0]?.message ?? 'Request failed');
      }
      const created = json.data ?? json;
      setCreatedPage(created);
      onCreated?.(created);
      resetForm();
    } catch (err) {
      setError(err.message);
      setCreatedPage(null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="card stack">
      <header>
        <h2>Create Page</h2>
        <p className="help-text">
          Draft a page and submit it to store it locally. Newly created pages open on the right so
          you can verify the content immediately.
        </p>
      </header>
      <form className="panel" onSubmit={handleSubmit}>
        <label htmlFor="page-slug">Slug</label>
        <input
          id="page-slug"
          name="slug"
          type="text"
          value={form.slug}
          onChange={handleChange}
          placeholder="Optional slug (letters, numbers, hyphens)"
        />

        <label htmlFor="page-title">Title</label>
        <input
          id="page-title"
          name="title"
          type="text"
          value={form.title}
          onChange={handleChange}
          placeholder="Give your page a name"
          required
        />

        <label htmlFor="page-summary">Summary</label>
        <textarea
          id="page-summary"
          name="summary"
          value={form.summary}
          onChange={handleChange}
          placeholder="Short description"
          rows={3}
        />

        <label htmlFor="page-content">Content</label>
        <textarea
          id="page-content"
          name="content"
          value={form.content}
          onChange={handleChange}
          placeholder="Long-form content (Markdown supported)"
          rows={6}
        />

        <label htmlFor="page-parent">Parent Page ID</label>
        <input
          id="page-parent"
          name="parentPageId"
          type="text"
          value={form.parentPageId}
          onChange={handleChange}
          placeholder="Optional parent page ID"
        />

        <label htmlFor="page-tags">Tags</label>
        <input
          id="page-tags"
          name="tags"
          type="text"
          value={form.tags}
          onChange={handleChange}
          placeholder="Comma separated list"
        />

        <label htmlFor="page-links">Linked Page IDs</label>
        <textarea
          id="page-links"
          name="linkedPageIds"
          value={form.linkedPageIds}
          onChange={handleChange}
          placeholder="Comma separated IDs of related pages"
          rows={2}
        />
        <p className="field-hint">
          Link pages together to build Notion-style relationships. Backlinks are created automatically when
          loading a page that other pages reference.
        </p>

        <button type="submit" disabled={loading}>
          {loading ? 'Creating…' : 'Create Page'}
        </button>
      </form>

      {error && <p className="error">{error}</p>}
      {createdPage && (
        <article className="card">
          <header>
            <h3>Page Created</h3>
            <small>ID: {createdPage.id}</small>
          </header>
          <pre>{JSON.stringify(createdPage, null, 2)}</pre>
        </article>
      )}
    </section>
  );
}
