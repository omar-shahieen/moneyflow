import { describe, it, expect } from "vitest";
import { queryKeys } from "./queryKeys";

describe("queryKeys", () => {
  describe("categories", () => {
    it("returns correct keys", () => {
      expect(queryKeys.categories.all).toEqual(["categories"]);
      expect(queryKeys.categories.lists()).toEqual(["categories", "list"]);
      expect(queryKeys.categories.detail("123")).toEqual([
        "categories",
        "detail",
        "123",
      ]);
    });
  });

  describe("transactions", () => {
    it("returns correct keys", () => {
      expect(queryKeys.transactions.all).toEqual(["transactions"]);
      expect(queryKeys.transactions.lists()).toEqual(["transactions", "list"]);
      expect(queryKeys.transactions.detail("123")).toEqual([
        "transactions",
        "detail",
        "123",
      ]);
      expect(queryKeys.transactions.summary("2024-01")).toEqual([
        "transactions",
        "summary",
        "2024-01",
      ]);
    });

    it("includes filters in list key", () => {
      const filters = { page: 1, type: "income" as const };
      expect(queryKeys.transactions.list(filters)).toEqual([
        "transactions",
        "list",
        filters,
      ]);
    });
  });

  describe("budgets", () => {
    it("returns correct keys", () => {
      expect(queryKeys.budgets.all).toEqual(["budgets"]);
      expect(queryKeys.budgets.lists()).toEqual(["budgets", "list"]);
      expect(queryKeys.budgets.detail("123")).toEqual([
        "budgets",
        "detail",
        "123",
      ]);
    });
  });

  describe("user", () => {
    it("returns correct keys", () => {
      expect(queryKeys.user.all).toEqual(["user"]);
      expect(queryKeys.user.detail()).toEqual(["user", "detail"]);
      expect(queryKeys.user.notifications()).toEqual(["user", "notifications"]);
    });
  });

  describe("subscription", () => {
    it("returns correct keys", () => {
      expect(queryKeys.subscription.all).toEqual(["subscription"]);
      expect(queryKeys.subscription.current()).toEqual([
        "subscription",
        "current",
      ]);
    });
  });

  describe("reports", () => {
    it("returns correct keys", () => {
      expect(queryKeys.reports.all).toEqual(["reports"]);
      expect(queryKeys.reports.lists()).toEqual(["reports", "list"]);
      expect(queryKeys.reports.detail("123")).toEqual([
        "reports",
        "detail",
        "123",
      ]);
    });
  });

  describe("imports", () => {
    it("returns correct keys", () => {
      expect(queryKeys.imports.all).toEqual(["imports"]);
      expect(queryKeys.imports.lists()).toEqual(["imports", "list"]);
      expect(queryKeys.imports.detail("123")).toEqual([
        "imports",
        "detail",
        "123",
      ]);
    });
  });
});
