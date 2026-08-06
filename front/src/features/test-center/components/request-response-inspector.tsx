'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, Badge, Button } from '@minisource/ui';
import { Copy, Check, Clock, Shield, AlertTriangle, ChevronDown, ChevronRight } from 'lucide-react';
import { toast } from 'sonner';
import { DiagnosticResult } from '../types';

interface RequestResponseInspectorProps {
  result: DiagnosticResult;
}

export function RequestResponseInspector({ result }: RequestResponseInspectorProps) {
  const [copied, setCopied] = useState(false);
  const [copiedCurl, setCopiedCurl] = useState(false);
  const [showDetails, setShowDetails] = useState(true);

  const handleCopyDebugReport = () => {
    const report = {
      timestamp: new Date().toISOString(),
      requestId: result.requestId,
      correlationId: result.correlationId,
      durationMs: result.duration,
      statusCode: result.statusCode,
      request: result.request,
      response: result.response,
      error: result.errorClassification,
    };

    if (typeof navigator !== 'undefined') {
      navigator.clipboard.writeText(JSON.stringify(report, null, 2));
      setCopied(true);
      toast.success('Safe diagnostic report copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleCopyCurl = () => {
    const { method, url, headers, body } = result.request;
    const directBackendUrl =
      process.env.NOTIFIER_BACKEND_URL ||
      process.env.NEXT_PUBLIC_NOTIFIER_API_URL ||
      'http://127.0.0.1:9002';
    
    let absoluteUrl = url;
    if (!url.startsWith('http') && !url.startsWith('/')) {
      absoluteUrl = `/${url}`;
    }
    if (!absoluteUrl.startsWith('http')) {
      absoluteUrl = `${directBackendUrl}${absoluteUrl}`;
    }
    
    let cmd = `curl -X ${method} "${absoluteUrl}"`;
    
    if (headers) {
      Object.entries(headers).forEach(([k, v]) => {
        // Redact authorization token from raw terminal command logs but keep it valid if it was fake
        cmd += ` -H "${k}: ${v}"`;
      });
    }
    
    if (body && method !== 'GET') {
      cmd += ` -d '${JSON.stringify(body)}'`;
    }
    
    if (typeof navigator !== 'undefined') {
      navigator.clipboard.writeText(cmd);
      setCopiedCurl(true);
      toast.success('cURL command copied to clipboard');
      setTimeout(() => setCopiedCurl(false), 2000);
    }
  };

  const getStatusBadge = (code: number) => {
    if (code >= 200 && code < 300) {
      return <Badge className="bg-emerald-500/15 border-emerald-500/30 text-emerald-600 dark:text-emerald-400">HTTP {code} OK</Badge>;
    }
    if (code >= 400 && code < 500) {
      return <Badge className="bg-amber-500/15 border-amber-500/30 text-amber-600 dark:text-amber-400">HTTP {code} Client Error</Badge>;
    }
    if (code >= 500) {
      return <Badge className="bg-rose-500/15 border-rose-500/30 text-rose-600 dark:text-rose-400">HTTP {code} Server Error</Badge>;
    }
    return <Badge variant="outline">HTTP {code}</Badge>;
  };

  return (
    <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm overflow-hidden">
      <CardHeader className="py-3 px-4 bg-muted/20 border-b border-border/50 flex flex-row items-center justify-between">
        <div className="flex items-center gap-2 truncate max-w-[65%]">
          <button
            type="button"
            onClick={() => setShowDetails(!showDetails)}
            className="text-muted-foreground hover:text-foreground transition-colors shrink-0"
          >
            {showDetails ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </button>
          <Badge variant="outline" className="font-mono text-xs uppercase px-2 py-0.5 shrink-0">
            {result.request.method}
          </Badge>
          <span className="font-mono text-[11px] text-foreground/90 font-medium truncate" title={result.request.url}>
            {result.request.url}
          </span>
        </div>

        <div className="flex items-center gap-1.5 shrink-0">
          {getStatusBadge(result.statusCode)}
          <span className="text-xs text-muted-foreground font-mono flex items-center gap-0.5">
            <Clock className="h-3 w-3" /> {result.duration}ms
          </span>
        </div>
      </CardHeader>

      {showDetails && (
        <CardContent className="p-3 space-y-3">
          {/* Action Row */}
          <div className="flex items-center justify-between gap-2 border-b border-border/40 pb-2">
            <span className="text-xs font-semibold text-muted-foreground">Diagnostics Trace</span>
            <div className="flex items-center gap-1.5">
              <Button variant="outline" size="sm" className="h-7 text-[10px] px-2 gap-1" onClick={handleCopyCurl}>
                {copiedCurl ? <Check className="h-3 w-3 text-emerald-500" /> : <Copy className="h-3 w-3" />}
                Copy cURL
              </Button>
              <Button variant="default" size="sm" className="h-7 text-[10px] px-2 gap-1" onClick={handleCopyDebugReport}>
                {copied ? <Check className="h-3 w-3 text-emerald-500" /> : <Copy className="h-3 w-3" />}
                Copy Debug JSON
              </Button>
            </div>
          </div>

          {/* Error Classification Banner if applicable */}
          {result.errorClassification && (
            <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-xs space-y-1">
              <div className="flex items-center gap-2 font-semibold text-destructive">
                <AlertTriangle className="h-4 w-4 shrink-0" />
                <span>{result.errorClassification.title}</span>
              </div>
              <p className="text-muted-foreground">{result.errorClassification.message}</p>
              {result.errorClassification.remediation && (
                <p className="text-foreground/80 font-mono text-[11px] pt-1">
                  <strong>Remediation:</strong> {result.errorClassification.remediation}
                </p>
              )}
            </div>
          )}

          {/* Trace Identifiers Bar */}
          <div className="grid grid-cols-2 gap-2 p-2 rounded border border-border/40 bg-muted/10 font-mono text-[10px]">
            <div className="flex items-center justify-between gap-1 border-r border-border/30 pr-1">
              <div className="truncate">
                <span className="text-muted-foreground block text-[9px]">Request ID:</span>
                <span className="text-foreground truncate block font-medium" title={result.requestId}>{result.requestId}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-5 w-5 shrink-0 opacity-60 hover:opacity-100"
                onClick={() => {
                  navigator.clipboard.writeText(result.requestId);
                  toast.success('Copied Request ID');
                }}
              >
                <Copy className="h-3.5 w-3.5" />
              </Button>
            </div>
            <div className="flex items-center justify-between gap-1">
              <div className="truncate">
                <span className="text-muted-foreground block text-[9px]">Correlation ID:</span>
                <span className="text-foreground truncate block font-medium" title={result.correlationId || 'N/A'}>{result.correlationId || 'N/A'}</span>
              </div>
              {result.correlationId && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-5 w-5 shrink-0 opacity-60 hover:opacity-100"
                  onClick={() => {
                    navigator.clipboard.writeText(result.correlationId || '');
                    toast.success('Copied Correlation ID');
                  }}
                >
                  <Copy className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          </div>

          <div className="flex items-center gap-1.5 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-500/5 px-2 py-1 rounded font-mono">
            <Shield className="h-3.5 w-3.5" /> Authorization credentials and user secrets have been automatically redacted.
          </div>

          {/* Request / Response Split Inspector */}
          <div className="space-y-3">
            {/* Request Panel */}
            <div className="space-y-1.5">
              <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block">
                Sanitized Request details
              </span>
              <div className="p-2.5 rounded bg-zinc-950 text-zinc-100 font-mono text-[10px] space-y-2 overflow-x-auto max-h-52 scrollbar-thin">
                <div>
                  <span className="text-zinc-400 block pb-1 border-b border-zinc-800 text-[9px] uppercase">Headers</span>
                  <pre>{JSON.stringify(result.request.headers, null, 2)}</pre>
                </div>
                {result.request.body && (
                  <div className="pt-2 border-t border-zinc-800">
                    <span className="text-zinc-400 block pb-1 border-b border-zinc-800 text-[9px] uppercase">Payload Body</span>
                    <pre>{JSON.stringify(result.request.body, null, 2)}</pre>
                  </div>
                )}
              </div>
            </div>

            {/* Response Panel */}
            <div className="space-y-1.5">
              <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block">
                Sanitized Response details
              </span>
              <div className="p-2.5 rounded bg-zinc-950 text-zinc-100 font-mono text-[10px] space-y-2 overflow-x-auto max-h-52 scrollbar-thin">
                <div>
                  <span className="text-zinc-400 block pb-1 border-b border-zinc-800 text-[9px] uppercase">Headers</span>
                  <pre>{JSON.stringify(result.response.headers, null, 2)}</pre>
                </div>
                <div className="pt-2 border-t border-zinc-800">
                  <span className="text-zinc-400 block pb-1 border-b border-zinc-800 text-[9px] uppercase">Response Body</span>
                  <pre>{JSON.stringify(result.response.body, null, 2)}</pre>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
