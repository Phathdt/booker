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

export interface IUser {
  id: string;
  email: string;
  role: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface IAuthResponse {
  user: IUser;
  access_token: string;
  expires_in: number;
}

export interface IRefreshResponse {
  access_token: string;
  expires_in: number;
}

export interface IWallet {
  id: string;
  user_id: string;
  asset_id: string;
  available: string;
  locked: string;
  updated_at: string;
}

export interface IOrder {
  id: string;
  user_id: string;
  pair_id: string;
  side: "buy" | "sell";
  type: "limit";
  price: string;
  quantity: string;
  filled_qty: string;
  status: "new" | "partial" | "filled" | "cancelled";
  created_at: string;
  updated_at: string;
}

export interface ITradingPair {
  id: string;
  base_asset: string;
  quote_asset: string;
  min_qty: string;
  tick_size: string;
  status: string;
}

export interface ITicker {
  pair: string;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  change_pct: string;
  last_price: string;
  ts: number;
}

export interface IMarketTrade {
  id: string;
  pair_id: string;
  price: string;
  quantity: string;
  buyer_id: string;
  seller_id: string;
  executed_at: string;
}

export interface INotification {
  id: string;
  user_id: string;
  event_key: string;
  type: string;
  title: string;
  body: string;
  is_read: boolean;
  metadata: Record<string, string>;
  created_at: string;
}
