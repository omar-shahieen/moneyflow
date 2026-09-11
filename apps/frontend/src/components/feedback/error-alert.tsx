import { AlertCircle, RefreshCw } from "lucide-react";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import type { NormalizedError } from "@/api/errors";

interface ErrorAlertProps {
  error: NormalizedError;
  onRetry?: () => void;
  className?: string;
}

export function ErrorAlert({ error, onRetry, className }: ErrorAlertProps) {
  const getErrorMessage = (status: number, code: string): string => {
    switch (code) {
      case "validation_failed":
        return "Please check your input and try again.";
      case "not_found":
        return "The requested resource was not found.";
      case "forbidden":
        return "You don't have permission to access this resource.";
      case "conflict":
        return "A conflict occurred. Please try again.";
      case "plan_limit_exceeded":
        return "You've reached your plan limit. Please upgrade to continue.";
      case "rate_limited":
        return "Too many requests. Please wait a moment and try again.";
      case "internal_error":
      default:
        return status >= 500
          ? "A server error occurred. Please try again later."
          : "An unexpected error occurred.";
    }
  };

  return (
    <Alert variant="destructive" className={className}>
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription className="flex items-center justify-between">
        <span>{getErrorMessage(error.status, error.code)}</span>
        {onRetry && (
          <Button
            variant="outline"
            size="sm"
            onClick={onRetry}
            className="ml-4 shrink-0"
          >
            <RefreshCw className="mr-2 h-3 w-3" />
            Retry
          </Button>
        )}
      </AlertDescription>
    </Alert>
  );
}
