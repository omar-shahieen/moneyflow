import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Save, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import {
  useUserSettings,
  useUpdateSettings,
  useNotificationPreferences,
  useUpdateNotificationPreferences,
  useExportAccount,
  TIMEZONES,
  CURRENCIES,
} from "./api";
import { toast } from "sonner";

const settingsSchema = z.object({
  display_name: z.string().min(1, "Name is required").max(100),
  default_currency: z.string().min(3),
  timezone: z.string().min(1),
});

type SettingsFormData = z.infer<typeof settingsSchema>;

export function SettingsPage() {
  const { data: settings, isLoading, error, refetch } = useUserSettings();
  const { data: notifications } = useNotificationPreferences();
  const updateSettings = useUpdateSettings();
  const updateNotifications = useUpdateNotificationPreferences();
  const exportAccount = useExportAccount();

  const [budgetAlerts, setBudgetAlerts] = useState(true);
  const [monthlySummary, setMonthlySummary] = useState(true);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<SettingsFormData>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      display_name: "",
      default_currency: "USD",
      timezone: "UTC",
    },
  });

  useEffect(() => {
    if (settings) {
      reset({
        display_name: settings.display_name,
        default_currency: settings.default_currency,
        timezone: settings.timezone,
      });
    }
  }, [settings, reset]);

  useEffect(() => {
    if (notifications) {
      setBudgetAlerts(notifications.budget_alerts);
      setMonthlySummary(notifications.monthly_summary);
    }
  }, [notifications]);

  if (isLoading) {
    return <LoadingPage message="Loading settings..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const onSaveSettings = async (data: SettingsFormData) => {
    const result = await updateSettings.mutateAsync(data);
    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Settings saved");
    }
  };

  const onSaveNotifications = async () => {
    const result = await updateNotifications.mutateAsync({
      budget_alerts: budgetAlerts,
      monthly_summary: monthlySummary,
    });
    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Notification preferences saved");
    }
  };

  const handleExport = async () => {
    const result = await exportAccount.mutateAsync();
    if (result instanceof Error) {
      toast.error(result.message);
    } else if (result.download_url) {
      window.open(result.download_url, "_blank");
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Update your personal information.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSaveSettings)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="display_name">Display Name</Label>
              <Input id="display_name" {...register("display_name")} />
              {errors.display_name && (
                <p className="text-sm text-destructive">
                  {errors.display_name.message}
                </p>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="default_currency">Default Currency</Label>
                <select
                  id="default_currency"
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                  {...register("default_currency")}
                >
                  {CURRENCIES.map((c) => (
                    <option key={c.code} value={c.code}>
                      {c.name} ({c.code})
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="timezone">Timezone</Label>
                <select
                  id="timezone"
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                  {...register("timezone")}
                >
                  {TIMEZONES.map((tz) => (
                    <option key={tz} value={tz}>
                      {tz}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <Button type="submit" disabled={!isDirty || updateSettings.isPending}>
              <Save className="mr-2 h-4 w-4" />
              Save Changes
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>
            Configure your notification preferences.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">Budget Alerts</p>
              <p className="text-sm text-muted-foreground">
                Get notified when you exceed a budget limit.
              </p>
            </div>
            <input
              type="checkbox"
              checked={budgetAlerts}
              onChange={(e) => setBudgetAlerts(e.target.checked)}
              className="h-4 w-4"
            />
          </div>

          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">Monthly Summary</p>
              <p className="text-sm text-muted-foreground">
                Receive a monthly summary of your finances.
              </p>
            </div>
            <input
              type="checkbox"
              checked={monthlySummary}
              onChange={(e) => setMonthlySummary(e.target.checked)}
              className="h-4 w-4"
            />
          </div>

          <Button
            onClick={onSaveNotifications}
            disabled={updateNotifications.isPending}
          >
            <Save className="mr-2 h-4 w-4" />
            Save Preferences
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Account Export</CardTitle>
          <CardDescription>
            Export all your data including transactions, budgets, and reports.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="outline"
            onClick={handleExport}
            disabled={exportAccount.isPending}
          >
            <Download className="mr-2 h-4 w-4" />
            {exportAccount.isPending ? "Exporting..." : "Export Data"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
