// --- Generated types (from OpenAPI spec via Zod schemas) ---
// Regenerate: cd web && pnpm generate:api

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

// TypeScript types — inferred from schema shape (all fields required)
export type IUser = {
  id: string;
  email: string;
  role: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type IAuthResponse = {
  user: IUser;
  access_token: string;
  expires_in: number;
};

export type IRefreshResponse = {
  access_token: string;
  expires_in: number;
};

export type IWallet = {
  id: string;
  user_id: string;
  asset_id: string;
  available: string;
  locked: string;
  updated_at: string;
};

export type IOrder = {
  id: string;
  user_id: string;
  pair_id: string;
  side: string;
  type: string;
  price: string;
  quantity: string;
  filled_qty: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type ITradingPair = {
  id: string;
  base_asset: string;
  quote_asset: string;
  min_qty: string;
  tick_size: string;
};

export type ITicker = {
  pair: string;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  change_pct: string;
  last_price: string;
  timestamp: number;
};

export type IMarketTrade = {
  trade_id: string;
  price: string;
  quantity: string;
  timestamp: number;
};

export type IOrderBookLevel = {
  price: string;
  quantity: string;
  order_count: number;
};

export type IOrderBook = {
  pair_id: string;
  bids: IOrderBookLevel[];
  asks: IOrderBookLevel[];
};

export type INotification = {
  id: string;
  type: string;
  title: string;
  body: string;
  metadata: Record<string, string> | null;
  is_read: boolean;
  created_at: string;
};

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
