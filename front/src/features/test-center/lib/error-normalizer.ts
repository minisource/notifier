import { ErrorClassification } from '../types';

export function classifyError(statusCode: number, responseBody?: any, errorMessage?: string): ErrorClassification {
  const message = errorMessage || responseBody?.error?.message || responseBody?.message || 'An error occurred';

  if (statusCode === 0 || message.includes('ECONNREFUSED') || message.includes('Network Error')) {
    return {
      category: 'NETWORK_ERROR',
      title: 'Connection Refused / Network Error',
      message: 'The target backend or gateway is unreachable. Verify service health and port availability.',
      remediation: 'Ensure Notifier backend (:9002) or Gateway (:8080) is running.',
    };
  }

  if (message.includes('timeout') || message.includes('ETIMEDOUT') || statusCode === 504) {
    return {
      category: 'TIMEOUT',
      title: 'Request Timed Out',
      message: 'The operation exceeded the configured timeout limit.',
      remediation: 'Check if the backend process is hanging or processing high load.',
    };
  }

  if (message.includes('CORS') || message.includes('Access-Control-Allow-Origin')) {
    return {
      category: 'CORS_FAILURE',
      title: 'CORS Preflight Failure',
      message: 'Browser blocked request due to missing CORS headers.',
      remediation: 'Verify backend AllowOrigins config in api.go or Traefik cors middleware.',
    };
  }

  if (statusCode === 401) {
    return {
      category: 'UNAUTHORIZED',
      title: '401 Unauthorized',
      message: 'Authentication token is missing, expired, or invalid.',
      remediation: 'Provide a valid Bearer token in the Authentication settings or sign in as Admin.',
    };
  }

  if (statusCode === 403) {
    return {
      category: 'FORBIDDEN',
      title: '403 Forbidden',
      message: 'The current identity does not hold required admin roles or scopes.',
      remediation: 'Check JWT roles/claims or admin permissions.',
    };
  }

  if (statusCode === 404) {
    return {
      category: 'NOT_FOUND',
      title: '404 Endpoint / Resource Not Found',
      message: 'The requested route or resource ID was not found.',
      remediation: 'Verify route path registration in backend router or test resource ID.',
    };
  }

  if (statusCode === 409) {
    return {
      category: 'CONFLICT',
      title: '409 Resource Conflict / Idempotency Duplicate',
      message: 'Resource with this key already exists or duplicate request detected.',
    };
  }

  if (statusCode === 429) {
    return {
      category: 'RATE_LIMIT',
      title: '429 Rate Limit Exceeded',
      message: 'Too many requests dispatched within the rate limit window.',
      remediation: 'Wait for rate limit window reset before retrying.',
    };
  }

  if (statusCode === 400 || statusCode === 422) {
    return {
      category: 'BACKEND_VALIDATION',
      title: 'Validation Error',
      message: message || 'Request payload parameters failed backend validation.',
      remediation: 'Check required request fields, DTO rules, and JSON types.',
    };
  }

  if (statusCode === 502 || statusCode === 503) {
    return {
      category: 'GATEWAY_ERROR',
      title: 'Gateway / Bad Gateway Error',
      message: 'Gateway proxy failed to forward request to upstream service.',
      remediation: 'Check Traefik routes.yml service upstream URL.',
    };
  }

  if (statusCode >= 500) {
    return {
      category: 'SERVER_ERROR',
      title: '500 Internal Server Error',
      message: message || 'Backend encountered an unhandled exception or panic.',
      remediation: 'Inspect backend stdout/stderr logs for stacktrace.',
    };
  }

  return {
    category: 'UNKNOWN',
    title: `HTTP ${statusCode}`,
    message: message,
  };
}
