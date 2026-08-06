'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { Sliders, Clock, Plus, ListFilter } from 'lucide-react';
import { ApprovedOperation, DiagnosticResult } from '../types';

interface PreferenceReminderTesterProps {
  onExecute: (
    operation: ApprovedOperation,
    pathParams: Record<string, string>,
    queryParams: Record<string, string>,
    body: any
  ) => Promise<DiagnosticResult | null>;
  executing: boolean;
}

export function PreferenceReminderTester({ onExecute, executing }: PreferenceReminderTesterProps) {
  const [targetChannel, setTargetChannel] = useState('email');
  const [channelEnabled, setChannelEnabled] = useState(true);

  const [reminderSubject, setReminderSubject] = useState('Diagnostic Reminder');
  const [reminderRecipient, setReminderRecipient] = useState('self@minisource.dev');

  const handleGetMyPreferences = () => {
    const op: ApprovedOperation = {
      id: 'preferences.get',
      domain: 'Preferences',
      name: 'Get My Preferences',
      description: 'Retrieves preferences for active user',
      method: 'GET',
      path: '/v1/me/preferences',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handlePatchChannelPreference = () => {
    const op: ApprovedOperation = {
      id: 'preferences.patch_channel',
      domain: 'Preferences',
      name: 'Update Channel Preference',
      description: 'Updates settings for specific channel',
      method: 'PATCH',
      path: '/v1/me/preferences/channel/:channel',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };

    onExecute(
      op,
      { channel: targetChannel },
      {},
      {
        is_enabled: channelEnabled,
        allow_instant: true,
        allow_digest: true,
        digest_frequency: 'daily',
      }
    );
  };

  const handleListReminders = () => {
    const op: ApprovedOperation = {
      id: 'reminders.list',
      domain: 'Reminders',
      name: 'List My Reminders',
      description: 'Queries scheduled reminders for active user',
      method: 'GET',
      path: '/v1/me/reminders',
      safetyClass: 'SAFE READ',
      requiresAuth: true,
    };
    onExecute(op, {}, {}, undefined);
  };

  const handleCreateReminder = () => {
    const op: ApprovedOperation = {
      id: 'reminders.create',
      domain: 'Reminders',
      name: 'Create My Reminder',
      description: 'Creates a scheduled reminder',
      method: 'POST',
      path: '/v1/me/reminders',
      safetyClass: 'LOCAL MUTATION',
      requiresAuth: true,
    };

    onExecute(
      op,
      {},
      {},
      {
        type: 'email',
        recipient: reminderRecipient,
        subject: reminderSubject,
        body: 'Automated test reminder from API Test Lab',
        scheduled_at: new Date(Date.now() + 3600000).toISOString(),
      }
    );
  };

  return (
    <div className="space-y-6">
      {/* Preferences Actions */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Sliders className="h-4 w-4 text-primary" /> User Preference Settings
          </CardTitle>
          <CardDescription className="text-xs">
            Query or update channel notification preferences (`/v1/me/preferences`).
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs font-medium">Channel</Label>
              <select
                value={targetChannel}
                onChange={(e) => setTargetChannel(e.target.value)}
                className="w-full h-8 px-3 rounded-md border border-input bg-background text-xs"
              >
                <option value="email">Email</option>
                <option value="sms">SMS</option>
                <option value="push">Push</option>
                <option value="in_app">In-App</option>
              </select>
            </div>
            <div className="space-y-1">
              <Label className="text-xs font-medium">Enabled Status</Label>
              <select
                value={channelEnabled ? 'true' : 'false'}
                onChange={(e) => setChannelEnabled(e.target.value === 'true')}
                className="w-full h-8 px-3 rounded-md border border-input bg-background text-xs"
              >
                <option value="true">Enabled (True)</option>
                <option value="false">Disabled (False)</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleGetMyPreferences}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> Get My Preferences
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handlePatchChannelPreference}
              disabled={executing}
            >
              <Sliders className="h-3.5 w-3.5 shrink-0" /> Update Channel Preference
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Reminders Actions */}
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Clock className="h-4 w-4 text-primary" /> Scheduled Reminders
          </CardTitle>
          <CardDescription className="text-xs">
            Query or create scheduled reminders for current user (`/v1/me/reminders`).
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs font-medium">Reminder Subject</Label>
              <Input
                value={reminderSubject}
                onChange={(e) => setReminderSubject(e.target.value)}
                className="h-8 text-xs"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs font-medium">Recipient Address</Label>
              <Input
                value={reminderRecipient}
                onChange={(e) => setReminderRecipient(e.target.value)}
                className="h-8 text-xs font-mono"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5"
              onClick={handleListReminders}
              disabled={executing}
            >
              <ListFilter className="h-3.5 w-3.5 text-primary shrink-0" /> List My Reminders
            </Button>
            <Button
              variant="default"
              size="sm"
              className="h-9 text-xs font-medium w-full justify-center gap-1.5 bg-amber-600 hover:bg-amber-700 text-white"
              onClick={handleCreateReminder}
              disabled={executing}
            >
              <Plus className="h-3.5 w-3.5 shrink-0" /> Create Scheduled Reminder
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
