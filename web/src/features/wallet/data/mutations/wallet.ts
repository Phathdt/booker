import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { WalletModel } from "../models";
import { WALLET_QUERY_KEYS } from "../queries";

export function useMutationDeposit() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ assetId, amount }: { assetId: string; amount: string }) =>
      WalletModel.deposit(assetId, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [WALLET_QUERY_KEYS.LIST] });
      toast.success("Deposit successful");
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Deposit failed");
    },
  });
}

export function useMutationWithdraw() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ assetId, amount }: { assetId: string; amount: string }) =>
      WalletModel.withdraw(assetId, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [WALLET_QUERY_KEYS.LIST] });
      toast.success("Withdrawal successful");
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Withdrawal failed");
    },
  });
}
