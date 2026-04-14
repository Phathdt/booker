// --- Generated types (from OpenAPI spec via Zod schemas) ---
// Regenerate: cd web && pnpm generate:api

import { z } from "zod";
import { schemas } from "./generated/schemas";

// Re-export Zod schemas for optional runtime validation
export const UserSchema = schemas.DtoUserResponse;
export const AuthResponseSchema = schemas.DtoAuthResponse;
export const TokenPairSchema = schemas.DtoTokenPairResponse;
export const WalletSchema = schemas.DtoWalletResponse;
export const OrderSchema = schemas.DtoOrderResponse;
export const TradingPairSchema = schemas.MarketPairResponse;
export const TickerSchema = schemas.MarketTickerResponse;
export const TradeSchema = schemas.MarketTradeResponse;
export const OrderBookLevelSchema = schemas.MarketOrderBookLevel;
export const OrderBookSchema = schemas.MarketOrderBookResponse;
export const NotificationSchema = schemas.DtoNotificationResponse;

// Inferred TypeScript types (from Zod schemas — single source of truth)
export type IUser = z.infer<typeof UserSchema>;
export type IAuthResponse = z.infer<typeof AuthResponseSchema>;
export type IRefreshResponse = z.infer<typeof TokenPairSchema>;
export type IWallet = z.infer<typeof WalletSchema>;
export type IOrder = z.infer<typeof OrderSchema>;
export type ITradingPair = z.infer<typeof TradingPairSchema>;
export type ITicker = z.infer<typeof TickerSchema>;
export type IMarketTrade = z.infer<typeof TradeSchema>;
export type IOrderBookLevel = z.infer<typeof OrderBookLevelSchema>;
export type IOrderBook = z.infer<typeof OrderBookSchema>;
export type INotification = z.infer<typeof NotificationSchema>;

// --- Utility types (client-side only, not from API spec) ---

export interface IApiResponse<T> {
  data: T;
  error?: { message: string; code?: string };
  trace_id?: string;
  request_id?: string;
}

export interface IHttpError {
  httpCode: number;
  message: string;
}
