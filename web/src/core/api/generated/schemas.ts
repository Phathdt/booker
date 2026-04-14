// @ts-nocheck
// AUTO-GENERATED — DO NOT EDIT (regenerate: pnpm generate:api)
import { makeApi, Zodios, type ZodiosOptions } from "@zodios/core";
import { z } from "zod";

type DtoAuthResponse = Partial<{
  access_token: string;
  expires_in: number;
  user: DtoUserResponse;
}>;
type DtoUserResponse = Partial<{
  created_at: string;
  email: string;
  id: string;
  role: string;
  status: string;
  updated_at: string;
}>;
type DtoNotificationListResponse = Partial<{
  notifications: Array<DtoNotificationResponse> | null;
}>;
type DtoNotificationResponse = Partial<{
  body: string;
  created_at: string;
  id: string;
  is_read: boolean;
  metadata: {};
  title: string;
  type: string;
}>;
type DtoOrderListResponse = Partial<{
  orders: Array<DtoOrderResponse> | null;
}>;
type DtoOrderResponse = Partial<{
  created_at: string;
  filled_qty: string;
  id: string;
  pair_id: string;
  price: string;
  quantity: string;
  side: string;
  status: string;
  type: string;
  updated_at: string;
  user_id: string;
}>;
type DtoUserListResponse = Partial<{
  total: number;
  users: Array<DtoUserResponse> | null;
}>;
type DtoWalletListResponse = Partial<{
  wallets: Array<DtoWalletResponse> | null;
}>;
type DtoWalletResponse = Partial<{
  asset_id: string;
  available: string;
  id: string;
  locked: string;
  updated_at: string;
  user_id: string;
}>;
type MarketOrderBookResponse = Partial<{
  asks: Array<MarketOrderBookLevel> | null;
  bids: Array<MarketOrderBookLevel> | null;
  pair_id: string;
}>;
type MarketOrderBookLevel = Partial<{
  order_count: number;
  price: string;
  quantity: string;
}>;

const DtoLoginDTO = z
  .object({ email: z.string(), password: z.string() })
  .passthrough();
const DtoUserResponse: z.ZodType<DtoUserResponse> = z
  .object({
    created_at: z.string(),
    email: z.string(),
    id: z.string(),
    role: z.string(),
    status: z.string(),
    updated_at: z.string(),
  })
  .passthrough();
const DtoAuthResponse: z.ZodType<DtoAuthResponse> = z
  .object({
    access_token: z.string(),
    expires_in: z.number().int(),
    user: DtoUserResponse,
  })
  .passthrough();
const DtoMessageResponse = z.object({ message: z.string() }).passthrough();
const DtoTokenPairResponse = z
  .object({ access_token: z.string(), expires_in: z.number().int() })
  .passthrough();
const DtoRegisterDTO = z
  .object({ email: z.string(), password: z.string() })
  .passthrough();
const MarketOrderBookLevel: z.ZodType<MarketOrderBookLevel> = z
  .object({
    order_count: z.number().int(),
    price: z.string(),
    quantity: z.string(),
  })
  .passthrough();
const MarketOrderBookResponse: z.ZodType<MarketOrderBookResponse> = z
  .object({
    asks: z.array(MarketOrderBookLevel).nullable(),
    bids: z.array(MarketOrderBookLevel).nullable(),
    pair_id: z.string(),
  })
  .passthrough();
const MarketPairResponse = z
  .object({
    base_asset: z.string(),
    id: z.string(),
    min_qty: z.string(),
    quote_asset: z.string(),
    tick_size: z.string(),
  })
  .passthrough();
const MarketTickerResponse = z
  .object({
    change_pct: z.string(),
    close: z.string(),
    high: z.string(),
    last_price: z.string(),
    low: z.string(),
    open: z.string(),
    pair: z.string(),
    timestamp: z.number().int(),
    volume: z.string(),
  })
  .passthrough();
const MarketTradeResponse = z
  .object({
    price: z.string(),
    quantity: z.string(),
    timestamp: z.number().int(),
    trade_id: z.string(),
  })
  .passthrough();
const DtoNotificationResponse: z.ZodType<DtoNotificationResponse> = z
  .object({
    body: z.string(),
    created_at: z.string(),
    id: z.string(),
    is_read: z.boolean(),
    metadata: z.record(z.string()).nullable(),
    title: z.string(),
    type: z.string(),
  })
  .passthrough();
const DtoNotificationListResponse: z.ZodType<DtoNotificationListResponse> = z
  .object({ notifications: z.array(DtoNotificationResponse).nullable() })
  .passthrough();
