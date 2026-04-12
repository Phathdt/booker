import { Layout } from "@/components/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BalanceTable } from "../containers/balance-table";

export function WalletPage() {
  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Wallet</h1>
          <p className="text-sm text-muted-foreground">
            Manage your asset balances
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Balances</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <BalanceTable />
          </CardContent>
        </Card>
      </div>
    </Layout>
  );
}
