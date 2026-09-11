import { createBrowserRouter, Navigate } from "react-router-dom";
import { RootLayout } from "@/components/layout/root-layout";
import { DashboardPage } from "@/pages/dashboard";
import { SignInPage } from "@/pages/sign-in";
import { SignUpPage } from "@/pages/sign-up";
import { NotFoundPage } from "@/pages/not-found";

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
    ],
  },
  {
    path: "*",
    element: <NotFoundPage />,
  },
]);
