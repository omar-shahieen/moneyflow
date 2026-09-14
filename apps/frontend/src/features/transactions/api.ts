import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";
import type { PaginatedResponse } from "@/api/helpers";

export type Transaction = {
  id: string;
  category_id: string;
  amount_minor: number;
  currency: string;
  note: string;
  receipt_key: string | null;
  occurred_at: string;
  created_at: string;
  category_name?: string;
  category_type?: "income" | "expense";
};

export type TransactionFilters = {
  page?: number;
  page_size?: number;
  from?: string;
  to?: string;
  category_id?: string;
  type?: "income" | "expense";
  sort?: string;
  order?: "asc" | "desc";
};

export type CreateTransactionInput = {
  category_id: string;
  amount_minor: number;
  currency: string;
  note: string;
  occurred_at: string;
};

export type UpdateTransactionInput = {
  category_id?: string;
  amount_minor?: number;
  currency?: string;
  note?: string;
  occurred_at?: string;
};

export type TransactionSummary = {
  total_income: number;
  total_expenses: number;
  balance: number;
  transaction_count: number;
  currency: string;
};

export function useTransactions(filters?: TransactionFilters) {
  return useQuery({
    queryKey: queryKeys.transactions.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.page) params.set("page", String(filters.page));
      if (filters?.page_size) params.set("page_size", String(filters.page_size));
      if (filters?.from) params.set("start_date", filters.from);
      if (filters?.to) params.set("end_date", filters.to);
      if (filters?.category_id) params.set("category_id", filters.category_id);
      if (filters?.type) params.set("type", filters.type);
      if (filters?.sort) params.set("sort", filters.sort);
      if (filters?.order) params.set("order", filters.order);

      const response = await apiClient.get(`/transactions?${params.toString()}`);
      return response.data as PaginatedResponse<Transaction>;
    },
  });
}

export function useTransactionSummary(month?: string) {
  return useQuery({
    queryKey: queryKeys.transactions.summary(month),
    queryFn: async () => {
      const params = month ? `?month=${month}` : "";
      const response = await apiClient.get(`/transactions/summary${params}`);
      return response.data as TransactionSummary;
    },
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateTransactionInput) => {
      const response = await apiClient.post("/transactions", input);
      return response.data as Transaction;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: UpdateTransactionInput & { id: string }) => {
      const response = await apiClient.patch(`/transactions/${id}`, input);
      return response.data as Transaction;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/transactions/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}
