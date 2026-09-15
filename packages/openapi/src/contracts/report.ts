import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZReportResponse,
  ZCreateReportRequest,
  schemaWithPagination,
} from "@moneyflow/zod";
import { ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const reportContract = c.router({
  listReports: {
    summary: "List reports",
    description: "List reports with pagination. Filterable by status and format.",
    path: "/api/v1/reports",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      status: z.enum(["pending", "processing", "ready", "failed"]).optional(),
      format: z.enum(["pdf", "csv"]).optional(),
    }),
    responses: {
      200: schemaWithPagination(ZReportResponse),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getReportById: {
    summary: "Get report by ID",
    description: "Get a report by ID with download URL if ready.",
    path: "/api/v1/reports/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZReportResponse,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createReport: {
    summary: "Create a report",
    description: "Create a new report generation request. Processing is asynchronous.",
    path: "/api/v1/reports",
    method: "POST",
    body: ZCreateReportRequest,
    responses: {
      202: ZReportResponse,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
