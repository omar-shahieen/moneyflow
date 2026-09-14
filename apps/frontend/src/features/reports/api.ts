import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";

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

export type CreateReportInput = {
  format: ReportFormat;
  period_start: string;
  period_end: string;
};

export function useReports() {
  return useQuery({
    queryKey: queryKeys.reports.lists(),
    queryFn: async () => {
      const response = await apiClient.get("/reports");
      return response.data as Report[];
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
