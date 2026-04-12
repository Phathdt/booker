import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export const TRADING_PAIRS = [
  { id: "BTC_USDT", label: "BTC / USDT" },
  { id: "ETH_USDT", label: "ETH / USDT" },
];

interface PairSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

export function PairSelector({ value, onChange }: PairSelectorProps) {
  return (
    <Select value={value} onValueChange={(v) => v && onChange(v)}>
      <SelectTrigger className="w-40">
        <SelectValue placeholder="Select pair" />
      </SelectTrigger>
      <SelectContent>
        {TRADING_PAIRS.map((pair) => (
          <SelectItem key={pair.id} value={pair.id}>
            {pair.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
