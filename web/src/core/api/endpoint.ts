export const AUTH_ENDPOINT = {
  LOGIN: "/api/v1/auth/login",
  REGISTER: "/api/v1/auth/register",
  LOGOUT: "/api/v1/auth/logout",
  REFRESH: "/api/v1/auth/refresh",
  ME: "/api/v1/auth/me",
};

export const WALLET_ENDPOINT = {
  LIST: "/api/v1/wallet",
  DEPOSIT: "/api/v1/wallet/deposit",
  WITHDRAW: "/api/v1/wallet/withdraw",
};

export const ORDER_ENDPOINT = {
  LIST: "/api/v1/orders",
  CREATE: "/api/v1/orders",
  DETAIL: (id: string) => `/api/v1/orders/${id}`,
  CANCEL: (id: string) => `/api/v1/orders/${id}`,
};

export const MARKET_ENDPOINT = {
  PAIRS: "/api/v1/market/pairs",
  TICKER_ALL: "/api/v1/market/ticker",
  TICKER: (pair: string) => `/api/v1/market/ticker/${pair}`,
  TRADES: (pair: string) => `/api/v1/market/trades/${pair}`,
  WS: "/ws",
};

export const NOTIFICATION_ENDPOINT = {
  LIST: "/api/v1/notifications",
  READ: (id: string) => `/api/v1/notifications/${id}/read`,
  READ_ALL: "/api/v1/notifications/read-all",
  UNREAD_COUNT: "/api/v1/notifications/unread-count",
  WS: "/api/v1/notifications/ws",
};
