import axios, { type AxiosError, isAxiosError } from "axios";
import { API_URL } from "@/config/env";
import type { ApiError, NormalizedError } from "./errors";

export const apiClient = axios.create({
  baseURL: `${API_URL}/api`,
  headers: {
    "Content-Type": "application/json",
  },
});

let getTokenFn: (() => Promise<string | null>) | null = null;

export function setTokenGetter(fn: () => Promise<string | null>) {
  getTokenFn = fn;
}

apiClient.interceptors.request.use(async (config) => {
  if (getTokenFn) {
    const token = await getTokenFn();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401 && !error.config?.headers["x-retry"]) {
      error.config!.headers["x-retry"] = "true";
      return apiClient.request(error.config!);
    }
    throw error;
  }
);

export function normalizeError(
  error: unknown,
  requestId?: string
): NormalizedError {
  if (isAxiosError(error)) {
    const response = error.response;
    const status = response?.status ?? 500;
    const data = response?.data as ApiError | undefined;

    return {
      status,
      code: data?.code ?? "internal_error",
      message: data?.message ?? error.message ?? "An unexpected error occurred",
      field: data?.field,
      requestId: requestId ?? (response?.headers?.["x-request-id"] as string),
    };
  }

  return {
    status: 500,
    code: "internal_error",
    message: error instanceof Error ? error.message : "An unexpected error occurred",
    requestId,
  };
}
