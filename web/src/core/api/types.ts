// --- Generated types (from OpenAPI spec via Zod schemas) ---
// Regenerate: cd web && pnpm generate:api

import { z } from "zod";
import {
  UserResponse,
  AuthResponse,
  TokenPairResponse,
  WalletResponse,
  OrderResponse,
  PairResponse,
  TickerResponse,
  TradeResponse,
  OrderBookLevel,
  OrderBookResponse,
  NotificationResponse,
} from "./generated/schemas";

// Re-export Zod schemas for optional runtime validation
export {
  UserResponse as UserSchema,
  AuthResponse as AuthResponseSchema,
  TokenPairResponse as TokenPairSchema,
  WalletResponse as WalletSchema,
  OrderResponse as OrderSchema,
  PairResponse as TradingPairSchema,
  TickerResponse as TickerSchema,
  TradeResponse as TradeSchema,
  OrderBookLevel as OrderBookLevelSchema,
  OrderBookResponse as OrderBookSchema,
  NotificationResponse as NotificationSchema,
} from "./generated/schemas";

// Inferred TypeScript types (replace manual interfaces)
export type IUser = z.infer<typeof UserResponse>;
export type IAuthResponse = z.infer<typeof AuthResponse>;
export type IRefreshResponse = z.infer<typeof TokenPairResponse>;
export type IWallet = z.infer<typeof WalletResponse>;
export type IOrder = z.infer<typeof OrderResponse>;
export type ITradingPair = z.infer<typeof PairResponse>;
export type ITicker = z.infer<typeof TickerResponse>;
export type IMarketTrade = z.infer<typeof TradeResponse>;
export type IOrderBookLevel = z.infer<typeof OrderBookLevel>;
export type IOrderBook = z.infer<typeof OrderBookResponse>;
export type INotification = z.infer<typeof NotificationResponse>;

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
