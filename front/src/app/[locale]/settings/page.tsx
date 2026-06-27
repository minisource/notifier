'use client';

import { useTranslations } from 'next-intl';
import { useState, useEffect } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useTheme } from 'next-themes';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Sun, Moon, Monitor, Globe, Shield, Wifi, Bell, Smartphone, Eye, EyeOff, Settings2, RefreshCw, Loader2, Save, Zap, Gauge } from 'lucide-react';
import { authAdapter, updateMockInLocalStorage } from '@/shared/auth/auth-adapter';
import { getRealtimeConfig } from '@/features/notifier/realtime/notifier-realtime-types';
import { refreshMockSession } from '@/shared/auth/auth-adapter';
import { useNotificationSettings, useUpdateNotificationSettings } from '@/features/settings/hooks/use-settings';
import { LoadingState } from '@/components/shared/loading-state';
import { ErrorState } from '@/components/shared/error-state';

export default function SettingsPage() {
  const t = useTranslations();
  const { theme, setTheme } = useTheme();
  const session = authAdapter.getSession();
  const realtimeConfig = getRealtimeConfig();
  const [showDebug, setShowDebug] = useState(false);
  const [mockRolesInput, setMockRolesInput] = useState(session.roles.join(', '));
  const [savedRoles, setSavedRoles] = useState(false);

  const { data: settings, isLoading, isError, refetch } = useNotificationSettings();
  const updateMutation = useUpdateNotificationSettings();
  const [retryAttempts, setRetryAttempts] = useState(3);
  const [retentionDays, setRetentionDays] = useState(90);
  const [ratePerMin, setRatePerMin] = useState(100);
  const [ratePerHour, setRatePerHour] = useState(1000);
  const [quietEnabled, setQuietEnabled] = useState(false);
  const [quietStart, setQuietStart] = useState('22:00');
  const [quietEnd, setQuietEnd] = useState('08:00');
  const [quietTz, setQuietTz] = useState('UTC');

  useEffect(() => {
    if (settings) {
      setRetryAttempts(settings.retryPolicy.maxAttempts);
      setRetentionDays(settings.retentionDays);
      setRatePerMin(settings.rateLimit.perMinute);
      setRatePerHour(settings.rateLimit.perHour);
      setQuietEnabled(settings.quietHours?.enabled ?? false);
      setQuietStart(settings.quietHours?.start ?? '22:00');
      setQuietEnd(settings.quietHours?.end ?? '08:00');
      setQuietTz(settings.quietHours?.timezone ?? 'UTC');
    }
  }, [settings]);

  const handleSaveRoles = () => {
    const roles = mockRolesInput.split(',').map(r => r.trim()).filter(Boolean);
    refreshMockSession({ roles: roles as any[] });
    updateMockInLocalStorage({ roles: roles as any[] });
    setSavedRoles(true);
    setTimeout(() => setSavedRoles(false), 2000);
  };

  const handleSaveRetryPolicy = () => {
    updateMutation.mutate({
      retryPolicy: {
        ...settings!.retryPolicy,
        maxAttempts: retryAttempts,
      },
    });
  };

  const handleSaveRetention = () => {
    updateMutation.mutate({ retentionDays });
  };

  const handleSaveRateLimit = () => {
    updateMutation.mutate({
      rateLimit: {
        enabled: true,
        perMinute: ratePerMin,
        perHour: ratePerHour,
      },
    });
  };

  const handleSaveQuietHours = () => {
    updateMutation.mutate({
      quietHours: {
        enabled: quietEnabled,
        timezone: quietTz,
        start: quietStart,
        end: quietEnd,
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('settings.title')} subtitle={t('settings.subtitle')} />

      {isLoading && <LoadingState rows={3} columns={2} />}

      {isError && (
        <ErrorState
          message="Failed to load settings from the server. Some features may be limited."
          onRetry={() => refetch()}
          autoRetrySeconds={30}
        />
      )}

      <div className="grid gap-6 md:grid-cols-2">
        {/* Language */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-4 w-4" />
              {t('settings.language')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button variant="outline" className="w-full justify-start" onClick={() => {
              const pathname = window.location.pathname;
              window.location.href = pathname.replace(/^\/(fa|en)/, '/fa');
            }}>
              <Globe className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              فارسی
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => {
              const pathname = window.location.pathname;
              window.location.href = pathname.replace(/^\/(fa|en)/, '/en');
            }}>
              <Globe className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              English
            </Button>
          </CardContent>
        </Card>

        {/* Theme */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Monitor className="h-4 w-4" />
              {t('settings.theme')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button variant={theme === 'light' ? 'default' : 'outline'} className="w-full justify-start" onClick={() => setTheme('light')}>
              <Sun className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              {t('settings.light')}
            </Button>
            <Button variant={theme === 'dark' ? 'default' : 'outline'} className="w-full justify-start" onClick={() => setTheme('dark')}>
              <Moon className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              {t('settings.dark')}
            </Button>
            <Button variant={theme === 'system' ? 'default' : 'outline'} className="w-full justify-start" onClick={() => setTheme('system')}>
              <Monitor className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              {t('settings.system')}
            </Button>
          </CardContent>
        </Card>

        {/* Notification Settings — Retry Policy */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4" />
              Retry Policy
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">Auto Retry</p>
                    <p className="text-xs text-muted-foreground">
                      {settings.retryPolicy.enabled ? 'Enabled' : 'Disabled'}
                    </p>
                  </div>
                  <Badge variant={settings.retryPolicy.enabled ? 'default' : 'secondary'}>
                    {settings.retryPolicy.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-2">
                  <Label className="text-xs text-muted-foreground">
                    Max Retry Attempts
                  </Label>
                  <div className="flex items-center gap-2">
                    <Input
                      type="number"
                      value={retryAttempts}
                      onChange={(e) => setRetryAttempts(Math.max(1, Math.min(10, parseInt(e.target.value) || 1)))}
                      min={1}
                      max={10}
                      className="h-8 w-20 text-sm"
                    />
                    <span className="text-xs text-muted-foreground">
                      Strategy: {settings.retryPolicy.backoffStrategy}
                    </span>
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveRetryPolicy} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    Save Retry Policy
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Enabled Channels */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="h-4 w-4" />
              Enabled Channels
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {settings ? (
              <div className="grid grid-cols-2 gap-2">
                {(['email', 'sms', 'push', 'webhook', 'inApp'] as const).map((ch) => (
                  <div
                    key={ch}
                    className={`flex items-center justify-between rounded-lg border p-3 ${settings.enabledChannels[ch] ? 'border-primary/30 bg-primary/5' : 'opacity-60'}`}
                  >
                    <span className="text-sm font-medium capitalize">{ch === 'inApp' ? 'In-App' : ch}</span>
                    <div
                      className={`h-5 w-9 rounded-full cursor-pointer transition-colors ${
                        settings.enabledChannels[ch] ? 'bg-primary' : 'bg-muted'
                      }`}
                      onClick={() =>
                        updateMutation.mutate({
                          enabledChannels: {
                            ...settings.enabledChannels,
                            [ch]: !settings.enabledChannels[ch],
                          },
                        })
                      }
                    >
                      <div
                        className={`h-4 w-4 rounded-full bg-white transition-transform ${
                          settings.enabledChannels[ch] ? 'translate-x-[18px]' : 'translate-x-0.5'
                        } mt-0.5`}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Loading channel configuration...</p>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Rate Limit */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Gauge className="h-4 w-4" />
              Rate Limit
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">Rate Limiting</p>
                    <p className="text-xs text-muted-foreground">
                      {settings.rateLimit.enabled ? 'Enabled' : 'Disabled'}
                    </p>
                  </div>
                  <Badge variant={settings.rateLimit.enabled ? 'default' : 'secondary'}>
                    {settings.rateLimit.enabled ? 'On' : 'Off'}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-3">
                  <div>
                    <Label className="text-xs text-muted-foreground">Per Minute: {ratePerMin}</Label>
                    <Input
                      type="number"
                      value={ratePerMin}
                      onChange={(e) => setRatePerMin(Math.max(1, Math.min(10000, parseInt(e.target.value) || 1)))}
                      min={1}
                      max={10000}
                      className="h-8 w-24 text-sm mt-1"
                    />
                  </div>
                  <div>
                    <Label className="text-xs text-muted-foreground">Per Hour: {ratePerHour}</Label>
                    <Input
                      type="number"
                      value={ratePerHour}
                      onChange={(e) => setRatePerHour(Math.max(1, Math.min(100000, parseInt(e.target.value) || 1)))}
                      min={1}
                      max={100000}
                      className="h-8 w-24 text-sm mt-1"
                    />
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveRateLimit} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    Save Rate Limits
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Quiet Hours */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Moon className="h-4 w-4" />
              Quiet Hours
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings ? (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">Quiet Hours</p>
                    <p className="text-xs text-muted-foreground">
                      {quietEnabled ? 'Notifications will be suppressed during quiet hours' : 'Disabled'}
                    </p>
                  </div>
                  <Badge
                    variant={quietEnabled ? 'default' : 'secondary'}
                    className="cursor-pointer"
                    onClick={() => setQuietEnabled(!quietEnabled)}
                  >
                    {quietEnabled ? 'On' : 'Off'}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-3">
                  <div className="flex items-center gap-2">
                    <div className="flex-1">
                      <Label className="text-xs text-muted-foreground">Start</Label>
                      <Input
                        type="time"
                        value={quietStart}
                        onChange={(e) => setQuietStart(e.target.value)}
                        className="h-8 text-sm mt-1"
                      />
                    </div>
                    <div className="flex-1">
                      <Label className="text-xs text-muted-foreground">End</Label>
                      <Input
                        type="time"
                        value={quietEnd}
                        onChange={(e) => setQuietEnd(e.target.value)}
                        className="h-8 text-sm mt-1"
                      />
                    </div>
                  </div>
                  <div>
                    <Label className="text-xs text-muted-foreground">Timezone</Label>
                    <Input
                      type="text"
                      value={quietTz}
                      onChange={(e) => setQuietTz(e.target.value)}
                      placeholder="UTC, Asia/Tehran, ..."
                      className="h-8 text-sm mt-1"
                    />
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveQuietHours} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    Save Quiet Hours
                  </Button>
                </div>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">Loading quiet hours configuration...</p>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Retention */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings2 className="h-4 w-4" />
              Data Retention
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="rounded-lg border p-3 space-y-2">
                  <Label className="text-xs text-muted-foreground">
                    Retention Period
                  </Label>
                  <div className="flex items-center gap-2">
                    <Input
                      type="number"
                      value={retentionDays}
                      onChange={(e) => setRetentionDays(Math.max(7, Math.min(365, parseInt(e.target.value) || 7)))}
                      min={7}
                      max={365}
                      className="h-8 w-24 text-sm"
                    />
                    <span className="text-xs text-muted-foreground">days (~{Math.round(retentionDays / 30)} months)</span>
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveRetention} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    Save Retention
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* API + Realtime */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wifi className="h-4 w-4" />
              {t('settings.api_mode')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="text-sm font-medium">{t('settings.api_mode')}</p>
                <p className="text-xs text-muted-foreground">
                  {process.env.NEXT_PUBLIC_API_MODE === 'mock' ? t('settings.mock') : t('settings.real')}
                </p>
              </div>
              <Badge variant={process.env.NEXT_PUBLIC_API_MODE === 'mock' ? 'secondary' : 'default'}>
                {process.env.NEXT_PUBLIC_API_MODE === 'mock' ? t('settings.mock') : t('settings.real')}
              </Badge>
            </div>
            <div className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="text-sm font-medium">{t('settings.backend_url')}</p>
                <p className="text-xs text-muted-foreground break-all">
                  {process.env.NEXT_PUBLIC_NOTIFIER_API_URL || 'http://localhost:9002/v1'}
                </p>
              </div>
            </div>
            <div className="flex items-center justify-between rounded-lg border p-3">
              <div className="flex items-center gap-2">
                <Bell className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-medium">{t('notifier.realtime.title') || 'Realtime'}</p>
                  <p className="text-xs text-muted-foreground">
                    {t('notifier.realtime.mode') || 'Mode'}: {realtimeConfig.mode}
                    {realtimeConfig.mode === 'polling' && ` (${realtimeConfig.pollIntervalMs / 1000}s interval)`}
                  </p>
                </div>
              </div>
              <Badge variant="outline">{realtimeConfig.mode}</Badge>
            </div>
          </CardContent>
        </Card>

        {/* Mock Session */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-4 w-4" />
              {t('settings.role')} — Mock Session
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="text-sm font-medium">{t('settings.role')}</p>
                <p className="text-xs text-muted-foreground">{session.roles.join(', ') || 'none'}</p>
              </div>
              <Badge variant="secondary">
                {session.roles.includes('admin') ? t('settings.admin') : session.roles.includes('operator') ? t('settings.operator') : t('settings.viewer')}
              </Badge>
            </div>
            <div className="rounded-lg border p-3 space-y-2">
              <Label className="text-xs text-muted-foreground">Mock Roles (comma-separated)</Label>
              <div className="flex gap-2">
                <Input
                  value={mockRolesInput}
                  onChange={(e) => setMockRolesInput(e.target.value)}
                  placeholder="admin, operator, user"
                  className="text-sm h-8"
                  aria-label="Mock roles"
                />
                <Button size="sm" className="h-8" onClick={handleSaveRoles}>
                  {savedRoles ? t('common.copied') || 'Saved' : t('common.save')}
                </Button>
              </div>
            </div>
            <div className="rounded-lg border p-3">
              <p className="text-xs text-muted-foreground">
                User ID: <code className="font-mono text-[10px]">{session.userId || 'none'}</code>
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Tenant: <code className="font-mono text-[10px]">{session.tenantId || 'none'}</code>
              </p>
            </div>
          </CardContent>
        </Card>

        {/* PWA */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Smartphone className="h-4 w-4" />
              {t('notifier.pwa.title') || 'PWA'}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="rounded-lg border p-3">
              <p className="text-sm font-medium">{t('notifier.pwa.installHint') || 'Install Notifier'}</p>
              <p className="text-xs text-muted-foreground mt-1">
                {t('notifier.pwa.description') || 'You can install this app on your device for a native-like experience.'}
              </p>
            </div>
            <div className="rounded-lg border p-3">
              <p className="text-xs text-muted-foreground">
                App: {process.env.NEXT_PUBLIC_APP_NAME || 'Notifier Admin'} v{process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0'}
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Debug */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Eye className="h-4 w-4" />
              {t('notifier.settings.debug') || 'Debug'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Button variant="outline" size="sm" className="w-full justify-start gap-2" onClick={() => setShowDebug(!showDebug)}>
              {showDebug ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              {showDebug ? 'Hide Debug Info' : 'Show Debug Info'}
            </Button>
            {showDebug && (
              <pre className="mt-3 rounded-lg bg-muted p-3 text-xs font-mono overflow-auto max-h-48">
{JSON.stringify({ session, realtimeConfig, env: { API_MODE: process.env.NEXT_PUBLIC_API_MODE, API_URL: process.env.NEXT_PUBLIC_NOTIFIER_API_URL, USE_MOCKS: process.env.NEXT_PUBLIC_NOTIFIER_USE_MOCKS } }, null, 2)}
              </pre>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
