// AUTO-GENERATED from docs/openapi.yaml — DO NOT EDIT
// Regenerate: cd web && pnpm generate:api
// Source: go run . openapi-export

import { z } from "zod";

// --- Auth ---

export const UserResponse = z.object({
  id: z.string(),
  email: z.string(),
  role: z.string(),
  status: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const AuthResponse = z.object({
  access_token: z.string(),
  expires_in: z.number().int(),
  user: UserResponse,
});

export const TokenPairResponse = z.object({
  access_token: z.string(),
  expires_in: z.number().int(),
});

export const MessageResponse = z.object({
  message: z.string(),
});

export const UserListResponse = z.object({
  users: z.array(UserResponse),
  total: z.number().int(),
});

export const LoginDTO = z.object({
  email: z.string(),
  password: z.string(),
});

export const RegisterDTO = z.object({
  email: z.string(),
  password: z.string(),
});

// --- Wallet ---

export const WalletResponse = z.object({
  id: z.string(),
  user_id: z.string(),
  asset_id: z.string(),
  available: z.string(),
  locked: z.string(),
  updated_at: z.string(),
});

export const WalletListResponse = z.object({
  wallets: z.array(WalletResponse),
});

export const DepositDTO = z.object({
  asset_id: z.string(),
  amount: z.string(),
});

export const WithdrawDTO = z.object({
  asset_id: z.string(),
  amount: z.string(),
});

// --- Orders ---

export const OrderResponse = z.object({
  id: z.string(),
  user_id: z.string(),
  pair_id: z.string(),
  side: z.string(),
  type: z.string(),
  price: z.string(),
  quantity: z.string(),
  filled_qty: z.string(),
  status: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const OrderListResponse = z.object({
  orders: z.array(OrderResponse),
});

export const CreateOrderDTO = z.object({
  pair_id: z.string(),
  side: z.string(),
  type: z.string(),
  price: z.string(),
  quantity: z.string(),
});

// --- Market ---

export const PairResponse = z.object({
  id: z.string(),
  base_asset: z.string(),
  quote_asset: z.string(),
  min_qty: z.string(),
  tick_size: z.string(),
});

export const TickerResponse = z.object({
  pair: z.string(),
  open: z.string(),
  high: z.string(),
  low: z.string(),
  close: z.string(),
  volume: z.string(),
  change_pct: z.string(),
  last_price: z.string(),
  timestamp: z.number().int(),
});

export const TradeResponse = z.object({
  trade_id: z.string(),
  price: z.string(),
  quantity: z.string(),
  timestamp: z.number().int(),
});

export const OrderBookLevel = z.object({
  price: z.string(),
  quantity: z.string(),
  order_count: z.number().int(),
});

export const OrderBookResponse = z.object({
  pair_id: z.string(),
  bids: z.array(OrderBookLevel),
  asks: z.array(OrderBookLevel),
});

// --- Notifications ---

export const NotificationResponse = z.object({
  id: z.string(),
  type: z.string(),
  title: z.string(),
  body: z.string(),
  metadata: z.record(z.string(), z.string()),
  is_read: z.boolean(),
  created_at: z.string(),
});

export const NotificationListResponse = z.object({
  notifications: z.array(NotificationResponse),
});

export const UnreadCountResponse = z.object({
  count: z.number().int(),
});
