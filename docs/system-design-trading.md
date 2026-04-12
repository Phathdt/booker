# System Design: Trading Model

## Overview

Booker is a CEX (Centralized Exchange) demo implementing a complete order-to-settlement trading flow across 6 microservices communicating via gRPC (synchronous) and NATS JetStream (asynchronous).

---

## System Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                   EXTERNAL LAYER                                     │
│                                                                                      │
│   ┌─────────┐         ┌──────────────────────────────────────────────────────────┐   │
│   │  Web UI │────────▶│                    Traefik (Reverse Proxy)               │   │
│   └─────────┘         └──────┬──────────────┬──────────────┬────────────────┬────┘   │
│                              │              │              │                │         │
└──────────────────────────────┼──────────────┼──────────────┼────────────────┼─────────┘
                               │ :8081        │ :8082        │ :8083          │ :8085
┌──────────────────────────────┼──────────────┼──────────────┼────────────────┼─────────┐
│                              ▼              ▼              ▼                ▼         │
│   HTTP LAYER          ┌───────────┐  ┌───────────┐  ┌───────────┐   ┌───────────┐   │
│   (Fiber REST)        │ users-svc │  │wallet-svc │  │ order-svc │   │market-svc │   │
│                       │  /auth/*  │  │ /wallet/* │  │ /orders/* │   │ /market/* │   │
│                       └─────┬─────┘  └─────┬─────┘  └─────┬─────┘   └───────────┘   │
│                             │              │              │                           │
└─────────────────────────────┼──────────────┼──────────────┼───────────────────────────┘
                              │ :50051       │ :50052       │ :50053
┌─────────────────────────────┼──────────────┼──────────────┼───────────────────────────┐
│                             ▼              ▼              ▼                           │
│   gRPC LAYER          ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌─────────────┐  │
│   (Inter-service)     │  Users    │  │  Wallet   │  │  Order    │  │  Matching   │  │
│                       │  gRPC     │  │  gRPC     │  │  gRPC     │  │  gRPC       │  │
│                       └───────────┘  └─────▲─────┘  └─────▲─────┘  │  :50054     │  │
│                                            │              │        └──────┬──────┘  │
│                                            │   gRPC calls │               │          │
│                                            └──────────────┴───────────────┘          │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                      │
│   EVENT LAYER         ┌──────────────────────────────────────────────────────────┐   │
│   (NATS JetStream)    │                  NATS JetStream                          │   │
│                       │                                                          │   │
│                       │   TRADES stream ──▶ trades.{pair}.executed               │   │
│                       │   ORDERS stream ──▶ orders.{user}.{status}              │   │
│                       │   WALLETS stream ─▶ wallets.{user}.{action}             │   │
│                       │                                                          │   │
│                       └──────────────────────────┬───────────────────────────────┘   │
│                                                  │                                   │
│                                                  ▼                                   │
│                       ┌──────────────────────────────────────────────────────────┐   │
│                       │              notification-svc :8086                      │   │
│                       │                                                          │   │
│                       │   ┌─────────────────┐    ┌──────────────────────────┐    │   │
│                       │   │  NATS Consumer   │──▶│  Notification Service    │    │   │
│                       │   │  (3 durable)     │   │  (persist + broadcast)   │    │   │
│                       │   └─────────────────┘    └────────────┬─────────────┘    │   │
│                       │                                       │                  │   │
│                       │                              ┌────────▼────────┐         │   │
│                       │                              │  WebSocket Hub  │──WS──▶ User │
│                       │                              └─────────────────┘         │   │
│                       └──────────────────────────────────────────────────────────┘   │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                      │
│   DATA LAYER          ┌────────────────┐  ┌─────────┐  ┌────────────────────────┐   │
│                       │   PostgreSQL   │  │  Redis  │  │  NATS File Storage    │   │
│                       │                │  │         │  │  (stream persistence) │   │
│                       │  - users       │  │  - JWT  │  └────────────────────────┘   │
│                       │  - wallets     │  │  - cache│                                │
│                       │  - orders      │  │         │  ┌────────────────────────┐   │
│                       │  - trades      │  └─────────┘  │  OTel Collector       │   │
│                       │  - assets      │               │  → Tempo (traces)     │   │
│                       │  - pairs       │               │  → Loki (logs)        │   │
│                       │  - notifs      │               │  → Grafana (dashboard)│   │
│                       └────────────────┘               └────────────────────────┘   │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

## End-to-End Trading Flow (Saga Pattern)

```
 Client                order-svc           wallet-svc          matching-svc         notification-svc
   │                      │                    │                     │                      │
   │  POST /orders        │                    │                     │                      │
   │─────────────────────▶│                    │                     │                      │
   │                      │                    │                     │                      │
   │                      │  1. HoldBalance()  │                     │                      │
   │                      │───────────────────▶│                     │                      │
   │                      │    available -= X  │                     │                      │
   │                      │    locked += X     │                     │                      │
   │                      │◀───────────────────│                     │                      │
   │                      │                    │                     │                      │
   │                      │  2. INSERT order (status=new)            │                      │
   │                      │─────────┐          │                     │                      │
   │                      │◀────────┘          │                     │                      │
   │                      │                    │                     │                      │
   │                      │  3. SubmitOrder()  │                     │                      │
   │                      │────────────────────────────────────────▶│                      │
   │  201 Created         │                    │                     │                      │
   │◀─────────────────────│                    │                     │                      │
   │                      │                    │                     │                      │
   │                      │                    │    4. Match()       │                      │
   │                      │                    │    ┌────────────────┤                      │
   │                      │                    │    │ Order Book     │                      │
   │                      │                    │    │ price crossing │                      │
   │                      │                    │    │ FIFO matching  │                      │
   │                      │                    │    │ self-trade chk │                      │
   │                      │                    │    └───────────────▶│                      │
   │                      │                    │                     │                      │
   │                      │                    │  5. SettleTrade()   │                      │
   │                      │                    │◀────────────────────│                      │
   │                      │                    │    locked -= X      │                      │
   │                      │                    │                     │                      │
   │                      │                    │  6. Deposit()       │                      │
   │                      │                    │◀────────────────────│                      │
   │                      │                    │    available += X   │                      │
   │                      │                    │                     │                      │
   │                      │  7. UpdateFill()   │                     │                      │
   │                      │◀───────────────────────────────────────│                      │
   │                      │    filled_qty += X │                     │                      │
   │                      │    status=partial  │                     │                      │
   │                      │                    │                     │                      │
   │                      │                    │  8. NATS Publish    │                      │
   │                      │                    │    trades.*         │                      │
   │                      │                    │    orders.*         │                      │
   │                      │                    │    wallets.*        │                      │
   │                      │                    │                     │  9. Consume events   │
   │                      │                    │                     │─────────────────────▶│
   │                      │                    │                     │                      │
   │                      │                    │                     │  10. Create notif    │
   │                      │                    │                     │      + persist DB    │
   │                      │                    │                     │                      │
   │  11. WebSocket push  │                    │                     │                      │
   │◀─────────────────────────────────────────────────────────────────────────────────────│
   │     (real-time)      │                    │                     │                      │
```

---

## Order Cancellation Flow

```
 Client                order-svc           matching-svc         wallet-svc
   │                      │                     │                    │
   │  DELETE /orders/:id  │                     │                    │
   │─────────────────────▶│                     │                    │
   │                      │                     │                    │
   │                      │  1. Verify owner    │                    │
   │                      │  2. Check status    │                    │
   │                      │     (new|partial)   │                    │
   │                      │                     │                    │
   │                      │  3. CancelOrder()   │                    │
   │                      │────────────────────▶│                    │
   │                      │    remove from book │                    │
   │                      │◀────────────────────│                    │
   │                      │                     │                    │
   │                      │  4. ReleaseBalance()│                    │
   │                      │────────────────────────────────────────▶│
   │                      │                     │  available += X   │
   │                      │                     │  locked -= X      │
   │                      │◀───────────────────────────────────────│
   │                      │                     │                    │
   │                      │  5. UPDATE status=cancelled             │
   │                      │  6. NATS: orders.{user}.cancelled       │
   │                      │                     │                    │
   │  200 OK              │                     │                    │
   │◀─────────────────────│                     │                    │
```

---

## Matching Engine Internals

```
                         Engine (1 goroutine per trading pair)
                         ┌─────────────────────────────────────────────────┐
                         │                                                 │
   Submit/Cancel ───────▶│  cmdCh (buffered channel)                      │
                         │    │                                            │
                         │    ▼                                            │
                         │  ┌──────────────────────────────────────────┐  │
                         │  │            Order Book                    │  │
                         │  │                                          │  │
                         │  │  BIDS (sorted DESC)    ASKS (sorted ASC) │  │
                         │  │  ┌────────────────┐    ┌────────────────┐│  │
                         │  │  │ 50,100 ────────┤    │ 50,200 ────────┤│  │
                         │  │  │  [ord3, ord7]  │    │  [ord1, ord4]  ││  │
                         │  │  ├────────────────┤    ├────────────────┤│  │
                         │  │  │ 50,000 ────────┤    │ 50,300 ────────┤│  │
                         │  │  │  [ord2, ord5]  │    │  [ord6]        ││  │
                         │  │  ├────────────────┤    ├────────────────┤│  │
                         │  │  │ 49,900 ────────┤    │ 50,500 ────────┤│  │
                         │  │  │  [ord8]        │    │  [ord9, ord10] ││  │
                         │  │  └────────────────┘    └────────────────┘│  │
                         │  │                                          │  │
                         │  │  INDEX: map[orderID] → node  (O(1) del) │  │
                         │  └──────────────────────────────────────────┘  │
                         │    │                                            │
                         │    ▼                                            │
                         │  resultCh ───────▶ []Trade                     │
                         │                                                 │
                         └─────────────────────────────────────────────────┘

   Match Example: Incoming BUY at 50,300
   ┌──────────────────────────────────────────────────────────────────┐
   │  1. Check ASK levels (ascending):                               │
   │     50,200 ≤ 50,300? YES → match ord1, ord4 (FIFO)            │
   │     50,300 ≤ 50,300? YES → match ord6                          │
   │     50,500 ≤ 50,300? NO  → stop                                │
   │                                                                  │
   │  2. Trade price = resting (maker) price                         │
   │     Trade1: price=50,200  Trade2: price=50,200  Trade3: 50,300  │
   │                                                                  │
   │  3. If incoming has remaining qty → add to BIDS at 50,300       │
   └──────────────────────────────────────────────────────────────────┘
```

---

## Wallet State Machine

```
                    ┌─────────────────────────────────────────────┐
                    │              Wallet Balance                  │
                    │                                             │
                    │   ┌─────────────┐      ┌─────────────┐     │
                    │   │  AVAILABLE  │      │   LOCKED    │     │
                    │   │             │      │             │     │
                    │   │  Free to    │ Hold │  Reserved   │     │
                    │   │  use/trade  │─────▶│  for open   │     │
                    │   │             │      │  orders     │     │
                    │   │             │◀─────│             │     │
                    │   │             │Release│            │     │
                    │   └──────▲──────┘      └──────┬──────┘     │
                    │          │                     │             │
                    │          │ Deposit             │ Settle      │
                    │          │ (receive            │ (trade      │
                    │          │  asset)             │  matched)   │
                    │          │                     │             │
                    │          │                     ▼             │
                    │   ┌──────┴──────┐      ┌─────────────┐     │
                    │   │  EXTERNAL   │      │  REMOVED    │     │
                    │   │  (trade     │      │  (funds     │     │
                    │   │   deposit)  │      │   consumed) │     │
                    │   └─────────────┘      └─────────────┘     │
                    │                                             │
                    └─────────────────────────────────────────────┘

   Example: BUY 1 BTC @ 50,000 USDT

   Phase 1: Order Created (Hold)          Phase 2: Trade Settled
   ┌────────────────────────────┐         ┌────────────────────────────┐
   │ Buyer USDT                 │         │ Buyer USDT                 │
   │   avail: 100k → 50k       │         │   avail: 50k  (unchanged)  │
   │   locked:  0  → 50k       │         │   locked: 50k → 0 (settle) │
   │                            │         │                            │
   │ Seller BTC                 │         │ Buyer BTC                  │
   │   avail: 1 → 0            │         │   avail: 0 → 1  (deposit)  │
   │   locked: 0 → 1           │         │                            │
   └────────────────────────────┘         │ Seller BTC                 │
                                          │   locked: 1 → 0  (settle)  │
                                          │                            │
                                          │ Seller USDT                │
                                          │   avail: 0 → 50k (deposit) │
                                          └────────────────────────────┘
```

---

## NATS JetStream Event Flow

```
   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
   │ matching-svc│   │  order-svc  │   │ wallet-svc  │
   │ (publisher) │   │ (publisher) │   │ (publisher) │
   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
          │                  │                  │
          │  trades.*        │  orders.*        │  wallets.*
          │  Msg-Id:         │  Msg-Id:         │  Msg-Id:
          │  {trade_id}      │  {ord_id}_{st}   │  {tx_id}
          ▼                  ▼                  ▼
   ┌──────────────────────────────────────────────────────┐
   │                  NATS JetStream                       │
   │                                                       │
   │   ┌──────────┐   ┌──────────┐   ┌──────────┐        │
   │   │ TRADES   │   │ ORDERS   │   │ WALLETS  │        │
   │   │ stream   │   │ stream   │   │ stream   │        │
   │   │ (file)   │   │ (file)   │   │ (file)   │        │
   │   └────┬─────┘   └────┬─────┘   └────┬─────┘        │
   │        │              │              │               │
   │   Dedup window   Dedup window   Dedup window         │
   │   (Nats-Msg-Id)  (Nats-Msg-Id)  (Nats-Msg-Id)       │
   └────────┼──────────────┼──────────────┼───────────────┘
            │              │              │
            ▼              ▼              ▼
   ┌──────────────────────────────────────────────────────┐
   │              notification-svc (consumers)             │
   │                                                       │
   │   ┌────────────────┐  ┌───────────────┐  ┌────────┐ │
   │   │notif-trades    │  │notif-orders   │  │notif-  │ │
   │   │(durable)       │  │(durable)      │  │wallets │ │
   │   │fetch:10 / 2s   │  │fetch:10 / 2s  │  │(dur.)  │ │
   │   └───────┬────────┘  └───────┬───────┘  └───┬────┘ │
   │           │                   │               │      │
   │           ▼                   ▼               ▼      │
   │   ┌─────────────────────────────────────────────┐    │
   │   │           Event Handler (router)            │    │
   │   │                                             │    │
   │   │  trades.* → handleTradeEvent()              │    │
   │   │    → 2 notifs (buyer + seller)              │    │
   │   │    → dedup: trade_{id}_{user}               │    │
   │   │                                             │    │
   │   │  orders.* → handleOrderEvent()              │    │
   │   │    → 1 notif (filled|cancelled)             │    │
   │   │    → dedup: order_{id}_{status}             │    │
   │   │                                             │    │
   │   │  wallets.* → handleWalletEvent()            │    │
   │   │    → 1 notif (deposit|withdrawal)           │    │
   │   │    → dedup: wallet_{tx_id}                  │    │
   │   └──────────────────┬──────────────────────────┘    │
   │                      │                                │
   │                      ▼                                │
   │   ┌──────────────────────────────────────────────┐   │
   │   │  Notification Service                        │   │
   │   │    1. INSERT notif (event_key UNIQUE)        │   │
   │   │    2. WebSocket Hub → SendToUser()           │   │
   │   └──────────────────────────────────────────────┘   │
   │                      │                                │
   │                      ▼                                │
   │   ┌──────────────────────────────────────────────┐   │
   │   │  WebSocket Hub                               │   │
   │   │    connections: map[userID][]safeConn         │   │
   │   │    → RLock, deep-copy, RUnlock               │   │
   │   │    → async write to each conn                │   │
   │   │    → auto-remove dead conns                  │   │
   │   └──────────────────┬───────────────────────────┘   │
   │                      │                                │
   └──────────────────────┼────────────────────────────────┘
                          │
                          ▼
                    ┌───────────┐
                    │  Client   │
                    │  (WS)     │
                    └───────────┘
```

---

## 1. Order Creation

**Endpoint:** `POST /api/v1/orders`
**Handler:** `cmd/http/order/create_order.go`

### Flow

```
Client ──▶ Fiber Handler ──▶ Order Service ──▶ Wallet Hold ──▶ DB Insert ──▶ Submit to Matching
```

### Step-by-Step

1. **Validate input** -- price > 0, qty >= minQty, price % tickSize == 0, side (buy/sell), type (limit)
2. **Hold balance** via gRPC to wallet-svc
   - BUY: hold `price x quantity` of quote asset (e.g. USDT)
   - SELL: hold `quantity` of base asset (e.g. BTC)
3. **Persist order** to Postgres (status = `new`, filled_qty = 0)
4. **Submit to matching engine** via gRPC (fire-and-forget from order-svc perspective)
5. **Return** `201 Created` with order details (matching happens async)

### Compensation

- If DB insert fails after hold -> release wallet hold
- If matching submit fails -> order stays in DB as `new` (can be cancelled later)

---

## 2. Matching Engine

**Core:** `modules/matching/engine/`

### Architecture

- One goroutine per trading pair, serialized via command channel
- Dual-sided order book: bids (descending), asks (ascending)
- Each price level holds a FIFO queue of orders
- O(1) cancellation via index map

```
Engine (per pair)
├── cmdCh: chan Command    // Submit, Cancel, Stop
├── OrderBook
│   ├── bids: sortedLevels (price DESC)
│   │   └── priceLevel → FIFO queue [order1, order2, ...]
│   ├── asks: sortedLevels (price ASC)
│   │   └── priceLevel → FIFO queue [order1, order2, ...]
│   └── index: map[orderID]*orderNode  // O(1) lookup
```

### Matching Algorithm

```
func Match(incoming Order) []Trade:
    oppositeSide = asks if BUY, bids if SELL
    
    for each priceLevel in oppositeSide (snapshot):
        if !pricesCross(incoming, level):
            break
        
        for each resting in level.queue:
            if resting.UserID == incoming.UserID:
                skip  // self-trade prevention
            
            fillQty = min(incoming.Remaining, resting.Remaining)
            tradePrice = resting.Price  // maker price
            
            create Trade{fillQty, tradePrice, buyer, seller}
            
            incoming.Remaining -= fillQty
            resting.Remaining -= fillQty
            
            if resting fully filled: remove from book
            if incoming fully filled: return trades
    
    if incoming has remaining: add to book
    return trades
```

### Price Crossing Rules

- BUY order matches when `incoming.Price >= resting.Price`
- SELL order matches when `incoming.Price <= resting.Price`
- Trade always executes at the **resting (maker) price**

---

## 3. Trade Settlement

**Service:** `modules/matching/application/services/matching_service.go`

### Flow

For each trade produced by the matching engine, 4 wallet operations execute via gRPC:

```
Trade: Buyer buys 1 BTC at 50,000 USDT from Seller

1. SettleTrade(buyer,  USDT, 50000)  // buyer's locked USDT -= 50000
2. Deposit    (buyer,  BTC,  1)      // buyer's available BTC += 1
3. SettleTrade(seller, BTC,  1)      // seller's locked BTC -= 1
4. Deposit    (seller, USDT, 50000)  // seller's available USDT += 50000
```

### Order Fill Update

After settlement, update both orders via gRPC to order-svc:

```
UpdateOrderFill(buyOrderID,  newFilledQty, "partial"|"filled")
UpdateOrderFill(sellOrderID, newFilledQty, "partial"|"filled")
```

**Guard:** `filled_qty` is monotonically increasing -- SQL rejects backward fills.

---

## 4. Wallet State Machine

**Service:** `modules/wallet/application/services/wallet_service.go`

### States

Each wallet has two balances: `available` and `locked`.

```
                    Hold
    available ──────────────▶ locked
        ▲                       │
        │      Release          │
        └───────────────────────┘
        │                       │
        │ Deposit          Settle
        │                       │
        ▲                       ▼
    (external)              (removed)
```

### Operations

| Operation | Available | Locked | When |
|-----------|-----------|--------|------|
| **Hold** | -= amount | += amount | Order created |
| **Release** | += amount | -= amount | Order cancelled |
| **Settle** | (no change) | -= amount | Trade matched (unlock maker's funds) |
| **Deposit** | += amount | (no change) | Trade matched (receive asset) |

### SQL Atomicity

All wallet operations use atomic SQL with balance guards:

```sql
-- Hold: fails if insufficient available
UPDATE wallets SET available = available - $3, locked = locked + $3
WHERE user_id = $1 AND asset_id = $2 AND available >= $3;

-- Settle: fails if insufficient locked
UPDATE wallets SET locked = locked - $3
WHERE user_id = $1 AND asset_id = $2 AND locked >= $3;
```

### Example: Buy 1 BTC at 50,000 USDT

| Phase | User | Asset | Available | Locked |
|-------|------|-------|-----------|--------|
| **Initial** | Buyer | USDT | 100,000 | 0 |
| | Seller | BTC | 1 | 0 |
| **After Hold** | Buyer | USDT | 50,000 | 50,000 |
| | Seller | BTC | 0 | 1 |
| **After Settlement** | Buyer | USDT | 50,000 | 0 |
| | Buyer | BTC | 1 | 0 |
| | Seller | BTC | 0 | 0 |
| | Seller | USDT | 50,000 | 0 |

---

## 5. Order Cancellation

**Endpoint:** `DELETE /api/v1/orders/:id`

### Flow

```
1. Verify ownership (user_id matches)
2. Check cancellable (status = new | partial)
3. Remove from matching engine order book (gRPC, best-effort)
4. Release wallet hold (gRPC)
   - BUY:  release price x remainingQty of quote asset
   - SELL: release remainingQty of base asset
5. Update order status = cancelled in DB
6. Publish cancellation event to NATS
```

### Compensation

- If wallet release fails -> cancel fails, balance stays locked (safe)
- If DB update fails after release -> attempt to re-hold (prevent double-release)

---

## 6. Event System (NATS JetStream)

**Publishers:** `pkg/nats/`

### Streams

| Stream | Subjects | Storage |
|--------|----------|---------|
| TRADES | `trades.>` | File |
| ORDERS | `orders.>` | File |
| WALLETS | `wallets.>` | File |

### Events Published

| Event | Subject | Dedup Key | Trigger |
|-------|---------|-----------|---------|
| Trade executed | `trades.{pair}.executed` | `trade_id` | After settlement |
| Order status change | `orders.{user}.{status}` | `order_id_status` | After fill update or cancel |
| Wallet action | `wallets.{user}.{action}` | `tx_id` | After deposit/withdrawal |

### Delivery Guarantees

- **At-least-once** delivery with JetStream durable consumers
- **Idempotency** via `Nats-Msg-Id` header (JetStream dedup window)
- **Retry:** NAK with 5s delay, max 5 deliveries
- **Poison pill:** Ack after max retries to prevent blocking

---

## 7. Notification System

**Consumer:** `modules/notification/infrastructure/consumer/`
**WebSocket:** `modules/notification/infrastructure/ws/`

### Consumer Architecture

```
NATS JetStream
├── notif-trades-consumer  (durable) ──▶ handleTradeEvent()
├── notif-orders-consumer  (durable) ──▶ handleOrderEvent()
└── notif-wallets-consumer (durable) ──▶ handleWalletEvent()
```

Each consumer fetches 10 messages with 2s timeout, processes sequentially.

### Notification Types

| NATS Event | Notification Type | Recipients |
|------------|-------------------|------------|
| Trade executed | `trade_executed` | Buyer + Seller (2 notifications) |
| Order filled | `order_filled` | Order owner |
| Order cancelled | `order_cancelled` | Order owner |
| Deposit confirmed | `deposit_confirmed` | Wallet owner |
| Withdrawal confirmed | `withdrawal_confirmed` | Wallet owner |

### Dedup Keys

- `trade_{tradeID}_{userID}` -- separate for buyer and seller
- `order_{orderID}_{status}` -- one per status change
- `wallet_{txID}` -- one per transaction

### WebSocket Broadcasting

```
Hub
├── connections: map[userID][]safeConn   // RWMutex protected
├── Register(userID, conn)
├── Unregister(userID, conn)
└── SendToUser(userID, notification)
    1. RLock -> deep-copy connections slice
    2. RUnlock
    3. Async write to each conn (failures don't block others)
    4. Auto-remove dead connections on write error
```

---

## 8. Inter-Service Communication

### Service Dependencies

```
                    ┌─────────────────┐
                    │   Traefik (LB)  │
                    └────────┬────────┘
                             │ HTTP
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
         users-svc      order-svc      wallet-svc
         :8081/:50051   :8083/:50053   :8082/:50052
                             │              ▲
                             │ gRPC         │ gRPC
                             ▼              │
                       matching-svc ────────┘
                       :8084/:50054
                             │
                             │ NATS
                             ▼
                    notification-svc ──WS──▶ clients
                    :8086
                             │
                             │ NATS
                             ▼
                       market-svc
                       :8085
```

### gRPC Calls

| Caller | Callee | Methods |
|--------|--------|---------|
| order-svc | wallet-svc | HoldBalance, ReleaseBalance |
| order-svc | matching-svc | SubmitOrder, CancelOrder |
| matching-svc | wallet-svc | SettleTrade, Deposit |
| matching-svc | order-svc | UpdateOrderFill |

### Error Mapping (gRPC -> Domain)

```
codes.InvalidArgument    -> ErrInsufficientBalance
codes.NotFound           -> ErrWalletNotFound / ErrOrderNotFound
codes.Unavailable        -> ErrServiceUnavailable
codes.DeadlineExceeded   -> ErrServiceUnavailable
```

---

## 9. Database Schema

### Core Tables

```sql
-- Orders
orders (id UUID PK, user_id, pair_id, side, type, price NUMERIC,
        quantity NUMERIC, filled_qty NUMERIC, status, created_at, updated_at)
-- Indexes: user_id, (pair_id, status), (user_id, created_at)

-- Trades
trades (id UUID PK, pair_id, buy_order_id, sell_order_id,
        price NUMERIC, quantity NUMERIC, buyer_id, seller_id, executed_at)
-- Indexes: (pair_id, executed_at), buyer_id, seller_id

-- Wallets
wallets (id UUID PK, user_id, asset_id, available NUMERIC, locked NUMERIC, updated_at)
-- Unique: (user_id, asset_id)

-- Notifications
notifications (id UUID PK, user_id, event_key UNIQUE, type, title, body,
              is_read, metadata JSONB, created_at)
```

### Financial Precision

- All monetary values use `NUMERIC` in Postgres (arbitrary precision)
- Go layer uses `shopspring/decimal` (no floating point)

---

## 10. Concurrency Model

| Component | Strategy | Purpose |
|-----------|----------|---------|
| Matching engine | 1 goroutine per pair + command channel | Serialize order book mutations |
| WebSocket hub | RWMutex + deep-copy before send | Thread-safe connection registry |
| NATS consumer | 1 goroutine per stream | Isolated processing per event type |
| Wallet operations | Atomic SQL with balance guards | Prevent negative balances |
| Order fills | Monotonic filled_qty guard | Prevent backward fills |

---

## 11. Key Invariants

1. **Balance consistency** -- `available + locked` is always conserved across hold/release cycles
2. **Fill monotonicity** -- `filled_qty` can only increase, SQL rejects backward updates
3. **Self-trade prevention** -- matching engine skips orders from the same user
4. **Idempotent notifications** -- `event_key` unique constraint prevents duplicate notifications
5. **NATS dedup** -- `Nats-Msg-Id` prevents duplicate event processing within dedup window
6. **Atomic wallet ops** -- SQL guards ensure `available >= amount` before hold, `locked >= amount` before settle
