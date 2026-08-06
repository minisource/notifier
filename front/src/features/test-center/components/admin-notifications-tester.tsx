'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { Shield, RefreshCw, XCircle, ListFilter, CheckCheck, History, Truck } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';
import { SecureUserPicker } from './secure-user-picker';

interface AdminNotificationsTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function AdminNotificationsTester({ onExecute, executing }: AdminNotificationsTesterProps) {
  const [targetNotificationId, setTargetNotificationId] = useState('');
  const [filterUserId, setFilterUserId] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterChannel, setFilterChannel] = useState('');

  const handleListAllGlobal = () => {
    const op: ApprovedOperation = {
      id: 'admin.notifications.list_all',
      domain: 'Admin Notifications',
      name: 'List All Notifications (Global)',
      description: 'Admin query across all users and tenants',
      method: 'GET',
      path: '/v1/admin/notifications',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    const q: Record<string, string> = {};
    if (filterUserId) q.userId = filterUserId;
    if (filterStatus) q.status = filterStatus;
    if (filterChannel) q.channel = filterChannel;
    onExecute(op, {}, q, undefined);
  };

  const handleGetAttempts = () => {
    if (!targetNotificationId) return;
    const op: ApprovedOperation = {
      id: 'admin.notifications.attempts',
      domain: 'Admin Notifications',
      name: 'Get Notification Delivery Attempts',
      description: 'Retrieve individual provider attempt history',
      method: 'GET',
      path: '/v1/admin/notifications/:notificationId/attempts',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: targetNotificationId }, {}, undefined);
  };

  const handleGetDeliveries = () => {
    if (!targetNotificationId) return;
    const op: ApprovedOperation = {
      id: 'admin.notifications.deliveries',
      domain: 'Admin Notifications',
      name: 'Get Notification Deliveries',
      description: 'Retrieve delivery logs for a notification',
      method: 'GET',
      path: '/v1/admin/notifications/:notificationId/deliveries',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: targetNotificationId }, {}, undefined);
  };

  const handleRetryNotification = () => {
    if (!targetNotificationId) return;
    const op: ApprovedOperation = {
      id: 'admin.notifications.retry',
      domain: 'Admin Notifications',
      name: 'Retry Notification (Admin)',
      description: 'Requeues failed notification for delivery',
      method: 'POST',
      path: '/v1/admin/notifications/:notificationId/retry',
      safetyClass: 'EXTERNAL DELIVERY',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: targetNotificationId }, {}, undefined);
  };

  const handleCancelNotification = () => {
    if (!targetNotificationId) return;
    const op: ApprovedOperation = {
      id: 'admin.notifications.cancel',
      domain: 'Admin Notifications',
      name: 'Cancel Notification (Admin)',
      description: 'Cancels pending/queued notification',
      method: 'POST',
      path: '/v1/admin/notifications/:notificationId/cancel',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: targetNotificationId }, {}, undefined);
  };

  const handleMarkAllUserRead = () => {
    if (!filterUserId) return;
    const op: ApprovedOperation = {
      id: 'admin.notifications.read_all_user',
      domain: 'Admin Notifications',
      name: 'Mark All Notifications Read for User',
      description: 'Admin bulk mark as read for specific user',
      method: 'POST',
      path: '/v1/admin/notifications/read-all',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, {}, { userId: filterUserId }, undefined);
  };

  return (
    <div className="space-y-6">
      {/* Global Notifications Query */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Shield className="h-4 w-4 text-primary" /> Admin Global Notification Feed Query
          </CardTitle>
          <CardDescription className="text-xs">
            Query across all tenants and users with optional filters (`/v1/admin/notifications`).
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="space-y-1">
              <Label className="text-xs font-medium">Filter User ID</Label>
              <SecureUserPicker
                value={filterUserId || null}
                onChange={(id) => setFilterUserId(id || '')}
                placeholder="Search user ID..."
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs font-medium">Status Filter</Label>
              <select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
                className="w-full h-8 px-3 rounded-md border border-input bg-background text-xs"
              >
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="queued">Queued</option>
                <option value="sent">Sent</option>
                <option value="failed">Failed</option>
              </select>
            </div>
            <div className="space-y-1">
              <Label className="text-xs font-medium">Channel Filter</Label>
              <select
                value={filterChannel}
                onChange={(e) => setFilterChannel(e.target.value)}
                className="w-full h-8 px-3 rounded-md border border-input bg-background text-xs"
              >
                <option value="">All Channels</option>
                <option value="in_app">In-App</option>
                <option value="email">Email</option>
                <option value="sms">SMS</option>
                <option value="push">Push</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleListAllGlobal}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> List All Global Notifications
            </Button>
            <Button
              variant="secondary"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleMarkAllUserRead}
              disabled={executing || !filterUserId}
            >
              <CheckCheck className="h-3.5 w-3.5 shrink-0 text-emerald-500" /> Mark All Read for User ID
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Single Notification Admin Operations */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <History className="h-4 w-4 text-primary" /> Admin Single Notification Diagnostics & Controls
          </CardTitle>
          <CardDescription className="text-xs">
            Inspect attempts, delivery records, or retry/cancel notification by ID.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="space-y-1">
            <Label className="text-xs font-medium">Target Notification ID</Label>
            <Input
              value={targetNotificationId}
              onChange={(e) => setTargetNotificationId(e.target.value)}
              placeholder="e.g. uuid-notification-id"
              className="h-9 text-xs font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetAttempts}
              disabled={executing || !targetNotificationId}
            >
              <History className="h-3.5 w-3.5 text-primary shrink-0" /> Get Provider Attempts
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetDeliveries}
              disabled={executing || !targetNotificationId}
            >
              <Truck className="h-3.5 w-3.5 text-blue-500 shrink-0" /> Get Delivery Logs
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5 text-amber-600 dark:text-amber-400 border-amber-500/30"
              onClick={handleRetryNotification}
              disabled={executing || !targetNotificationId}
            >
              <RefreshCw className="h-3.5 w-3.5 shrink-0" /> Retry Failed Notification
            </Button>
            <Button
              variant="destructive"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleCancelNotification}
              disabled={executing || !targetNotificationId}
            >
              <XCircle className="h-3.5 w-3.5 shrink-0" /> Cancel Notification
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
