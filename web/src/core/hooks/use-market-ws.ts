import { useEffect, useRef, useState } from "react";
import type { IMarketTrade, ITicker } from "@/core/api/types";
import { getWsBaseUrl } from "./ws-utils";

interface WsTickerMessage {
  type: "ticker";
  pair: string;
  data: {
    open: string;
    high: string;
    low: string;
    close: string;
    volume: string;
    change_pct: string;
    last_price: string;
    ts: number;
  };
}

interface WsTradeMessage {
  type: "trade";
  pair: string;
  data: {
    id: string;
    price: string;
    quantity: string;
    executed_at: string;
  };
}

type WsMessage = WsTickerMessage | WsTradeMessage;

const MAX_TRADES = 50;
const RECONNECT_DELAY_MS = 3000;

export interface UseMarketWSResult {
  ticker: ITicker | null;
  trades: IMarketTrade[];
  connected: boolean;
}

/**
 * useMarketWS — subscribes to ticker and trades channels for the given pair.
 *
 * @example
 * const { ticker, trades, connected } = useMarketWS("BTC_USDT");
 */
export function useMarketWS(pair: string): UseMarketWSResult {
  const [ticker, setTicker] = useState<ITicker | null>(null);
  const [trades, setTrades] = useState<IMarketTrade[]>([]);
  const [connected, setConnected] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // Track whether the hook is still mounted so reconnect logic doesn't fire after unmount
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

      clearReconnectTimer();

      const wsUrl = `${getWsBaseUrl()}/ws`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!mountedRef.current) {
          ws.close();
          return;
        }
        setConnected(true);
        ws.send(JSON.stringify({ op: "subscribe", channel: "ticker", pair }));
        ws.send(JSON.stringify({ op: "subscribe", channel: "trades", pair }));
      };

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg = JSON.parse(event.data as string) as WsMessage;

          if (msg.type === "ticker") {
            setTicker({
              pair: msg.pair,
              open: msg.data.open,
              high: msg.data.high,
              low: msg.data.low,
              close: msg.data.close,
              volume: msg.data.volume,
              change_pct: msg.data.change_pct,
              last_price: msg.data.last_price,
              ts: msg.data.ts,
            });
          } else if (msg.type === "trade") {
            const trade: IMarketTrade = {
              id: msg.data.id,
              pair_id: msg.pair,
              price: msg.data.price,
              quantity: msg.data.quantity,
              buyer_id: "",
              seller_id: "",
              executed_at: msg.data.executed_at,
            };
            setTrades((prev) => [trade, ...prev].slice(0, MAX_TRADES));
          }
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onerror = () => {
        // onclose will fire after onerror and handle reconnect
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
        // Unsubscribe before closing
        if (ws.readyState === WebSocket.OPEN) {
          try {
            ws.send(JSON.stringify({ op: "unsubscribe", channel: "ticker", pair }));
            ws.send(JSON.stringify({ op: "unsubscribe", channel: "trades", pair }));
          } catch {
            // Ignore send errors during cleanup
          }
        }
        ws.onopen = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.onclose = null;
        ws.close();
        wsRef.current = null;
      }
    };
  }, [pair]);

  return { ticker, trades, connected };
}
