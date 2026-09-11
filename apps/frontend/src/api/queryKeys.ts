export const queryKeys = {
  health: {
    all: ["health"] as const,
    status: () => [...queryKeys.health.all, "status"] as const,
  },
  categories: {
    all: ["categories"] as const,
    lists: () => [...queryKeys.categories.all, "list"] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.categories.lists(), filters] as const,
    details: () => [...queryKeys.categories.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.categories.details(), id] as const,
  },
  transactions: {
    all: ["transactions"] as const,
    lists: () => [...queryKeys.transactions.all, "list"] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.transactions.lists(), filters] as const,
    details: () => [...queryKeys.transactions.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.transactions.details(), id] as const,
    summary: (month?: string) =>
      [...queryKeys.transactions.all, "summary", month] as const,
  },
  budgets: {
    all: ["budgets"] as const,
    lists: () => [...queryKeys.budgets.all, "list"] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.budgets.lists(), filters] as const,
    details: () => [...queryKeys.budgets.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.budgets.details(), id] as const,
  },
  subscription: {
    all: ["subscription"] as const,
    current: () => [...queryKeys.subscription.all, "current"] as const,
  },
  reports: {
    all: ["reports"] as const,
    lists: () => [...queryKeys.reports.all, "list"] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.reports.lists(), filters] as const,
    details: () => [...queryKeys.reports.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.reports.details(), id] as const,
  },
  imports: {
    all: ["imports"] as const,
    details: () => [...queryKeys.imports.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.imports.details(), id] as const,
  },
} as const;
