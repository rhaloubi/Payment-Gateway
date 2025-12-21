"use client";

import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { CheckCircle2 } from "lucide-react";
import Link from "next/link";

export default function SuccessPage() {
  const searchParams = useSearchParams();
  const paymentIntentId = searchParams.get("payment_intent");

  return (
    <div className="min-h-screen bg-gradient-to-b from-green-50 to-gray-50 dark:from-green-950 dark:to-gray-900 flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="rounded-full bg-green-100 dark:bg-green-900 p-3">
              <CheckCircle2 className="h-12 w-12 text-green-600 dark:text-green-400" />
            </div>
          </div>
          <CardTitle className="text-2xl">Payment Successful!</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-muted-foreground">
            Your payment has been processed successfully. You will receive a confirmation email shortly.
          </p>

          {paymentIntentId && (
            <div className="bg-muted p-4 rounded-lg">
              <p className="text-sm text-muted-foreground mb-1">Payment Reference</p>
              <p className="text-xs font-mono break-all">{paymentIntentId}</p>
            </div>
          )}

          <div className="flex flex-col gap-2 pt-4">
            <Button asChild className="w-full">
              <Link href="/">Return to Store</Link>
            </Button>
            <Button asChild variant="outline" className="w-full">
              <Link href={`mailto:${process.env.NEXT_PUBLIC_SUPPORT_EMAIL || "support@example.com"}`}>
                Contact Support
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}