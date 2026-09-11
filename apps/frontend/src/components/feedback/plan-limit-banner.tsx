import { AlertTriangle } from "lucide-react";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface PlanLimitBannerProps {
  feature: string;
  currentPlan: string;
  onUpgrade?: () => void;
  className?: string;
}

export function PlanLimitBanner({
  feature,
  currentPlan,
  onUpgrade,
  className,
}: PlanLimitBannerProps) {
  return (
    <Alert variant="warning" className={cn("", className)}>
      <AlertTriangle className="h-4 w-4" />
      <AlertTitle>Plan Limit Reached</AlertTitle>
      <AlertDescription className="flex items-center justify-between">
        <span>
          You've reached the {feature} limit on the{" "}
          <strong className="capitalize">{currentPlan}</strong> plan.
        </span>
        {onUpgrade && (
          <Button
            size="sm"
            onClick={onUpgrade}
            className="ml-4 shrink-0"
          >
            Upgrade Plan
          </Button>
        )}
      </AlertDescription>
    </Alert>
  );
}
