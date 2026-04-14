import { useQueryMarketTrades } from "../data/queries";
import type { IMarketTrade } from "@/core/api/types";

interface TradeHistoryProps {
  pairId: string;
}

const MAX_TRADES = 20;

function formatTime(timestamp: number): string {
  try {
    return new Date(timestamp).toLocaleTimeString(undefined, {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  } catch {
    return "--";
  }
}

function formatPrice(price: string): string {
  const num = parseFloat(price);
  return num.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 8,
  });
}

function formatQty(qty: string): string {
  const num = parseFloat(qty);
  return num.toLocaleString(undefined, {
    minimumFractionDigits: 4,
    maximumFractionDigits: 8,
  });
}

function getPriceColor(trade: IMarketTrade, prevTrade: IMarketTrade | undefined): string {
  if (!prevTrade) return "text-foreground";
  const current = parseFloat(trade.price);
  const prev = parseFloat(prevTrade.price);
  if (current > prev) return "text-green-500";
  if (current < prev) return "text-red-500";
  return "text-foreground";
}

/**
 * TradeHistory — recent trades table for the selected pair.
 * Usage: <TradeHistory pairId="BTC-USDT" />
 */
export function TradeHistory({ pairId }: TradeHistoryProps) {
  const { data, isLoading } = useQueryMarketTrades(pairId);

  const trades = (data?.trades ?? []).slice(0, MAX_TRADES);

  if (isLoading) {
    return (
      <div className="flex flex-col rounded-lg border border-border bg-card">
        <div className="border-b border-border px-3 py-2">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Trade History
          </p>
        </div>
        <div className="flex flex-col gap-1 p-3">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex justify-between gap-2">
              <div className="h-3 w-16 animate-pulse rounded bg-muted" />
              <div className="h-3 w-20 animate-pulse rounded bg-muted" />
              <div className="h-3 w-14 animate-pulse rounded bg-muted" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col rounded-lg border border-border bg-card">
      {/* Header */}
      <div className="border-b border-border px-3 py-2">
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Trade History
        </p>
      </div>

      {/* Column labels */}
      <div className="grid grid-cols-3 gap-2 border-b border-border px-3 py-1">
        <span className="text-xs text-muted-foreground">Time</span>
        <span className="text-right text-xs text-muted-foreground">Price</span>
        <span className="text-right text-xs text-muted-foreground">Qty</span>
      </div>

      {/* Rows */}
      <div className="flex flex-col overflow-y-auto">
        {trades.length === 0 ? (
          <div className="flex items-center justify-center py-8">
            <p className="text-xs text-muted-foreground">No recent trades</p>
          </div>
        ) : (
          trades.map((trade, index) => {
            const prevTrade = trades[index + 1];
            const priceColor = getPriceColor(trade, prevTrade);
            return (
              <div
                key={trade.trade_id}
                className="grid grid-cols-3 gap-2 px-3 py-1 hover:bg-muted/40"
              >
                <span className="text-xs text-muted-foreground">
                  {formatTime(trade.timestamp)}
                </span>
                <span className={`text-right text-xs font-medium tabular-nums ${priceColor}`}>
                  {formatPrice(trade.price)}
                </span>
                <span className="text-right text-xs tabular-nums text-foreground">
                  {formatQty(trade.quantity)}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
