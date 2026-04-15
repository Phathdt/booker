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
import { useDeposit, getGetBalancesQueryKey } from "@/core/api/generated/wallet/wallet";

interface DepositDialogProps {
  assetId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DepositDialog({ assetId, open, onOpenChange }: DepositDialogProps) {
  const [amount, setAmount] = useState("");
  const queryClient = useQueryClient();
  const { mutate, isPending } = useDeposit({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetBalancesQueryKey() });
        toast.success("Deposit successful");
        setAmount("");
        onOpenChange(false);
      },
      onError: (err: unknown) => {
        const message = err instanceof Error ? err.message : "Deposit failed";
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
          <DialogTitle>Deposit {assetId}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="deposit-amount">Amount</Label>
            <Input
              id="deposit-amount"
              type="number"
              placeholder="0.00"
              step="any"
              min="0"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
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
              {isPending ? "Depositing..." : "Deposit"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
