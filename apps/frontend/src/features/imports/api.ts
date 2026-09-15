import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";

export type ImportStatus = "pending" | "processing" | "completed" | "failed";

export type ImportError = {
  row: number;
  reason: string;
};

export type Import = {
  id: string;
  status: ImportStatus;
  total_rows: number;
  success_rows: number;
  failed_rows: ImportError[];
  created_at: string;
  completed_at: string | null;
};

const MAX_CSV_SIZE_MB = 10;
const ALLOWED_CSV_TYPES = ["text/csv", "application/vnd.ms-excel"];

export function validateCsvFile(file: File): string | null {
  if (!ALLOWED_CSV_TYPES.includes(file.type) && !file.name.endsWith(".csv")) {
    return "File must be a CSV file";
  }
  if (file.size > MAX_CSV_SIZE_MB * 1024 * 1024) {
    return `File size must be less than ${MAX_CSV_SIZE_MB}MB`;
  }
  return null;
}

export function useImports() {
  return useQuery({
    queryKey: queryKeys.imports.lists(),
    queryFn: async () => {
      const response = await apiClient.get("/imports");
      return (response.data as { data: Import[] }).data;
    },
  });
}

export function useImport(id: string) {
  return useQuery({
    queryKey: queryKeys.imports.detail(id),
    queryFn: async () => {
      const response = await apiClient.get(`/imports/${id}`);
      return response.data as Import;
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

export function useCreateImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (file: File) => {
      const validationError = validateCsvFile(file);
      if (validationError) {
        throw new Error(validationError);
      }

      const formData = new FormData();
      formData.append("file", file);

      const response = await apiClient.post("/imports", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      return response.data as Import;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.imports.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}
