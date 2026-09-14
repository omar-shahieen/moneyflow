import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCategories } from "@/features/categories/api";
import { useBudgets, useCreateBudget, useUpdateBudget } from "./api";
import { parseMoneyInput } from "@/lib/money";
import { toast } from "sonner";

const budgetSchema = z.object({
  category_id: z.string().min(1, "Category is required"),
  amount: z.string().min(1, "Amount is required"),
  currency: z.string().min(3, "Currency is required").default("USD"),
});

type BudgetFormData = z.infer<typeof budgetSchema>;

interface BudgetFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  budgetId?: string;
}

export function BudgetFormDialog({
  open,
  onOpenChange,
  budgetId,
}: BudgetFormDialogProps) {
  const { data: categoriesData } = useCategories({ page_size: 100 });
  const { data: budgetsData } = useBudgets({ page_size: 100 });
  const createBudget = useCreateBudget();
  const updateBudget = useUpdateBudget();
  const isEditing = !!budgetId;
  const categories = categoriesData?.data ?? [];
  const budgets = budgetsData?.data ?? [];
  const budget = budgets.find((b) => b.id === budgetId);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<BudgetFormData>({
    resolver: zodResolver(budgetSchema),
    defaultValues: {
      category_id: "",
      amount: "",
      currency: "USD",
    },
  });

  useEffect(() => {
    if (budget) {
      reset({
        category_id: budget.category_id,
        amount: String(budget.monthly_limit_minor / 100),
        currency: budget.currency,
      });
    } else {
      reset({
        category_id: "",
        amount: "",
        currency: "USD",
      });
    }
  }, [budget, reset]);

  const onSubmit = async (data: BudgetFormData) => {
    const amountMinor = parseMoneyInput(data.amount);
    if (amountMinor === null || amountMinor === 0) {
      toast.error("Please enter a valid amount");
      return;
    }

    try {
      if (isEditing && budgetId) {
        await updateBudget.mutateAsync({
          id: budgetId,
          monthly_limit_minor: amountMinor,
          currency: data.currency,
        });
        toast.success("Budget updated");
      } else {
        await createBudget.mutateAsync({
          category_id: data.category_id,
          monthly_limit_minor: amountMinor,
          currency: data.currency,
        });
        toast.success("Budget created");
      }
      onOpenChange(false);
    } catch {
      toast.error("Something went wrong");
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Budget" : "Create Budget"}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="category_id">Category</Label>
            <select
              id="category_id"
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              {...register("category_id")}
              disabled={isEditing}
            >
              <option value="">Select a category</option>
              {categories
                .filter((c) => c.type === "expense")
                .map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                ))}
            </select>
            {errors.category_id && (
              <p className="text-sm text-destructive">
                {errors.category_id.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="amount">Monthly Limit</Label>
              <Input
                id="amount"
                type="number"
                step="0.01"
                placeholder="0.00"
                {...register("amount")}
              />
              {errors.amount && (
                <p className="text-sm text-destructive">
                  {errors.amount.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="currency">Currency</Label>
              <Input
                id="currency"
                placeholder="USD"
                {...register("currency")}
              />
              {errors.currency && (
                <p className="text-sm text-destructive">
                  {errors.currency.message}
                </p>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createBudget.isPending || updateBudget.isPending}
            >
              {isEditing ? "Save Changes" : "Create"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
