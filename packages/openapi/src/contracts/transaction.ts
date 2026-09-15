import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZTransaction,
  ZCreateTransactionRequest,
  ZUpdateTransactionRequest,
  ZTransactionSummary,
  ZPresignedUploadResponse,
  schemaWithPagination,
} from "@moneyflow/zod";
import { ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const transactionContract = c.router({
  listTransactions: {
    summary: "List transactions",
    description: "List transactions with filtering and pagination.",
    path: "/api/v1/transactions",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      category_id: z.string().uuid().optional(),
      min_amount: z.number().optional(),
      max_amount: z.number().optional(),
      from: z.string().datetime().optional(),
      to: z.string().datetime().optional(),
      type: z.enum(["income", "expense"]).optional(),
    }),
    responses: {
      200: schemaWithPagination(ZTransaction),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getTransactionById: {
    summary: "Get transaction by ID",
    path: "/api/v1/transactions/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZTransaction,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getTransactionSummary: {
    summary: "Get transaction summary",
    description: "Get a summary of transactions for a given month (total income, total expense, by category, by currency).",
    path: "/api/v1/transactions/summary",
    method: "GET",
    query: z.object({
      month: z.string().regex(/^\d{4}-\d{2}$/),
    }),
    responses: {
      200: ZTransactionSummary,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createTransaction: {
    summary: "Create a transaction",
    path: "/api/v1/transactions",
    method: "POST",
    body: ZCreateTransactionRequest,
    responses: {
      201: ZTransaction,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  updateTransaction: {
    summary: "Update a transaction",
    path: "/api/v1/transactions/:id",
    method: "PATCH",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: ZUpdateTransactionRequest,
    responses: {
      200: ZTransaction,
      400: ZErrorResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  deleteTransaction: {
    summary: "Delete a transaction",
    path: "/api/v1/transactions/:id",
    method: "DELETE",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      204: z.void(),
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getReceiptUploadUrl: {
    summary: "Get receipt upload URL",
    description: "Get a presigned URL for uploading a receipt image for a transaction.",
    path: "/api/v1/transactions/:id/receipt",
    method: "POST",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: z.void(),
    responses: {
      200: ZPresignedUploadResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
