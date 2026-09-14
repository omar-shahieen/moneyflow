import { Crown, ArrowUpRight } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useCreateCheckout, PLAN_PRICES } from "./api";
import { toast } from "sonner";
import type { PlanType } from "./api";

interface PlanLimitPromptProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  feature?: string;
  currentPlan?: PlanType;
}

export function PlanLimitPrompt({
  open,
  onOpenChange,
  feature,
  currentPlan = "free",
}: PlanLimitPromptProps) {
  const createCheckout = useCreateCheckout();

  const handleUpgrade = async (plan: PlanType) => {
    const result = await createCheckout.mutateAsync({ plan, billing_period: "monthly" });
    if (result instanceof Error) {
      toast.error(result.message);
    } else if (result.checkout_url) {
      window.open(result.checkout_url, "_blank");
    }
  };

  const nextPlan: PlanType = currentPlan === "free" ? "pro" : "vip";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Crown className="h-5 w-5 text-primary" />
            <DialogTitle>Upgrade Required</DialogTitle>
          </div>
          <DialogDescription>
            {feature
              ? `You've reached the ${feature} limit on the ${currentPlan} plan.`
              : `You've reached a limit on the ${currentPlan} plan.`}{" "}
            Upgrade to continue using this feature.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-4">
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div>
              <p className="font-medium capitalize">{nextPlan} Plan</p>
              <p className="text-sm text-muted-foreground">
                {PLAN_PRICES[nextPlan].monthly === 0
                  ? "Free"
                  : `$${PLAN_PRICES[nextPlan].monthly}/month`}
              </p>
            </div>
            <Button
              onClick={() => handleUpgrade(nextPlan)}
              disabled={createCheckout.isPending}
            >
              Upgrade
              <ArrowUpRight className="ml-2 h-4 w-4" />
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Maybe Later
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}


