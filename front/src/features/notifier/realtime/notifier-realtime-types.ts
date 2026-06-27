export type RealtimeMode = 'websocket' | 'sse' | 'polling' | 'disabled';

export interface RealtimeConfig {
  mode: RealtimeMode;
  url?: string;
  pollIntervalMs: number;
}

export function getRealtimeConfig(): RealtimeConfig {
  const mode = (process.env.NEXT_PUBLIC_NOTIFIER_REALTIME_MODE ?? 'polling') as RealtimeMode;
  const url = process.env.NEXT_PUBLIC_NOTIFIER_REALTIME_URL;
  const pollIntervalMs = Number(process.env.NEXT_PUBLIC_NOTIFIER_REALTIME_POLL_INTERVAL ?? 30000);
  return { mode, url, pollIntervalMs };
}

export function shouldShowLiveToasts(): boolean {
  return process.env.NEXT_PUBLIC_NOTIFIER_SHOW_LIVE_TOASTS !== 'false';
}
