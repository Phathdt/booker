import {
  getApiV1MarketPairs,
  getApiV1MarketTicker,
  getApiV1MarketTickerPair,
  getApiV1MarketTradesPair,
  getApiV1MarketOrderbookPair,
} from "@/core/api/generated/market/market";
import type { ITicker, IMarketTrade, ITradingPair, IOrderBook } from "@/core/api/types";

export class MarketModel {
  static getPairs(): Promise<ITradingPair[]> {
    return getApiV1MarketPairs() as Promise<ITradingPair[]>;
  }

  static getAllTickers(): Promise<ITicker[]> {
    return getApiV1MarketTicker() as Promise<ITicker[]>;
  }

  static getTicker(pair: string): Promise<ITicker> {
    return getApiV1MarketTickerPair(pair) as Promise<ITicker>;
  }

  static getTrades(pair: string): Promise<IMarketTrade[]> {
    return getApiV1MarketTradesPair(pair) as Promise<IMarketTrade[]>;
  }

  static getOrderBook(pair: string): Promise<IOrderBook> {
    return getApiV1MarketOrderbookPair(pair) as Promise<IOrderBook>;
  }
}
