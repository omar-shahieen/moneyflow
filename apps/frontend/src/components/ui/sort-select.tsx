import { ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";

type SortOption = {
  value: string;
  label: string;
};

type SortSelectProps = {
  options: SortOption[];
  value?: string;
  order?: "asc" | "desc";
  onChange: (sort: string, order: "asc" | "desc") => void;
  className?: string;
};

export function SortSelect({
  options,
  value,
  order = "desc",
  onChange,
  className,
}: SortSelectProps) {
  const toggleOrder = () => {
    onChange(value ?? options[0]?.value ?? "", order === "asc" ? "desc" : "asc");
  };

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Select
        value={value ?? ""}
        onChange={(e) => onChange(e.target.value || (options[0]?.value ?? ""), order)}
        className="w-[160px]"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </Select>
      <Button
        variant="outline"
        size="icon"
        className="h-9 w-9 shrink-0"
        onClick={toggleOrder}
        title={order === "asc" ? "Ascending" : "Descending"}
      >
        {order === "asc" ? (
          <ArrowUp className="h-4 w-4" />
        ) : (
          <ArrowDown className="h-4 w-4" />
        )}
      </Button>
    </div>
  );
}
