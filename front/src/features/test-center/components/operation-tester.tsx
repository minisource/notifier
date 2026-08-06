'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label, Badge } from '@minisource/ui';
import { Play } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface OperationTesterProps {
  operations: ApprovedOperation[];
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function OperationTester({ operations, onExecute, executing }: OperationTesterProps) {
  const [selectedOpId, setSelectedOpId] = useState(operations[0]?.id || '');
  const [pathParamValues, setPathParamValues] = useState<Record<string, string>>({});
  const [queryParamValues, setQueryParamValues] = useState<Record<string, string>>({});
  const [jsonBodyText, setJsonBodyText] = useState('');
  const [jsonError, setJsonError] = useState<string | null>(null);

  const selectedOp = operations.find((op) => op.id === selectedOpId) || operations[0];

  const handleSelectOp = (opId: string) => {
    setSelectedOpId(opId);
    const targetOp = operations.find((o) => o.id === opId);
    if (targetOp) {
      setPathParamValues({});
      setQueryParamValues({});
      setJsonBodyText(targetOp.defaultBody ? JSON.stringify(targetOp.defaultBody, null, 2) : '');
      setJsonError(null);
    }
  };

  const handleRun = async () => {
    let parsedBody: any = undefined;
    if (jsonBodyText.trim() && selectedOp.method !== 'GET') {
      try {
        parsedBody = JSON.parse(jsonBodyText);
        setJsonError(null);
      } catch (err: any) {
        setJsonError('Invalid JSON body: ' + err.message);
        return;
      }
    }

    await onExecute(selectedOp, pathParamValues, queryParamValues, parsedBody);
  };

  const getSafetyBadge = (safetyClass: string) => {
    switch (safetyClass) {
      case 'SAFE READ':
        return <Badge className="bg-blue-500/15 border-blue-500/30 text-blue-600 dark:text-blue-400">SAFE READ</Badge>;
      case 'SAFE VALIDATION':
        return <Badge className="bg-emerald-500/15 border-emerald-500/30 text-emerald-600 dark:text-emerald-400">SAFE VALIDATION</Badge>;
      case 'LOCAL MUTATION':
        return <Badge className="bg-amber-500/15 border-amber-500/30 text-amber-600 dark:text-amber-400">LOCAL MUTATION</Badge>;
      case 'EXTERNAL DELIVERY':
        return <Badge className="bg-purple-500/15 border-purple-500/30 text-purple-600 dark:text-purple-400">EXTERNAL DELIVERY</Badge>;
      default:
        return <Badge variant="outline">{safetyClass}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Play className="h-4 w-4 text-primary" /> Individual API Operation Diagnostics
            </span>
            {selectedOp && getSafetyBadge(selectedOp.safetyClass)}
          </CardTitle>
          <CardDescription className="text-xs">
            Select an approved operation from the Notifier API catalog to inspect request parameters and live responses.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4 pt-4">
          {/* Operation Selector & Domain Filter */}
          <div className="space-y-1.5">
            <Label className="text-xs font-medium">Select API Operation</Label>
            <select
              value={selectedOpId}
              onChange={(e) => handleSelectOp(e.target.value)}
              className="w-full h-9 px-3 rounded-lg border border-input bg-background text-xs font-mono focus:outline-none focus:ring-2 focus:ring-ring"
            >
              {operations.map((op) => (
                <option key={op.id} value={op.id}>
                  [{op.domain}] {op.method} {op.path} — {op.name}
                </option>
              ))}
            </select>
          </div>

          {selectedOp && (
            <div className="p-3 rounded-lg bg-muted/20 border border-border/40 space-y-1">
              <p className="text-xs text-foreground/90">{selectedOp.description}</p>
              <div className="flex items-center gap-2 font-mono text-[11px] text-muted-foreground pt-1">
                <span>Method: <strong>{selectedOp.method}</strong></span>
                <span>•</span>
                <span>Path: <strong>{selectedOp.path}</strong></span>
              </div>
            </div>
          )}

          {/* Dynamic Path Parameters */}
          {selectedOp?.pathParams && selectedOp.pathParams.length > 0 && (
            <div className="space-y-2 pt-2 border-t border-border/40">
              <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Path Parameters
              </Label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {selectedOp.pathParams.map((param) => (
                  <div key={param.key} className="space-y-1">
                    <Label className="text-[11px]">{param.label} (:{param.key})</Label>
                    <Input
                      placeholder={param.placeholder || param.key}
                      value={pathParamValues[param.key] || ''}
                      onChange={(e) =>
                        setPathParamValues((prev) => ({ ...prev, [param.key]: e.target.value }))
                      }
                      className="h-8 text-xs font-mono"
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* JSON Body Editor for POST/PUT/PATCH */}
          {selectedOp?.method !== 'GET' && (
            <div className="space-y-1.5 pt-2 border-t border-border/40">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  Request Payload Body (JSON)
                </Label>
                {jsonError && <span className="text-[11px] text-rose-500 font-mono">{jsonError}</span>}
              </div>
              <textarea
                value={jsonBodyText}
                onChange={(e) => {
                  setJsonBodyText(e.target.value);
                  setJsonError(null);
                }}
                rows={7}
                className="w-full p-3 rounded-lg border border-input bg-zinc-950 text-zinc-100 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-ring scrollbar-thin"
                placeholder="{}"
              />
            </div>
          )}

          <Button
            onClick={handleRun}
            disabled={executing}
            className="w-full h-10 gap-2 text-xs font-semibold bg-primary hover:bg-primary/90"
          >
            {executing ? 'Executing Request...' : `Execute ${selectedOp?.method} ${selectedOp?.name}`}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
