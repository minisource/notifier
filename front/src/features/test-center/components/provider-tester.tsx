'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { Server, Activity, Plus, ListFilter } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface ProviderTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function ProviderTester({ onExecute, executing }: ProviderTesterProps) {
  const [providerId, setProviderId] = useState('');

  const handleListProviders = () => {
    const op: ApprovedOperation = {
      id: 'providers.list',
      domain: 'Providers',
      name: 'List Notification Providers',
      description: 'Retrieves all configured SMS, Email, and Push providers',
      method: 'GET',
      path: '/v1/providers',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleGetProvider = () => {
    if (!providerId) return;
    const op: ApprovedOperation = {
      id: 'providers.get',
      domain: 'Providers',
      name: 'Get Provider Details',
      description: 'Fetches detailed metadata for a provider',
      method: 'GET',
      path: '/v1/providers/:providerId',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, { providerId }, {}, undefined);
  };

  const handleTestConnectivity = () => {
    if (!providerId) return;
    const op: ApprovedOperation = {
      id: 'providers.test',
      domain: 'Providers',
      name: 'Test Provider Connectivity',
      description: 'Validates credentials and connection without sending real message',
      method: 'POST',
      path: '/v1/providers/:providerId/test',
      safetyClass: 'SAFE VALIDATION',
      requiresAuth: true,
    };
    onExecute(op, { providerId }, {}, undefined);
  };

  const handleCreateTestProvider = () => {
    const op: ApprovedOperation = {
      id: 'providers.create',
      domain: 'Providers',
      name: 'Create Notification Provider',
      description: 'Registers a new provider configuration',
      method: 'POST',
      path: '/v1/providers',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };

    const provKey = `test_prov_${Math.random().toString(36).substring(2, 7)}`;
    onExecute(
      op,
      {},
      {},
      {
        name: `SMTP Test Provider ${provKey}`,
        type: 'email',
        provider_key: provKey,
        is_active: true,
        priority: 10,
        config: {
          host: 'smtp.mailtrap.io',
          port: 2525,
          from_email: 'test@minisource.dev',
        },
      }
    );
  };

  return (
    <div className="space-y-6">
      {/* Provider Quick Actions */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Server className="h-4 w-4 text-primary" /> Provider Overview & Creation
          </CardTitle>
          <CardDescription className="text-xs">
            List configured delivery providers or create a new test provider (`/v1/providers`).
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleListProviders}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> List All Providers
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleCreateTestProvider}
              disabled={executing}
            >
              <Plus className="h-3.5 w-3.5 shrink-0" /> Create Test SMTP Provider
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Provider Test & Inspection */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Activity className="h-4 w-4 text-primary" /> Provider Connection Diagnostics
          </CardTitle>
          <CardDescription className="text-xs">
            Test provider credentials and network handshake without sending live messages.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="space-y-1">
            <Label className="text-xs font-medium">Provider ID</Label>
            <Input
              value={providerId}
              onChange={(e) => setProviderId(e.target.value)}
              placeholder="e.g. prov-001 or uuid"
              className="h-8 text-xs font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetProvider}
              disabled={executing || !providerId}
            >
              <Server className="h-3.5 w-3.5 text-primary shrink-0" /> Get Provider Metadata
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5 bg-purple-600 hover:bg-purple-700 text-white"
              onClick={handleTestConnectivity}
              disabled={executing || !providerId}
            >
              <Activity className="h-3.5 w-3.5 shrink-0" /> Test Provider Connection
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
