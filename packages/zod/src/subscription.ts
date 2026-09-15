import { z } from "zod";

export const ZSubscriptionResponse = z.object({
  id: z.string().uuid(),
  user_id: z.string(),
  plan: z.enum(["free", "pro", "vip"]),
  status: z.enum(["active", "past_due", "canceled"]),
  payment_provider: z.string().optional(),
  provider_customer_id: z.string().optional(),
  provider_subscription_id: z.string().optional(),
  current_period_end: z.string().datetime().optional(),
});
