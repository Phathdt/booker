import { Service } from "@/core/api/service";
import { ORDER_ENDPOINT } from "@/core/api/endpoint";
import type { IOrder } from "@/core/api/types";

interface OrderListResponse {
  orders: IOrder[];
}

interface CreateOrderPayload {
  pair_id: string;
  side: "buy" | "sell";
  type: "limit";
  price: string;
  quantity: string;
}

interface OrderListParams {
  pair_id?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export class OrderModel {
  private static service = new Service(ORDER_ENDPOINT.LIST);

  static getAll(params?: OrderListParams): Promise<OrderListResponse> {
    return OrderModel.service.get<OrderListResponse>(ORDER_ENDPOINT.LIST, params);
  }

  static getById(id: string): Promise<IOrder> {
    return OrderModel.service.get<IOrder>(`${ORDER_ENDPOINT.LIST}/${id}`);
  }

  static create(payload: CreateOrderPayload): Promise<IOrder> {
    return OrderModel.service.post<IOrder>(payload, ORDER_ENDPOINT.CREATE);
  }

  static cancel(id: string): Promise<IOrder> {
    return OrderModel.service.delete<IOrder>(`${ORDER_ENDPOINT.LIST}/${id}`);
  }
}
