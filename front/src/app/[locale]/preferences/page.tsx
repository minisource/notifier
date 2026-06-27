'use client';

import { useTranslations } from 'next-intl';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { usePreferences, useUpdatePreference } from '@/features/preferences/hooks/use-preferences';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { Bell, BellOff, Save } from 'lucide-react';

const CHANNELS = ['sms', 'email', 'push', 'in_app'] as const;
const FREQUENCIES = ['daily', 'weekly', 'monthly'] as const;

export default function PreferencesPage() {
  const t = useTranslations();
  const { data: preferences, isLoading, isError, error, refetch } = usePreferences();
  const updateMutation = useUpdatePreference();
  const [saving, setSaving] = useState(false);

  // Local toggle state
  const [channelStates, setChannelStates] = useState<Record<string, { isEnabled: boolean; allowInstant: boolean; allowDigest: boolean; digestFrequency: string }>>({});

  // Initialize from loaded data
  useEffect(() => {
    if (preferences && Object.keys(channelStates).length === 0 && preferences.length > 0) {
      const initial: Record<string, { isEnabled: boolean; allowInstant: boolean; allowDigest: boolean; digestFrequency: string }> = {};
      for (const p of preferences) {
        initial[p.type] = {
          isEnabled: p.isEnabled,
          allowInstant: p.allowInstant,
          allowDigest: p.allowDigest,
          digestFrequency: p.digestFrequency || 'daily',
        };
      }
      if (Object.keys(initial).length > 0) {
        setChannelStates(initial);
      }
    }
  }, [preferences]);

  const handleSave = async () => {
    setSaving(true);
    try {
      // Save each channel preference
      const userId = 'current';
      for (const [channel, state] of Object.entries(channelStates)) {
        await updateMutation.mutateAsync({
          userId,
          type: channel,
          input: {
            isEnabled: state.isEnabled,
            allowInstant: state.allowInstant,
            allowDigest: state.allowDigest,
            digestFrequency: state.digestFrequency as 'daily' | 'weekly' | 'monthly',
          },
        });
      }
      toast.success(t('common.save') as string);
    } catch {
      toast.error(t('errors.generic'));
    }
    setSaving(false);
  };

  if (isLoading) return <PageSkeleton context="preferences" layout="detail" />;

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title={t('preferences.title')} subtitle={t('preferences.subtitle')} />
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || t('errors.generic')}
          onRetry={() => refetch()}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader title={t('preferences.title')} subtitle={t('preferences.subtitle')}>
        <Button size="sm" onClick={handleSave} disabled={saving}>
          <Save className="ml-1.5 h-4 w-4" />
          {saving ? t('common.loading') : t('common.save')}
        </Button>
      </PageHeader>

      <div className="space-y-6">
        {/* Channel Preferences */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-4 w-4" />
              {t('channels.email')} & {t('channels.sms')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {CHANNELS.map((channel) => {
              const pref = preferences?.find(p => p.type === channel);
              const state = channelStates[channel] || {
                isEnabled: pref?.isEnabled ?? true,
                allowInstant: pref?.allowInstant ?? true,
                allowDigest: pref?.allowDigest ?? false,
                digestFrequency: pref?.digestFrequency || 'daily',
              };

              if (!channelStates[channel] && pref) {
                // Will be set by effect above
              }

              return (
                <div key={channel}>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {state.isEnabled ? <Bell className="h-4 w-4 text-primary" /> : <BellOff className="h-4 w-4 text-muted-foreground" />}
                      <div>
                        <p className="text-sm font-medium">{t(`channels.${channel}`)}</p>
                        <p className="text-xs text-muted-foreground">
                          {state.isEnabled ? 'Enabled' : 'Disabled'}
                        </p>
                      </div>
                    </div>
                    <Switch
                      checked={state.isEnabled}
                      onCheckedChange={(checked) => {
                        setChannelStates(prev => ({
                          ...prev,
                          [channel]: { ...prev[channel], isEnabled: checked },
                        }));
                      }}
                    />
                  </div>

                  {state.isEnabled && (
                    <div className="mt-3 ml-9 grid gap-4 sm:grid-cols-3">
                      <div className="flex items-center justify-between">
                        <Label className="text-xs">{t('preferences.allow_instant')}</Label>
                        <Switch
                          checked={state.allowInstant}
                          onCheckedChange={(checked) => {
                            setChannelStates(prev => ({
                              ...prev,
                              [channel]: { ...prev[channel], allowInstant: checked },
                            }));
                          }}
                        />
                      </div>
                      <div className="flex items-center justify-between">
                        <Label className="text-xs">{t('preferences.allow_digest')}</Label>
                        <Switch
                          checked={state.allowDigest}
                          onCheckedChange={(checked) => {
                            setChannelStates(prev => ({
                              ...prev,
                              [channel]: { ...prev[channel], allowDigest: checked },
                            }));
                          }}
                        />
                      </div>
                      <div className="flex items-center justify-between">
                        <Label className="text-xs">{t('preferences.digest_frequency')}</Label>
                        <Select
                          value={state.digestFrequency}
                          onValueChange={(value) => {
                            setChannelStates(prev => ({
                              ...prev,
                              [channel]: { ...prev[channel], digestFrequency: value },
                            }));
                          }}
                        >
                          <SelectTrigger className="w-[110px]">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {FREQUENCIES.map(f => (
                              <SelectItem key={f} value={f}>{t(`preferences.${f}`)}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                  )}
                  <Separator className="mt-4" />
                </div>
              );
            })}
          </CardContent>
        </Card>

        {/* Header actions are inline */}
      </div>
    </PageContainer>
  );
}
