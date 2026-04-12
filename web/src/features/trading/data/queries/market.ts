import { useQuery } from "@tanstack/react-query";
import { MarketModel } from "../models";

export const MARKET_QUERY_KEYS = {
  PAIRS: "market-pairs",
  TICKER: "market-ticker",
  TRADES: "market-trades",
};

export function useQueryPairs() {
  return useQuery({
    queryKey: [MARKET_QUERY_KEYS.PAIRS],
    queryFn: () => MarketModel.getPairs(),
    staleTime: 60000,
  });
}

export function useQueryTicker(pair: string) {
  return useQuery({
    queryKey: [MARKET_QUERY_KEYS.TICKER, pair],
    queryFn: () => MarketModel.getTicker(pair),
    refetchInterval: 3000,
    enabled: Boolean(pair),
  });
}

export function useQueryMarketTrades(pair: string) {
  return useQuery({
    queryKey: [MARKET_QUERY_KEYS.TRADES, pair],
    queryFn: () => MarketModel.getTrades(pair),
    refetchInterval: 5000,
    enabled: Boolean(pair),
  });
}
