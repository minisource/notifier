'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { Send, Bell, CheckCheck, Eye, MousePointer, Hash, ListFilter } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface InAppTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function InAppTester({ onExecute, executing }: InAppTesterProps) {
  // Form states for individual action cards
  const [inAppSubject, setInAppSubject] = useState('In-App Diagnostic Test Alert');
  const [inAppContent, setInAppContent] = useState('This is a real-time in-app notification test sent to my active identity.');
  const [inAppPriority, setInAppPriority] = useState('high');

  const [singleNotificationId, setSingleNotificationId] = useState('');

  // 1-Click Operations
  const handleSendInAppToSelf = () => {
    const op: ApprovedOperation = {
      id: 'me.notifications.send_inapp',
      domain: 'In-App Notifications',
      name: 'Send In-App Notification to Self',
      description: 'Dispatches an in-app notification to myself',
      method: 'POST',
      path: '/v1/notifications/send',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };

    onExecute(
      op,
      {},
      {},
      {
        type: 'in_app',
        recipient: 'local-admin',
        subject: inAppSubject,
        content: inAppContent,
        priority: inAppPriority,
      }
    );
  };

  const handleFetchUnreadCount = () => {
    const op: ApprovedOperation = {
      id: 'me.notifications.unread_count',
      domain: 'In-App Notifications',
      name: 'Get My Unread Count',
      description: 'Queries total unread notifications count',
      method: 'GET',
      path: '/v1/me/notifications/unread-count',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleFetchUnreadList = () => {
    const op: ApprovedOperation = {
      id: 'me.notifications.unread',
      domain: 'In-App Notifications',
      name: 'Get My Unread Notifications',
      description: 'Queries unread notifications list',
      method: 'GET',
      path: '/v1/me/notifications/unread',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleFetchAllNotifications = () => {
    const op: ApprovedOperation = {
      id: 'me.notifications.list',
      domain: 'In-App Notifications',
      name: 'List My Notifications',
      description: 'Queries all user notifications',
      method: 'GET',
      path: '/v1/me/notifications',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleMarkAllRead = () => {
    const op: ApprovedOperation = {
      id: 'me.notifications.read_all',
      domain: 'In-App Notifications',
      name: 'Mark All My Notifications as Read',
      description: 'Marks all unread notifications as read',
      method: 'POST',
      path: '/v1/me/notifications/read-all',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleMarkSingleRead = () => {
    if (!singleNotificationId) return;
    const op: ApprovedOperation = {
      id: 'me.notifications.mark_read',
      domain: 'In-App Notifications',
      name: 'Mark Notification as Read',
      description: 'Marks single notification as read',
      method: 'PUT',
      path: '/v1/me/notifications/:notificationId/read',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: singleNotificationId }, {}, undefined);
  };

  const handleMarkSingleSeen = () => {
    if (!singleNotificationId) return;
    const op: ApprovedOperation = {
      id: 'me.notifications.mark_seen',
      domain: 'In-App Notifications',
      name: 'Mark Notification as Seen',
      description: 'Marks single notification as seen',
      method: 'POST',
      path: '/v1/me/notifications/:notificationId/seen',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: singleNotificationId }, {}, undefined);
  };

  const handleMarkSingleClicked = () => {
    if (!singleNotificationId) return;
    const op: ApprovedOperation = {
      id: 'me.notifications.mark_click',
      domain: 'In-App Notifications',
      name: 'Mark Notification Link Clicked',
      description: 'Marks notification link as clicked',
      method: 'POST',
      path: '/v1/me/notifications/:notificationId/click',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };
    onExecute(op, { notificationId: singleNotificationId }, {}, undefined);
  };

  return (
    <div className="space-y-6">
      {/* Action Card 1: Dispatch In-App Notification to Self */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Send className="h-4 w-4 text-primary" /> Send In-App Notification to Self
          </CardTitle>
          <CardDescription className="text-xs">
            Dispatches a real-time In-App notification (type: in_app) to your active user session.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs font-medium">Notification Subject</Label>
              <Input
                value={inAppSubject}
                onChange={(e) => setInAppSubject(e.target.value)}
                placeholder="Subject line..."
                className="h-8 text-xs"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs font-medium">Priority Level</Label>
              <select
                value={inAppPriority}
                onChange={(e) => setInAppPriority(e.target.value)}
                className="w-full h-8 px-3 rounded-md border border-input bg-background text-xs"
              >
                <option value="low">Low</option>
                <option value="normal">Normal</option>
                <option value="high">High</option>
                <option value="urgent">Urgent</option>
              </select>
            </div>
          </div>
          <div className="space-y-1">
            <Label className="text-xs font-medium">Message Body Content</Label>
            <textarea
              value={inAppContent}
              onChange={(e) => setInAppContent(e.target.value)}
              rows={2}
              className="w-full p-2.5 rounded-md border border-input bg-background text-xs font-mono focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <Button
            onClick={handleSendInAppToSelf}
            disabled={executing}
            size="sm"
            className="w-full h-9 text-xs font-semibold bg-primary hover:bg-primary/90 gap-1.5"
          >
            <Send className="h-3.5 w-3.5" /> {executing ? 'Sending Notification...' : 'Send In-App Notification to Me'}
          </Button>
        </CardContent>
      </Card>

      {/* Action Card 2: Read / Unread Status Queries & Bulk Actions */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Hash className="h-4 w-4 text-primary" /> Unread Count & Notification Feed Actions
          </CardTitle>
          <CardDescription className="text-xs">
            Query user-scoped notification status endpoints (`/v1/me/notifications`) or mark unread feed as read.
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleFetchUnreadCount}
              disabled={executing}
            >
              <Hash className="h-3.5 w-3.5 text-primary shrink-0" /> Get Unread Count
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleFetchUnreadList}
              disabled={executing}
            >
              <Bell className="h-3.5 w-3.5 text-primary shrink-0" /> List Unread Only
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleFetchAllNotifications}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> List All My Notifications
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white"
              onClick={handleMarkAllRead}
              disabled={executing}
            >
              <CheckCheck className="h-3.5 w-3.5 shrink-0" /> Mark All as Read
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Action Card 3: Single Notification State Mutations */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <CheckCheck className="h-4 w-4 text-primary" /> Single Notification Status Transitions
          </CardTitle>
          <CardDescription className="text-xs">
            Enter a Notification ID to test individual state transitions (Read, Seen, Clicked).
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="space-y-1">
            <Label className="text-xs font-medium">Target Notification ID</Label>
            <Input
              value={singleNotificationId}
              onChange={(e) => setSingleNotificationId(e.target.value)}
              placeholder="e.g. 123e4567-e89b-12d3-a456-426614174000"
              className="h-9 text-xs font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleMarkSingleRead}
              disabled={executing || !singleNotificationId}
            >
              <CheckCheck className="h-3.5 w-3.5 text-emerald-500 shrink-0" /> Mark as Read
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleMarkSingleSeen}
              disabled={executing || !singleNotificationId}
            >
              <Eye className="h-3.5 w-3.5 text-blue-500 shrink-0" /> Mark as Seen
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleMarkSingleClicked}
              disabled={executing || !singleNotificationId}
            >
              <MousePointer className="h-3.5 w-3.5 text-purple-500 shrink-0" /> Mark Link Clicked
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
