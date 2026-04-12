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
  refresh_token: string;
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
