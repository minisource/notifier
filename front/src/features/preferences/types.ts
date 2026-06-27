export interface Preference {
  id: string;
  userId: string;
  type: string;
  isEnabled: boolean;
  allowInstant: boolean;
  allowDigest: boolean;
  digestFrequency: 'daily' | 'weekly' | 'monthly';
  quietHours?: { start: string; end: string; timezone: string };
  categorySettings?: Record<string, boolean>;
  updatedAt: string;
}

export interface UpdatePreferenceInput {
  isEnabled?: boolean;
  allowInstant?: boolean;
  allowDigest?: boolean;
  digestFrequency?: 'daily' | 'weekly' | 'monthly';
}
