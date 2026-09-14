import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";

type SelectOption = {
  value: string;
  label: string;
};

type SelectFilterProps = {
  label?: string;
  options: SelectOption[];
  value?: string;
  onChange: (value: string | undefined) => void;
  placeholder?: string;
  className?: string;
};

export function SelectFilter({
  label,
  options,
  value,
  onChange,
  placeholder = "All",
  className,
}: SelectFilterProps) {
  return (
    <div className={cn("flex-1 min-w-[160px]", className)}>
      {label && <Label>{label}</Label>}
      <Select
        value={value ?? ""}
        onChange={(e) => onChange(e.target.value || undefined)}
      >
        <option value="">{placeholder}</option>
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </Select>
    </div>
  );
}
