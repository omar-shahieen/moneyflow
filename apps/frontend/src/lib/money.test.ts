import { describe, it, expect } from "vitest";
import { formatMoney, formatMoneyCompact, parseMoneyInput } from "./money";

describe("formatMoney", () => {
  it("formats USD amounts correctly", () => {
    expect(formatMoney(1000, "USD")).toBe("$10.00");
    expect(formatMoney(0, "USD")).toBe("$0.00");
    expect(formatMoney(99, "USD")).toBe("$0.99");
    expect(formatMoney(100000, "USD")).toBe("$1,000.00");
  });

  it("formats EUR amounts correctly", () => {
    expect(formatMoney(1000, "EUR")).toContain("10.00");
  });

  it("handles negative amounts", () => {
    expect(formatMoney(-1000, "USD")).toBe("-$10.00");
  });

  it("handles zero", () => {
    expect(formatMoney(0, "USD")).toBe("$0.00");
  });
});

describe("formatMoneyCompact", () => {
  it("formats small amounts normally", () => {
    expect(formatMoneyCompact(1000, "USD")).toBe("$10.00");
    expect(formatMoneyCompact(99, "USD")).toBe("$0.99");
  });

  it("formats thousands with K suffix", () => {
    expect(formatMoneyCompact(100000, "USD")).toBe("1.0K USD");
    expect(formatMoneyCompact(1500000, "USD")).toBe("15.0K USD");
  });

  it("formats millions with M suffix", () => {
    expect(formatMoneyCompact(100000000, "USD")).toBe("1.0M USD");
  });
});

describe("parseMoneyInput", () => {
  it("parses valid decimal input", () => {
    expect(parseMoneyInput("10.00")).toBe(1000);
    expect(parseMoneyInput("0.99")).toBe(99);
    expect(parseMoneyInput("1000")).toBe(100000);
  });

  it("strips non-numeric characters", () => {
    expect(parseMoneyInput("$10")).toBe(1000);
    expect(parseMoneyInput("abc")).toBeNull();
  });

  it("returns null for empty or non-parseable input", () => {
    expect(parseMoneyInput("")).toBeNull();
    expect(parseMoneyInput("   ")).toBeNull();
  });

  it("handles zero", () => {
    expect(parseMoneyInput("0")).toBe(0);
    expect(parseMoneyInput("0.00")).toBe(0);
  });

  it("handles integer input", () => {
    expect(parseMoneyInput("10")).toBe(1000);
    expect(parseMoneyInput("100")).toBe(10000);
  });
});
