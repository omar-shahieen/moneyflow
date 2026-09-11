import { useEffect } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { useAuth, UserButton } from "@clerk/clerk-react";
import { Loader2, LayoutDashboard } from "lucide-react";

function LoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );
}

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isLoaded, isSignedIn } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isLoaded && !isSignedIn) {
      navigate("/sign-in", { replace: true });
    }
  }, [isLoaded, isSignedIn, navigate]);

  if (!isLoaded) {
    return <LoadingScreen />;
  }

  if (!isSignedIn) {
    return null;
  }

  return <>{children}</>;
}

function Sidebar() {
  return (
    <aside className="hidden w-64 border-r bg-muted/30 lg:block">
      <div className="flex h-16 items-center gap-2 border-b px-6">
        <LayoutDashboard className="h-6 w-6" />
        <span className="text-lg font-semibold">MoneyFlow</span>
      </div>
      <nav className="space-y-1 p-4">
        <a
          href="/dashboard"
          className="flex items-center gap-3 rounded-md bg-primary/10 px-3 py-2 text-sm font-medium text-primary"
        >
          <LayoutDashboard className="h-4 w-4" />
          Dashboard
        </a>
      </nav>
    </aside>
  );
}

function TopBar() {
  return (
    <header className="flex h-16 items-center justify-between border-b bg-background px-6">
      <div className="lg:hidden">
        <LayoutDashboard className="h-6 w-6" />
      </div>
      <div className="flex-1" />
      <UserButton
        afterSignOutUrl="/sign-in"
        appearance={{
          elements: {
            avatarBox: "h-8 w-8",
          },
        }}
      />
    </header>
  );
}

export function RootLayout() {
  return (
    <AuthGuard>
      <div className="flex min-h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <TopBar />
          <main className="flex-1 p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </AuthGuard>
  );
}
