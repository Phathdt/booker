import { useQueryNotifications } from "../data/queries";
import { useMutationMarkRead, useMutationMarkAllRead } from "../data/mutations";
import { Button } from "@/components/ui/button";
import type { INotification } from "@/core/api/types";

function formatTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
}

interface NotificationItemProps {
  notification: INotification;
  onMarkRead: (id: string) => void;
}

function NotificationItem({ notification, onMarkRead }: NotificationItemProps) {
  return (
    <button
      type="button"
      className={`w-full text-left px-4 py-3 flex gap-3 hover:bg-muted/50 transition-colors border-b border-border last:border-b-0 ${
        !notification.is_read ? "bg-muted/20" : ""
      }`}
      onClick={() => {
        if (!notification.is_read) {
          onMarkRead(notification.id);
        }
      }}
    >
      <div className="mt-1.5 shrink-0">
        <span
          className={`block h-2 w-2 rounded-full ${
            notification.is_read ? "bg-transparent" : "bg-blue-500"
          }`}
          aria-label={notification.is_read ? "read" : "unread"}
        />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-foreground truncate">
          {notification.title}
        </p>
        <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
          {notification.body}
        </p>
        <p className="text-xs text-muted-foreground mt-1">
          {formatTime(notification.created_at)}
        </p>
      </div>
    </button>
  );
}

export function NotificationList() {
  const { data, isLoading } = useQueryNotifications();
  const markRead = useMutationMarkRead();
  const markAllRead = useMutationMarkAllRead();

  const notifications = (data?.notifications ?? []).slice(0, 10);

  return (
    <div className="flex flex-col" role="region" aria-label="Notifications">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border">
        <span className="text-sm font-semibold">Notifications</span>
        <Button
          variant="ghost"
          size="sm"
          className="text-xs h-auto py-1"
          onClick={() => markAllRead.mutate()}
          disabled={markAllRead.isPending || notifications.every((n) => n.is_read)}
        >
          Mark all read
        </Button>
      </div>

      <div className="overflow-y-auto max-h-80">
        {isLoading ? (
          <div className="flex items-center justify-center h-20">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-muted border-t-primary" />
          </div>
        ) : notifications.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-8">
            No notifications
          </p>
        ) : (
          notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onMarkRead={(id) => markRead.mutate(id)}
            />
          ))
        )}
      </div>
    </div>
  );
}
