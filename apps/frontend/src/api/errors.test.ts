import { describe, it, expect, vi } from "vitest";

vi.mock("@/config/env", () => ({
  ENV: "test",
  API_URL: "http://localhost:8080",
  CLERK_PUBLISHABLE_KEY: "pk_test_placeholder",
}));

import { normalizeError } from "./client";
import {
  isValidationError,
  isPlanLimitError,
  isAuthError,
  isForbiddenError,
  isNotFoundError,
  isConflictError,
  isRateLimitError,
  isServerError,
} from "./errors";
import type { NormalizedError } from "./errors";

describe("normalizeError", () => {
  it("normalizes an Axios error with response", () => {
    const axiosError = {
      isAxiosError: true,
      response: {
        status: 404,
        data: {
          code: "not_found",
          message: "Resource not found",
        },
        headers: { "x-request-id": "req-123" },
      },
      message: "Request failed with status code 404",
    };

    const result = normalizeError(axiosError);
    expect(result).toEqual({
      status: 404,
      code: "not_found",
      message: "Resource not found",
      field: undefined,
      requestId: "req-123",
    });
  });

  it("normalizes a non-Axios error", () => {
    const error = new Error("Network error");
    const result = normalizeError(error);
    expect(result).toEqual({
      status: 500,
      code: "internal_error",
      message: "Network error",
      requestId: undefined,
    });
  });

  it("normalizes an unknown error", () => {
    const result = normalizeError("something went wrong");
    expect(result).toEqual({
      status: 500,
      code: "internal_error",
      message: "An unexpected error occurred",
      requestId: undefined,
    });
  });

  it("includes requestId when provided", () => {
    const result = normalizeError(new Error("test"), "req-456");
    expect(result.requestId).toBe("req-456");
  });
});

describe("error type guards", () => {
  const validationError: NormalizedError = {
    status: 422,
    code: "validation_failed",
    message: "Invalid input",
    field: "email",
  };

  const planLimitError: NormalizedError = {
    status: 402,
    code: "plan_limit_exceeded",
    message: "Plan limit exceeded",
  };

  const authError: NormalizedError = {
    status: 401,
    code: "internal_error",
    message: "Unauthorized",
  };

  const forbiddenError: NormalizedError = {
    status: 403,
    code: "forbidden",
    message: "Forbidden",
  };

  const notFoundError: NormalizedError = {
    status: 404,
    code: "not_found",
    message: "Not found",
  };

  const conflictError: NormalizedError = {
    status: 409,
    code: "conflict",
    message: "Conflict",
  };

  const rateLimitError: NormalizedError = {
    status: 429,
    code: "rate_limited",
    message: "Too many requests",
  };

  const serverError: NormalizedError = {
    status: 500,
    code: "internal_error",
    message: "Internal server error",
  };

  it("identifies validation errors", () => {
    expect(isValidationError(validationError)).toBe(true);
    expect(isValidationError(serverError)).toBe(false);
  });

  it("identifies plan limit errors", () => {
    expect(isPlanLimitError(planLimitError)).toBe(true);
    expect(isPlanLimitError(serverError)).toBe(false);
  });

  it("identifies auth errors", () => {
    expect(isAuthError(authError)).toBe(true);
    expect(isAuthError(serverError)).toBe(false);
  });

  it("identifies forbidden errors", () => {
    expect(isForbiddenError(forbiddenError)).toBe(true);
    expect(isForbiddenError(serverError)).toBe(false);
  });

  it("identifies not found errors", () => {
    expect(isNotFoundError(notFoundError)).toBe(true);
    expect(isNotFoundError(serverError)).toBe(false);
  });

  it("identifies conflict errors", () => {
    expect(isConflictError(conflictError)).toBe(true);
    expect(isConflictError(serverError)).toBe(false);
  });

  it("identifies rate limit errors", () => {
    expect(isRateLimitError(rateLimitError)).toBe(true);
    expect(isRateLimitError(serverError)).toBe(false);
  });

  it("identifies server errors", () => {
    expect(isServerError(serverError)).toBe(true);
    expect(isServerError(notFoundError)).toBe(false);
  });
});
