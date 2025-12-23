import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "~/components/ui/card";
import { formatCurrency } from "~/lib/utils/format";
import { Badge } from "~/components/ui/badge";

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
    <div className="sticky top-8 space-y-4">
      <Card className="border-border/40 bg-card/50 backdrop-blur-sm">
        <CardHeader>
          <CardTitle>Order Summary</CardTitle>
          <CardDescription>Premium Digital Bundle</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">merchant Name</span>
              <span>{merchantName}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Shipping</span>
              <span>{formatCurrency(0.00, currency)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Tax</span>
              <span>{formatCurrency(0.00, currency)}</span>
            </div>
            <div className="border-t border-border/40 pt-3 flex justify-between font-semibold">
              <span>Total</span>
              <span className="text-lg">{formatCurrency(amount, currency)}</span>
            </div>
          </div>

          <div className="bg-primary/10 border border-primary/20 rounded-lg p-3 space-y-2">
            <div className="flex items-center justify-between">
              <Badge variant="secondary" className="bg-green-500/20 text-green-700 border-green-500/50">
                ✓ Secure
              </Badge>
              <Badge variant="secondary" className="bg-blue-500/20 text-blue-700 border-blue-500/50">
                SSL Encrypted
              </Badge>
            </div>
          </div>

          <div className="space-y-2 text-xs text-muted-foreground">
            <p>✓ 30-day money-back guarantee</p>
            <p>✓ 24/7 customer support</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
