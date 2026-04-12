import { useMemo } from "react";
import { useQueryMarketTrades } from "../data/queries";

interface OrderBookProps {
  pairId: string;
}

interface OrderLevel {
  price: string;
  qty: string;
  total: string;
}

const LEVELS = 10;

/**
 * Simulates order book depth from recent market trades.
 * Groups trades into price levels and splits them into bids/asks.
 */
function buildBookFromTrades(
  prices: { price: string; quantity: string }[]
): { bids: OrderLevel[]; asks: OrderLevel[] } {
  if (prices.length === 0) return { bids: [], asks: [] };

  // Build a price→qty map
  const priceMap = new Map<string, number>();
  for (const t of prices) {
    const p = parseFloat(t.price).toFixed(2);
    priceMap.set(p, (priceMap.get(p) ?? 0) + parseFloat(t.quantity));
  }

  const sorted = Array.from(priceMap.entries())
    .map(([price, qty]) => ({ price: parseFloat(price), qty }))
    .sort((a, b) => b.price - a.price);

  const mid = Math.floor(sorted.length / 2);
  const askEntries = sorted.slice(0, mid).slice(0, LEVELS);
  const bidEntries = sorted.slice(mid).slice(0, LEVELS);

  const toLevel = (entries: typeof sorted): OrderLevel[] => {
    let cumTotal = 0;
    return entries.map(({ price, qty }) => {
      cumTotal += qty;
      return {
        price: price.toLocaleString(undefined, {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2,
        }),
        qty: qty.toFixed(4),
        total: cumTotal.toFixed(4),
      };
    });
  };

  return { asks: toLevel(askEntries), bids: toLevel(bidEntries) };
}

/**
 * OrderBook — simulated order book using market trades data.
 * Usage: <OrderBook pairId="BTC-USDT" />
 */
export function OrderBook({ pairId }: OrderBookProps) {
  const { data, isLoading } = useQueryMarketTrades(pairId);

  const { bids, asks } = useMemo(
    () => buildBookFromTrades(data?.trades ?? []),
    [data]
  );

  const colHeader = (
    <div className="grid grid-cols-3 gap-1 border-b border-border px-2 py-1">
      <span className="text-xs text-muted-foreground">Price</span>
      <span className="text-right text-xs text-muted-foreground">Qty</span>
      <span className="text-right text-xs text-muted-foreground">Total</span>
    </div>
  );

  const skeletonRows = (
    <div className="flex flex-col gap-1 p-2">
      {Array.from({ length: LEVELS }).map((_, i) => (
        <div key={i} className="grid grid-cols-3 gap-1">
          <div className="h-3 animate-pulse rounded bg-muted" />
          <div className="h-3 animate-pulse rounded bg-muted" />
          <div className="h-3 animate-pulse rounded bg-muted" />
        </div>
      ))}
    </div>
  );

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card">
      {/* Title */}
      <div className="border-b border-border px-3 py-2">
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Order Book
        </p>
      </div>

      <div className="grid flex-1 grid-cols-2 divide-x divide-border overflow-hidden">
        {/* Bids */}
        <div className="flex flex-col overflow-hidden">
          <div className="border-b border-border px-2 py-1">
            <span className="text-xs font-medium text-green-500">Bids</span>
          </div>
          {colHeader}
          {isLoading ? (
            skeletonRows
          ) : (
            <div className="overflow-y-auto">
              {bids.length === 0 ? (
                <p className="px-2 py-4 text-center text-xs text-muted-foreground">
                  No data
                </p>
              ) : (
                bids.map((level, i) => (
                  <div
                    key={i}
                    className="grid grid-cols-3 gap-1 px-2 py-0.5 hover:bg-green-500/5"
                  >
                    <span className="text-xs font-medium tabular-nums text-green-500">
                      {level.price}
                    </span>
                    <span className="text-right text-xs tabular-nums text-foreground">
                      {level.qty}
                    </span>
                    <span className="text-right text-xs tabular-nums text-muted-foreground">
                      {level.total}
                    </span>
                  </div>
                ))
              )}
            </div>
          )}
        </div>

        {/* Asks */}
        <div className="flex flex-col overflow-hidden">
          <div className="border-b border-border px-2 py-1">
            <span className="text-xs font-medium text-red-500">Asks</span>
          </div>
          {colHeader}
          {isLoading ? (
            skeletonRows
          ) : (
            <div className="overflow-y-auto">
              {asks.length === 0 ? (
                <p className="px-2 py-4 text-center text-xs text-muted-foreground">
                  No data
                </p>
              ) : (
                asks.map((level, i) => (
                  <div
                    key={i}
                    className="grid grid-cols-3 gap-1 px-2 py-0.5 hover:bg-red-500/5"
                  >
                    <span className="text-xs font-medium tabular-nums text-red-500">
                      {level.price}
                    </span>
                    <span className="text-right text-xs tabular-nums text-foreground">
                      {level.qty}
                    </span>
                    <span className="text-right text-xs tabular-nums text-muted-foreground">
                      {level.total}
                    </span>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
