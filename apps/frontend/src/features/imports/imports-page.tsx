import { useState, useRef } from "react";
import { Upload, FileText, CheckCircle, XCircle, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPage } from "@/components/feedback/loading";
import { ErrorAlert } from "@/components/feedback/error-alert";
import { useImports, useImport, useCreateImport } from "./api";
import { toast } from "sonner";

export function ImportsPage() {
  const { data: imports, isLoading, error, refetch } = useImports();
  const createImport = useCreateImport();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [activeImportId, setActiveImportId] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data: activeImport } = useImport(activeImportId ?? "");

  if (isLoading) {
    return <LoadingPage message="Loading imports..." />;
  }

  if (error) {
    return <ErrorAlert error={error} onRetry={() => refetch()} />;
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    const result = await createImport.mutateAsync(selectedFile);
    if (result instanceof Error) {
      toast.error(result.message);
    } else {
      toast.success("Import started processing");
      setActiveImportId(result.id);
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "completed":
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
        <h1 className="text-3xl font-bold tracking-tight">CSV Imports</h1>
        <p className="text-muted-foreground">
          Bulk upload transactions from a CSV file.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Upload CSV</CardTitle>
          <CardDescription>
            Select a CSV file to import transactions. Maximum file size: 10MB.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <Label htmlFor="csv-file" className="sr-only">
                Choose file
              </Label>
              <input
                ref={fileInputRef}
                id="csv-file"
                type="file"
                accept=".csv"
                onChange={handleFileSelect}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              />
            </div>
            <Button
              onClick={handleUpload}
              disabled={!selectedFile || createImport.isPending}
            >
              <Upload className="mr-2 h-4 w-4" />
              {createImport.isPending ? "Uploading..." : "Upload"}
            </Button>
          </div>

          {activeImport && (
            <div className="rounded-lg border p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Current Import</span>
                <Badge variant={activeImport.status === "completed" ? "default" : "secondary"}>
                  {activeImport.status}
                </Badge>
              </div>
              <div className="text-sm text-muted-foreground">
                {activeImport.success_rows} of {activeImport.total_rows} rows processed
                {activeImport.failed_rows.length > 0 && (
                  <span className="text-destructive">
                    {" "}({activeImport.failed_rows.length} failed)
                  </span>
                )}
              </div>
              {activeImport.failed_rows.length > 0 && (
                <div className="mt-2 space-y-1">
                  {activeImport.failed_rows.slice(0, 5).map((err, i) => (
                    <p key={i} className="text-xs text-destructive">
                      Row {err.row}: {err.reason}
                    </p>
                  ))}
                  {activeImport.failed_rows.length > 5 && (
                    <p className="text-xs text-muted-foreground">
                      ...and {activeImport.failed_rows.length - 5} more errors
                    </p>
                  )}
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Import History</h2>
        {!imports || imports.length === 0 ? (
          <EmptyState
            icon={<FileText className="h-8 w-8 text-muted-foreground" />}
            title="No imports yet"
            description="Upload your first CSV file to get started."
          />
        ) : (
          <div className="space-y-3">
            {imports.map((imp) => (
              <Card key={imp.id}>
                <CardContent className="flex items-center justify-between py-4">
                  <div className="flex items-center gap-3">
                    {getStatusIcon(imp.status)}
                    <div>
                      <p className="text-sm font-medium">
                        Import #{imp.id.slice(0, 8)}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(imp.created_at).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right text-sm">
                      <span className={imp.failed_rows.length > 0 ? "text-destructive" : ""}>
                        {imp.success_rows}/{imp.total_rows} rows
                      </span>
                    </div>
                    <Badge variant={imp.status === "completed" ? "default" : "secondary"}>
                      {imp.status}
                    </Badge>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
