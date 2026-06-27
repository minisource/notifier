'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { toast } from 'sonner';

interface WebSocketMessage {
  type: 'notification' | 'status' | 'connected';
  data?: Record<string, unknown>;
  notification?: {
    id: string;
    subject?: string;
    type: string;
    body: string;
  };
}

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export function useWebSocket(userId?: string) {
  const ws = useRef<WebSocket | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const reconnectTimeout = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const reconnectAttempts = useRef(0);

  const connect = useCallback(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) return;

    const baseUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:9002';
    const url = `${baseUrl}/ws?token=${token}`;

    setStatus('connecting');
    ws.current = new WebSocket(url);

    ws.current.onopen = () => {
      setStatus('connected');
      reconnectAttempts.current = 0;
    };

    ws.current.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        setLastMessage(message);

        if (message.type === 'notification' && message.notification) {
          toast(`New ${message.notification.type} notification: ${message.notification.subject || 'No subject'}`, {
            description: message.notification.body,
          });
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.current.onclose = () => {
      setStatus('disconnected');
      // Reconnect with exponential backoff
      if (reconnectAttempts.current < 5) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
        reconnectTimeout.current = setTimeout(() => {
          reconnectAttempts.current++;
          connect();
        }, delay);
      }
    };

    ws.current.onerror = () => {
      setStatus('error');
    };
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    ws.current?.close();
    ws.current = null;
    setStatus('disconnected');
  }, []);

  useEffect(() => {
    if (userId) {
      connect();
    }
    return () => disconnect();
  }, [userId, connect, disconnect]);

  return { status, lastMessage, connect, disconnect };
}
