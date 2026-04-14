import { useGetApiV1MarketTickerPair } from "@/core/api/generated/market/market";

interface TickerBarProps {
  pairId: string;
}

/**
 * TickerBar — compact horizontal bar showing 24h market stats for a trading pair.
 * Usage: <TickerBar pairId="BTC-USDT" />
 */
export function TickerBar({ pairId }: TickerBarProps) {
  const { data: ticker, isLoading } = useGetApiV1MarketTickerPair(pairId, { query: { refetchInterval: 3000, enabled: Boolean(pairId) } });

  if (isLoading) {
    return (
      <div className="flex h-12 items-center gap-6 rounded-lg border border-border bg-card px-4">
        <div className="h-4 w-24 animate-pulse rounded bg-muted" />
        <div className="h-4 w-16 animate-pulse rounded bg-muted" />
        <div className="h-4 w-20 animate-pulse rounded bg-muted" />
      </div>
    );
  }

  if (!ticker) {
    return (
      <div className="flex h-12 items-center rounded-lg border border-border bg-card px-4">
        <p className="text-sm text-muted-foreground">No ticker data</p>
      </div>
    );
  }

  const changePct = parseFloat(ticker.change_pct);
  const isPositive = changePct >= 0;
  const changeColor = isPositive ? "text-green-500" : "text-red-500";
  const changeSign = isPositive ? "+" : "";

  return (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-1 rounded-lg border border-border bg-card px-4 py-2">
      {/* Current price */}
      <div className="flex flex-col">
        <span className="text-xs text-muted-foreground">Last Price</span>
        <span className={`text-base font-semibold ${changeColor}`}>
          {parseFloat(ticker.last_price).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 8,
          })}
        </span>
      </div>

      {/* 24h change */}
      <div className="flex flex-col">
        <span className="text-xs text-muted-foreground">24h Change</span>
        <span className={`text-sm font-medium ${changeColor}`}>
          {changeSign}{changePct.toFixed(2)}%
        </span>
      </div>

      {/* 24h high */}
      <div className="flex flex-col">
        <span className="text-xs text-muted-foreground">24h High</span>
        <span className="text-sm font-medium text-foreground">
          {parseFloat(ticker.high).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 8,
          })}
        </span>
      </div>

      {/* 24h low */}
      <div className="flex flex-col">
        <span className="text-xs text-muted-foreground">24h Low</span>
        <span className="text-sm font-medium text-foreground">
          {parseFloat(ticker.low).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 8,
          })}
        </span>
      </div>

      {/* 24h volume */}
      <div className="flex flex-col">
        <span className="text-xs text-muted-foreground">24h Volume</span>
        <span className="text-sm font-medium text-foreground">
          {parseFloat(ticker.volume).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          })}
        </span>
      </div>
    </div>
  );
}
