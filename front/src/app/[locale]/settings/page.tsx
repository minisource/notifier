'use client';

import { useTranslations } from 'next-intl';
import { useState, useEffect } from 'react';
import { PageHeader } from '@minisource/ui';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Badge } from '@minisource/ui';
import { useTheme } from 'next-themes';
import { Input } from '@minisource/ui';
import { Label } from '@minisource/ui';
import { Sun, Moon, Monitor, Globe, User, Smartphone, Settings2, RefreshCw, Loader2, Save, Zap, Gauge } from 'lucide-react';
import { authAdapter } from '@/shared/auth/auth-adapter';
import { useNotificationSettings, useUpdateNotificationSettings } from '@/features/settings/hooks/use-settings';
import { LoadingState } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';

export default function SettingsPage() {
  const t = useTranslations();
  const { theme, setTheme } = useTheme();
  const session = authAdapter.getSession();

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
      <PageHeader title={t('settings.title')} description={t('settings.subtitle')} />

      {isLoading && <LoadingState />}

      {isError && (
        <ErrorState
          message={t('settings.failed_load')}
          onRetry={() => refetch()}
          autoRetrySeconds={30}
        />
      )}

      <div className="grid gap-6 md:grid-cols-2">
        {/* Account */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-4 w-4" />
              {t('settings.account')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                {(session.name || session.email || '?').charAt(0).toUpperCase()}
              </div>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{session.name || '—'}</p>
                <p className="truncate text-xs text-muted-foreground">
                  {t('settings.email')}: {session.email || '—'}
                </p>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-1.5">
              {session.roles.length > 0 ? (
                session.roles.map((role) => (
                  <Badge key={role} variant={role === 'super_admin' ? 'default' : 'secondary'}>
                    {role === 'super_admin'
                      ? t('settings.super_admin')
                      : role === 'admin'
                        ? t('settings.admin')
                        : role === 'operator'
                          ? t('settings.operator')
                          : role}
                  </Badge>
                ))
              ) : (
                <span className="text-xs text-muted-foreground">{t('settings.viewer')}</span>
              )}
            </div>
          </CardContent>
        </Card>

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
              // Locale is cookie-driven now (URLs are locale-free)
              document.cookie = `NEXT_LOCALE=fa; path=/; max-age=31536000; samesite=lax`;
              window.location.reload();
            }}>
              <Globe className="h-4 w-4 ltr:mr-2 rtl:ml-2" />
              فارسی
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => {
              document.cookie = `NEXT_LOCALE=en; path=/; max-age=31536000; samesite=lax`;
              window.location.reload();
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
              {t('settings.retry_policy')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">{t('settings.auto_retry')}</p>
                    <p className="text-xs text-muted-foreground">
                      {settings.retryPolicy.enabled ? t('settings.enabled') : t('settings.disabled')}
                    </p>
                  </div>
                  <Badge variant={settings.retryPolicy.enabled ? 'default' : 'secondary'}>
                    {settings.retryPolicy.enabled ? t('settings.enabled') : t('settings.disabled')}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-2">
                  <Label className="text-xs text-muted-foreground">
                    {t('settings.max_retry_attempts')}
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
                      {t('settings.strategy')}: {settings.retryPolicy.backoffStrategy}
                    </span>
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveRetryPolicy} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    {t('settings.save_retry_policy')}
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
              {t('settings.enabled_channels')}
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
              <p className="text-sm text-muted-foreground">{t('settings.loading_channels')}</p>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Rate Limit */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Gauge className="h-4 w-4" />
              {t('settings.rate_limit')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">{t('settings.rate_limiting')}</p>
                    <p className="text-xs text-muted-foreground">
                      {settings.rateLimit.enabled ? t('settings.enabled') : t('settings.disabled')}
                    </p>
                  </div>
                  <Badge variant={settings.rateLimit.enabled ? 'default' : 'secondary'}>
                    {settings.rateLimit.enabled ? t('settings.on') : t('settings.off')}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-3">
                  <div>
                    <Label className="text-xs text-muted-foreground">{t('settings.per_minute')}: {ratePerMin}</Label>
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
                    <Label className="text-xs text-muted-foreground">{t('settings.per_hour')}: {ratePerHour}</Label>
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
                    {t('settings.save_rate_limits')}
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
              {t('settings.quiet_hours')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings ? (
              <>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="text-sm font-medium">{t('settings.quiet_hours')}</p>
                    <p className="text-xs text-muted-foreground">
                      {quietEnabled ? t('settings.quiet_hours_suppressed') : t('settings.disabled')}
                    </p>
                  </div>
                  <Badge
                    variant={quietEnabled ? 'default' : 'secondary'}
                    className="cursor-pointer"
                    onClick={() => setQuietEnabled(!quietEnabled)}
                  >
                    {quietEnabled ? t('settings.on') : t('settings.off')}
                  </Badge>
                </div>

                <div className="rounded-lg border p-3 space-y-3">
                  <div className="flex items-center gap-2">
                    <div className="flex-1">
                      <Label className="text-xs text-muted-foreground">{t('settings.start')}</Label>
                      <Input
                        type="time"
                        value={quietStart}
                        onChange={(e) => setQuietStart(e.target.value)}
                        className="h-8 text-sm mt-1"
                      />
                    </div>
                    <div className="flex-1">
                      <Label className="text-xs text-muted-foreground">{t('settings.end')}</Label>
                      <Input
                        type="time"
                        value={quietEnd}
                        onChange={(e) => setQuietEnd(e.target.value)}
                        className="h-8 text-sm mt-1"
                      />
                    </div>
                  </div>
                  <div>
                    <Label className="text-xs text-muted-foreground">{t('settings.timezone')}</Label>
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
                    {t('settings.save_quiet_hours')}
                  </Button>
                </div>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">{t('settings.loading_quiet_hours')}</p>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings — Retention */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings2 className="h-4 w-4" />
              {t('settings.data_retention')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {settings && (
              <>
                <div className="rounded-lg border p-3 space-y-2">
                  <Label className="text-xs text-muted-foreground">
                    {t('settings.retention_period')}
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
                    <span className="text-xs text-muted-foreground">
                      {t('settings.days_approx', { months: Math.round(retentionDays / 30) })}
                    </span>
                  </div>
                  <Button size="sm" className="mt-2 w-full" onClick={handleSaveRetention} disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? (
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1" />
                    )}
                    {t('settings.save_retention')}
                  </Button>
                </div>
              </>
            )}
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
                {t('settings.version')}: {process.env.NEXT_PUBLIC_APP_NAME || 'Notifier Admin'} v{process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0'}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
