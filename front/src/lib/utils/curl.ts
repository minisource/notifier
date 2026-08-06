// Builds a copyable curl command from sanitized provider-request data.
//
// The request data in provider-attempt logs is ALREADY sanitized (API keys,
// OTP tokens and other secrets are replaced with [REDACTED], recipients are
// masked). This utility only assembles those sanitized values into a curl
// command the operator can copy, edit, and re-run — it never introduces raw
// secrets. Any [REDACTED] marker the user leaves in place will simply be sent
// as-is (which fails against the provider), so the UI must hint that
// placeholders must be replaced first.

export interface CurlCommandOptions {
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  body?: string;
  // Optional base URL used to make relative request URLs absolute. When the
  // request URL already starts with http(s):// it is used verbatim.
  baseUrl?: string;
}

export function buildCurlCommand(opts: CurlCommandOptions): string | null {
  const { method, url, headers, body, baseUrl } = opts;
  if (!method || !url) return null;

  const verb = method.toUpperCase();

  // Resolve the target URL: keep absolute URLs as-is, otherwise (relative or
  // scheme-less pseudo-URLs like "sms://...") resolve against the base URL.
  let target = url;
  if (!/^https?:\/\//i.test(target)) {
    if (/^[a-z][a-z0-9+.-]*:\/\//i.test(target)) {
      // Non-HTTP pseudo-URL (e.g. "smtp://host/send") — no meaningful curl.
      return null;
    }
    if (baseUrl) {
      target = `${baseUrl.replace(/\/+$/, '')}/${target.replace(/^\/+/, '')}`;
    }
  }

  let cmd = `curl --request ${verb}`;
  cmd += ` \\\n  --url "${escapeDoubleQuotes(target)}"`;

  const hdrs = { ...(headers || {}) };

  // Infer a Content-Type when the request has a body and none was recorded.
  if (body && !Object.keys(hdrs).some((k) => k.toLowerCase() === 'content-type')) {
    const trimmed = body.trim();
    const isJson = trimmed.startsWith('{') || trimmed.startsWith('[');
    hdrs['Content-Type'] = isJson ? 'application/json' : 'application/x-www-form-urlencoded';
  }

  for (const [k, v] of Object.entries(hdrs)) {
    if (!v) continue;
    cmd += ` \\\n  -H "${escapeDoubleQuotes(k)}: ${escapeDoubleQuotes(v)}"`;
  }

  if (body && verb !== 'GET' && verb !== 'HEAD' && verb !== 'OPTIONS') {
    const trimmed = body.trim();
    const isJson = trimmed.startsWith('{') || trimmed.startsWith('[');
    cmd += isJson ? ` \\\n  --data '${escapeSingleQuotes(body)}'` : ` \\\n  --data-raw '${escapeSingleQuotes(body)}'`;
  }

  return cmd;
}

function escapeSingleQuotes(s: string): string {
  // Standard bash single-quote escape: 'foo' -> '\''  (close, escaped quote, reopen)
  return s.replace(/'/g, `'\\''`);
}

function escapeDoubleQuotes(s: string): string {
  return s.replace(/"/g, '\\"');
}
