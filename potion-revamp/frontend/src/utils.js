import { nanoid } from './vendor.js';

export function prepareBlocksForSave(blocks) {
  return blocks.map((block, index) => ({
    id: block.id || nanoid(),
    type: block.type,
    position: index,
    data: sanitizeData(block.data || {}),
  }));
}

export function sanitizeData(data) {
  if (!data || typeof data !== 'object') return {};
  return JSON.parse(JSON.stringify(data));
}

export function normalizeViewOptions(view, properties) {
  if (!view) return view;
  const copy = { ...view, options: { ...(view.options || {}) } };
  if (copy.type === 'kanban' && copy.options.groupBy) {
    copy.options.groupBy = resolvePropertyId(properties, copy.options.groupBy);
  }
  if (copy.type === 'gallery' && copy.options.coverProperty) {
    copy.options.coverProperty = resolvePropertyId(properties, copy.options.coverProperty);
  }
  return copy;
}

export function buildPropertyPayload(property, position) {
  const payload = {
    name: property.name,
    type: property.type,
    position,
  };
  if (property.type === 'select' || property.type === 'multi_select') {
    payload.options = {
      options: (property.optionsText || '')
        .split(',')
        .map((option) => option.trim())
        .filter(Boolean)
        .map((option) => ({ id: slugify(option), name: option })),
    };
  }
  return payload;
}

export function resolvePropertyId(properties, candidate) {
  if (!candidate) return candidate;
  const direct = properties.find((property) => property.id === candidate);
  if (direct) return direct.id;
  const byName = properties.find((property) => property.name.toLowerCase() === String(candidate).toLowerCase());
  return byName ? byName.id : candidate;
}

export function resolvePropertyLabel(properties, propertyId) {
  const property = properties.find((candidate) => candidate.id === propertyId);
  return property ? property.name : propertyId;
}

export function resolvePropertyValue(entry, propertyId) {
  if (!entry) return '';
  if (propertyId === 'title') return entry.title;
  const value = entry.properties?.[propertyId];
  if (value == null) return '';
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (Array.isArray(value)) {
    return value.map((item) => item?.name || item).join(', ');
  }
  if (typeof value === 'object') {
    return value.name || value.title || JSON.stringify(value);
  }
  return value;
}

export function groupEntriesBySelect(entries, propertyId) {
  if (!propertyId) return [{ key: 'default', label: 'Ungrouped', items: entries }];
  const buckets = new Map();
  entries.forEach((entry) => {
    const value = entry.properties?.[propertyId];
    if (Array.isArray(value)) {
      value.forEach((item) => {
        const key = item?.id || item?.name || 'unsorted';
        const label = item?.name || 'Untitled';
        if (!buckets.has(key)) {
          buckets.set(key, { key, label, items: [] });
        }
        buckets.get(key).items.push(entry);
      });
    } else {
      const key = value?.id || value?.name || 'unsorted';
      const label = value?.name || value || 'Untitled';
      if (!buckets.has(key)) {
        buckets.set(key, { key, label, items: [] });
      }
      buckets.get(key).items.push(entry);
    }
  });
  if (!buckets.size) {
    buckets.set('unsorted', { key: 'unsorted', label: 'Unsorted', items: entries });
  }
  return Array.from(buckets.values());
}

function slugify(value) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 24) || nanoid();
}
