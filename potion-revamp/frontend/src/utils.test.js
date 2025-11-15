import { describe, expect, it } from 'vitest';
import {
  prepareBlocksForSave,
  groupEntriesBySelect,
  buildPropertyPayload,
  resolvePropertyValue,
} from './utils.js';

describe('prepareBlocksForSave', () => {
  it('assigns ids and positions', () => {
    const blocks = [
      { type: 'markdown', data: { markdown: 'Hello' } },
      { id: 'custom', type: 'heading', data: { text: 'World' } },
    ];
    const prepared = prepareBlocksForSave(blocks);
    expect(prepared).toHaveLength(2);
    expect(prepared[0].id).toBeDefined();
    expect(prepared[0].position).toBe(0);
    expect(prepared[1].id).toBe('custom');
    expect(prepared[1].position).toBe(1);
  });
});

describe('groupEntriesBySelect', () => {
  it('groups select values into columns', () => {
    const entries = [
      { id: 'a', title: 'One', properties: { status: { id: 'todo', name: 'To Do' } } },
      { id: 'b', title: 'Two', properties: { status: { id: 'done', name: 'Done' } } },
      { id: 'c', title: 'Three', properties: { status: { id: 'todo', name: 'To Do' } } },
    ];
    const grouped = groupEntriesBySelect(entries, 'status');
    expect(grouped).toHaveLength(2);
    const todo = grouped.find((column) => column.key === 'todo');
    expect(todo.items).toHaveLength(2);
  });
});

describe('buildPropertyPayload', () => {
  it('parses select options', () => {
    const payload = buildPropertyPayload({ name: 'Status', type: 'select', optionsText: 'To Do, Done' }, 0);
    expect(payload.options.options.map((option) => option.name)).toEqual(['To Do', 'Done']);
  });
});

describe('resolvePropertyValue', () => {
  it('returns formatted values for multi-select', () => {
    const entry = { properties: { tags: [{ id: 't1', name: 'Home' }, { id: 't2', name: 'Work' }] } };
    expect(resolvePropertyValue(entry, 'tags')).toBe('Home, Work');
  });
});
