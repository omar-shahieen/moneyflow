import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { App } from "./app/App";

vi.mock("@/config/env", () => ({
  ENV: "test",
  API_URL: "http://localhost:8080",
  CLERK_PUBLISHABLE_KEY: "pk_test_placeholder",
}));

vi.mock("@clerk/clerk-react", () => ({
  ClerkProvider: ({ children }: { children: React.ReactNode }) => children,
  useAuth: () => ({ isLoaded: true, isSignedIn: true }),
  UserButton: () => null,
  SignIn: () => null,
  SignUp: () => null,
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    RouterProvider: () => <div>Router Provider</div>,
  };
});

describe("App", () => {
  it("renders without crashing", () => {
    render(<App />);
    expect(screen.getByText("Router Provider")).toBeInTheDocument();
  });
});
