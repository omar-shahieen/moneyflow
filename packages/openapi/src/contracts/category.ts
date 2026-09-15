import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZCategory,
  ZCreateCategoryRequest,
  ZUpdateCategoryRequest,
  schemaWithPagination,
} from "@moneyflow/zod";
import { ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const categoryContract = c.router({
  listCategories: {
    summary: "List categories",
    description: "List all categories for the authenticated user with pagination and optional type filter.",
    path: "/api/v1/categories",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      type: z.enum(["income", "expense"]).optional(),
    }),
    responses: {
      200: schemaWithPagination(ZCategory),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getCategoryById: {
    summary: "Get category by ID",
    path: "/api/v1/categories/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZCategory,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createCategory: {
    summary: "Create a category",
    description: "Create a new income or expense category.",
    path: "/api/v1/categories",
    method: "POST",
    body: ZCreateCategoryRequest,
    responses: {
      201: ZCategory,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  updateCategory: {
    summary: "Update a category",
    path: "/api/v1/categories/:id",
    method: "PATCH",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: ZUpdateCategoryRequest,
    responses: {
      200: ZCategory,
      400: ZErrorResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  deleteCategory: {
    summary: "Delete a category",
    path: "/api/v1/categories/:id",
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
