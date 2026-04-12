import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { NotificationModel } from "../models";
import { NOTIFICATION_QUERY_KEYS } from "../queries";

export function useMutationMarkRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => NotificationModel.markRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATION_QUERY_KEYS.LIST] });
      queryClient.invalidateQueries({ queryKey: [NOTIFICATION_QUERY_KEYS.UNREAD_COUNT] });
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Failed to mark notification as read");
    },
  });
}

export function useMutationMarkAllRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => NotificationModel.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATION_QUERY_KEYS.LIST] });
      queryClient.invalidateQueries({ queryKey: [NOTIFICATION_QUERY_KEYS.UNREAD_COUNT] });
      toast.success("All notifications marked as read");
    },
    onError: (err: { message: string }) => {
      toast.error(err.message ?? "Failed to mark all notifications as read");
    },
  });
}
