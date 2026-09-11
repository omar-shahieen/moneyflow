import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";

export type PlanType = "free" | "pro" | "vip";

export type SubscriptionStatus = "active" | "past_due" | "canceled";

export type Subscription = {
  id: string;
  plan: PlanType;
  status: SubscriptionStatus;
  payment_provider: string | null;
  current_period_end: string | null;
  created_at: string;
  updated_at: string;
};

export type PlanLimits = {
  categories: number | null;
  transactions_per_month: number | null;
  budgets: number | null;
  csv_import_rows: number;
  reports_format: "csv_only" | "pdf_and_csv";
  report_history: number | null;
  multi_currency: boolean;
  receipt_attachments: boolean;
  recurring_rules: boolean;
  shared_budgets: boolean;
  full_account_export: boolean;
};

export const PLAN_LIMITS: Record<PlanType, PlanLimits> = {
  free: {
    categories: 5,
    transactions_per_month: 50,
    budgets: 1,
    csv_import_rows: 0,
    reports_format: "csv_only",
    report_history: 3,
    multi_currency: false,
    receipt_attachments: false,
    recurring_rules: false,
    shared_budgets: false,
    full_account_export: false,
  },
  pro: {
    categories: 50,
    transactions_per_month: 500,
    budgets: 10,
    csv_import_rows: 100,
    reports_format: "pdf_and_csv",
    report_history: 30,
    multi_currency: true,
    receipt_attachments: true,
    recurring_rules: true,
    shared_budgets: true,
    full_account_export: false,
  },
  vip: {
    categories: null,
    transactions_per_month: null,
    budgets: null,
    csv_import_rows: 10000,
    reports_format: "pdf_and_csv",
    report_history: null,
    multi_currency: true,
    receipt_attachments: true,
    recurring_rules: true,
    shared_budgets: true,
    full_account_export: true,
  },
};

export const PLAN_PRICES: Record<PlanType, { monthly: number; yearly: number }> = {
  free: { monthly: 0, yearly: 0 },
  pro: { monthly: 9, yearly: 90 },
  vip: { monthly: 29, yearly: 290 },
};

export function useSubscription() {
  return useQuery({
    queryKey: queryKeys.subscription.current(),
    queryFn: async () => {
      const response = await apiClient.get("/subscription");
      return response.data as Subscription;
    },
    retry: false,
  });
}

export function useCreateCheckout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { plan: PlanType; billing_period: "monthly" | "yearly" }) => {
      const response = await apiClient.post("/subscription/checkout", input);
      return response.data as { checkout_url: string };
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.subscription.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useCreatePortal() {
  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post("/subscription/portal");
      return response.data as { portal_url: string };
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useCurrentPlanLimits() {
  const { data: subscription } = useSubscription();
  const plan = subscription?.plan ?? "free";
  return PLAN_LIMITS[plan];
}
