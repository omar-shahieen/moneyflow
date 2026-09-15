import { z } from "zod";

export const ZCategory = z.object({
  id: z.string().uuid(),
  user_id: z.string(),
  name: z.string().max(100),
  type: z.enum(["income", "expense"]),
  created_at: z.string().datetime(),
});

export const ZCreateCategoryRequest = z.object({
  name: z.string().min(1).max(100),
  type: z.enum(["income", "expense"]),
});

export const ZUpdateCategoryRequest = z.object({
  name: z.string().min(1).max(100),
  type: z.enum(["income", "expense"]),
});
