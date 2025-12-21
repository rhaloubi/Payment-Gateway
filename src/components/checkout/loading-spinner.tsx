import { Card, CardContent } from "~/components/ui/card";
import { Loader2 } from "lucide-react";

export function LoadingSpinner() {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Loader2 className="h-12 w-12 animate-spin text-primary mb-4" />
        <p className="text-lg font-medium">Loading payment details...</p>
        <p className="text-sm text-muted-foreground mt-2">Please wait</p>
      </CardContent>
    </Card>
  );
}