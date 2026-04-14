import {
  getApiV1Notifications,
  patchApiV1NotificationsIdRead,
  postApiV1NotificationsReadAll,
  getApiV1NotificationsUnreadCount,
} from "@/core/api/generated/notifications/notifications";
import type { INotification } from "@/core/api/types";

interface NotificationListResponse {
  notifications: INotification[];
}

interface UnreadCountResponse {
  count: number;
}

export class NotificationModel {
  static getAll(): Promise<NotificationListResponse> {
    return getApiV1Notifications() as Promise<NotificationListResponse>;
  }

  static getUnreadCount(): Promise<UnreadCountResponse> {
    return getApiV1NotificationsUnreadCount() as Promise<UnreadCountResponse>;
  }

  static markRead(id: string): Promise<INotification> {
    return patchApiV1NotificationsIdRead(id) as unknown as Promise<INotification>;
  }

  static markAllRead(): Promise<void> {
    return postApiV1NotificationsReadAll() as unknown as Promise<void>;
  }
}
