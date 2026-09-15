import { z } from "zod";

export const ZImportRowError = z.object({
  row: z.number(),
  reason: z.string(),
});

export const ZImportResponse = z.object({
  id: z.string().uuid(),
  status: z.enum(["pending", "processing", "completed", "failed"]),
  total_rows: z.number(),
  success_rows: z.number(),
  failed_rows: z.array(ZImportRowError).optional(),
  created_at: z.string().datetime(),
});

export const ZListImportsQuery = z.object({
  page: z.number().int().positive().optional(),
  page_size: z.number().int().min(1).max(100).optional(),
  sort: z.string().optional(),
  order: z.enum(["asc", "desc"]).optional(),
  status: z.enum(["pending", "processing", "completed", "failed"]).optional(),
});
