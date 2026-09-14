import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { queryKeys } from "@/api/queryKeys";
import { normalizeError } from "@/api/client";
import type { NormalizedError } from "@/api/errors";

export type UserSettings = {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string | null;
  default_currency: string;
  timezone: string;
  created_at: string;
  updated_at: string;
};

export type UpdateSettingsInput = {
  display_name?: string;
  default_currency?: string;
  timezone?: string;
};

export type NotificationPreferences = {
  budget_alerts: boolean;
  monthly_summary: boolean;
};

export function useUserSettings() {
  return useQuery({
    queryKey: queryKeys.user.detail(),
    queryFn: async () => {
      const response = await apiClient.get("/user");
      return response.data as UserSettings;
    },
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateSettingsInput) => {
      const response = await apiClient.patch("/user", input);
      return response.data as UserSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.user.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useNotificationPreferences() {
  return useQuery({
    queryKey: queryKeys.user.notifications(),
    queryFn: async () => {
      const response = await apiClient.get("/user/notifications");
      return response.data as NotificationPreferences;
    },
  });
}

export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: NotificationPreferences) => {
      const response = await apiClient.put("/user/notifications", input);
      return response.data as NotificationPreferences;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.user.all });
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export function useExportAccount() {
  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post("/user/export");
      return response.data as { download_url: string };
    },
    onError: (error: unknown): NormalizedError => {
      return normalizeError(error);
    },
  });
}

export const TIMEZONES = [
  "UTC",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Asia/Tokyo",
  "Asia/Shanghai",
  "Australia/Sydney",
];

export const CURRENCIES = [
  { code: "USD", name: "US Dollar" },
  { code: "EUR", name: "Euro" },
  { code: "GBP", name: "British Pound" },
  { code: "JPY", name: "Japanese Yen" },
  { code: "CAD", name: "Canadian Dollar" },
  { code: "AUD", name: "Australian Dollar" },
];
