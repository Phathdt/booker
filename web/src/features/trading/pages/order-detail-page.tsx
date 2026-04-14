import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { Layout } from "@/components/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getApiV1OrdersId, useDeleteApiV1OrdersId } from "@/core/api/generated/orders/orders";
import type { IOrder } from "@/core/api/types";

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

interface DetailRowProps {
  label: string;
  value: React.ReactNode;
}

function DetailRow({ label, value }: DetailRowProps) {
  return (
    <div className="flex items-center justify-between border-b py-3 last:border-0">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: order, isLoading, isError } = useQuery({
    queryKey: ["order", id],
    queryFn: () => getApiV1OrdersId(id!),
    enabled: Boolean(id),
  });

  const { mutate: cancel, isPending: isCancelling } = useDeleteApiV1OrdersId({
    mutation: {
      onSuccess: () => navigate("/trade"),
    },
  });

  const canCancel = order?.status === "new" || order?.status === "partial";

  if (isLoading) {
    return (
      <Layout>
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
        </div>
      </Layout>
    );
  }

  if (isError || !order) {
    return (
      <Layout>
        <div className="flex h-64 flex-col items-center justify-center gap-4">
          <p className="text-muted-foreground">Order not found.</p>
          <Button variant="outline" onClick={() => navigate("/trade")}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Trade
          </Button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="mx-auto max-w-lg space-y-4">
        {/* Back button */}
        <Button variant="ghost" size="sm" onClick={() => navigate("/trade")}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Trade
        </Button>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">Order Detail</CardTitle>
              <Badge variant={statusVariant(order.status)}>{order.status}</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <DetailRow label="Order ID" value={<span className="font-mono text-xs">{order.id}</span>} />
            <DetailRow label="Pair" value={order.pairId} />
            <DetailRow
              label="Side"
              value={
                <span
                  className={
                    order.side === "buy"
                      ? "font-semibold text-green-600"
                      : "font-semibold text-red-600"
                  }
                >
                  {order.side.toUpperCase()}
                </span>
              }
            />
            <DetailRow label="Type" value={order.type.toUpperCase()} />
            <DetailRow label="Price" value={<span className="font-mono">{order.price}</span>} />
            <DetailRow label="Quantity" value={<span className="font-mono">{order.quantity}</span>} />
            <DetailRow label="Filled" value={<span className="font-mono">{order.filledQty}</span>} />
            <DetailRow label="Created" value={formatDate(order.createdAt)} />
            <DetailRow label="Updated" value={formatDate(order.updatedAt)} />
          </CardContent>
        </Card>

        {canCancel && (
          <Button
            variant="destructive"
            className="w-full"
            disabled={isCancelling}
            onClick={() => {
              cancel({ id: order.id });
            }}
          >
            {isCancelling ? "Cancelling..." : "Cancel Order"}
          </Button>
        )}
      </div>
    </Layout>
  );
}
