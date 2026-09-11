import { describe, it, expect } from "vitest";
import {
  isPlanLimitError,
  isNotFoundError,
  isForbiddenError,
  isConflictError,
  isValidationError,
} from "./hooks";

describe("isPlanLimitError", () => {
  it("returns true for plan_limit_exceeded error", () => {
    expect(isPlanLimitError({ code: "plan_limit_exceeded" })).toBe(true);
  });

  it("returns false for other errors", () => {
    expect(isPlanLimitError({ code: "not_found" })).toBe(false);
    expect(isPlanLimitError({ code: "internal_error" })).toBe(false);
    expect(isPlanLimitError({})).toBe(false);
    expect(isPlanLimitError(null)).toBe(false);
  });
});

describe("isNotFoundError", () => {
  it("returns true for not_found error", () => {
    expect(isNotFoundError({ code: "not_found" })).toBe(true);
  });

  it("returns false for other errors", () => {
    expect(isNotFoundError({ code: "plan_limit_exceeded" })).toBe(false);
  });
});

describe("isForbiddenError", () => {
  it("returns true for forbidden error", () => {
    expect(isForbiddenError({ code: "forbidden" })).toBe(true);
  });

  it("returns false for other errors", () => {
    expect(isForbiddenError({ code: "not_found" })).toBe(false);
  });
});

describe("isConflictError", () => {
  it("returns true for conflict error", () => {
    expect(isConflictError({ code: "conflict" })).toBe(true);
  });

  it("returns false for other errors", () => {
    expect(isConflictError({ code: "not_found" })).toBe(false);
  });
});

describe("isValidationError", () => {
  it("returns true for validation_failed error", () => {
    expect(isValidationError({ code: "validation_failed" })).toBe(true);
  });

  it("returns false for other errors", () => {
    expect(isValidationError({ code: "not_found" })).toBe(false);
  });
});
