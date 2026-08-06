import { describe, it, expect } from 'vitest';
import { redactSensitiveData } from '../redaction';

describe('Redaction Utility', () => {
  it('should redact sensitive keys in flat objects', () => {
    const input = {
      user: 'alice',
      password: 'super-secret-password',
      token: 'jwt-token-value',
      authorization: 'Bearer token-123',
    };

    const redacted = redactSensitiveData(input);
    expect(redacted.user).toBe('alice');
    expect(redacted.password).toBe('••••••••');
    expect(redacted.token).toBe('••••••••');
    expect(redacted.authorization).toBe('••••••••');
  });

  it('should recursively redact nested objects and arrays case-insensitively', () => {
    const input = {
      data: {
        API_KEY: 'secret-key-123',
        nestedArray: [
          { RefreshToken: 'ref-token' },
          { normalField: 'public' },
        ],
      },
    };

    const redacted = redactSensitiveData(input);
    expect(redacted.data.API_KEY).toBe('••••••••');
    expect(redacted.data.nestedArray[0].RefreshToken).toBe('••••••••');
    expect(redacted.data.nestedArray[1].normalField).toBe('public');
  });

  it('should mask standalone Bearer token strings', () => {
    const headerVal = 'Bearer eyJhbGciOiJKV1QiLCJhbGci...';
    expect(redactSensitiveData(headerVal)).toBe('Bearer •••••••••');
  });
});
