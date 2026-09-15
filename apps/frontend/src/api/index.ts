import { apiContract } from "@moneyflow/openapi/contracts";
import { useAuth } from "@clerk/clerk-react";
import { initClient } from "@ts-rest/core";
import { apiClient, setTokenGetter } from "./client";

type Headers = Awaited<
  ReturnType<NonNullable<Parameters<typeof initClient>[1]["api"]>>
>["headers"];

export type TApiClient = ReturnType<typeof useApiClient>;

export const useApiClient = ({ isBlob = false }: { isBlob?: boolean } = {}) => {
  const { getToken } = useAuth();

  setTokenGetter(() => getToken({ template: "test" }));

  return initClient(apiContract, {
    baseUrl: "",
    baseHeaders: {
      "Content-Type": "application/json",
    },
    api: async ({ path, method, headers, body }) => {
      const makeRequest = async (retryCount = 0): Promise<unknown> => {
        try {
          const result = await apiClient.request({
            method: method as string,
            url: path,
            headers: {
              ...headers,
            },
            data: body,
            ...(isBlob ? { responseType: "blob" } : {}),
          });
          return {
            status: result.status,
            body: result.data,
            headers: result.headers as unknown as Headers,
          };
        } catch (e: unknown) {
          const error = e as {
            response?: { status: number; data: unknown; headers: Record<string, string> };
          };

          if (error.response?.status === 401 && retryCount < 2) {
            return makeRequest(retryCount + 1);
          }

          if (error.response) {
            return {
              status: error.response.status,
              body: error.response.data || { message: "Internal server error" },
              headers: (error.response.headers as unknown as Headers) || {},
            };
          }
          throw e;
        }
      };

      return makeRequest();
    },
  });
};
