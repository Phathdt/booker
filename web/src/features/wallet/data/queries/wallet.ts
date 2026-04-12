import { useQuery } from "@tanstack/react-query";
import { WalletModel } from "../models";

export const WALLET_QUERY_KEYS = {
  LIST: "wallets",
};

export function useQueryWallets() {
  return useQuery({
    queryKey: [WALLET_QUERY_KEYS.LIST],
    queryFn: () => WalletModel.getAll(),
  });
}
