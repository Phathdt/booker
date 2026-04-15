// --- Generated types (from OpenAPI spec via orval) ---
// Regenerate: cd web && pnpm generate:api

// Re-export generated model types with I-prefix aliases
export type { DtoUserResponse as IUser } from "./generated/models";
export type { DtoAuthResponse as IAuthResponse } from "./generated/models";
export type { DtoTokenPairResponse as IRefreshResponse } from "./generated/models";
export type { DtoWalletResponse as IWallet } from "./generated/models";
export type { DtoOrderResponse as IOrder } from "./generated/models";
export type { MarketPairResponse as ITradingPair } from "./generated/models";
export type { MarketTickerResponse as ITicker } from "./generated/models";
export type { MarketTradeResponse as IMarketTrade } from "./generated/models";
export type { MarketOrderBookLevel as IOrderBookLevel } from "./generated/models";
export type { MarketOrderBookResponse as IOrderBook } from "./generated/models";
export type { DtoNotificationResponse as INotification } from "./generated/models";

// --- Utility types (client-side only, not from API spec) ---

export interface IApiResponse<T> {
  data: T;
  error?: { message: string; code?: string };
  traceId?: string;
  requestId?: string;
}

export interface IHttpError {
  httpCode: number;
  message: string;
}
