import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";
import type { PaginatedResponse } from "@/api/helpers";

export type Category = {
  id: string;
  name: string;
  type: "income" | "expense";
  created_at: string;
};

export type CategoryFilters = {
  page?: number;
  page_size?: number;
  search?: string;
  type?: "income" | "expense";
  sort?: string;
  order?: "asc" | "desc";
};

export type CreateCategoryInput = {
  name: string;
  type: "income" | "expense";
};

export type UpdateCategoryInput = {
  name?: string;
  type?: "income" | "expense";
};

export function useCategories(filters?: CategoryFilters) {
  return useQuery({
    queryKey: queryKeys.categories.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.page) params.set("page", String(filters.page));
      if (filters?.page_size) params.set("page_size", String(filters.page_size));
      if (filters?.search) params.set("search", filters.search);
      if (filters?.type) params.set("type", filters.type);
      if (filters?.sort) params.set("sort", filters.sort);
      if (filters?.order) params.set("order", filters.order);

      const response = await apiClient.get(`/categories?${params.toString()}`);
      return response.data as PaginatedResponse<Category>;
    },
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateCategoryInput) => {
      const response = await apiClient.post("/categories", input);
      return response.data as Category;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useUpdateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, ...input }: UpdateCategoryInput & { id: string }) => {
      const response = await apiClient.patch(`/categories/${id}`, input);
      return response.data as Category;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/categories/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}
