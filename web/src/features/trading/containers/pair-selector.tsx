import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useGetApiV1MarketPairs } from "@/core/api/generated/market/market";

interface PairSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

export function PairSelector({ value, onChange }: PairSelectorProps) {
  const { data: pairs = [], isLoading } = useGetApiV1MarketPairs({ query: { staleTime: 60000 } });

  return (
    <Select value={value} onValueChange={(v) => v && onChange(v)} disabled={isLoading}>
      <SelectTrigger className="w-40">
        <SelectValue placeholder={isLoading ? "Loading..." : "Select pair"} />
      </SelectTrigger>
      <SelectContent>
        {pairs.map((pair) => (
          <SelectItem key={pair.id} value={pair.id}>
            {pair.baseAsset} / {pair.quoteAsset}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
