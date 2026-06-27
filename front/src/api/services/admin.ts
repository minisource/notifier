import { api } from '../client';
import { BaseApi } from '../base';
import type { Notification, PaginatedResponse, Setting } from '@/types';

class AdminApi extends BaseApi {
  constructor() { super('/admin'); }

  async getNotifications(page = 1, pageSize = 20, filters?: Record<string, string>): Promise<PaginatedResponse<Notification>> {
    const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    if (filters) Object.entries(filters).forEach(([k, v]) => params.set(k, v));
    return api.get<PaginatedResponse<Notification>>(`/admin/notifications?${params}`);
  }

  async getNotificationLogs(notificationId: string) {
    return api.get(this.url(`/notifications/${notificationId}/logs`));
  }

  async getSettings(): Promise<Setting[]> {
    return api.get<Setting[]>(this.url('/settings'));
  }

  async updateSetting(id: string, value: string): Promise<Setting> {
    return api.put<Setting>(this.url(`/settings/${id}`), { value });
  }

  async getUsers(page = 1, pageSize = 20): Promise<PaginatedResponse<{ id: string; email: string; name: string }>> {
    return api.get(`/admin/users?page=${page}&pageSize=${pageSize}`);
  }
}

export const adminApi = new AdminApi();
