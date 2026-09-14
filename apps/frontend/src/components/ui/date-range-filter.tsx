import { X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

type DateRangeFilterProps = {
  from?: string;
  to?: string;
  onChange: (range: { from?: string; to?: string }) => void;
  className?: string;
};

export function DateRangeFilter({
  from,
  to,
  onChange,
  className,
}: DateRangeFilterProps) {
  const hasValue = from || to;

  return (
    <div className={cn("flex flex-wrap items-end gap-4", className)}>
      <div className="flex-1 min-w-[160px]">
        <Label htmlFor="date-from">From</Label>
        <Input
          id="date-from"
          type="date"
          value={from ?? ""}
          onChange={(e) =>
            onChange({ from: e.target.value || undefined, to })
          }
        />
      </div>
      <div className="flex-1 min-w-[160px]">
        <Label htmlFor="date-to">To</Label>
        <Input
          id="date-to"
          type="date"
          value={to ?? ""}
          onChange={(e) =>
            onChange({ from, to: e.target.value || undefined })
          }
        />
      </div>
      {hasValue && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange({ from: undefined, to: undefined })}
          className="h-9"
        >
          <X className="mr-1 h-3 w-3" />
          Clear
        </Button>
      )}
    </div>
  );
}
