import { useQuery } from "@tanstack/react-query";
import { NotificationModel } from "../models";

export const NOTIFICATION_QUERY_KEYS = {
  LIST: "notifications",
  UNREAD_COUNT: "notifications-unread-count",
};

export function useQueryNotifications() {
  return useQuery({
    queryKey: [NOTIFICATION_QUERY_KEYS.LIST],
    queryFn: () => NotificationModel.getAll(),
  });
}

export function useQueryUnreadCount() {
  return useQuery({
    queryKey: [NOTIFICATION_QUERY_KEYS.UNREAD_COUNT],
    queryFn: () => NotificationModel.getUnreadCount(),
    refetchInterval: 10000,
  });
}
