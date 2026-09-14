import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";
import type { PaginatedResponse } from "@/api/helpers";

export type ReportFormat = "pdf" | "csv";
export type ReportStatus = "pending" | "processing" | "ready" | "failed";

export type Report = {
  id: string;
  format: ReportFormat;
  period_start: string;
  period_end: string;
  status: ReportStatus;
  storage_key: string | null;
  download_url: string | null;
  created_at: string;
  completed_at: string | null;
};

export type ReportFilters = {
  page?: number;
  page_size?: number;
};

export type CreateReportInput = {
  format: ReportFormat;
  period_start: string;
  period_end: string;
};

export function useReports(filters?: ReportFilters) {
  return useQuery({
    queryKey: queryKeys.reports.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.page) params.set("page", String(filters.page));
      if (filters?.page_size) params.set("page_size", String(filters.page_size));

      const response = await apiClient.get(`/reports?${params.toString()}`);
      return response.data as PaginatedResponse<Report>;
    },
  });
}

export function useReport(id: string) {
  return useQuery({
    queryKey: queryKeys.reports.detail(id),
    queryFn: async () => {
      const response = await apiClient.get(`/reports/${id}`);
      return response.data as Report;
    },
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === "pending" || status === "processing") {
        return 2000;
      }
      return false;
    },
  });
}

export function useCreateReport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateReportInput) => {
      const response = await apiClient.post("/reports", input);
      return response.data as Report;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reports.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}
