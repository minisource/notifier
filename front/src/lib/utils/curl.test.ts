import { describe, expect, it } from 'vitest';
import { buildCurlCommand } from './curl';

describe('buildCurlCommand', () => {
  it('returns null without a method or url', () => {
    expect(buildCurlCommand({})).toBeNull();
    expect(buildCurlCommand({ method: 'POST' })).toBeNull();
    expect(buildCurlCommand({ url: 'https://x.test' })).toBeNull();
  });

  it('builds a POST form command with inferred content-type', () => {
    const cmd = buildCurlCommand({
      method: 'POST',
      url: 'https://api.kavenegar.com/v1/[REDACTED]/sms/send.json',
      body: 'receptor=%2B98%2A%2A%2A41&template=verify&token=[REDACTED]',
    });
    expect(cmd).toContain('curl --request POST');
    expect(cmd).toContain('--url "https://api.kavenegar.com/v1/[REDACTED]/sms/send.json"');
    expect(cmd).toContain('-H "Content-Type: application/x-www-form-urlencoded"');
    expect(cmd).toContain(`--data-raw 'receptor=%2B98%2A%2A%2A41&template=verify&token=[REDACTED]'`);
  });

  it('builds a JSON command with --data and json content-type', () => {
    const cmd = buildCurlCommand({
      method: 'POST',
      url: '/v1/admin/notifications',
      headers: { Authorization: 'Bearer <token>' },
      body: '{"channel":"sms"}',
      baseUrl: 'http://localhost:8080',
    });
    expect(cmd).toContain('--url "http://localhost:8080/v1/admin/notifications"');
    expect(cmd).toContain('-H "Authorization: Bearer <token>"');
    expect(cmd).toContain('-H "Content-Type: application/json"');
    expect(cmd).toContain(`--data '{"channel":"sms"}'`);
  });

  it('keeps absolute URLs verbatim and ignores baseUrl', () => {
    const cmd = buildCurlCommand({
      method: 'GET',
      url: 'https://api.kavenegar.com/v1/[REDACTED]/account/info.json',
      baseUrl: 'http://localhost:8080',
    });
    expect(cmd).toContain('--url "https://api.kavenegar.com/v1/[REDACTED]/account/info.json"');
    expect(cmd).not.toContain('localhost');
  });

  it('omits body for GET requests', () => {
    const cmd = buildCurlCommand({
      method: 'GET',
      url: 'https://api.example.com/status',
      body: 'should-not-appear',
    });
    expect(cmd).not.toContain('--data');
  });

  it('returns null for non-HTTP pseudo URLs (e.g. smtp://)', () => {
    expect(
      buildCurlCommand({ method: 'SMTP', url: 'smtp://provider/send' })
    ).toBeNull();
  });

  it('escapes quotes in body and header values', () => {
    const cmd = buildCurlCommand({
      method: 'POST',
      url: 'https://api.example.com/hook',
      headers: { 'X-Note': 'say "hi"' },
      body: "it's a test",
    });
    expect(cmd).toContain('-H "X-Note: say \\"hi\\""');
    // bash single-quote escape for the apostrophe
    expect(cmd).toContain(`--data-raw 'it'\\''s a test'`);
  });

  it('escapes double quotes inside the URL', () => {
    const cmd = buildCurlCommand({
      method: 'GET',
      url: 'https://api.example.com/hook?q="quoted"',
    });
    // String.raw keeps the literal backslash-quote: the curl output contains
    //  --url "https://api.example.com/hook?q=\"quoted\""  (escaped quotes).
    expect(cmd).toContain(String.raw`q=\"quoted\"`);
  });
});
