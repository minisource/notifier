'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { FileText, Eye, Plus, ListFilter, Search } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface TemplateTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function TemplateTester({ onExecute, executing }: TemplateTesterProps) {
  const [templateKey, setTemplateKey] = useState('welcome_user_v1');
  const [previewVarsJson, setPreviewVarsJson] = useState('{\n  "userName": "Alice Smith",\n  "companyName": "Acme Corp"\n}');

  const handleListTemplates = () => {
    const op: ApprovedOperation = {
      id: 'templates.list',
      domain: 'Templates',
      name: 'List Notification Templates',
      description: 'Retrieves all registered templates',
      method: 'GET',
      path: '/v1/templates',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleGetByKey = () => {
    if (!templateKey) return;
    const op: ApprovedOperation = {
      id: 'templates.get_by_key',
      domain: 'Templates',
      name: 'Get Template By Key',
      description: 'Retrieves template definition by key',
      method: 'GET',
      path: '/v1/templates/key/:key',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, { key: templateKey }, {}, undefined);
  };

  const handleRenderPreview = () => {
    if (!templateKey) return;
    let varsObj = {};
    try {
      varsObj = JSON.parse(previewVarsJson);
    } catch {
      // fallback
    }

    const op: ApprovedOperation = {
      id: 'templates.render_preview',
      domain: 'Templates',
      name: 'Render Template Preview',
      description: 'Performs dry-run rendering and variable replacement',
      method: 'POST',
      path: '/v1/templates/render-preview',
      safetyClass: 'SAFE VALIDATION',
      requiresAuth: true,
    };

    onExecute(op, {}, {}, { key: templateKey, variables: varsObj });
  };

  const handleCreateTestTemplate = () => {
    const op: ApprovedOperation = {
      id: 'templates.create',
      domain: 'Templates',
      name: 'Create Notification Template',
      description: 'Registers a new email/SMS/push template',
      method: 'POST',
      path: '/v1/templates',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };

    const newKey = `test_tmpl_${Math.random().toString(36).substring(2, 7)}`;
    onExecute(
      op,
      {},
      {},
      {
        key: newKey,
        name: `Test Template ${newKey}`,
        channel: 'email',
        subject: 'Hello {{userName}} from MiniSource',
        content: 'Your account balance is {{balance}} USD.',
        is_active: true,
        variables: ['userName', 'balance'],
      }
    );
  };

  return (
    <div className="space-y-6">
      {/* Template Quick Actions */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <FileText className="h-4 w-4 text-primary" /> Template Management & List
          </CardTitle>
          <CardDescription className="text-xs">
            Query registered templates or create a test template definition (`/v1/templates`).
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleListTemplates}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> List All Templates
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleCreateTestTemplate}
              disabled={executing}
            >
              <Plus className="h-3.5 w-3.5 shrink-0" /> Create Random Test Template
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Render Preview & Key Lookup */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Eye className="h-4 w-4 text-primary" /> Dry-Run Render Preview & Key Lookup
          </CardTitle>
          <CardDescription className="text-xs">
            Perform live template rendering and variable substitution without side-effects.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="space-y-1">
            <Label className="text-xs font-medium">Template Key</Label>
            <Input
              value={templateKey}
              onChange={(e) => setTemplateKey(e.target.value)}
              placeholder="e.g. welcome_user_v1"
              className="h-8 text-xs font-mono"
            />
          </div>

          <div className="space-y-1">
            <Label className="text-xs font-medium">Variables Payload (JSON)</Label>
            <textarea
              value={previewVarsJson}
              onChange={(e) => setPreviewVarsJson(e.target.value)}
              rows={3}
              className="w-full p-2.5 rounded-md border border-input bg-background text-xs font-mono focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetByKey}
              disabled={executing || !templateKey}
            >
              <Search className="h-3.5 w-3.5 text-primary shrink-0" /> Get Template by Key
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5 bg-blue-600 hover:bg-blue-700 text-white"
              onClick={handleRenderPreview}
              disabled={executing || !templateKey}
            >
              <Eye className="h-3.5 w-3.5 shrink-0" /> Render Dry-Run Preview
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
