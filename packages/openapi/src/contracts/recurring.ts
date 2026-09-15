import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZRecurringRule,
  ZCreateRecurringRuleRequest,
  ZUpdateRecurringRuleRequest,
  schemaWithPagination,
} from "@moneyflow/zod";
import { ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const recurringRuleContract = c.router({
  listRecurringRules: {
    summary: "List recurring rules",
    description: "List recurring rules with pagination, filterable by frequency.",
    path: "/api/v1/recurring-rules",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      frequency: z.enum(["weekly", "monthly"]).optional(),
      search: z.string().optional(),
    }),
    responses: {
      200: schemaWithPagination(ZRecurringRule),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getRecurringRuleById: {
    summary: "Get recurring rule by ID",
    path: "/api/v1/recurring-rules/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZRecurringRule,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createRecurringRule: {
    summary: "Create a recurring rule",
    path: "/api/v1/recurring-rules",
    method: "POST",
    body: ZCreateRecurringRuleRequest,
    responses: {
      201: ZRecurringRule,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  updateRecurringRule: {
    summary: "Update a recurring rule",
    path: "/api/v1/recurring-rules/:id",
    method: "PATCH",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: ZUpdateRecurringRuleRequest,
    responses: {
      200: ZRecurringRule,
      400: ZErrorResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  deleteRecurringRule: {
    summary: "Delete a recurring rule",
    path: "/api/v1/recurring-rules/:id",
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
});
