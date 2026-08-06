import { describe, it, expect } from 'vitest';
import { classifyError } from '../error-normalizer';

describe('Error Normalizer', () => {
  it('should classify connection refusal / network error (status 0)', () => {
    const err = classifyError(0, null, 'ECONNREFUSED 127.0.0.1:9002');
    expect(err.category).toBe('NETWORK_ERROR');
    expect(err.title).toContain('Connection Refused');
  });

  it('should classify 401 Unauthorized', () => {
    const err = classifyError(401, { error: { message: 'Authentication required' } });
    expect(err.category).toBe('UNAUTHORIZED');
  });

  it('should classify 403 Forbidden', () => {
    const err = classifyError(403, { error: { message: 'Admin role required' } });
    expect(err.category).toBe('FORBIDDEN');
  });

  it('should classify 429 Rate Limit Exceeded', () => {
    const err = classifyError(429);
    expect(err.category).toBe('RATE_LIMIT');
  });

  it('should classify 502 Gateway error', () => {
    const err = classifyError(502);
    expect(err.category).toBe('GATEWAY_ERROR');
  });

  it('should classify 500 Internal Server Error', () => {
    const err = classifyError(500, { message: 'Database connection failed' });
    expect(err.category).toBe('SERVER_ERROR');
  });
});
