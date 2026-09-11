export type ErrorCode =
  | "validation_failed"
  | "not_found"
  | "forbidden"
  | "conflict"
  | "plan_limit_exceeded"
  | "rate_limited"
  | "internal_error";

export type ApiError = {
  code: ErrorCode;
  message: string;
  field?: string;
};

export type NormalizedError = {
  status: number;
  code: ErrorCode;
  message: string;
  field?: string;
  requestId?: string;
};

export type ValidationError = NormalizedError & {
  code: "validation_failed";
  field: string;
};

export function isValidationError(
  error: NormalizedError
): error is ValidationError {
  return error.code === "validation_failed" && error.field !== undefined;
}

export function isPlanLimitError(
  error: NormalizedError
): error is NormalizedError & { code: "plan_limit_exceeded" } {
  return error.code === "plan_limit_exceeded";
}

export function isAuthError(
  error: NormalizedError
): error is NormalizedError & { status: 401 } {
  return error.status === 401;
}

export function isForbiddenError(
  error: NormalizedError
): error is NormalizedError & { status: 403 } {
  return error.status === 403;
}

export function isNotFoundError(
  error: NormalizedError
): error is NormalizedError & { status: 404 } {
  return error.status === 404;
}

export function isConflictError(
  error: NormalizedError
): error is NormalizedError & { status: 409 } {
  return error.status === 409;
}

export function isRateLimitError(
  error: NormalizedError
): error is NormalizedError & { status: 429 } {
  return error.status === 429;
}

export function isServerError(
  error: NormalizedError
): error is NormalizedError & { status: 500 } {
  return error.status >= 500;
}
