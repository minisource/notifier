import type { Preference, UpdatePreferenceInput } from './types';
import type { PreferenceResponse } from '@/features/notifier/api/notifier-types';
import { adminPreferencesApi } from '@/features/notifier/api/notifier-api-mode';

function mapPreference(p: PreferenceResponse): Preference {
  return {
    id: p.id,
    userId: p.userId,
    type: p.type,
    isEnabled: p.isEnabled,
    allowInstant: p.allowInstant,
    allowDigest: p.allowDigest,
    digestFrequency: p.digestFrequency as Preference['digestFrequency'],
    quietHours: p.quietHours,
    categorySettings: p.categorySettings,
    updatedAt: p.updatedAt || new Date().toISOString(),
  };
}

export async function listPreferences(userId?: string): Promise<Preference[]> {
  const uid = userId || 'current';
  const result = await adminPreferencesApi.list(uid);
  return (result || []).map(mapPreference);
}

export async function updatePreference(userId: string, input: UpdatePreferenceInput, type: string): Promise<void> {
  await adminPreferencesApi.update(userId, type, {
    isEnabled: input.isEnabled,
    allowInstant: input.allowInstant,
    allowDigest: input.allowDigest,
    digestFrequency: input.digestFrequency,
  });
}
