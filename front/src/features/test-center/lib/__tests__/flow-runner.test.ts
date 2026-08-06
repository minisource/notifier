import { describe, it, expect } from 'vitest';
import { extractPathValue, substituteContextVars, evaluateAssertions } from '../flow-runner';
import { DiagnosticResult } from '../../types';

describe('Flow Runner Engine', () => {
  it('should extract values by dot-separated JSON path', () => {
    const obj = {
      data: {
        id: 'notif-12345',
        nested: {
          key: 'welcome_v1',
        },
      },
    };

    expect(extractPathValue(obj, 'data.id')).toBe('notif-12345');
    expect(extractPathValue(obj, 'data.nested.key')).toBe('welcome_v1');
    expect(extractPathValue(obj, 'nonexistent.field')).toBeUndefined();
  });

  it('should substitute context variables in nested objects & templates', () => {
    const context = {
      providerId: 'prov-99',
      templateKey: 'order_conf_v1',
    };

    const targetBody = {
      provider: '{{providerId}}',
      template: '{{templateKey}}',
      staticVal: 'hello',
    };

    const substituted = substituteContextVars(targetBody, context);
    expect(substituted.provider).toBe('prov-99');
    expect(substituted.template).toBe('order_conf_v1');
    expect(substituted.staticVal).toBe('hello');
  });

  it('should evaluate assertions against HTTP status and body fields', () => {
    const result: DiagnosticResult = {
      success: true,
      statusCode: 200,
      duration: 45,
      requestId: 'req-001',
      request: { url: '/v1/health', method: 'GET', headers: {}, body: null },
      response: {
        headers: {},
        body: { data: { status: 'healthy', count: 10 } },
      },
    };

    const failures = evaluateAssertions(result, [
      { field: 'statusCode', operator: 'equals', expected: 200 },
      { field: 'response.body.data.status', operator: 'equals', expected: 'healthy' },
      { field: 'response.body.data.count', operator: 'in', expected: [10, 20] },
    ]);

    expect(failures).toHaveLength(0);
  });

  it('should detect failing assertions', () => {
    const result: DiagnosticResult = {
      success: false,
      statusCode: 400,
      duration: 30,
      requestId: 'req-002',
      request: { url: '/v1/send', method: 'POST', headers: {}, body: null },
      response: { headers: {}, body: { error: 'Invalid payload' } },
    };

    const failures = evaluateAssertions(result, [
      { field: 'statusCode', operator: 'equals', expected: 200 },
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("expected '200', got '400'");
  });
});
