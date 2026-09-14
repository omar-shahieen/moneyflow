import { useState } from "react";
import { Plus, Pencil, Trash2, ArrowLeftRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Pagination } from "@/components/ui/pagination";
import { DateRangeFilter } from "@/components/ui/date-range-filter";
import { SelectFilter } from "@/components/ui/select-filter";
import { SortSelect } from "@/components/ui/sort-select";
import { SearchInput } from "@/components/ui/search-input";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import { ConfirmDialog } from "@/components/feedback/confirm-dialog";
import { useTransactions, useDeleteTransaction } from "./api";
import { useCategories } from "@/features/categories/api";
import { TransactionFormDialog } from "./transaction-form-dialog";
import { formatMoney } from "@/lib/money";
import { useUrlFilters } from "@/hooks/use-url-filters";
import { toast } from "sonner";
import type { TransactionFilters } from "./api";

const SORT_OPTIONS = [
  { value: "occurred_at", label: "Date" },
  { value: "amount_minor", label: "Amount" },
  { value: "created_at", label: "Created" },
];

const DEFAULT_FILTERS: TransactionFilters = {
  page: 1,
  page_size: 20,
};

export function TransactionsPage() {
  const { filters, setFilter, setPage, setPageSize, resetFilters } =
    useUrlFilters<TransactionFilters>({
      defaults: DEFAULT_FILTERS,
    });

  const { data, isLoading, error, refetch } = useTransactions(filters);
  const { data: categoriesData } = useCategories({ page_size: 100 });
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
  const totalPages = data?.totalPages ?? 1;
  const page = data?.page ?? filters.page ?? 1;
  const limit = data?.limit ?? filters.page_size ?? 20;

  const categories = categoriesData?.data ?? [];
  const categoryOptions = categories.map((c) => ({ value: c.id, label: c.name }));

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

      <div className="flex flex-wrap items-end gap-4">
        <SearchInput
          value={filters.search}
          onChange={(v) => setFilter("search" as any, v)}
          placeholder="Search transactions..."
          className="w-64"
        />
        <DateRangeFilter
          from={filters.from}
          to={filters.to}
          onChange={(range) =>
            setFilter("from" as any, range.from) ||
            setFilter("to" as any, range.to)
          }
        />
        <SelectFilter
          label="Type"
          options={[
            { value: "income", label: "Income" },
            { value: "expense", label: "Expense" },
          ]}
          value={filters.type}
          onChange={(v) => setFilter("type" as any, v)}
        />
        <SelectFilter
          label="Category"
          options={categoryOptions}
          value={filters.category_id}
          onChange={(v) => setFilter("category_id" as any, v)}
        />
        <div className="flex items-end gap-2">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Min Amount</label>
            <Input
              type="number"
              placeholder="0"
              value={filters.min_amount ?? ""}
              onChange={(e) => setFilter("min_amount" as any, e.target.value || undefined)}
              className="w-28"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Max Amount</label>
            <Input
              type="number"
              placeholder="∞"
              value={filters.max_amount ?? ""}
              onChange={(e) => setFilter("max_amount" as any, e.target.value || undefined)}
              className="w-28"
            />
          </div>
        </div>
        <SortSelect
          options={SORT_OPTIONS}
          value={filters.sort}
          order={filters.order}
          onChange={(sort, order) => {
            setFilter("sort" as any, sort);
            setFilter("order" as any, order);
          }}
        />
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

          <Pagination
            page={page}
            totalPages={totalPages}
            total={total}
            limit={limit}
            onPageChange={setPage}
            onLimitChange={(size) => setPageSize(size)}
          />
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
