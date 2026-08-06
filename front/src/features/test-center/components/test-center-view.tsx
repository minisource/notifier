'use client';

import React, { useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { Badge, toast } from '@minisource/ui';
import { Wrench, Shield, Server, Globe } from 'lucide-react';
import { NOTIFIER_OPERATIONS_CATALOG } from '../lib/catalog';
import { NOTIFIER_FLOW_DEFINITIONS } from '../lib/flow-definitions';
import { HeaderConfigBuilder } from './header-config-builder';
import { OperationTester } from './operation-tester';
import { FlowTester } from './flow-tester';
import { InAppTester } from './inapp-tester';
import { AdminNotificationsTester } from './admin-notifications-tester';
import { TemplateTester } from './template-tester';
import { ProviderTester } from './provider-tester';
import { PreferenceReminderTester } from './preference-reminder-tester';
import { SystemObservabilityTester } from './system-observability-tester';
import { RequestResponseInspector } from './request-response-inspector';
import { ExecutionHistoryList } from './execution-history-list';
import { ExternalDeliveryModal } from './external-delivery-modal';
import { ApprovedOperation, AuthMode, DiagnosticResult, ExecutionHistoryItem, FlowDefinition, HeaderPair } from '../types';

export function TestCenterView() {
  const searchParams = useSearchParams();
  const rawTab = searchParams.get('tab');
  const activeTab = (rawTab || 'inapp') as
    | 'inapp'
    | 'admin-notifs'
    | 'templates'
    | 'providers'
    | 'preferences'
    | 'system'
    | 'operations'
    | 'flows'
    | 'config';

  // Request Configuration State
  const [targetMode, setTargetMode] = useState<'direct' | 'gateway'>('direct');
  const [authMode, setAuthMode] = useState<AuthMode>('current-session');
  const [customToken, setCustomToken] = useState('');
  const [headers, setHeaders] = useState<HeaderPair[]>([
    { key: 'X-Correlation-ID', value: `corr-${Math.random().toString(36).substring(2, 9)}` },
    { key: 'X-Tenant-ID', value: 'tenant-default' },
  ]);

  // Execution State
  const [executing, setExecuting] = useState(false);
  const [latestResult, setLatestResult] = useState<DiagnosticResult | null>(null);
  const [history, setHistory] = useState<ExecutionHistoryItem[]>([]);
  const [selectedHistoryItem, setSelectedHistoryItem] = useState<ExecutionHistoryItem | null>(null);

  // Safeguard Modal State
  const [pendingConfirmation, setPendingConfirmation] = useState<{
    flow?: FlowDefinition;
    op?: ApprovedOperation;
    pathParams?: Record<string, string>;
    queryParams?: Record<string, string>;
    body?: any;
  } | null>(null);

  const addHistoryItem = (item: Omit<ExecutionHistoryItem, 'id' | 'timestamp'>) => {
    const newItem: ExecutionHistoryItem = {
      ...item,
      id: Math.random().toString(36).substring(2, 9),
      timestamp: new Date().toLocaleTimeString(),
    };
    setHistory((prev) => [newItem, ...prev]);
  };

  const getHeadersMap = (): Record<string, string> => {
    const map: Record<string, string> = {};
    headers.forEach((h) => {
      if (h.key.trim() && h.value.trim()) {
        map[h.key.trim()] = h.value.trim();
      }
    });
    return map;
  };

  const handleExecuteOperation = async (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ): Promise<DiagnosticResult | null> => {
    if (operation.safetyClass === 'EXTERNAL DELIVERY') {
      setPendingConfirmation({ op: operation, pathParams, queryParams, body });
      return null;
    }

    return await runOperationRequest(operation, pathParams, queryParams, body);
  };

  const runOperationRequest = async (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ): Promise<DiagnosticResult | null> => {
    setExecuting(true);
    try {
      const clientHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('accessToken') || sessionStorage.getItem('accessToken');
        if (token) {
          clientHeaders['Authorization'] = token.startsWith('Bearer ') ? token : `Bearer ${token}`;
        }
      }

      const res = await fetch('/api/admin/test-center/execute', {
        method: 'POST',
        headers: clientHeaders,
        body: JSON.stringify({
          operationId: operation.id,
          input: body,
          pathParams,
          queryParams,
          authMode,
          customToken,
          headers: getHeadersMap(),
          targetMode,
        }),
      });

      const result: DiagnosticResult = await res.json();
      setLatestResult(result);
      setSelectedHistoryItem(null); // Clear selected history to show latest active run details
      addHistoryItem({
        operationId: operation.id,
        operationName: operation.name,
        domain: operation.domain,
        success: result.success,
        statusCode: result.statusCode,
        duration: result.duration,
        safetyClass: operation.safetyClass,
        request: result.request,
        response: result.response,
        requestId: result.requestId,
        correlationId: result.correlationId,
      });

      if (result.success) {
        toast.success(`Executed ${operation.name} successfully (${result.duration}ms)`);
      } else {
        toast.error(`Execution returned HTTP ${result.statusCode}`);
      }

      return result;
    } catch (err: any) {
      toast.error('Execution request failed: ' + err.message);
      return null;
    } finally {
      setExecuting(false);
    }
  };

  return (
    <div className="w-full max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
      {/* Top Header Bar */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border pb-5">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2.5">
            <Wrench className="h-6 w-6 text-primary" /> Notifier API Test Lab
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Production-available diagnostics playground for authorized administrators. Real-time endpoint testing, in-app dispatches, and trace inspection.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline" className="px-2.5 py-1 bg-amber-500/10 border-amber-500/30 text-amber-700 dark:text-amber-400 flex items-center gap-1.5 text-xs">
            <Shield className="h-3.5 w-3.5" /> Admin Authorized Area
          </Badge>
          <Badge variant="outline" className="px-2.5 py-1 font-mono text-xs flex items-center gap-1">
            {targetMode === 'direct' ? <Server className="h-3.5 w-3.5 text-primary" /> : <Globe className="h-3.5 w-3.5 text-primary" />}
            Target: {targetMode === 'direct' ? 'Direct Backend (:9002)' : 'Gateway Proxy (:8080)'}
          </Badge>
        </div>
      </div>



      {/* Grid Split View: Scenarios (Left) & Diagnostics (Right) */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* Left Side: Scenarios and Configuration */}
        <div className="lg:col-span-8 space-y-6">
          {activeTab === 'inapp' && (
            <InAppTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'admin-notifs' && (
            <AdminNotificationsTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'templates' && (
            <TemplateTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'providers' && (
            <ProviderTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'preferences' && (
            <PreferenceReminderTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'system' && (
            <SystemObservabilityTester onExecute={handleExecuteOperation} executing={executing} />
          )}

          {activeTab === 'operations' && (
            <OperationTester
              operations={NOTIFIER_OPERATIONS_CATALOG}
              onExecute={handleExecuteOperation}
              executing={executing}
            />
          )}

          {activeTab === 'flows' && (
            <FlowTester
              flows={NOTIFIER_FLOW_DEFINITIONS}
              requestConfig={{
                authMode,
                customToken,
                headersMap: getHeadersMap(),
                targetMode,
              }}
              onRequestConfirmation={(flow) => setPendingConfirmation({ flow })}
            />
          )}

          {activeTab === 'config' && (
            <HeaderConfigBuilder
              authMode={authMode}
              setAuthMode={setAuthMode}
              customToken={customToken}
              setCustomToken={setCustomToken}
              headers={headers}
              setHeaders={setHeaders}
              targetMode={targetMode}
              setTargetMode={setTargetMode}
            />
          )}
        </div>

        {/* Right Side: Diagnostics Inspector and Local History */}
        <div className="lg:col-span-4 space-y-6">
          <ExecutionHistoryList
            history={history}
            onSelect={(item) => {
              setSelectedHistoryItem(item);
              setLatestResult(null); // Clear active latest run output to focus on selected history record
            }}
            onClear={() => {
              setHistory([]);
              setSelectedHistoryItem(null);
              setLatestResult(null);
            }}
            selectedId={selectedHistoryItem?.id}
          />

          {(selectedHistoryItem || latestResult) && (
            <RequestResponseInspector
              result={selectedHistoryItem ? (selectedHistoryItem as any) : latestResult!}
            />
          )}
        </div>
      </div>

      {/* External Delivery Safeguard Modal */}
      <ExternalDeliveryModal
        isOpen={!!pendingConfirmation}
        onClose={() => setPendingConfirmation(null)}
        onConfirm={() => {
          if (pendingConfirmation?.op) {
            runOperationRequest(
              pendingConfirmation.op,
              pendingConfirmation.pathParams || {},
              pendingConfirmation.queryParams || {},
              pendingConfirmation.body
            );
          }
        }}
        operationName={pendingConfirmation?.op?.name || pendingConfirmation?.flow?.name || 'Notification Dispatch'}
      />
    </div>
  );
}
