import { useState } from "react";
import { Plus, Pencil, Trash2, PiggyBank } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import { ConfirmDialog } from "@/components/feedback/confirm-dialog";
import { useBudgets, useDeleteBudget } from "./api";
import { BudgetFormDialog } from "./budget-form-dialog";
import { formatMoney } from "@/lib/money";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

export function BudgetsPage() {
  const { data: budgets, isLoading, error, refetch } = useBudgets();
  const deleteBudget = useDeleteBudget();
  const [createOpen, setCreateOpen] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  if (isLoading) {
    return <LoadingPage message="Loading budgets..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const handleDelete = async () => {
    if (!deleteId) return;
    const result = await deleteBudget.mutateAsync(deleteId);
    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Budget deleted");
      setDeleteId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Budgets</h1>
          <p className="text-muted-foreground">
            Set monthly spending limits for your categories.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Budget
        </Button>
      </div>

      {!budgets || budgets.length === 0 ? (
        <EmptyState
          icon={<PiggyBank className="h-8 w-8 text-muted-foreground" />}
          title="No budgets"
          description="Create a budget to start tracking your spending limits."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Budget
            </Button>
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {budgets.map((budget) => {
            const usage = budget.current_usage_minor ?? 0;
            const percentage = budget.percentage_used ?? 0;
            const isExceeded = budget.is_exceeded ?? percentage > 100;

            return (
              <Card key={budget.id}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">
                      {budget.category_name ?? "Unknown"}
                    </CardTitle>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setEditId(budget.id)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteId(budget.id)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                  <CardDescription>
                    {formatMoney(budget.monthly_limit_minor, budget.currency)} / month
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Spent</span>
                      <span className={cn("font-medium", isExceeded && "text-destructive")}>
                        {formatMoney(usage, budget.currency)}
                      </span>
                    </div>
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full transition-all",
                          isExceeded ? "bg-destructive" : "bg-primary"
                        )}
                        style={{ width: `${Math.min(percentage, 100)}%` }}
                      />
                    </div>
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>{Math.round(percentage)}% used</span>
                      {isExceeded && (
                        <Badge variant="destructive">Exceeded</Badge>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <BudgetFormDialog open={createOpen} onOpenChange={setCreateOpen} />

      {editId && (
        <BudgetFormDialog
          open={!!editId}
          onOpenChange={(open) => {
            if (!open) setEditId(null);
          }}
          budgetId={editId}
        />
      )}

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title="Delete Budget"
        description="Are you sure you want to delete this budget?"
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}
