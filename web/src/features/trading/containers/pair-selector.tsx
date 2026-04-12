import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useQueryPairs } from "../data/queries";

interface PairSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

export function PairSelector({ value, onChange }: PairSelectorProps) {
  const { data, isLoading } = useQueryPairs();
  const pairs = data?.pairs ?? [];

  return (
    <Select value={value} onValueChange={(v) => v && onChange(v)} disabled={isLoading}>
      <SelectTrigger className="w-40">
        <SelectValue placeholder={isLoading ? "Loading..." : "Select pair"} />
      </SelectTrigger>
      <SelectContent>
        {pairs.map((pair) => (
          <SelectItem key={pair.id} value={pair.id}>
            {pair.base_asset} / {pair.quote_asset}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
