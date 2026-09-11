export function isPlanLimitError(error: unknown): boolean {
  const normalized = error as { code?: string };
  return normalized?.code === "plan_limit_exceeded";
}
