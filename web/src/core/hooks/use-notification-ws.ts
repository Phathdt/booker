import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getAccessToken } from "@/core/api/service";
import type { INotification } from "@/core/api/types";
import { getListNotificationsQueryKey, getGetUnreadCountQueryKey } from "@/core/api/generated/notifications/notifications";
import { getWsBaseUrl } from "./ws-utils";

interface WsNotificationMessage {
  type: string;
  data: INotification;
}

const RECONNECT_DELAY_MS = 3000;

export interface UseNotificationWSResult {
  lastNotification: INotification | null;
  connected: boolean;
}

/**
 * useNotificationWS — connects to the notification WebSocket and invalidates
 * react-query caches when a new notification arrives.
 *
 * @example
 * const { lastNotification, connected } = useNotificationWS();
 */
export function useNotificationWS(): UseNotificationWSResult {
  const [lastNotification, setLastNotification] = useState<INotification | null>(null);
  const [connected, setConnected] = useState(false);

  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;

    function clearReconnectTimer() {
      if (reconnectTimerRef.current !== null) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
    }

    function connect() {
      if (!mountedRef.current) return;

      const token = getAccessToken();
      if (!token) {
        // No token yet — retry after delay
        reconnectTimerRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
        return;
      }

      clearReconnectTimer();

      const wsUrl = `${getWsBaseUrl()}/api/v1/notifications/ws?token=${encodeURIComponent(token)}`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!mountedRef.current) {
          ws.close();
          return;
        }
        setConnected(true);
      };

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg = JSON.parse(event.data as string) as WsNotificationMessage;

          if (msg.data) {
            setLastNotification(msg.data);
            queryClient.invalidateQueries({ queryKey: getListNotificationsQueryKey() });
            queryClient.invalidateQueries({ queryKey: getGetUnreadCountQueryKey() });
          }
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onerror = () => {
        // onclose fires after onerror and handles reconnect
      };

      ws.onclose = () => {
        if (!mountedRef.current) return;
        setConnected(false);
        reconnectTimerRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
      };
    }

    connect();

    return () => {
      mountedRef.current = false;
      clearReconnectTimer();

      const ws = wsRef.current;
      if (ws) {
        ws.onopen = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.onclose = null;
        ws.close();
        wsRef.current = null;
      }
    };
  }, [queryClient]);

  return { lastNotification, connected };
}
