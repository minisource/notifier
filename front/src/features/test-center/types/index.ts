export type OperationSafetyClass =
  | 'SAFE READ'
  | 'SAFE VALIDATION'
  | 'LOCAL MUTATION'
  | 'EXTERNAL DELIVERY'
  | 'DESTRUCTIVE'
  | 'ADMIN/SECURITY-SENSITIVE'
  | 'NOT TESTABLE FROM BROWSER';

export type AuthMode = 'current-session' | 'bearer-token' | 'none';

export interface ApprovedOperation {
  id: string;
  domain: string;
  name: string;
  description: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  path: string; // e.g. /v1/notifications
  safetyClass: OperationSafetyClass;
  requiresAuth: boolean;
  defaultHeaders?: Record<string, string>;
  defaultBody?: any;
  queryParams?: Array<{ key: string; label: string; placeholder?: string; required?: boolean }>;
  pathParams?: Array<{ key: string; label: string; placeholder?: string; required?: boolean }>;
}

export interface HeaderPair {
  key: string;
  value: string;
}

export interface RequestConfig {
  authMode: AuthMode;
  customToken?: string;
  headers: HeaderPair[];
  targetMode?: 'direct' | 'gateway';
}

export interface DiagnosticRequestInfo {
  url: string;
  method: string;
  headers: Record<string, string>;
  body: any;
}

export interface DiagnosticResponseInfo {
  headers: Record<string, string>;
  body: any;
}

export interface DiagnosticResult {
  success: boolean;
  statusCode: number;
  duration: number;
  requestId: string;
  correlationId?: string;
  traceId?: string;
  request: DiagnosticRequestInfo;
  response: DiagnosticResponseInfo;
  errorClassification?: ErrorClassification;
}

export type ErrorCategory =
  | 'CLIENT_VALIDATION'
  | 'NETWORK_ERROR'
  | 'TIMEOUT'
  | 'CORS_FAILURE'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'RATE_LIMIT'
  | 'BACKEND_VALIDATION'
  | 'SERVER_ERROR'
  | 'GATEWAY_ERROR'
  | 'CONTRACT_MISMATCH'
  | 'UNKNOWN';

export interface ErrorClassification {
  category: ErrorCategory;
  title: string;
  message: string;
  remediation?: string;
}

export interface ExecutionHistoryItem {
  id: string;
  timestamp: string;
  operationId: string;
  operationName: string;
  domain: string;
  success: boolean;
  statusCode: number;
  duration: number;
  safetyClass: OperationSafetyClass;
  request: DiagnosticRequestInfo;
  response: DiagnosticResponseInfo;
  requestId?: string;
  correlationId?: string;
}

export type StepState = 'idle' | 'running' | 'success' | 'failed' | 'cancelled' | 'skipped';

export interface FlowStepDefinition {
  id: string;
  title: string;
  description: string;
  operationId: string;
  pathParams?: Record<string, string>; // May contain dynamic references e.g. {{providerId}}
  queryParams?: Record<string, string>;
  body?: any; // May contain dynamic references e.g. {{templateKey}}
  extractValues?: Array<{
    targetVar: string;
    jsonPath: string; // e.g. "data.id" or "data.key" or "id"
  }>;
  assertions?: Array<{
    field: string; // e.g. "statusCode" or "response.body.status" or "response.body.data.id"
    operator: 'equals' | 'contains' | 'exists' | 'gte' | 'lte' | 'in';
    expected: any;
  }>;
  allowFailure?: boolean;
}

export interface FlowDefinition {
  id: string;
  name: string;
  domain: string;
  description: string;
  safetyClass: OperationSafetyClass;
  requiresConfirmation?: boolean;
  steps: FlowStepDefinition[];
  cleanupSteps?: FlowStepDefinition[];
}

export interface FlowStepResult {
  stepId: string;
  title: string;
  state: StepState;
  duration?: number;
  statusCode?: number;
  result?: DiagnosticResult;
  extractedValues?: Record<string, any>;
  assertionFailures?: string[];
  errorMessage?: string;
}

export interface FlowExecutionState {
  flowId: string;
  state: 'idle' | 'running' | 'passed' | 'failed' | 'cancelled';
  currentStepIndex: number;
  stepResults: FlowStepResult[];
  extractedContext: Record<string, any>;
  createdResourceIds: Array<{ type: string; id: string }>;
  cleanupResults?: FlowStepResult[];
  startTime?: string;
  endTime?: string;
  totalDuration?: number;
}
