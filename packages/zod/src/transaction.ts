import { z } from "zod";

export const ZTransaction = z.object({
  id: z.string().uuid(),
  user_id: z.string(),
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  currency: z.string().length(3),
  note: z.string(),
  receipt_key: z.string().optional(),
  occurred_at: z.string().datetime(),
  created_at: z.string().datetime(),
});

export const ZCreateTransactionRequest = z.object({
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  currency: z.string().length(3).optional(),
  note: z.string().max(500).optional(),
  receipt_key: z.string().optional(),
  occurred_at: z.string().datetime(),
});

export const ZUpdateTransactionRequest = z.object({
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  currency: z.string().length(3).optional(),
  note: z.string().max(500).optional(),
  receipt_key: z.string().optional(),
  occurred_at: z.string().datetime(),
});

export const ZCategoryTotal = z.object({
  category_id: z.string().uuid(),
  category_name: z.string(),
  total: z.number(),
  type: z.string(),
});

export const ZCurrencyTotal = z.object({
  currency: z.string(),
  total_income: z.number(),
  total_expense: z.number(),
});

export const ZTransactionSummary = z.object({
  total_income: z.number(),
  total_expense: z.number(),
  by_category: z.array(ZCategoryTotal),
  by_currency: z.array(ZCurrencyTotal),
});

export const ZPresignedUploadResponse = z.object({
  upload_url: z.string().url(),
  storage_key: z.string(),
  expires_at: z.string().datetime(),
});
