import { useMemo } from "react";
import { useQueryOrderBook } from "../data/queries";
import type { IOrderBookLevel } from "@/core/api/types";

interface OrderBookProps {
  pairId: string;
}

interface OrderLevel {
  price: string;
  qty: string;
  total: string;
  cumQty: number;
}

const MAX_LEVELS = 15;

function buildLevels(entries: IOrderBookLevel[]): OrderLevel[] {
  let cumQty = 0;
  return entries.slice(0, MAX_LEVELS).map((entry) => {
    const qty = parseFloat(entry.quantity);
    cumQty += qty;
    return {
      price: parseFloat(entry.price).toLocaleString(undefined, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }),
      qty: qty.toFixed(4),
      total: cumQty.toFixed(4),
      cumQty,
    };
  });
}

interface BookSideProps {
  levels: OrderLevel[];
  side: "bid" | "ask";
  maxCumQty: number;
}

function BookSide({ levels, side, maxCumQty }: BookSideProps) {
  const isBid = side === "bid";
  const colorClass = isBid ? "text-green-500" : "text-red-500";
  const bgClass = isBid ? "bg-green-500/10" : "bg-red-500/10";
  const label = isBid ? "Bids" : "Asks";

  return (
    <div className="flex flex-col overflow-hidden">
      {/* Side header */}
      <div className="border-b border-border px-2 py-1">
        <span className={`text-xs font-medium ${colorClass}`}>{label}</span>
      </div>

      {/* Column headers */}
      <div className="grid grid-cols-3 gap-1 border-b border-border px-2 py-1">
        <span className="text-xs text-muted-foreground">Price</span>
        <span className="text-right text-xs text-muted-foreground">Qty</span>
        <span className="text-right text-xs text-muted-foreground">Total</span>
      </div>

      {/* Rows */}
      <div className="overflow-y-auto">
        {levels.length === 0 ? (
          <p className="px-2 py-4 text-center text-xs text-muted-foreground">
            No orders
          </p>
        ) : (
          levels.map((level, i) => {
            const depthPct =
              maxCumQty > 0 ? (level.cumQty / maxCumQty) * 100 : 0;
            return (
              <div key={i} className="relative px-2 py-0.5">
                {/* Depth background bar */}
                <div
                  className={`absolute inset-y-0 ${isBid ? "left-0" : "right-0"} ${bgClass}`}
                  style={{ width: `${depthPct}%` }}
                  aria-hidden="true"
                />
                {/* Row content */}
                <div className="relative grid grid-cols-3 gap-1">
                  <span
                    className={`text-xs font-medium tabular-nums ${colorClass}`}
                  >
                    {level.price}
                  </span>
                  <span className="text-right text-xs tabular-nums text-foreground">
                    {level.qty}
                  </span>
                  <span className="text-right text-xs tabular-nums text-muted-foreground">
                    {level.total}
                  </span>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

/**
 * OrderBook — real order book using the market orderbook API.
 * Usage: <OrderBook pairId="BTC_USDT" />
 */
export function OrderBook({ pairId }: OrderBookProps) {
  const { data, isLoading } = useQueryOrderBook(pairId);

  const { bids, asks, bothEmpty } = useMemo(() => {
    const bids = buildLevels(data?.bids ?? []);
    const asks = buildLevels(data?.asks ?? []);
    return { bids, asks, bothEmpty: bids.length === 0 && asks.length === 0 };
  }, [data]);

  const maxBidCumQty = bids.length > 0 ? bids[bids.length - 1].cumQty : 0;
  const maxAskCumQty = asks.length > 0 ? asks[asks.length - 1].cumQty : 0;

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card">
      {/* Title */}
      <div className="border-b border-border px-3 py-2">
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Order Book
        </p>
      </div>

      {isLoading ? (
        <div className="flex flex-1 items-center justify-center">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
        </div>
      ) : bothEmpty && data !== undefined ? (
        <div className="flex flex-1 items-center justify-center">
          <p className="text-xs text-muted-foreground">No orders</p>
        </div>
      ) : (
        <div className="grid flex-1 grid-cols-2 divide-x divide-border overflow-hidden">
          <BookSide levels={bids} side="bid" maxCumQty={maxBidCumQty} />
          <BookSide levels={asks} side="ask" maxCumQty={maxAskCumQty} />
        </div>
      )}
    </div>
  );
}
