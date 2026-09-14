import { createBrowserRouter, Navigate } from "react-router-dom";
import { RootLayout } from "@/components/layout/root-layout";
import { DashboardPage } from "@/pages/dashboard";
import { SignInPage } from "@/pages/sign-in";
import { SignUpPage } from "@/pages/sign-up";
import { NotFoundPage } from "@/pages/not-found";
import { CategoriesPage } from "@/features/categories/categories-page";
import { TransactionsPage } from "@/features/transactions/transactions-page";
import { BudgetsPage } from "@/features/budgets/budgets-page";
import { SubscriptionPage } from "@/features/subscription/subscription-page";
import { ImportsPage } from "@/features/imports/imports-page";
import { ReportsPage } from "@/features/reports/reports-page";
import { SettingsPage } from "@/features/settings/settings-page";

export const router = createBrowserRouter([
  {
    path: "/sign-in/*",
    element: <SignInPage />,
  },
  {
    path: "/sign-up/*",
    element: <SignUpPage />,
  },
  {
    path: "/",
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: "dashboard",
        element: <DashboardPage />,
      },
      {
        path: "categories",
        element: <CategoriesPage />,
      },
      {
        path: "transactions",
        element: <TransactionsPage />,
      },
      {
        path: "budgets",
        element: <BudgetsPage />,
      },
      {
        path: "imports",
        element: <ImportsPage />,
      },
      {
        path: "reports",
        element: <ReportsPage />,
      },
      {
        path: "subscription",
        element: <SubscriptionPage />,
      },
      {
        path: "settings",
        element: <SettingsPage />,
      },
    ],
  },
  {
    path: "*",
    element: <NotFoundPage />,
  },
]);
