import { z } from "zod";

export const ZFieldError = z.object({
  field: z.string(),
  message: z.string(),
});

export const ZErrorResponse = z.object({
  code: z.string(),
  message: z.string(),
  details: z.array(ZFieldError).optional(),
});

export const ZUnauthorizedResponse = z.object({
  code: z.literal("UNAUTHORIZED"),
  message: z.string(),
});

export const ZRateLimitResponse = z.object({
  code: z.literal("RATE_LIMIT_EXCEEDED"),
  message: z.string(),
});

export const ZNotFoundResponse = z.object({
  code: z.literal("NOT_FOUND"),
  message: z.string(),
});

export const ZValidationErrorResponse = z.object({
  code: z.literal("VALIDATION_ERROR"),
  message: z.string(),
  details: z.array(ZFieldError),
});

export const ZInternalErrorResponse = z.object({
  code: z.literal("INTERNAL_ERROR"),
  message: z.string(),
});
