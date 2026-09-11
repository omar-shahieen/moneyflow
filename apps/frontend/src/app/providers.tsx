import { useEffect } from "react";
import { ClerkProvider, useAuth } from "@clerk/clerk-react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Toaster } from "sonner";
import { CLERK_PUBLISHABLE_KEY, ENV } from "@/config/env";
import { setTokenGetter, setOnUnauthorized } from "@/api/client";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function AuthSync() {
  const { getToken, signOut } = useAuth();

  useEffect(() => {
    setTokenGetter(() => getToken({ template: "custom" }));
    setOnUnauthorized(() => {
      signOut();
    });
  }, [getToken, signOut]);

  return null;
}

function InnerProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthSync />
      {children}
      {ENV === "development" && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  );
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider publishableKey={CLERK_PUBLISHABLE_KEY}>
      <InnerProviders>
        {children}
        <Toaster position="top-right" richColors />
      </InnerProviders>
    </ClerkProvider>
  );
}
