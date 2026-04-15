import { useState } from "react";
import Big from "big.js";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateOrder, getListOrdersQueryKey } from "@/core/api/generated/orders/orders";
import { cn } from "@/lib/utils";

interface OrderFormProps {
  pairId: string;
}

function computeTotal(price: string, quantity: string): string {
  try {
    const p = new Big(price);
    const q = new Big(quantity);
    if (p.lte(0) || q.lte(0)) return "0.00";
    return p.times(q).toFixed(8);
  } catch {
    return "0.00";
  }
}

interface SideFormProps {
  side: "buy" | "sell";
  pairId: string;
}

function SideForm({ side, pairId }: SideFormProps) {
  const [price, setPrice] = useState("");
  const [quantity, setQuantity] = useState("");
  const queryClient = useQueryClient();
  const { mutate, isPending } = useCreateOrder({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getListOrdersQueryKey() });
        toast.success("Order placed");
        setPrice("");
        setQuantity("");
      },
      onError: (err: unknown) => {
        const message = err instanceof Error ? err.message : "Failed to place order";
        toast.error(message);
      },
    },
  });

  const total = computeTotal(price, quantity);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!price || !quantity) return;
    mutate({ data: { pairId: pairId, side, type: "limit", price, quantity } });
  };

  const isBuy = side === "buy";
  const buttonClass = isBuy
    ? "w-full bg-green-600 hover:bg-green-700 text-white"
    : "w-full bg-red-600 hover:bg-red-700 text-white";

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="space-y-1.5">
        <Label htmlFor={`${side}-price`}>Price (USDT)</Label>
        <Input
          id={`${side}-price`}
          type="number"
          placeholder="0.00"
          step="any"
          min="0"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          required
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor={`${side}-quantity`}>
          Quantity ({pairId.split("_")[0]})
        </Label>
        <Input
          id={`${side}-quantity`}
          type="number"
          placeholder="0.00000000"
          step="any"
          min="0"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          required
        />
      </div>
      <div className="rounded-md bg-muted px-3 py-2 text-sm">
        <span className="text-muted-foreground">Total: </span>
        <span className="font-mono font-medium">{total} USDT</span>
      </div>
      <Button
        type="submit"
        disabled={isPending}
        className={cn(buttonClass)}
      >
        {isPending
          ? "Placing..."
          : `${isBuy ? "Buy" : "Sell"} ${pairId.split("_")[0]}`}
      </Button>
    </form>
  );
}

export function OrderForm({ pairId }: OrderFormProps) {
  return (
    <Tabs defaultValue="buy">
      <TabsList className="w-full">
        <TabsTrigger
          value="buy"
          className="flex-1 data-[state=active]:bg-green-600 data-[state=active]:text-white"
        >
          Buy
        </TabsTrigger>
        <TabsTrigger
          value="sell"
          className="flex-1 data-[state=active]:bg-red-600 data-[state=active]:text-white"
        >
          Sell
        </TabsTrigger>
      </TabsList>
      <TabsContent value="buy" className="pt-3">
        <SideForm side="buy" pairId={pairId} />
      </TabsContent>
      <TabsContent value="sell" className="pt-3">
        <SideForm side="sell" pairId={pairId} />
      </TabsContent>
    </Tabs>
  );
}
