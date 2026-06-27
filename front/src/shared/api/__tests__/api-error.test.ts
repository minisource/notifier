import { describe, it, expect } from 'vitest';
import { ApiError } from '@/shared/api/api-error';
import type { ApiErrorResponse } from '@/shared/api/api-error';

function createMockResponse(status: number, statusText: string): Response {
  return { status, statusText, ok: status >= 200 && status < 300 } as Response;
}

describe('ApiError', () => {
  it('creates from response with error body', () => {
    const response = createMockResponse(400, 'Bad Request');
    const body: ApiErrorResponse = {
      error: { code: 'VALIDATION_ERROR', message: 'Invalid input' },
      requestId: 'req-123',
    };
    const error = ApiError.fromResponse(response, body);
    expect(error.status).toBe(400);
    expect(error.code).toBe('VALIDATION_ERROR');
    expect(error.message).toBe('Invalid input');
    expect(error.requestId).toBe('req-123');
  });

  it('creates from response without error body', () => {
    const response = createMockResponse(500, 'Internal Server Error');
    const body: ApiErrorResponse = {
      error: { code: '', message: '' },
    };
    const error = ApiError.fromResponse(response, body);
    expect(error.status).toBe(500);
    expect(error.message).toBe('Internal Server Error');
  });

  it('creates network error', () => {
    const error = ApiError.networkError(new Error('Network failure'));
    expect(error.status).toBe(0);
    expect(error.code).toBe('NETWORK_ERROR');
    expect(error.message).toBe('Network failure');
  });

  it('creates timeout error', () => {
    const error = ApiError.timeout();
    expect(error.status).toBe(0);
    expect(error.code).toBe('TIMEOUT');
    expect(error.message).toBe('Request timed out');
  });

  it('detects client error', () => {
    const error = new ApiError(400, 'BAD_REQUEST', 'Bad request');
    expect(error.isClientError()).toBe(true);
    expect(error.isServerError()).toBe(false);
  });

  it('detects server error', () => {
    const error = new ApiError(500, 'SERVER_ERROR', 'Server error');
    expect(error.isServerError()).toBe(true);
    expect(error.isClientError()).toBe(false);
  });

  it('detects rate limited', () => {
    const byStatus = new ApiError(429, 'RATE_LIMITED', 'Too many');
    const byCode = new ApiError(403, 'RATE_LIMITED', 'Too many');
    expect(byStatus.isRateLimited()).toBe(true);
    expect(byCode.isRateLimited()).toBe(true);
  });

  it('detects unauthorized', () => {
    const byStatus = new ApiError(401, 'UNAUTHORIZED', 'Unauthorized');
    const byCode = new ApiError(403, 'UNAUTHORIZED', 'No auth');
    expect(byStatus.isUnauthorized()).toBe(true);
    expect(byCode.isUnauthorized()).toBe(true);
  });

  it('detects forbidden', () => {
    const byStatus = new ApiError(403, 'FORBIDDEN', 'Forbidden');
    expect(byStatus.isForbidden()).toBe(true);
  });

  it('detects not found', () => {
    const byStatus = new ApiError(404, 'NOT_FOUND', 'Not found');
    expect(byStatus.isNotFound()).toBe(true);
  });

  it('detects conflict', () => {
    const byStatus = new ApiError(409, 'CONFLICT', 'Conflict');
    expect(byStatus.isConflict()).toBe(true);
  });
});
