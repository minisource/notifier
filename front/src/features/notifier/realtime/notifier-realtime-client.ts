import { getRealtimeConfig } from './notifier-realtime-types';
import type { RealtimeConfig } from './notifier-realtime-types';

export type RealtimeEventType = 'notification' | 'unread_change' | 'health_change' | 'queue_change' | 'worker_change';

export interface RealtimeEvent {
  type: RealtimeEventType;
  payload: unknown;
  timestamp: string;
}

type EventHandler = (event: RealtimeEvent) => void;

class NotifierRealtimeClient {
  private config: RealtimeConfig;
  private handlers: Map<RealtimeEventType, Set<EventHandler>> = new Map();
  private ws: WebSocket | null = null;
  private es: EventSource | null = null;
  private pollTimer: ReturnType<typeof setInterval> | null = null;
  private visibleNotifications: Set<string> = new Set();

  constructor() {
    this.config = getRealtimeConfig();
  }

  connect(): void {
    this.disconnect();
    switch (this.config.mode) {
      case 'websocket':
        this.connectWebSocket();
        break;
      case 'sse':
        this.connectSSE();
        break;
      case 'polling':
        this.startPolling();
        break;
      case 'disabled':
        break;
    }
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
    this.es?.close();
    this.es = null;
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
    }
  }

  on(eventType: RealtimeEventType, handler: EventHandler): () => void {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, new Set());
    }
    this.handlers.get(eventType)!.add(handler);
    return () => this.handlers.get(eventType)?.delete(handler);
  }

  off(eventType: RealtimeEventType, handler: EventHandler): void {
    this.handlers.get(eventType)?.delete(handler);
  }

  emitDeduplicatedId(id: string): boolean {
    if (this.visibleNotifications.has(id)) return false;
    this.visibleNotifications.add(id);
    setTimeout(() => this.visibleNotifications.delete(id), 60000);
    return true;
  }

  getConfig(): RealtimeConfig {
    return this.config;
  }

  isPolling(): boolean {
    return this.config.mode === 'polling' && this.pollTimer !== null;
  }

  getPollInterval(): number {
    return this.config.pollIntervalMs;
  }

  private emit(event: RealtimeEvent): void {
    const handlers = this.handlers.get(event.type);
    if (handlers) {
      handlers.forEach(h => h(event));
    }
  }

  private connectWebSocket(): void {
    if (!this.config.url) return;
    try {
      this.ws = new WebSocket(this.config.url);
      this.ws.onmessage = (msg) => {
        try {
          const event = JSON.parse(msg.data) as RealtimeEvent;
          this.emit(event);
        } catch { /* ignore parse errors */ }
      };
      this.ws.onclose = () => setTimeout(() => this.connect(), 5000);
    } catch { /* ignore */ }
  }

  private connectSSE(): void {
    if (!this.config.url) return;
    try {
      this.es = new EventSource(this.config.url);
      this.es.onmessage = (msg) => {
        try {
          const event = JSON.parse(msg.data) as RealtimeEvent;
          this.emit(event);
        } catch { /* ignore */ }
      };
    } catch { /* ignore */ }
  }

  private startPolling(): void {
    this.pollTimer = setInterval(() => {
      this.emit({ type: 'notification', payload: null, timestamp: new Date().toISOString() });
      this.emit({ type: 'unread_change', payload: null, timestamp: new Date().toISOString() });
    }, this.config.pollIntervalMs);
  }
}

export const realtimeClient = new NotifierRealtimeClient();
