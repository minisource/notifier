import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { JsonViewer } from '@/shared/components/json-viewer';
import { WithIntl } from './test-utils';

describe('JsonViewer', () => {
  it('renders simple JSON object', () => {
    render(<WithIntl><JsonViewer data={{ name: 'test', value: 123 }} /></WithIntl>);
    expect(screen.getByText('name')).toBeDefined();
    expect(screen.getByText('test')).toBeDefined();
  });

  it('redacts sensitive keys', () => {
    const data = { password: 'secret123', token: 'abc', name: 'user' };
    render(<WithIntl><JsonViewer data={data} /></WithIntl>);
    expect(screen.getByText('name')).toBeDefined();
    expect(screen.getByText('user')).toBeDefined();
    expect(screen.queryByText('secret123')).toBeNull();
    expect(screen.queryByText('abc')).toBeNull();
  });

  it('handles nested objects', () => {
    render(<WithIntl><JsonViewer data={{ nested: { key: 'value' }, flat: true }} /></WithIntl>);
    expect(screen.getByText('nested')).toBeDefined();
    expect(screen.getByText('key')).toBeDefined();
  });

  it('handles null data', () => {
    render(<WithIntl><JsonViewer data={null as unknown as Record<string, unknown>} /></WithIntl>);
    expect(screen.getByText('null')).toBeDefined();
  });

  it('handles empty object', () => {
    render(<WithIntl><JsonViewer data={{}} /></WithIntl>);
    expect(screen.getByText('{}')).toBeDefined();
  });

  it('handles array data', () => {
    render(<WithIntl><JsonViewer data={{ items: ['a', 'b'] }} /></WithIntl>);
    expect(screen.getByText('items')).toBeDefined();
  });
});
