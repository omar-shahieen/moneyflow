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
import { Textarea } from "@/components/ui/textarea";
import { useCategories } from "@/features/categories/api";
import {
  useTransactions,
  useCreateTransaction,
  useUpdateTransaction,
} from "./api";
import { parseMoneyInput } from "@/lib/money";
import { toast } from "sonner";

const transactionSchema = z.object({
  category_id: z.string().min(1, "Category is required"),
  amount: z.string().min(1, "Amount is required"),
  currency: z.string().min(3, "Currency is required").default("USD"),
  note: z.string().min(1, "Note is required").max(500),
  occurred_at: z.string().min(1, "Date is required"),
});

type TransactionFormData = z.infer<typeof transactionSchema>;

interface TransactionFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  transactionId?: string;
}

export function TransactionFormDialog({
  open,
  onOpenChange,
  transactionId,
}: TransactionFormDialogProps) {
  const { data: categoriesData } = useCategories({ page_size: 100 });
  const { data: transactions } = useTransactions();
  const createTransaction = useCreateTransaction();
  const updateTransaction = useUpdateTransaction();
  const isEditing = !!transactionId;
  const transaction = transactions?.data?.find((t) => t.id === transactionId);
  const categories = categoriesData?.data ?? [];

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TransactionFormData>({
    resolver: zodResolver(transactionSchema),
    defaultValues: {
      category_id: "",
      amount: "",
      currency: "USD",
      note: "",
      occurred_at: new Date().toISOString().split("T")[0],
    },
  });

  useEffect(() => {
    if (transaction) {
      reset({
        category_id: transaction.category_id,
        amount: String(transaction.amount_minor / 100),
        currency: transaction.currency,
        note: transaction.note,
        occurred_at: transaction.occurred_at.split("T")[0],
      });
    } else {
      reset({
        category_id: "",
        amount: "",
        currency: "USD",
        note: "",
        occurred_at: new Date().toISOString().split("T")[0],
      });
    }
  }, [transaction, reset]);

  const onSubmit = async (data: TransactionFormData) => {
    const amountMinor = parseMoneyInput(data.amount);
    if (amountMinor === null || amountMinor === 0) {
      toast.error("Please enter a valid amount");
      return;
    }

    try {
      if (isEditing && transactionId) {
        await updateTransaction.mutateAsync({
          id: transactionId,
          category_id: data.category_id,
          amount_minor: amountMinor,
          currency: data.currency,
          note: data.note,
          occurred_at: data.occurred_at,
        });
        toast.success("Transaction updated");
      } else {
        await createTransaction.mutateAsync({
          category_id: data.category_id,
          amount_minor: amountMinor,
          currency: data.currency,
          note: data.note,
          occurred_at: data.occurred_at,
        });
        toast.success("Transaction created");
      }
      onOpenChange(false);
    } catch {
      toast.error("Something went wrong");
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Transaction" : "New Transaction"}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="category_id">Category</Label>
            <select
              id="category_id"
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              {...register("category_id")}
            >
              <option value="">Select a category</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>
                  {cat.name} ({cat.type})
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
              <Label htmlFor="amount">Amount</Label>
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

          <div className="space-y-2">
            <Label htmlFor="occurred_at">Date</Label>
            <Input id="occurred_at" type="date" {...register("occurred_at")} />
            {errors.occurred_at && (
              <p className="text-sm text-destructive">
                {errors.occurred_at.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="note">Note</Label>
            <Textarea
              id="note"
              placeholder="What was this transaction for?"
              {...register("note")}
            />
            {errors.note && (
              <p className="text-sm text-destructive">{errors.note.message}</p>
            )}
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
              disabled={createTransaction.isPending || updateTransaction.isPending}
            >
              {isEditing ? "Save Changes" : "Create"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
