import { Crown, ExternalLink, Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import {
  useSubscription,
  useCreateCheckout,
  useCreatePortal,
  PLAN_LIMITS,
  PLAN_PRICES,
} from "./api";
import { toast } from "sonner";
import type { PlanType } from "./api";

const PLAN_FEATURES: { key: string; label: string; getValue: (plan: PlanType) => string | boolean }[] = [
  { key: "categories", label: "Categories", getValue: (p) => PLAN_LIMITS[p].categories === null ? "Unlimited" : `${PLAN_LIMITS[p].categories} max` },
  { key: "transactions", label: "Transactions / month", getValue: (p) => PLAN_LIMITS[p].transactions_per_month === null ? "Unlimited" : `${PLAN_LIMITS[p].transactions_per_month} max` },
  { key: "budgets", label: "Budgets", getValue: (p) => PLAN_LIMITS[p].budgets === null ? "Unlimited" : `${PLAN_LIMITS[p].budgets} max` },
  { key: "recurring", label: "Recurring rules", getValue: (p) => PLAN_LIMITS[p].recurring_rules },
  { key: "shared", label: "Shared / family budgets", getValue: (p) => PLAN_LIMITS[p].shared_budgets },
  { key: "csv", label: "CSV import", getValue: (p) => PLAN_LIMITS[p].csv_import_rows > 0 ? `${PLAN_LIMITS[p].csv_import_rows} rows/upload` : false },
  { key: "receipts", label: "Receipt attachments", getValue: (p) => PLAN_LIMITS[p].receipt_attachments },
  { key: "reports", label: "Reports", getValue: (p) => PLAN_LIMITS[p].reports_format === "pdf_and_csv" ? "PDF + CSV" : "CSV only" },
  { key: "history", label: "Report history", getValue: (p) => PLAN_LIMITS[p].report_history === null ? "Unlimited" : `${PLAN_LIMITS[p].report_history} reports` },
  { key: "currency", label: "Multi-currency", getValue: (p) => PLAN_LIMITS[p].multi_currency },
  { key: "export", label: "Full account export", getValue: (p) => PLAN_LIMITS[p].full_account_export },
];

export function SubscriptionPage() {
  const { data: subscription, isLoading, error, refetch } = useSubscription();
  const createCheckout = useCreateCheckout();
  const createPortal = useCreatePortal();

  if (isLoading) {
    return <LoadingPage message="Loading subscription..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const currentPlan = subscription?.plan ?? "free";
  const currentStatus = subscription?.status ?? "active";

  const handleUpgrade = async (plan: PlanType) => {
    const result = await createCheckout.mutateAsync({ plan, billing_period: "monthly" });
    if (result instanceof Error) {
      toast.error(result.message);
    } else if (result.checkout_url) {
      window.open(result.checkout_url, "_blank");
    }
  };

  const handleManageBilling = async () => {
    const result = await createPortal.mutateAsync();
    if (result instanceof Error) {
      toast.error(result.message);
    } else if (result.portal_url) {
      window.open(result.portal_url, "_blank");
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Subscription</h1>
        <p className="text-muted-foreground">
          Manage your plan and billing.
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Crown className="h-6 w-6 text-primary" />
            <div>
              <CardTitle className="capitalize">{currentPlan} Plan</CardTitle>
              <CardDescription>
                Status:{" "}
                <Badge variant={currentStatus === "active" ? "default" : "destructive"}>
                  {currentStatus}
                </Badge>
                {subscription?.current_period_end && currentStatus === "active" && (
                  <span className="ml-2">
                    Renews {new Date(subscription.current_period_end).toLocaleDateString()}
                  </span>
                )}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {currentPlan !== "free" && (
            <Button variant="outline" onClick={handleManageBilling} disabled={createPortal.isPending}>
              <ExternalLink className="mr-2 h-4 w-4" />
              Manage Billing
            </Button>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-3">
        {(["free", "pro", "vip"] as PlanType[]).map((plan) => {
          const isCurrent = plan === currentPlan;
          const price = PLAN_PRICES[plan];

          return (
            <Card key={plan} className={isCurrent ? "border-primary" : ""}>
              <CardHeader>
                <CardTitle className="capitalize">{plan}</CardTitle>
                <CardDescription>
                  {price.monthly === 0 ? "Free" : `$${price.monthly}/month`}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {isCurrent ? (
                  <Badge className="w-full justify-center">Current Plan</Badge>
                ) : (
                  <Button
                    className="w-full"
                    onClick={() => handleUpgrade(plan)}
                    disabled={createCheckout.isPending || (plan === "free")}
                  >
                    {plan === "free" ? "Downgrade" : "Upgrade"}
                  </Button>
                )}

                <ul className="space-y-2 text-sm">
                  {PLAN_FEATURES.map((feature) => {
                    const value = feature.getValue(plan);
                    return (
                      <li key={feature.key} className="flex items-center gap-2">
                        {value === false || value === 0 ? (
                          <X className="h-4 w-4 text-muted-foreground shrink-0" />
                        ) : (
                          <Check className="h-4 w-4 text-primary shrink-0" />
                        )}
                        <span className={value === false || value === 0 ? "text-muted-foreground" : ""}>
                          {feature.label}:{" "}
                          {typeof value === "boolean" ? (value ? "Yes" : "No") : value}
                        </span>
                      </li>
                    );
                  })}
                </ul>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
