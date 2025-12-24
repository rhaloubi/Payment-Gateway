"use client";

import { useEffect, useState, use } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { PaymentForm } from "~/components/checkout/payment-form";
import { PaymentSummary } from "~/components/checkout/payment-summary";
import { LoadingSpinner } from "~/components/checkout/loading-spinner";
import { PaymentSuccess } from "~/components/checkout/payment-success";
import { Alert, AlertDescription, AlertTitle } from "~/components/ui/alert";
import { getPaymentIntent, confirmPaymentIntent, isValidClientSecret } from "~/lib/api/payment-intent";
import type { PaymentIntent, CardData } from "~/types";
import { AlertCircle, CheckCircle2 } from "lucide-react";

export default function CheckoutPage({ params }: { params: Promise<{ id: string }> }) {
  const { id: intentId } = use(params);
  
  const router = useRouter();
  const searchParams = useSearchParams();
  const clientSecret = searchParams.get("client_secret");

  const [intent, setIntent] = useState<PaymentIntent | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showSuccessAnimation, setShowSuccessAnimation] = useState(false);

  // Load payment intent
  useEffect(() => {
    async function loadIntent() {
      try {
        // Validate client secret
        if (!clientSecret || !isValidClientSecret(clientSecret)) {
          setError("Invalid payment session. Please try again.");
          setLoading(false);
          return;
        }

        // Fetch payment intent
        const response = await getPaymentIntent(intentId);

        if (!response.success || !response.data) {
          throw new Error(response.error || "Failed to load payment details");
        }

        const intentData = response.data;

        // Check status
        if (intentData.status === "expired") {
          setError("This payment session has expired. Please create a new payment.");
           window.location.href = intentData.cancel_url || '';
          setLoading(false);
          return;
        }

        if (intentData.status === "authorized" || intentData.status === "captured") {
          setError(`Payment already ${intentData.status}. Redirecting...`);
          setTimeout(() => {
            window.location.href = intentData.success_url || '';
          }, 1500);
          return;
        }

        if (intentData.status === "failed" || intentData.status === "canceled") {
          setError(`This payment has ${intentData.status}. Please create a new payment.`);
          window.location.href = intentData.cancel_url || '';
          setLoading(false);
          return;
        }

        setIntent(intentData);
        setLoading(false);
      } catch (err) {
        console.error("Failed to load payment intent:", err);
        setError(err instanceof Error ? err.message : "Failed to load payment details");
        setLoading(false);
      }
    }

    loadIntent();
  }, [intentId, clientSecret, router, ]);

  // Handle payment submission
  const handlePayment = async (cardData: CardData) => {
    if (!clientSecret || !intent) return;

    setSubmitting(true);
    setError(null);

    try {
      const response = await confirmPaymentIntent(
        intentId,
        clientSecret,
        cardData
      );

      if (!response.success || !response.data) {
        throw new Error(response.error || `Payment failed`);
      }

      const payment = response.data;

      if (payment.status === "authorized" || payment.status === "captured") {
        // Success - Show animation
        setShowSuccessAnimation(true);
      } else {
        throw new Error("Payment was declined");
      }
    } catch (err) {
      console.error("Payment failed:", err);
      setError(err instanceof Error ? err.message : "Payment failed. Please try again.");
      setSubmitting(false);
    }
  };

  // Loading state
  if (loading) {
    return (
      <div className="min-h-screen bg-background py-12">
        <div className="container max-w-xl mx-auto px-4">
          <LoadingSpinner />
        </div>
      </div>
    );
  }

  // Error state
  if (error && !intent) {
    return (
      <div className="min-h-screen bg-background py-12">
        <div className="container max-w-xl mx-auto px-4">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  // Already processed
  if (intent && intent.status !== "awaiting_payment_method") {
    return (
      <div className="min-h-screen bg-background py-12">
        <div className="container max-w-xl mx-auto px-4">
          <Alert variant="default">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <AlertTitle>Payment Processed</AlertTitle>
            <AlertDescription>
              This payment has already been {intent.status}. Redirecting...
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  // Checkout form
  return (
    <div className="min-h-screen bg-background py-12 font-sans">
      {showSuccessAnimation && <PaymentSuccess successURL={intent?.success_url || ''} />}
      
      <div className="container mx-auto px-4 lg:px-8 max-w-6xl">
        <div className="text-center mb-12">
          <h1 className="text-3xl font-bold mb-2 tracking-tight">Secure Checkout</h1>
          <p className="text-muted-foreground">
            Complete your payment securely
          </p>
        </div>

        <div className="grid gap-8 lg:grid-cols-12 items-start">
          {/* Main Form Area */}
          <div className="lg:col-span-7 space-y-8">
            <PaymentForm
              onSubmit={handlePayment}
              isSubmitting={submitting}
              error={error}
            />
            
            <div className="text-center text-sm text-muted-foreground">
              <p>
                By completing this payment, you agree to our{" "}
                <a href="#" className="underline hover:text-primary transition-colors">
                  Terms of Service
                </a>
              </p>
            </div>
          </div>

          {/* Sidebar Summary */}
          <div className="lg:col-span-5">
             {intent && (
              <PaymentSummary
                amount={intent.amount}
                currency={intent.currency}
                
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
