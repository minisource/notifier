const SENSITIVE_KEYS = new Set([
  'password',
  'current_password',
  'new_password',
  'otp',
  'code',
  'token',
  'accesstoken',
  'refreshtoken',
  'access_token',
  'refresh_token',
  'client_secret',
  'secret',
  'api_key',
  'apikey',
  'authorization',
  'cookie',
  'set-cookie',
  'private_key',
  'smtp_password',
  'webhook_secret',
  'signing_key',
  'auth_token',
  'credentials',
]);

/**
 * Recursively masks sensitive fields in objects, arrays, and headers.
 * Applies case-insensitively.
 */
export function redactSensitiveData(val: any): any {
  if (val === null || val === undefined) return val;

  if (typeof val === 'string') {
    // Mask raw Bearer tokens or potential secret strings
    if (val.toLowerCase().startsWith('bearer ')) {
      return 'Bearer •••••••••';
    }
    return val;
  }

  if (Array.isArray(val)) {
    return val.map(redactSensitiveData);
  }

  if (typeof val === 'object') {
    const redactedCopy: Record<string, any> = {};
    for (const key of Object.keys(val)) {
      const lowerKey = key.toLowerCase();
      if (SENSITIVE_KEYS.has(lowerKey)) {
        redactedCopy[key] = '••••••••';
      } else {
        redactedCopy[key] = redactSensitiveData(val[key]);
      }
    }
    return redactedCopy;
  }

  return val;
}
