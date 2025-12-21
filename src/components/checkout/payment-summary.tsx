import { Card, CardContent } from "~/components/ui/card";
import { formatCurrency } from "~/lib/utils/format";
import { Lock } from "lucide-react";

interface PaymentSummaryProps {
  amount: number;
  currency: string;
  merchantName?: string;
}

export function PaymentSummary({
  amount,
  currency,
  merchantName = "Merchant",
}: PaymentSummaryProps) {
  return (
    <Card className="mb-6">
      <CardContent className="pt-6">
        <div className="space-y-4">
          {/* Merchant Name */}
          <div className="text-center">
            <p className="text-sm text-muted-foreground">Payment to</p>
            <h2 className="text-xl font-semibold">{merchantName}</h2>
          </div>

          {/* Amount */}
          <div className="text-center border-t pt-4">
            <p className="text-sm text-muted-foreground">Amount</p>
            <p className="text-3xl font-bold">{formatCurrency(amount, currency)}</p>
          </div>

          {/* Security Badge */}
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground pt-4 border-t">
            <Lock className="h-3 w-3" />
            <span>Secured payment</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}