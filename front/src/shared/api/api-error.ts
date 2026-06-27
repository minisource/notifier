export interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
  requestId?: string;
}

export class ApiError extends Error {
  public readonly status: number;
  public readonly code: string;
  public readonly requestId?: string;
  public readonly details?: unknown;

  constructor(status: number, code: string, message: string, requestId?: string, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.details = details;
  }

  static fromResponse(response: Response, body: ApiErrorResponse): ApiError {
    return new ApiError(
      response.status,
      body.error?.code || 'UNKNOWN_ERROR',
      body.error?.message || response.statusText || 'Unknown error',
      body.requestId,
      body.error?.details,
    );
  }

  static networkError(error: Error): ApiError {
    return new ApiError(0, 'NETWORK_ERROR', error.message || 'Network error');
  }

  static timeout(): ApiError {
    return new ApiError(0, 'TIMEOUT', 'Request timed out');
  }

  isClientError(): boolean {
    return this.status >= 400 && this.status < 500;
  }

  isServerError(): boolean {
    return this.status >= 500;
  }

  isRateLimited(): boolean {
    return this.status === 429 || this.code === 'RATE_LIMITED';
  }

  isUnauthorized(): boolean {
    return this.status === 401 || this.code === 'UNAUTHORIZED';
  }

  isForbidden(): boolean {
    return this.status === 403 || this.code === 'FORBIDDEN';
  }

  isNotFound(): boolean {
    return this.status === 404 || this.code === 'NOT_FOUND';
  }

  isConflict(): boolean {
    return this.status === 409 || this.code === 'CONFLICT';
  }
}
