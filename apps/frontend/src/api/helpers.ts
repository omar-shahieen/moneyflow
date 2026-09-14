import { type QueryClient } from "@tanstack/react-query";

export type PaginatedResponse<T> = {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
};

export function getPaginationParams(filters?: {
  page?: number;
  limit?: number;
}) {
  return {
    page: filters?.page ?? 1,
    limit: filters?.limit ?? 20,
  };
}

export function getTotalPages(total: number, limit: number): number {
  return Math.ceil(total / limit);
}

export function invalidateQueries(
  queryClient: QueryClient,
  queryKey: readonly unknown[]
) {
  return queryClient.invalidateQueries({ queryKey });
}

export function setQueryData<T>(
  queryClient: QueryClient,
  queryKey: readonly unknown[],
  data: T
) {
  return queryClient.setQueryData(queryKey, data);
}

export function getQueryData<T>(
  queryClient: QueryClient,
  queryKey: readonly unknown[]
): T | undefined {
  return queryClient.getQueryData<T>(queryKey);
}
