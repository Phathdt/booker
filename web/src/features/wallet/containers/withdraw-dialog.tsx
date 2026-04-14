import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { usePostApiV1WalletWithdraw, getGetApiV1WalletQueryKey } from "@/core/api/generated/wallet/wallet";

interface WithdrawDialogProps {
  assetId: string;
  available: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function WithdrawDialog({
  assetId,
  available,
  open,
  onOpenChange,
}: WithdrawDialogProps) {
  const [amount, setAmount] = useState("");
  const queryClient = useQueryClient();
  const { mutate, isPending } = usePostApiV1WalletWithdraw({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1WalletQueryKey() });
        toast.success("Withdrawal successful");
        setAmount("");
        onOpenChange(false);
      },
      onError: (err: unknown) => {
        const message = err instanceof Error ? err.message : "Withdrawal failed";
        toast.error(message);
      },
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || Number(amount) <= 0) return;
    mutate({ data: { assetId: assetId, amount } });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Withdraw {assetId}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="withdraw-amount">Amount</Label>
            <Input
              id="withdraw-amount"
              type="number"
              placeholder="0.00"
              step="any"
              min="0"
              max={available}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
            <p className="text-xs text-muted-foreground">
              Available: {available} {assetId}
            </p>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Withdrawing..." : "Withdraw"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
