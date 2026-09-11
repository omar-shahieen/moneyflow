import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome back! Here&apos;s your financial overview.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Balance</CardDescription>
            <CardTitle className="text-2xl">$0.00</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">+0% from last month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Income</CardDescription>
            <CardTitle className="text-2xl text-green-600">$0.00</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">This month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Expenses</CardDescription>
            <CardTitle className="text-2xl text-red-600">$0.00</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">This month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Transactions</CardDescription>
            <CardTitle className="text-2xl">0</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">This month</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
          <CardDescription>Your latest financial activity</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No transactions yet. Start by adding your first transaction.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
