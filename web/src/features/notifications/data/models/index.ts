import { Service } from "@/core/api/service";
import { NOTIFICATION_ENDPOINT } from "@/core/api/endpoint";
import type { INotification } from "@/core/api/types";

interface NotificationListResponse {
  notifications: INotification[];
}

interface UnreadCountResponse {
  count: number;
}

export class NotificationModel {
  private static service = new Service(NOTIFICATION_ENDPOINT.LIST);

  static getAll(): Promise<NotificationListResponse> {
    return NotificationModel.service.get<NotificationListResponse>(
      NOTIFICATION_ENDPOINT.LIST
    );
  }

  static getUnreadCount(): Promise<UnreadCountResponse> {
    return NotificationModel.service.get<UnreadCountResponse>(
      NOTIFICATION_ENDPOINT.UNREAD_COUNT
    );
  }

  static markRead(id: string): Promise<INotification> {
    return NotificationModel.service.post<INotification>(
      {},
      NOTIFICATION_ENDPOINT.READ(id)
    );
  }

  static markAllRead(): Promise<void> {
    return NotificationModel.service.post<void>({}, NOTIFICATION_ENDPOINT.READ_ALL);
  }
}
