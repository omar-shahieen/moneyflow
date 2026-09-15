import { z } from "zod";

export const ZRecurringRule = z.object({
  id: z.string().uuid(),
  user_id: z.string(),
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  frequency: z.enum(["weekly", "monthly"]),
  next_run_date: z.string().date(),
  last_generated_date: z.string().date().optional(),
  end_date: z.string().date().optional(),
});

export const ZCreateRecurringRuleRequest = z.object({
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  frequency: z.enum(["weekly", "monthly"]),
  next_run_date: z.string().date(),
  end_date: z.string().date().optional(),
});

export const ZUpdateRecurringRuleRequest = z.object({
  category_id: z.string().uuid(),
  amount_minor: z.number(),
  frequency: z.enum(["weekly", "monthly"]),
  next_run_date: z.string().date(),
  end_date: z.string().date().optional(),
});
