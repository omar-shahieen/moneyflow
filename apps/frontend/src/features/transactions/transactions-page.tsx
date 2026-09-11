import { useState } from "react";
import { Plus, Pencil, Trash2, ArrowLeftRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import { ConfirmDialog } from "@/components/feedback/confirm-dialog";
import { useTransactions, useDeleteTransaction } from "./api";
import { TransactionFormDialog } from "./transaction-form-dialog";
import { formatMoney } from "@/lib/money";
import { toast } from "sonner";
import type { TransactionFilters } from "./api";

export function TransactionsPage() {
  const [filters, setFilters] = useState<TransactionFilters>({
    page: 1,
    limit: 20,
  });
  const { data, isLoading, error, refetch } = useTransactions(filters);
  const deleteTransaction = useDeleteTransaction();
  const [createOpen, setCreateOpen] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  if (isLoading) {
    return <LoadingPage message="Loading transactions..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const transactions = data?.data ?? [];
  const total = data?.total ?? 0;

  const handleDelete = async () => {
    if (!deleteId) return;
    const result = await deleteTransaction.mutateAsync(deleteId);
    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Transaction deleted");
      setDeleteId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Transactions</h1>
          <p className="text-muted-foreground">
            Track your income and expenses.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Transaction
        </Button>
      </div>

      <div className="flex flex-wrap gap-4">
        <div className="flex-1 min-w-[200px]">
          <Label htmlFor="from">From</Label>
          <Input
            id="from"
            type="date"
            value={filters.from ?? ""}
            onChange={(e) =>
              setFilters((prev) => ({ ...prev, from: e.target.value || undefined }))
            }
          />
        </div>
        <div className="flex-1 min-w-[200px]">
          <Label htmlFor="to">To</Label>
          <Input
            id="to"
            type="date"
            value={filters.to ?? ""}
            onChange={(e) =>
              setFilters((prev) => ({ ...prev, to: e.target.value || undefined }))
            }
          />
        </div>
        <div className="flex-1 min-w-[200px]">
          <Label htmlFor="type">Type</Label>
          <select
            id="type"
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
            value={filters.type ?? ""}
            onChange={(e) =>
              setFilters((prev) => ({
                ...prev,
                type: (e.target.value as "income" | "expense") || undefined,
              }))
            }
          >
            <option value="">All</option>
            <option value="income">Income</option>
            <option value="expense">Expense</option>
          </select>
        </div>
      </div>

      {transactions.length === 0 ? (
        <EmptyState
          icon={<ArrowLeftRight className="h-8 w-8 text-muted-foreground" />}
          title="No transactions"
          description="Record your first transaction to start tracking."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Transaction
            </Button>
          }
        />
      ) : (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Note</TableHead>
                  <TableHead>Category</TableHead>
                  <TableHead className="text-right">Amount</TableHead>
                  <TableHead className="w-[100px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell className="text-muted-foreground">
                      {new Date(tx.occurred_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="font-medium">{tx.note}</TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {tx.category_name ?? "Unknown"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <span
                        className={
                          tx.category_type === "income"
                            ? "text-green-600"
                            : "text-red-600"
                        }
                      >
                        {tx.category_type === "income" ? "+" : "-"}
                        {formatMoney(tx.amount_minor, tx.currency)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setEditId(tx.id)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setDeleteId(tx.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              Showing {transactions.length} of {total} transactions
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={filters.page === 1}
                onClick={() =>
                  setFilters((prev) => ({ ...prev, page: (prev.page ?? 1) - 1 }))
                }
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={transactions.length < (filters.limit ?? 20)}
                onClick={() =>
                  setFilters((prev) => ({ ...prev, page: (prev.page ?? 1) + 1 }))
                }
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}

      <TransactionFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
      />

      {editId && (
        <TransactionFormDialog
          open={!!editId}
          onOpenChange={(open) => {
            if (!open) setEditId(null);
          }}
          transactionId={editId}
        />
      )}

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title="Delete Transaction"
        description="Are you sure you want to delete this transaction? This action cannot be undone."
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}
