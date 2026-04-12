import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { OrderModel } from "../models";
import { ORDER_QUERY_KEYS } from "../queries";

export function useMutationCreateOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: {
      pair_id: string;
      side: "buy" | "sell";
      type: "limit";
      price: string;
      quantity: string;
    }) => OrderModel.create(payload),
    onSuccess: (order) => {
      queryClient.invalidateQueries({ queryKey: [ORDER_QUERY_KEYS.LIST] });
      toast.success(
        `${order.side === "buy" ? "Buy" : "Sell"} order placed successfully`
      );
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Failed to place order");
    },
  });
}

export function useMutationCancelOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => OrderModel.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [ORDER_QUERY_KEYS.LIST] });
      toast.success("Order cancelled");
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Failed to cancel order");
    },
  });
}
