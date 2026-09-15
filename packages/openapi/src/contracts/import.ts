import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { ZImportResponse, ZErrorResponse, ZListImportsQuery, schemaWithPagination } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const importContract = c.router({
  listImports: {
    summary: "List imports",
    description: "List CSV imports with pagination. Filterable by status.",
    path: "/api/v1/imports",
    method: "GET",
    query: ZListImportsQuery,
    responses: {
      200: schemaWithPagination(ZImportResponse),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createImport: {
    summary: "Create a CSV import",
    description: "Upload a CSV file to import transactions asynchronously. CSV format: header + rows with columns category_name, amount, currency, note, date.",
    path: "/api/v1/imports",
    method: "POST",
    body: z.object({
      file: z.any().describe("CSV file to upload"),
    }),
    responses: {
      202: ZImportResponse,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getImportById: {
    summary: "Get import status",
    description: "Get import status and results by ID.",
    path: "/api/v1/imports/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZImportResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
