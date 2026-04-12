import { useState } from "react";
import { Layout } from "@/components/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PairSelector, TRADING_PAIRS } from "../containers/pair-selector";
import { OrderForm } from "../containers/order-form";
import { OpenOrders } from "../containers/open-orders";
import { OrderBook } from "../containers/order-book";

export function TradingPage() {
  const [selectedPair, setSelectedPair] = useState(TRADING_PAIRS[0].id);

  return (
    <Layout>
      <div className="space-y-4">
        {/* Header row with pair selector */}
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold tracking-tight">Trade</h1>
          <PairSelector value={selectedPair} onChange={setSelectedPair} />
        </div>

        {/* Top section: order book + order form */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {/* Order book takes 2/3 */}
          <div className="lg:col-span-2">
            <Card className="h-full">
              <CardHeader className="pb-2">
                <CardTitle className="text-base">Order Book</CardTitle>
              </CardHeader>
              <CardContent>
                <OrderBook />
              </CardContent>
            </Card>
          </div>

          {/* Order form takes 1/3 */}
          <div>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-base">{selectedPair}</CardTitle>
              </CardHeader>
              <CardContent>
                <OrderForm pairId={selectedPair} />
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Bottom section: open orders */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Open Orders</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <OpenOrders pairId={selectedPair} />
          </CardContent>
        </Card>
      </div>
    </Layout>
  );
}
