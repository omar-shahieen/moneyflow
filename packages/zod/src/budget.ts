import { z } from "zod";

export const ZBudgetMember = z.object({
  budget_id: z.string().uuid(),
  user_id: z.string(),
  role: z.enum(["owner", "member"]),
});

export const ZBudget = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  monthly_limit_minor: z.number(),
  currency: z.string(),
  created_at: z.string().datetime(),
});

export const ZBudgetWithMembers = ZBudget.extend({
  members: z.array(ZBudgetMember),
});

export const ZBudgetResponse = ZBudgetWithMembers.extend({
  usage: z.number(),
  usage_percent: z.number(),
  exceeded: z.boolean(),
});

export const ZCreateBudgetRequest = z.object({
  category_id: z.string().uuid(),
  monthly_limit_minor: z.number().positive(),
  currency: z.string().length(3).optional(),
});

export const ZUpdateBudgetRequest = z.object({
  category_id: z.string().uuid(),
  monthly_limit_minor: z.number().positive(),
  currency: z.string().length(3).optional(),
});

export const ZAddMemberRequest = z.object({
  user_id: z.string(),
});
