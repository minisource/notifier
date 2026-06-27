import { api } from '../client';
import { BaseApi } from '../base';
import type {
  Notification, PaginatedResponse, CreateNotificationDto, BatchNotificationDto,
  DashboardStats, NotificationLog,
} from '@/types';

class NotificationsApi extends BaseApi {
  constructor() { super('/notifications'); }

  async getAll(userId: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Notification>> {
    return api.get<PaginatedResponse<Notification>>(this.url(`/user/${userId}?page=${page}&pageSize=${pageSize}`));
  }

  async getUnread(userId: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Notification>> {
    return api.get<PaginatedResponse<Notification>>(this.url(`/user/${userId}/unread?page=${page}&pageSize=${pageSize}`));
  }

  async create(data: CreateNotificationDto): Promise<Notification> {
    return api.post<Notification>(this.url('/'), data);
  }

  async createBatch(data: BatchNotificationDto): Promise<{ successCount: number; failedCount: number; successIds: string[] }> {
    return api.post(this.url('/batch'), data);
  }

  async markAsRead(notificationId: string): Promise<void> {
    return api.put(this.url(`/${notificationId}/read`));
  }

  async getLogs(notificationId: string): Promise<NotificationLog[]> {
    return api.get<NotificationLog[]>(`/admin/notifications/${notificationId}/logs`);
  }

  async getStats(): Promise<DashboardStats> {
    return api.get<DashboardStats>('/admin/notifications/stats');
  }

  async retryFailed(notificationId: string): Promise<void> {
    return api.post(this.url(`/admin/retry/${notificationId}`));
  }

  async cancel(notificationId: string, reason?: string): Promise<void> {
    return api.post(this.url(`/admin/cancel/${notificationId}`), { reason });
  }
}

export const notificationsApi = new NotificationsApi();
