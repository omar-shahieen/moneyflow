import { z } from "zod";

export const ZReportResponse = z.object({
  id: z.string().uuid(),
  format: z.enum(["pdf", "csv"]),
  period_start: z.string().date(),
  period_end: z.string().date(),
  status: z.enum(["pending", "processing", "ready", "failed"]),
  download_url: z.string().url().optional(),
  created_at: z.string().datetime(),
  completed_at: z.string().datetime().optional(),
});

export const ZCreateReportRequest = z.object({
  format: z.enum(["pdf", "csv"]),
  period_start: z.string().date(),
  period_end: z.string().date(),
});
