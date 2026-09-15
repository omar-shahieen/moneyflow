import { z } from "zod";

export const ZUserAccount = z.object({
  id: z.string(),
  email: z.string().email(),
  display_name: z.string(),
  avatar_url: z.string().url().optional(),
  default_currency: z.string().length(3),
  timezone: z.string(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

export const ZUpdateUserRequest = z.object({
  display_name: z.string().min(1).max(100).optional(),
  avatar_url: z.string().url().optional(),
  default_currency: z.string().length(3).optional(),
  timezone: z.string().optional(),
});
