import { Link } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useGetApiV1Orders, useDeleteApiV1OrdersId, getGetApiV1OrdersQueryKey } from "@/core/api/generated/orders/orders";
import type { IOrder } from "@/core/api/types";

interface OpenOrdersProps {
  pairId: string;
}

function statusVariant(
  status: IOrder["status"]
): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "new":
      return "default";
    case "partial":
      return "secondary";
    case "filled":
      return "outline";
    case "cancelled":
      return "destructive";
    default:
      return "outline";
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString();
}

export function OpenOrders({ pairId }: OpenOrdersProps) {
  const queryClient = useQueryClient();
  const { data, isLoading } = useGetApiV1Orders(
    { pairId },
    { query: { refetchInterval: 5000 } }
  );
  const { mutate: cancel, isPending: isCancelling } = useDeleteApiV1OrdersId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1OrdersQueryKey() });
        toast.success("Order cancelled");
      },
      onError: (err: unknown) => {
        const message = err instanceof Error ? err.message : "Failed to cancel order";
        toast.error(message);
      },
    },
  });

  const orders = data?.orders ?? [];

  if (isLoading) {
    return (
      <div className="flex h-24 items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="flex h-24 items-center justify-center text-sm text-muted-foreground">
        No orders found
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Time</TableHead>
          <TableHead>Pair</TableHead>
          <TableHead>Side</TableHead>
          <TableHead className="text-right">Price</TableHead>
          <TableHead className="text-right">Qty</TableHead>
          <TableHead className="text-right">Filled</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Action</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {orders.map((order) => (
          <TableRow key={order.id}>
            <TableCell className="text-xs text-muted-foreground">
              {formatDate(order.createdAt)}
            </TableCell>
            <TableCell className="font-medium">
              <Link
                to={`/orders/${order.id}`}
                className="hover:underline"
              >
                {order.pairId}
              </Link>
            </TableCell>
            <TableCell>
              <span
                className={
                  order.side === "buy"
                    ? "font-medium text-green-600"
                    : "font-medium text-red-600"
                }
              >
                {order.side.toUpperCase()}
              </span>
            </TableCell>
            <TableCell className="text-right font-mono">{order.price}</TableCell>
            <TableCell className="text-right font-mono">{order.quantity}</TableCell>
            <TableCell className="text-right font-mono">{order.filledQty}</TableCell>
            <TableCell>
              <Badge variant={statusVariant(order.status)}>
                {order.status}
              </Badge>
            </TableCell>
            <TableCell className="text-right">
              {(order.status === "new" || order.status === "partial") && (
                <Button
                  size="sm"
                  variant="destructive"
                  disabled={isCancelling}
                  onClick={() => cancel({ id: order.id })}
                >
                  Cancel
                </Button>
              )}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
