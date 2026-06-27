import { api } from '../client';
import { BaseApi } from '../base';
import type { NotificationTemplate, CreateTemplateDto, PaginatedResponse } from '@/types';

class TemplatesApi extends BaseApi {
  constructor() { super('/templates'); }

  async getAll(page = 1, pageSize = 50): Promise<PaginatedResponse<NotificationTemplate>> {
    return api.get<PaginatedResponse<NotificationTemplate>>(this.url(`?page=${page}&pageSize=${pageSize}`));
  }

  async getById(id: string): Promise<NotificationTemplate> {
    return api.get<NotificationTemplate>(this.url(`/${id}`));
  }

  async create(data: CreateTemplateDto): Promise<NotificationTemplate> {
    return api.post<NotificationTemplate>(this.url('/'), data);
  }

  async update(id: string, data: CreateTemplateDto): Promise<NotificationTemplate> {
    return api.put<NotificationTemplate>(this.url(`/${id}`), data);
  }

  async delete(id: string): Promise<void> {
    return api.delete(this.url(`/${id}`));
  }
}

export const templatesApi = new TemplatesApi();
