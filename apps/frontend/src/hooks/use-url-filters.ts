import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";

type FilterValue = string | number | boolean | undefined;

type UseUrlFiltersOptions<T extends Record<string, FilterValue>> = {
  defaults: T;
  pageKey?: string;
  pageSizeKey?: string;
};

export function useUrlFilters<T extends Record<string, FilterValue>>({
  defaults,
  pageKey = "page",
  pageSizeKey = "page_size",
}: UseUrlFiltersOptions<T>) {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => {
    const result = { ...defaults } as Record<string, FilterValue>;

    for (const [key] of Object.entries(defaults)) {
      const urlValue = searchParams.get(key);
      if (urlValue !== null) {
        const defaultValue = defaults[key];
        if (typeof defaultValue === "number") {
          const parsed = Number(urlValue);
          if (!isNaN(parsed) && parsed > 0) {
            result[key] = parsed;
          }
        } else if (typeof defaultValue === "boolean") {
          result[key] = urlValue === "true";
        } else {
          result[key] = urlValue as string;
        }
      }
    }

    return result as T;
  }, [searchParams, defaults]);

  const setFilter = useCallback(
    (key: keyof T, value: FilterValue) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);

        if (value === undefined || value === "" || value === null) {
          next.delete(key as string);
        } else {
          next.set(key as string, String(value));
        }

        // Reset to page 1 when filters change (but not when changing page itself)
        if (key !== pageKey) {
          next.delete(pageKey);
        }

        return next;
      });
    },
    [setSearchParams, pageKey]
  );

  const setFilters = useCallback(
    (updates: Partial<T>) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);

        for (const [key, value] of Object.entries(updates)) {
          if (value === undefined || value === "" || value === null) {
            next.delete(key);
          } else {
            next.set(key, String(value));
          }
        }

        // Reset to page 1 when filters change (but not when changing page)
        if (!("page" in updates)) {
          next.delete(pageKey);
        }

        return next;
      });
    },
    [setSearchParams, pageKey]
  );

  const resetFilters = useCallback(() => {
    setSearchParams(new URLSearchParams());
  }, [setSearchParams]);

  const setPage = useCallback(
    (page: number) => {
      setFilter(pageKey as keyof T, page === 1 ? undefined : page);
    },
    [setFilter, pageKey]
  );

  const setPageSize = useCallback(
    (size: number) => {
      setFilters({
        [pageSizeKey]: size === 20 ? undefined : size,
        [pageKey]: undefined,
      } as Partial<T>);
    },
    [setFilters, pageSizeKey, pageKey]
  );

  return {
    filters,
    setFilter,
    setFilters,
    resetFilters,
    setPage,
    setPageSize,
  };
}
