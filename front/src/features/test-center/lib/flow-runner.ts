import {
  DiagnosticResult,
  FlowStepDefinition,
  FlowStepResult,
} from '../types';

/**
 * Utility to extract a value from a nested JSON object by dot-separated path.
 * e.g. extractPathValue({ data: { id: "123" } }, "data.id") => "123"
 */
export function extractPathValue(obj: any, path: string): any {
  if (!obj || !path) return undefined;
  const parts = path.split('.');
  let curr = obj;
  for (const part of parts) {
    if (curr === null || curr === undefined) return undefined;
    curr = curr[part];
  }
  return curr;
}

/**
 * Replaces placeholders in strings, objects, or arrays with extracted context values.
 * e.g. "Hello {{userName}}" => "Hello Alice"
 */
export function substituteContextVars(target: any, context: Record<string, any>): any {
  if (target === null || target === undefined) return target;

  if (typeof target === 'string') {
    let result = target;
    for (const [key, val] of Object.entries(context)) {
      const placeholder = `{{${key}}}`;
      if (result.includes(placeholder)) {
        result = result.replaceAll(placeholder, String(val ?? ''));
      }
    }
    return result;
  }

  if (Array.isArray(target)) {
    return target.map((item) => substituteContextVars(item, context));
  }

  if (typeof target === 'object') {
    const copy: Record<string, any> = {};
    for (const [k, v] of Object.entries(target)) {
      copy[k] = substituteContextVars(v, context);
    }
    return copy;
  }

  return target;
}

/**
 * Evaluates step assertion rules against HTTP status and response payload.
 */
export function evaluateAssertions(
  result: DiagnosticResult,
  assertions?: FlowStepDefinition['assertions']
): string[] {
  if (!assertions || assertions.length === 0) return [];
  const failures: string[] = [];

  for (const rule of assertions) {
    let actualVal: any;
    if (rule.field === 'statusCode') {
      actualVal = result.statusCode;
    } else if (rule.field.startsWith('response.body.')) {
      const path = rule.field.replace('response.body.', '');
      actualVal = extractPathValue(result.response.body, path);
    } else if (rule.field.startsWith('response.headers.')) {
      const path = rule.field.replace('response.headers.', '');
      actualVal = extractPathValue(result.response.headers, path);
    } else {
      actualVal = extractPathValue(result, rule.field);
    }

    switch (rule.operator) {
      case 'equals':
        if (actualVal !== rule.expected) {
          failures.push(`Assertion failed: ${rule.field} expected '${rule.expected}', got '${actualVal}'`);
        }
        break;
      case 'contains':
        if (typeof actualVal === 'string' && !actualVal.includes(rule.expected)) {
          failures.push(`Assertion failed: ${rule.field} expected to contain '${rule.expected}', got '${actualVal}'`);
        }
        break;
      case 'exists':
        if (actualVal === undefined || actualVal === null) {
          failures.push(`Assertion failed: ${rule.field} expected to exist, got '${actualVal}'`);
        }
        break;
      case 'in':
        if (Array.isArray(rule.expected) && !rule.expected.includes(actualVal)) {
          failures.push(`Assertion failed: ${rule.field} expected to be one of [${rule.expected.join(', ')}], got '${actualVal}'`);
        }
        break;
    }
  }

  return failures;
}

/**
 * Executes a single step in a flow using the server-side execute proxy API route.
 */
export async function executeFlowStep(
  step: FlowStepDefinition,
  context: Record<string, any>,
  requestConfig: any
): Promise<{ stepResult: FlowStepResult; newExtracted: Record<string, any> }> {
  const startTime = performance.now();
  const substitutedPathParams = substituteContextVars(step.pathParams, context);
  const substitutedQueryParams = substituteContextVars(step.queryParams, context);
  const substitutedBody = substituteContextVars(step.body, context);

  try {
    const res = await fetch('/api/admin/test-center/execute', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        operationId: step.operationId,
        input: substitutedBody,
        pathParams: substitutedPathParams,
        queryParams: substitutedQueryParams,
        authMode: requestConfig.authMode,
        customToken: requestConfig.customToken,
        headers: requestConfig.headersMap,
        targetMode: requestConfig.targetMode,
      }),
    });

    const result: DiagnosticResult = await res.json();
    const duration = Math.round(performance.now() - startTime);

    const assertionFailures = evaluateAssertions(result, step.assertions);
    const stepPassed = result.success && assertionFailures.length === 0;

    const newExtracted: Record<string, any> = {};
    if (stepPassed && step.extractValues) {
      for (const extractRule of step.extractValues) {
        const val = extractPathValue(result.response.body, extractRule.jsonPath);
        if (val !== undefined) {
          newExtracted[extractRule.targetVar] = val;
        }
      }
    }

    return {
      stepResult: {
        stepId: step.id,
        title: step.title,
        state: stepPassed ? 'success' : 'failed',
        duration,
        statusCode: result.statusCode,
        result,
        extractedValues: newExtracted,
        assertionFailures,
        errorMessage: stepPassed ? undefined : assertionFailures[0] || result.errorClassification?.message || 'Step execution failed',
      },
      newExtracted,
    };
  } catch (err: any) {
    const duration = Math.round(performance.now() - startTime);
    return {
      stepResult: {
        stepId: step.id,
        title: step.title,
        state: 'failed',
        duration,
        statusCode: 0,
        errorMessage: err.message || 'Network error during step execution',
      },
      newExtracted: {},
    };
  }
}
