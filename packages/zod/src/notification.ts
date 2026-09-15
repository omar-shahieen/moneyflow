import { z } from "zod";

export const ZNotification = z.object({
  id: z.string().uuid(),
  user_id: z.string(),
  title: z.string(),
  message: z.string(),
  type: z.enum(["info", "warning", "error", "success"]),
  read: z.boolean(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

export const ZCreateNotificationRequest = z.object({
  title: z.string().min(1).max(200),
  message: z.string().min(1).max(1000),
  type: z.enum(["info", "warning", "error", "success"]).optional(),
});

export const ZUnreadCountResponse = z.object({
  count: z.number(),
});
