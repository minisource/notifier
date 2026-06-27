import { describe, it, expect } from 'vitest';
import { notifierMock } from '@/features/notifier/api/notifier-mocks';

describe('notifierMock', () => {
  it('getDashboardOverview returns expected shape', async () => {
    const data = await notifierMock.getDashboardOverview();
    expect(data).toBeDefined();
    expect(data.totalNotifications).toBeGreaterThan(0);
    expect(data.successRate).toBeGreaterThan(0);
    expect(data.channelBreakdown).toBeInstanceOf(Array);
    expect(data.statusBreakdown).toBeInstanceOf(Array);
    expect(data.dailyTrend).toBeInstanceOf(Array);
    expect(data.recentFailures).toBeInstanceOf(Array);
    expect(data.queue).toBeDefined();
    expect(data.providers).toBeDefined();
  });

  it('listNotifications returns paginated response', async () => {
    const result = await notifierMock.listNotifications({ page: 1, pageSize: 5 });
    expect(result).toBeDefined();
    expect(result.data).toBeInstanceOf(Array);
    expect(result.total).toBeGreaterThan(0);
    expect(result.page).toBe(1);
    expect(result.pageSize).toBe(5);
    expect(result.data.length).toBeLessThanOrEqual(5);
  });

  it('listNotifications filters by status', async () => {
    const result = await notifierMock.listNotifications({ status: 'failed' });
    expect(result.data.every(n => n.status === 'failed')).toBe(true);
  });

  it('getNotification returns single notification', async () => {
    const all = await notifierMock.listNotifications();
    if (all.data.length > 0) {
      const notification = await notifierMock.getNotification(all.data[0].id);
      expect(notification).toBeDefined();
      expect(notification.id).toBe(all.data[0].id);
    }
  });

  it('listTemplates returns templates', async () => {
    const templates = await notifierMock.listTemplates();
    expect(templates).toBeInstanceOf(Array);
    expect(templates.length).toBeGreaterThan(0);
  });

  it('listProviders returns providers with health data', async () => {
    const providers = await notifierMock.listProviders();
    expect(providers).toBeInstanceOf(Array);
    expect(providers.length).toBeGreaterThan(0);
    expect(providers[0]).toHaveProperty('successRate');
    expect(providers[0]).toHaveProperty('status');
  });

  it('getHealth returns observability health', async () => {
    const health = await notifierMock.getHealth();
    expect(health).toBeDefined();
    expect(health.status).toBe('healthy');
    expect(health.dependencies).toBeInstanceOf(Array);
  });

  it('getQueueOverview returns queue data', async () => {
    const queue = await notifierMock.getQueueOverview();
    expect(queue).toBeDefined();
    expect(queue.queuedCount).toBeGreaterThanOrEqual(0);
    expect(queue.throughputPerMinute).toBeGreaterThan(0);
  });

  it('getWorkersOverview returns workers', async () => {
    const workers = await notifierMock.getWorkersOverview();
    expect(workers).toBeDefined();
    expect(workers.workers).toBeInstanceOf(Array);
    expect(workers.activeCount).toBeGreaterThan(0);
  });
});
