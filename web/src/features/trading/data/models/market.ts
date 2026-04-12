import { Service } from "@/core/api/service";
import { MARKET_ENDPOINT } from "@/core/api/endpoint";
import type { ITicker, IMarketTrade, ITradingPair } from "@/core/api/types";

interface MarketTradesResponse {
  trades: IMarketTrade[];
}

interface PairsResponse {
  pairs: ITradingPair[];
}

export class MarketModel {
  private static service = new Service(MARKET_ENDPOINT.PAIRS);

  static getPairs(): Promise<PairsResponse> {
    return MarketModel.service.get<PairsResponse>(MARKET_ENDPOINT.PAIRS);
  }

  static getTicker(pair: string): Promise<ITicker> {
    return MarketModel.service.get<ITicker>(MARKET_ENDPOINT.TICKER(pair));
  }

  static getTrades(pair: string): Promise<MarketTradesResponse> {
    return MarketModel.service.get<MarketTradesResponse>(MARKET_ENDPOINT.TRADES(pair));
  }
}