const V2Map = z.object({}).passthrough();
const DtoUnreadCountResponse = z
  .object({ count: z.number().int() })
  .passthrough();
const DtoOrderResponse: z.ZodType<DtoOrderResponse> = z
  .object({
    created_at: z.string(),
    filled_qty: z.string(),
    id: z.string(),
    pair_id: z.string(),
    price: z.string(),
    quantity: z.string(),
    side: z.string(),
    status: z.string(),
    type: z.string(),
    updated_at: z.string(),
    user_id: z.string(),
  })
  .passthrough();
const DtoOrderListResponse: z.ZodType<DtoOrderListResponse> = z
  .object({ orders: z.array(DtoOrderResponse).nullable() })
  .passthrough();
const DtoCreateOrderDTO = z
  .object({
    pair_id: z.string(),
    price: z.string(),
    quantity: z.string(),
    side: z.string(),
    type: z.string(),
  })
  .passthrough();
const DtoUserListResponse: z.ZodType<DtoUserListResponse> = z
  .object({
    total: z.number().int(),
    users: z.array(DtoUserResponse).nullable(),
  })
  .passthrough();
const DtoWalletResponse: z.ZodType<DtoWalletResponse> = z
  .object({
    asset_id: z.string(),
    available: z.string(),
    id: z.string(),
    locked: z.string(),
    updated_at: z.string(),
    user_id: z.string(),
  })
  .passthrough();
const DtoWalletListResponse: z.ZodType<DtoWalletListResponse> = z
  .object({ wallets: z.array(DtoWalletResponse).nullable() })
  .passthrough();
const DtoDepositDTO = z
  .object({ amount: z.string(), asset_id: z.string() })
  .passthrough();
const DtoWithdrawDTO = z
  .object({ amount: z.string(), asset_id: z.string() })
  .passthrough();

export const schemas = {
  DtoLoginDTO,
  DtoUserResponse,
  DtoAuthResponse,
  DtoMessageResponse,
  DtoTokenPairResponse,
  DtoRegisterDTO,
  MarketOrderBookLevel,
  MarketOrderBookResponse,
  MarketPairResponse,
  MarketTickerResponse,
  MarketTradeResponse,
  DtoNotificationResponse,
  DtoNotificationListResponse,
  V2Map,
  DtoUnreadCountResponse,
  DtoOrderResponse,
  DtoOrderListResponse,
  DtoCreateOrderDTO,
  DtoUserListResponse,
  DtoWalletResponse,
  DtoWalletListResponse,
  DtoDepositDTO,
  DtoWithdrawDTO,
};

