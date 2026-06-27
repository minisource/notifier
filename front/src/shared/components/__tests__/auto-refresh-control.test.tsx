import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AutoRefreshControl } from '@/shared/components/auto-refresh-control';
import { WithIntl } from './test-utils';

describe('AutoRefreshControl', () => {
  it('renders refresh button with aria-label', () => {
    render(
      <WithIntl>
        <AutoRefreshControl
          isRefreshing={false}
          onRefresh={vi.fn()}
          autoRefreshEnabled={false}
          onToggleAutoRefresh={vi.fn()}
        />
      </WithIntl>
    );
    expect(screen.getByLabelText('Refresh')).toBeDefined();
  });

  it('calls onRefresh when refresh button clicked', () => {
    const onRefresh = vi.fn();
    render(
      <WithIntl>
        <AutoRefreshControl
          isRefreshing={false}
          onRefresh={onRefresh}
          autoRefreshEnabled={false}
          onToggleAutoRefresh={vi.fn()}
        />
      </WithIntl>
    );
    fireEvent.click(screen.getByLabelText('Refresh'));
    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it('shows last updated text when provided', () => {
    render(
      <WithIntl>
        <AutoRefreshControl
          isRefreshing={false}
          onRefresh={vi.fn()}
          lastUpdated="12:30:00"
          autoRefreshEnabled={false}
          onToggleAutoRefresh={vi.fn()}
        />
      </WithIntl>
    );
    expect(screen.getByText('12:30:00')).toBeDefined();
  });

  it('calls onToggleAutoRefresh when checkbox clicked', () => {
    const onToggle = vi.fn();
    render(
      <WithIntl>
        <AutoRefreshControl
          isRefreshing={false}
          onRefresh={vi.fn()}
          autoRefreshEnabled={false}
          onToggleAutoRefresh={onToggle}
        />
      </WithIntl>
    );
    fireEvent.click(screen.getByLabelText('Toggle auto refresh'));
    expect(onToggle).toHaveBeenCalledWith(true);
  });
});
