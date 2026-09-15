import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZBudget,
  ZBudgetWithMembers,
  ZBudgetResponse,
  ZCreateBudgetRequest,
  ZUpdateBudgetRequest,
  ZAddMemberRequest,
  schemaWithPagination,
} from "@moneyflow/zod";
import { ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const budgetContract = c.router({
  listBudgets: {
    summary: "List budgets",
    description: "List budgets with pagination and optional search by category name.",
    path: "/api/v1/budgets",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      search: z.string().optional(),
    }),
    responses: {
      200: schemaWithPagination(ZBudgetResponse),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getBudgetById: {
    summary: "Get budget by ID",
    description: "Get a budget with usage stats and membership info.",
    path: "/api/v1/budgets/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZBudgetResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createBudget: {
    summary: "Create a budget",
    path: "/api/v1/budgets",
    method: "POST",
    body: ZCreateBudgetRequest,
    responses: {
      201: ZBudgetWithMembers,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  updateBudget: {
    summary: "Update a budget",
    path: "/api/v1/budgets/:id",
    method: "PATCH",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: ZUpdateBudgetRequest,
    responses: {
      200: ZBudget,
      400: ZErrorResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  addBudgetMember: {
    summary: "Add budget member",
    description: "Add a member to a budget.",
    path: "/api/v1/budgets/:id/members",
    method: "POST",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: ZAddMemberRequest,
    responses: {
      204: z.void(),
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  removeBudgetMember: {
    summary: "Remove budget member",
    description: "Remove a member from a budget.",
    path: "/api/v1/budgets/:id/members/:userId",
    method: "DELETE",
    pathParams: z.object({
      id: z.string().uuid(),
      userId: z.string(),
    }),
    responses: {
      204: z.void(),
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