const endpoints = makeApi([
  {
    method: "post",
    path: "/api/v1/auth/login",
    alias: "postApiv1authlogin",
    description: `Login with email and password`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DtoLoginDTO,
      },
    ],
    response: DtoAuthResponse,
  },
  {
    method: "post",
    path: "/api/v1/auth/logout",
    alias: "postApiv1authlogout",
    description: `Logout (revoke all tokens)`,
    requestFormat: "json",
    response: z.object({ message: z.string() }).passthrough(),
  },
  {
    method: "get",
    path: "/api/v1/auth/me",
    alias: "getApiv1authme",
    description: `Get current authenticated user`,
    requestFormat: "json",
    response: DtoUserResponse,
  },
  {
    method: "post",
    path: "/api/v1/auth/refresh",
    alias: "postApiv1authrefresh",
    description: `Refresh access token`,
    requestFormat: "json",
    response: DtoTokenPairResponse,
  },
  {
    method: "post",
    path: "/api/v1/auth/register",
    alias: "postApiv1authregister",
    description: `Register a new user`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DtoRegisterDTO,
      },
    ],
    response: DtoAuthResponse,
  },
  {
    method: "get",
    path: "/api/v1/market/orderbook/:pair",
    alias: "getApiv1marketorderbookPair",
    description: `Get order book depth for a trading pair`,
    requestFormat: "json",
    parameters: [
      {
        name: "depth",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "pair",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: MarketOrderBookResponse,
  },
  {
    method: "get",
    path: "/api/v1/market/pairs",
    alias: "getApiv1marketpairs",
    description: `List active trading pairs`,
    requestFormat: "json",
    response: z.array(MarketPairResponse),
  },
  {
    method: "get",
    path: "/api/v1/market/ticker",
    alias: "getApiv1marketticker",
    description: `Get all pair tickers`,
    requestFormat: "json",
    response: z.array(MarketTickerResponse),
  },
  {
    method: "get",
    path: "/api/v1/market/ticker/:pair",
    alias: "getApiv1markettickerPair",
    description: `Get ticker for a single pair`,
    requestFormat: "json",
    parameters: [
      {
        name: "pair",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: MarketTickerResponse,
  },
  {
    method: "get",
    path: "/api/v1/market/trades/:pair",
    alias: "getApiv1markettradesPair",
    description: `Get recent trades for a pair`,
    requestFormat: "json",
    parameters: [
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "pair",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.array(MarketTradeResponse),
  },
  {
    method: "get",
    path: "/api/v1/notifications/",
    alias: "getApiv1notifications",
    description: `List notifications for current user`,
    requestFormat: "json",
    parameters: [
      {
        name: "cursor",
        type: "Query",
        schema: z.string().optional(),
      },
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "only_unread",
        type: "Query",
        schema: z.boolean().optional(),
      },
    ],
    response: DtoNotificationListResponse,
  },
  {
    method: "patch",
    path: "/api/v1/notifications/:id/read",
    alias: "patchApiv1notificationsIdread",
    description: `Mark a notification as read`,
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.object({}).passthrough().nullable(),
  },
  {
    method: "post",
    path: "/api/v1/notifications/read-all",
    alias: "postApiv1notificationsreadAll",
    description: `Mark all notifications as read`,
    requestFormat: "json",
    response: z.object({}).passthrough().nullable(),
  },
  {
    method: "get",
    path: "/api/v1/notifications/unread-count",
    alias: "getApiv1notificationsunreadCount",
    description: `Get unread notification count`,
    requestFormat: "json",
    response: z.object({ count: z.number().int() }).passthrough(),
  },
  {
    method: "get",
    path: "/api/v1/orders/",
    alias: "getApiv1orders",
    description: `List orders for current user`,
    requestFormat: "json",
    parameters: [
      {
        name: "pair_id",
        type: "Query",
        schema: z.string().optional(),
      },
      {
        name: "status",
        type: "Query",
        schema: z.string().optional(),
      },
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "offset",
        type: "Query",
        schema: z.number().int().optional(),
      },
    ],
    response: DtoOrderListResponse,
  },
  {
    method: "post",
    path: "/api/v1/orders/",
    alias: "postApiv1orders",
    description: `Create a new limit order`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DtoCreateOrderDTO,
      },
    ],
    response: DtoOrderResponse,
  },
  {
    method: "delete",
    path: "/api/v1/orders/:id",
    alias: "deleteApiv1ordersId",
    description: `Cancel an order`,
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: DtoOrderResponse,
  },
  {
    method: "get",
    path: "/api/v1/orders/:id",
    alias: "getApiv1ordersId",
    description: `Get a single order by ID`,
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: DtoOrderResponse,
  },
  {
    method: "get",
    path: "/api/v1/users/",
    alias: "getApiv1users",
    description: `List users`,
    requestFormat: "json",
    parameters: [
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "offset",
        type: "Query",
        schema: z.number().int().optional(),
      },
    ],
    response: DtoUserListResponse,
  },
  {
    method: "get",
    path: "/api/v1/users/:id",
    alias: "getApiv1usersId",
    description: `Get user by ID`,
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: DtoUserResponse,
  },
  {
    method: "get",
    path: "/api/v1/wallet/",
    alias: "getApiv1wallet",
    description: `Get all wallet balances for current user`,
    requestFormat: "json",
    response: DtoWalletListResponse,
  },
  {
    method: "get",
    path: "/api/v1/wallet/:asset_id",
    alias: "getApiv1walletAsset_id",
    description: `Get wallet balance for a specific asset`,
    requestFormat: "json",
    parameters: [
      {
        name: "asset_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: DtoWalletResponse,
  },
  {
    method: "post",
    path: "/api/v1/wallet/deposit",
    alias: "postApiv1walletdeposit",
    description: `Deposit funds to wallet`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DtoDepositDTO,
      },
    ],
    response: DtoWalletResponse,
  },
  {
    method: "post",
    path: "/api/v1/wallet/withdraw",
    alias: "postApiv1walletwithdraw",
    description: `Withdraw funds from wallet`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DtoWithdrawDTO,
      },
    ],
    response: DtoWalletResponse,
  },
]);

export const api = new Zodios(endpoints);

export function createApiClient(baseUrl: string, options?: ZodiosOptions) {
  return new Zodios(baseUrl, endpoints, options);
}
