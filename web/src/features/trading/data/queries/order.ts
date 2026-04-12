import { useQuery } from "@tanstack/react-query";
import { OrderModel } from "../models";

export const ORDER_QUERY_KEYS = {
  LIST: "orders",
};

export function useQueryOrders(pairId?: string) {
  return useQuery({
    queryKey: [ORDER_QUERY_KEYS.LIST, pairId],
    queryFn: () =>
      OrderModel.getAll(pairId ? { pair_id: pairId } : undefined),
    refetchInterval: 5000,
  });
}
