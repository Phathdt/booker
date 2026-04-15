import { useMemo, useState } from "react";
import { Layout } from "@/components/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PairSelector } from "../containers/pair-selector";
import { OrderForm } from "../containers/order-form";
import { OpenOrders } from "../containers/open-orders";
import { OrderBook } from "../containers/order-book";
import { TickerBar } from "../containers/ticker-bar";
import { TradeHistory } from "../containers/trade-history";
import { useGetPairs } from "@/core/api/generated/market/market";

export function TradingPage() {
  const { data: pairs = [], isLoading } = useGetPairs({ query: { staleTime: 60000 } });

  const [selectedPair, setSelectedPair] = useState<string>("");

  // Default to first pair once loaded (derived state, no effect needed)
  const activePair = useMemo(
    () => selectedPair || (pairs.length > 0 ? pairs[0].id : ""),
    [selectedPair, pairs],
  );

  if (isLoading && !activePair) {
    return (
      <Layout>
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-4">
        {/* Header row with pair selector + ticker */}
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold tracking-tight">Trade</h1>
          <PairSelector value={activePair} onChange={setSelectedPair} />
        </div>

        {/* Ticker bar */}
        <TickerBar pairId={activePair} />

        {/* Top section: order book + order form */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {/* Order book takes 2/3 */}
          <div className="lg:col-span-2">
            <Card className="h-full">
              <CardHeader className="pb-2">
                <CardTitle className="text-base">Order Book</CardTitle>
              </CardHeader>
              <CardContent>
                <OrderBook pairId={activePair} />
              </CardContent>
            </Card>
          </div>

          {/* Order form takes 1/3 */}
          <div>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-base">{activePair}</CardTitle>
              </CardHeader>
              <CardContent>
                <OrderForm pairId={activePair} />
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Bottom section: trade history + open orders */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Recent Trades</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <TradeHistory pairId={activePair} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Open Orders</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <OpenOrders pairId={activePair} />
            </CardContent>
          </Card>
        </div>
      </div>
    </Layout>
  );
}
