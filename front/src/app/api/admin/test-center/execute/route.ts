import { NextRequest, NextResponse } from 'next/server';
import { NOTIFIER_OPERATIONS_CATALOG } from '@/features/test-center/lib/catalog';
import { redactSensitiveData } from '@/features/test-center/lib/redaction';
import { classifyError } from '@/features/test-center/lib/error-normalizer';
import crypto from 'crypto';

function signHS256(payload: any, secret: string): string {
  const header = { alg: 'HS256', typ: 'JWT' };
  const base64UrlEncode = (obj: any) => {
    return Buffer.from(JSON.stringify(obj))
      .toString('base64')
      .replace(/=/g, '')
      .replace(/\+/g, '-')
      .replace(/\//g, '_');
  };
  const tokenHeader = base64UrlEncode(header);
  const tokenPayload = base64UrlEncode(payload);
  const signature = crypto
    .createHmac('sha256', secret)
    .update(`${tokenHeader}.${tokenPayload}`)
    .digest('base64')
    .replace(/=/g, '')
    .replace(/\+/g, '-')
    .replace(/\//g, '_');
  return `${tokenHeader}.${tokenPayload}.${signature}`;
}

export async function POST(req: NextRequest) {
  const timestamp = new Date().toISOString();
  const requestId = req.headers.get('x-request-id') || `req-${Math.random().toString(36).substring(2, 9)}`;
  const correlationId = req.headers.get('x-correlation-id') || requestId;

  let body: any;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json(
      { success: false, error: { code: 'BAD_REQUEST', message: 'Invalid JSON payload' } },
      { status: 400 }
    );
  }

  const {
    operationId,
    input,
    pathParams = {},
    queryParams = {},
    authMode = 'current-session',
    customToken,
    headers: reqHeaders = {},
    targetMode = 'direct',
  } = body;

  const operation = NOTIFIER_OPERATIONS_CATALOG.find((op) => op.id === operationId);
  if (!operation) {
    return NextResponse.json(
      { success: false, error: { code: 'BAD_REQUEST', message: `Unknown operation ID: ${operationId}` } },
      { status: 400 }
    );
  }

  // Resolve Base URL dynamically
  const directBackendUrl =
    process.env.NOTIFIER_BACKEND_URL ||
    process.env.NEXT_PUBLIC_NOTIFIER_API_URL ||
    'http://127.0.0.1:9002';
  const gatewayUrl = process.env.GATEWAY_API_URL || 'http://127.0.0.1:8080';

  const baseUrl = targetMode === 'gateway' ? gatewayUrl : directBackendUrl.replace(/\/v1\/?$/, '');

  // Substitute path parameters
  let resolvedPath = operation.path;
  for (const [key, val] of Object.entries(pathParams)) {
    resolvedPath = resolvedPath.replace(`:${key}`, String(val));
  }

  // Append query parameters
  const queryParts: string[] = [];
  for (const [key, val] of Object.entries(queryParams)) {
    if (val !== undefined && val !== '') {
      queryParts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(val))}`);
    }
  }
  const fullPath = queryParts.length > 0 ? `${resolvedPath}?${queryParts.join('&')}` : resolvedPath;
  const targetUrl = `${baseUrl}${fullPath}`;

  // Prepare Headers
  const finalHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Request-ID': requestId,
    'X-Correlation-ID': correlationId,
    'X-Tenant-ID': reqHeaders['X-Tenant-ID'] || reqHeaders['x-tenant-id'] || 'tenant-default',
  };

  // Trace propagation headers
  const traceparent = req.headers.get('traceparent');
  if (traceparent) finalHeaders['traceparent'] = traceparent;

  // Custom allowed headers
  const allowedHeaders = ['x-tenant-id', 'idempotency-key', 'accept-language', 'x-language', 'x-correlation-id'];
  for (const [k, v] of Object.entries(reqHeaders)) {
    if (allowedHeaders.includes(k.toLowerCase()) && typeof v === 'string') {
      finalHeaders[k] = v;
    }
  }

  // Auth Header Integration
  const clientAuthHeader = req.headers.get('authorization');
  if (authMode === 'current-session') {
    if (clientAuthHeader) {
      finalHeaders['Authorization'] = clientAuthHeader;
    } else {
      const fallbackPayload = {
        userId: "00000000-0000-0000-0000-000000000001",
        tenantId: "00000000-0000-0000-0000-000000000001",
        email: "admin@minisource.dev",
        roles: ["admin", "super_admin"],
        permissions: ["notifier:admin", "notifier:templates:manage"],
        sessionId: "00000000-0000-0000-0000-000000000001",
        tokenType: "access",
        iss: "minisource-auth",
        sub: "00000000-0000-0000-0000-000000000001",
        aud: ["minisource"],
        exp: 2524608000
      };
      const secret = process.env.AUTH_JWT_SECRET || 'change-me-in-production';
      const token = signHS256(fallbackPayload, secret);
      finalHeaders['Authorization'] = `Bearer ${token}`;
    }
  } else if (authMode === 'bearer-token' && customToken) {
    finalHeaders['Authorization'] = customToken.startsWith('Bearer ') ? customToken : `Bearer ${customToken}`;
  }

  const sanitizedRequestHeaders = redactSensitiveData(finalHeaders);
  const sanitizedRequestBody = redactSensitiveData(input);

  // Execute Request
  const startTime = performance.now();
  let statusCode = 500;
  let responseData: any = null;
  let responseHeaders: Record<string, string> = {};
  let duration = 0;
  let isSuccess = false;
  let errorMsg = '';

  try {
    const fetchOptions: RequestInit = {
      method: operation.method,
      headers: finalHeaders,
      body: operation.method !== 'GET' && input ? JSON.stringify(input) : undefined,
    };

    const res = await fetch(targetUrl, fetchOptions);
    statusCode = res.status;
    duration = Math.round(performance.now() - startTime);

    res.headers.forEach((v, k) => {
      const lower = k.toLowerCase();
      if (['content-type', 'x-request-id', 'x-correlation-id', 'traceparent', 'x-tenant-id'].includes(lower)) {
        responseHeaders[k] = v;
      }
    });

    const contentType = res.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      responseData = await res.json();
    } else {
      const text = await res.text();
      responseData = { rawText: text };
    }

    isSuccess = statusCode >= 200 && statusCode < 300;
  } catch (err: any) {
    duration = Math.round(performance.now() - startTime);
    statusCode = 0;
    errorMsg = err.message || 'Fetch execution failed';
  }

  const sanitizedResponseHeaders = redactSensitiveData(responseHeaders);
  const sanitizedResponseBody = redactSensitiveData(responseData);
  const errorClassification = isSuccess ? undefined : classifyError(statusCode, responseData, errorMsg);

  // Audit Log Entry
  const auditEvent = {
    event: 'NOTIFIER_TEST_CENTER_EXECUTION',
    timestamp,
    operationId: operation.id,
    targetUrl,
    method: operation.method,
    statusCode,
    durationMs: duration,
    isSuccess,
    requestId,
    correlationId,
  };
  console.log('[Notifier Test Center Audit]', JSON.stringify(auditEvent));

  return NextResponse.json({
    success: isSuccess,
    statusCode,
    duration,
    requestId,
    correlationId,
    request: {
      url: fullPath,
      method: operation.method,
      headers: sanitizedRequestHeaders,
      body: sanitizedRequestBody,
    },
    response: {
      headers: sanitizedResponseHeaders,
      body: sanitizedResponseBody,
    },
    errorClassification,
  });
}
