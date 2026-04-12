import { Link } from "react-router-dom";
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
import { useQueryOrders } from "../data/queries";
import { useMutationCancelOrder } from "../data/mutations";
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
  const { data, isLoading } = useQueryOrders(pairId);
  const { mutate: cancel, isPending: isCancelling } = useMutationCancelOrder();

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
              {formatDate(order.created_at)}
            </TableCell>
            <TableCell className="font-medium">
              <Link
                to={`/orders/${order.id}`}
                className="hover:underline"
              >
                {order.pair_id}
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
            <TableCell className="text-right font-mono">{order.filled_qty}</TableCell>
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
                  onClick={() => cancel(order.id)}
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
