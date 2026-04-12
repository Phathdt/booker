import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { useQueryWallets } from "../data/queries";
import { DepositDialog } from "./deposit-dialog";
import { WithdrawDialog } from "./withdraw-dialog";
import type { IWallet } from "@/core/api/types";

const KNOWN_ASSETS = [
  { id: "BTC", name: "Bitcoin" },
  { id: "ETH", name: "Ethereum" },
  { id: "USDT", name: "Tether" },
];

interface AssetRow {
  id: string;
  name: string;
  available: string;
  locked: string;
}

function mergeWallets(wallets: IWallet[]): AssetRow[] {
  return KNOWN_ASSETS.map((asset) => {
    const wallet = wallets.find((w) => w.asset_id === asset.id);
    return {
      id: asset.id,
      name: asset.name,
      available: wallet?.available ?? "0.00",
      locked: wallet?.locked ?? "0.00",
    };
  });
}

export function BalanceTable() {
  const { data, isLoading } = useQueryWallets();
  const [depositAsset, setDepositAsset] = useState<string | null>(null);
  const [withdrawAsset, setWithdrawAsset] = useState<string | null>(null);

  const rows = mergeWallets(data?.wallets ?? []);

  if (isLoading) {
    return (
      <div className="flex h-32 items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  const withdrawRow = rows.find((r) => r.id === withdrawAsset);

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Asset</TableHead>
            <TableHead>Name</TableHead>
            <TableHead className="text-right">Available</TableHead>
            <TableHead className="text-right">Locked</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.id}>
              <TableCell className="font-medium">{row.id}</TableCell>
              <TableCell className="text-muted-foreground">{row.name}</TableCell>
              <TableCell className="text-right font-mono">{row.available}</TableCell>
              <TableCell className="text-right font-mono text-muted-foreground">
                {row.locked}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setDepositAsset(row.id)}
                  >
                    Deposit
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setWithdrawAsset(row.id)}
                  >
                    Withdraw
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {depositAsset && (
        <DepositDialog
          assetId={depositAsset}
          open={!!depositAsset}
          onOpenChange={(open) => !open && setDepositAsset(null)}
        />
      )}

      {withdrawAsset && withdrawRow && (
        <WithdrawDialog
          assetId={withdrawAsset}
          available={withdrawRow.available}
          open={!!withdrawAsset}
          onOpenChange={(open) => !open && setWithdrawAsset(null)}
        />
      )}
    </>
  );
}
