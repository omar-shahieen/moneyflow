import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";

export type Budget = {
  id: string;
  category_id: string;
  monthly_limit_minor: number;
  currency: string;
  created_at: string;
  category_name?: string;
  current_usage_minor?: number;
  percentage_used?: number;
  is_exceeded?: boolean;
};

export type CreateBudgetInput = {
  category_id: string;
  monthly_limit_minor: number;
  currency: string;
};

export type UpdateBudgetInput = {
  monthly_limit_minor?: number;
  currency?: string;
};

export function useBudgets() {
  return useQuery({
    queryKey: queryKeys.budgets.lists(),
    queryFn: async () => {
      const response = await apiClient.get("/budgets");
      return response.data as Budget[];
    },
  });
}

export function useCreateBudget() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateBudgetInput) => {
      const response = await apiClient.post("/budgets", input);
      return response.data as Budget;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useUpdateBudget() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: UpdateBudgetInput & { id: string }) => {
      const response = await apiClient.patch(`/budgets/${id}`, input);
      return response.data as Budget;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useDeleteBudget() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/budgets/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.budgets.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}
