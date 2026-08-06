'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button } from '@minisource/ui';
import { Activity, Server, LayoutDashboard, Settings } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface SystemObservabilityTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function SystemObservabilityTester({ onExecute, executing }: SystemObservabilityTesterProps) {
  const handleHealthCheck = () => {
    const op: ApprovedOperation = {
      id: 'health.check',
      domain: 'System',
      name: 'Backend Health Check',
      description: 'Queries backend health status (/v1/health)',
      method: 'GET',
      path: '/v1/health',
      safetyClass: 'SAFE READ',
      requiresAuth: false,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleReadinessCheck = () => {
    const op: ApprovedOperation = {
      id: 'health.readiness',
      domain: 'System',
      name: 'Readiness Probe',
      description: 'Queries service readiness status (/ready)',
      method: 'GET',
      path: '/ready',
      safetyClass: 'SAFE READ',
      requiresAuth: false,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleObservabilityOverview = () => {
    const op: ApprovedOperation = {
      id: 'observability.overview',
      domain: 'Observability',
      name: 'Observability Health Summary',
      description: 'Retrieves metrics, queue status, and active worker overview',
      method: 'GET',
      path: '/v1/observability/health',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleQueueStatus = () => {
    const op: ApprovedOperation = {
      id: 'observability.queue',
      domain: 'Observability',
      name: 'Queue Status Overview',
      description: 'Inspects active notification dispatch queues',
      method: 'GET',
      path: '/v1/observability/queue',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleDashboardOverview = () => {
    const op: ApprovedOperation = {
      id: 'dashboard.overview',
      domain: 'Dashboard',
      name: 'Dashboard Overview Analytics',
      description: 'Retrieves delivery metrics and volume trends',
      method: 'GET',
      path: '/v1/dashboard/overview',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleGetSystemSettings = () => {
    const op: ApprovedOperation = {
      id: 'settings.get',
      domain: 'System Settings',
      name: 'Get Global Notification Settings',
      description: 'Queries system settings, quiet hours, and rate limits',
      method: 'GET',
      path: '/v1/admin/settings/notifications',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  return (
    <div className="space-y-6">
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Activity className="h-4 w-4 text-primary" /> System Health & Observability Diagnostics
          </CardTitle>
          <CardDescription className="text-xs">
            Run diagnostic probes across platform health, worker queue depth, analytics, and system settings.
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleHealthCheck}
              disabled={executing}
            >
              <Server className="h-3.5 w-3.5 text-primary shrink-0" /> Backend Health Probe
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleReadinessCheck}
              disabled={executing}
            >
              <Server className="h-3.5 w-3.5 text-emerald-500 shrink-0" /> Service Readiness Probe
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleObservabilityOverview}
              disabled={executing}
            >
              <Activity className="h-3.5 w-3.5 text-purple-500 shrink-0" /> Observability Summary
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleQueueStatus}
              disabled={executing}
            >
              <Activity className="h-3.5 w-3.5 text-amber-500 shrink-0" /> Queue Depth Overview
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleDashboardOverview}
              disabled={executing}
            >
              <LayoutDashboard className="h-3.5 w-3.5 text-blue-500 shrink-0" /> Dashboard Analytics
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetSystemSettings}
              disabled={executing}
            >
              <Settings className="h-3.5 w-3.5 text-slate-400 shrink-0" /> System Notification Settings
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
