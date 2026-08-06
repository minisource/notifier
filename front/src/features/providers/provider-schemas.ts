// Declarative per-provider form schemas for the Notifier frontend.
// Each provider type maps to the exact config keys the backend expects:
//   - SMS:  internal/platform/sms/sms.go  -> ProviderConfig
//   - Email: internal/platform/email/email.go -> ProviderConfig
//   - Push:  internal/platform/push/push.go -> ProviderConfig
// Fields with `extraIndex` are positional entries in the backend's `extra` array.

import type { NotificationChannel } from '@/features/notifier/api/notifier-types';

export type ProviderFieldInput = 'text' | 'password' | 'number' | 'boolean' | 'textarea';

export interface ProviderFieldDef {
  /** config / secretConfig key the backend expects (e.g. 'apiKey', 'accessId') */
  key: string;
  input: ProviderFieldInput;
  required?: boolean;
  /** secret values are sent as `secretConfig` and never returned by the API */
  secret?: boolean;
  defaultValue?: string | number | boolean;
  /** maps to config.extra[index] (positional) */
  extraIndex?: number;
}

export interface ProviderTypeDef {
  type: string;
  channels: NotificationChannel[];
  fields: ProviderFieldDef[];
}

const SMS_CHANNELS: NotificationChannel[] = ['sms'];

export const PROVIDER_TYPES: Record<string, ProviderTypeDef> = {
  kavenegar: {
    type: 'kavenegar',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'apiKey', input: 'password', required: true, secret: true },
      { key: 'template', input: 'text' },
      { key: 'senderId', input: 'text' },
    ],
  },
  twilio: {
    type: 'twilio',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'template', input: 'text' },
    ],
  },
  tencent: {
    type: 'tencent',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'sign', input: 'text', required: true },
      { key: 'template', input: 'text', required: true },
      { key: 'appId', input: 'text', required: true, extraIndex: 0 },
    ],
  },
  huawei: {
    type: 'huawei',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'sign', input: 'text', required: true },
      { key: 'template', input: 'text', required: true },
      { key: 'apiAddress', input: 'text', required: true, extraIndex: 0 },
      { key: 'sender', input: 'text', required: true, extraIndex: 1 },
    ],
  },
  infobip: {
    type: 'infobip',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'senderId', input: 'text', required: true },
      { key: 'apiKey', input: 'password', required: true, secret: true },
      { key: 'template', input: 'text' },
      { key: 'baseUrl', input: 'text', required: true, extraIndex: 0 },
    ],
  },
  msg91: {
    type: 'msg91',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'senderId', input: 'text', required: true },
      { key: 'apiKey', input: 'password', required: true, secret: true },
      { key: 'template', input: 'text', required: true },
    ],
  },
  netgsm: {
    type: 'netgsm',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'senderId', input: 'text', required: true },
      { key: 'template', input: 'text', required: true },
    ],
  },
  oson: {
    type: 'oson',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'senderId', input: 'text' },
      { key: 'template', input: 'text' },
    ],
  },
  smsbao: {
    type: 'smsbao',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'sign', input: 'text', required: true },
      { key: 'template', input: 'text', required: true },
      { key: 'goodsid', input: 'text', extraIndex: 0 },
    ],
  },
  submail: {
    type: 'submail',
    channels: SMS_CHANNELS,
    fields: [
      { key: 'accessId', input: 'text', required: true },
      { key: 'accessKey', input: 'password', required: true, secret: true },
      { key: 'template', input: 'text', required: true },
    ],
  },
  smtp: {
    type: 'smtp',
    channels: ['email'],
    fields: [
      { key: 'host', input: 'text', required: true },
      { key: 'port', input: 'number', defaultValue: 587 },
      { key: 'username', input: 'text' },
      { key: 'password', input: 'password', secret: true },
      { key: 'from', input: 'text', required: true },
      { key: 'fromName', input: 'text' },
      { key: 'useTls', input: 'boolean', defaultValue: true },
    ],
  },
  fcm: {
    type: 'fcm',
    channels: ['push'],
    fields: [
      { key: 'serverKey', input: 'password', required: true, secret: true },
      { key: 'projectId', input: 'text' },
      { key: 'privateKeyJson', input: 'textarea', secret: true },
    ],
  },
  custom: {
    type: 'custom',
    channels: ['sms', 'email', 'push', 'webhook', 'in_app'],
    fields: [],
  },
};

/** Deterministic display order for the type selector */
const TYPE_ORDER: string[] = [
  'kavenegar', 'twilio', 'tencent', 'huawei', 'infobip', 'msg91',
  'netgsm', 'oson', 'smsbao', 'submail', 'smtp', 'fcm', 'custom',
];

export function typesForChannel(channel: string): ProviderTypeDef[] {
  return TYPE_ORDER
    .filter((t) => PROVIDER_TYPES[t].channels.includes(channel as NotificationChannel))
    .map((t) => PROVIDER_TYPES[t]);
}

/**
 * Builds the `config` + `secretConfig` payloads from the form's field values,
 * honoring secret vs public split and positional `extra` entries.
 */
export function buildProviderConfig(
  type: string,
  values: Record<string, unknown>,
): { config: Record<string, unknown>; secretConfig: Record<string, unknown> } {
  const def = PROVIDER_TYPES[type];
  // The backend client factories switch on the "provider" key in config
  // (e.g. sms.NewClientFromConfig). Always declare it so health checks and
  // test sends work for providers created through the UI.
  const config: Record<string, unknown> = { provider: type };
  const secretConfig: Record<string, unknown> = {};
  const extra: unknown[] = [];
  let hasExtra = false;

  for (const field of def?.fields ?? []) {
    const raw = values[field.key];
    if (raw === undefined || raw === null || raw === '') continue;

    if (field.input === 'number') {
      const n = Number(raw);
      if (Number.isNaN(n)) continue;
      if (field.secret) secretConfig[field.key] = n;
      else config[field.key] = n;
      continue;
    }
    if (field.input === 'boolean') {
      if (field.secret) secretConfig[field.key] = !!raw;
      else config[field.key] = !!raw;
      continue;
    }
    if (field.extraIndex !== undefined) {
      extra[field.extraIndex] = String(raw);
      hasExtra = true;
      continue;
    }
    const val = String(raw);
    if (field.secret) secretConfig[field.key] = val;
    else config[field.key] = val;
  }

  if (hasExtra) config.extra = extra;
  return { config, secretConfig };
}

/**
 * Seeds form field values from an existing (redacted) config. Secret fields are
 * never prefilled — in edit mode they stay empty so the backend keeps the
 * existing values.
 */
export function configToFieldValues(
  type: string,
  config?: Record<string, unknown>,
): Record<string, unknown> {
  const def = PROVIDER_TYPES[type];
  const values: Record<string, unknown> = {};
  for (const field of def?.fields ?? []) {
    if (field.secret) continue;
    if (field.extraIndex !== undefined) {
      const extra = config?.extra;
      values[field.key] = Array.isArray(extra) ? (extra[field.extraIndex] ?? '') : '';
      continue;
    }
    values[field.key] =
      config?.[field.key] ?? field.defaultValue ?? (field.input === 'boolean' ? false : '');
  }
  return values;
}

export function fieldLabelKey(type: string, key: string): string {
  return `providers.fields.${type}.${key}`;
}

export function typeLabelKey(type: string): string {
  return `providers.types.${type}`;
}
