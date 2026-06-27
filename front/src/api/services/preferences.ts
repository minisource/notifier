import { api } from '../client';
import { BaseApi } from '../base';
import type { NotificationPreference, UpdatePreferenceDto } from '@/types';

class PreferencesApi extends BaseApi {
  constructor() { super('/preferences'); }

  async getByUser(userId: string): Promise<NotificationPreference[]> {
    return api.get<NotificationPreference[]>(this.url(`/user/${userId}`));
  }

  async update(userId: string, data: UpdatePreferenceDto): Promise<NotificationPreference> {
    return api.put<NotificationPreference>(this.url(`/user/${userId}`), data);
  }
}

export const preferencesApi = new PreferencesApi();
