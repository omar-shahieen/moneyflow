import { useState } from "react";
import {
  Download,
  FileText,
  CheckCircle,
  XCircle,
  Clock,
} from "lucide-react";
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
import { Badge } from "@/components/ui/badge";
import { Pagination } from "@/components/ui/pagination";
import { SearchInput } from "@/components/ui/search-input";
import { SelectFilter } from "@/components/ui/select-filter";
import { SortSelect } from "@/components/ui/sort-select";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import { useReports, useReport, useCreateReport } from "./api";
import { useUrlFilters } from "@/hooks/use-url-filters";
import { toast } from "sonner";
import type { ReportFormat, ReportFilters } from "./api";

const DEFAULT_FILTERS: ReportFilters = {
  page: 1,
  page_size: 10,
};

export function ReportsPage() {
  const { filters, setFilter, setPage, setPageSize } =
    useUrlFilters<ReportFilters>({
      defaults: DEFAULT_FILTERS,
    });

  const { data, isLoading, error, refetch } = useReports(filters);
  const createReport = useCreateReport();
  const [format, setFormat] = useState<ReportFormat>("csv");
  const [periodStart, setPeriodStart] = useState("");
  const [periodEnd, setPeriodEnd] = useState("");
  const [activeReportId, setActiveReportId] = useState<string | null>(null);

  const { data: activeReport } = useReport(activeReportId ?? "");

  if (isLoading) {
    return <LoadingPage message="Loading reports..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const reports = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = data?.totalPages ?? 1;
  const page = data?.page ?? filters.page ?? 1;
  const limit = data?.limit ?? filters.page_size ?? 10;

  const handleGenerate = async () => {
    if (!periodStart || !periodEnd) {
      toast.error("Please select a date range");
      return;
    }

    const result = await createReport.mutateAsync({
      format,
      period_start: periodStart,
      period_end: periodEnd,
    });

    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Report generation started");
      setActiveReportId(result.id);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "ready":
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "failed":
        return <XCircle className="h-4 w-4 text-destructive" />;
      case "processing":
      case "pending":
        return <Clock className="h-4 w-4 text-muted-foreground" />;
      default:
        return null;
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Reports</h1>
        <p className="text-muted-foreground">
          Generate PDF or CSV reports of your transactions.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Generate Report</CardTitle>
          <CardDescription>
            Select a format and date range to generate a report.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Format</Label>
              <div className="flex gap-2">
                <Button
                  variant={format === "csv" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFormat("csv")}
                >
                  CSV
                </Button>
                <Button
                  variant={format === "pdf" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFormat("pdf")}
                >
                  PDF
                </Button>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="period-start">Start Date</Label>
              <Input
                id="period-start"
                type="date"
                value={periodStart}
                onChange={(e) => setPeriodStart(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="period-end">End Date</Label>
              <Input
                id="period-end"
                type="date"
                value={periodEnd}
                onChange={(e) => setPeriodEnd(e.target.value)}
              />
            </div>
          </div>

          <Button
            onClick={handleGenerate}
            disabled={createReport.isPending || !periodStart || !periodEnd}
          >
            <FileText className="mr-2 h-4 w-4" />
            {createReport.isPending ? "Generating..." : "Generate Report"}
          </Button>

          {activeReport && (
            <div className="rounded-lg border p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">
                  Current Report ({activeReport.format.toUpperCase()})
                </span>
                <Badge
                  variant={
                    activeReport.status === "ready" ? "default" : "secondary"
                  }
                >
                  {activeReport.status}
                </Badge>
              </div>
              {activeReport.status === "ready" && activeReport.download_url && (
                <Button size="sm" asChild>
                  <a
                    href={activeReport.download_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Download className="mr-2 h-4 w-4" />
                    Download
                  </a>
                </Button>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Report History</h2>
        <div className="flex flex-wrap items-end gap-4">
          <SearchInput
            value={filters.search}
            onChange={(v) => setFilter("search" as any, v)}
            placeholder="Search reports..."
            className="w-64"
          />
          <SelectFilter
            label="Status"
            options={[
              { value: "pending", label: "Pending" },
              { value: "processing", label: "Processing" },
              { value: "ready", label: "Ready" },
              { value: "failed", label: "Failed" },
            ]}
            value={filters.status}
            onChange={(v) => setFilter("status" as any, v)}
          />
          <SelectFilter
            label="Format"
            options={[
              { value: "csv", label: "CSV" },
              { value: "pdf", label: "PDF" },
            ]}
            value={filters.format}
            onChange={(v) => setFilter("format" as any, v)}
          />
          <SortSelect
            options={[
              { value: "created_at", label: "Created" },
              { value: "period_start", label: "Period Start" },
              { value: "period_end", label: "Period End" },
              { value: "status", label: "Status" },
            ]}
            value={filters.sort}
            order={filters.order}
            onChange={(sort, order) => {
              setFilter("sort" as any, sort);
              setFilter("order" as any, order);
            }}
          />
        </div>
        {reports.length === 0 ? (
          <EmptyState
            icon={<FileText className="h-8 w-8 text-muted-foreground" />}
            title="No reports yet"
            description="Generate your first report to get started."
          />
        ) : (
          <>
            <div className="space-y-3">
              {reports.map((report) => (
                <Card key={report.id}>
                  <CardContent className="flex items-center justify-between py-4">
                    <div className="flex items-center gap-3">
                      {getStatusIcon(report.status)}
                      <div>
                        <p className="text-sm font-medium">
                          {report.format.toUpperCase()} Report
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(report.period_start).toLocaleDateString()} -{" "}
                          {new Date(report.period_end).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <Badge
                        variant={
                          report.status === "ready" ? "default" : "secondary"
                        }
                      >
                        {report.status}
                      </Badge>
                      {report.status === "ready" && report.download_url && (
                        <Button size="sm" variant="ghost" asChild>
                          <a
                            href={report.download_url}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            <Download className="h-4 w-4" />
                          </a>
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>

            <Pagination
              page={page}
              totalPages={totalPages}
              total={total}
              limit={limit}
              onPageChange={setPage}
              onLimitChange={(size) => setPageSize(size)}
            />
          </>
        )}
      </div>
    </div>
  );
}
