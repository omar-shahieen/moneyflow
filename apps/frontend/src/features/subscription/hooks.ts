export function isPlanLimitError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "plan_limit_exceeded";
}

export function isNotFoundError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "not_found";
}

export function isForbiddenError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "forbidden";
}

export function isConflictError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "conflict";
}

export function isValidationError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "validation_failed";
}
