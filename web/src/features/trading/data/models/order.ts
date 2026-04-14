import {
  getApiV1Orders,
  postApiV1Orders,
  getApiV1OrdersId,
  deleteApiV1OrdersId,
} from "@/core/api/generated/orders/orders";
import type { IOrder } from "@/core/api/types";
import type { GetApiV1OrdersParams } from "@/core/api/generated/models";

interface OrderListResponse {
  orders: IOrder[];
}

export class OrderModel {
  static getAll(params?: GetApiV1OrdersParams): Promise<OrderListResponse> {
    return getApiV1Orders(params) as Promise<OrderListResponse>;
  }

  static getById(id: string): Promise<IOrder> {
    return getApiV1OrdersId(id) as Promise<IOrder>;
  }

  static create(payload: {
    pair_id: string;
    side: string;
    type: string;
    price: string;
    quantity: string;
  }): Promise<IOrder> {
    return postApiV1Orders(payload) as Promise<IOrder>;
  }

  static cancel(id: string): Promise<IOrder> {
    return deleteApiV1OrdersId(id) as Promise<IOrder>;
  }
}
