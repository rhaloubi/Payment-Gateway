"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { PaymentForm } from "~/components/checkout/payment-form";
import { PaymentSummary } from "~/components/checkout/payment-summary";
import { LoadingSpinner } from "~/components/checkout/loading-spinner";
import { Alert, AlertDescription, AlertTitle } from "~/components/ui/alert";
import { getPaymentIntent, confirmPaymentIntent, isValidClientSecret } from "~/lib/api/payment-intent";
import type { PaymentIntent, CardData } from "~/types";
import { AlertCircle, CheckCircle2 } from "lucide-react";

export default function CheckoutPage({ params }: { params: { id: string } }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const clientSecret = searchParams.get("client_secret");

  const [intent, setIntent] = useState<PaymentIntent | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
        const response = await getPaymentIntent(params.id);

        if (!response.success || !response.data) {
          throw new Error(response.error || "Failed to load payment details");
        }

        const intentData = response.data;

        // Check status
        if (intentData.status === "expired") {
          setError("This payment session has expired. Please create a new payment.");
          setLoading(false);
          return;
        }

        if (intentData.status !== "awaiting_payment_method") {
          setError(`Payment already ${intentData.status}. Redirecting...`);
          setTimeout(() => {
            router.push(`/success?payment_intent=${params.id}`);
          }, 2000);
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
  }, [params.id, clientSecret, router]);

  // Handle payment submission
  const handlePayment = async (cardData: CardData) => {
    if (!clientSecret || !intent) return;

    setSubmitting(true);
    setError(null);

    try {
      const response = await confirmPaymentIntent(
        params.id,
        clientSecret,
        cardData
      );

      if (!response.success || !response.data) {
        throw new Error(response.error || "Payment failed");
      }

      const payment = response.data;

      if (payment.status === "authorized" || payment.status === "captured") {
        // Success - redirect
        router.push(`/success?payment_intent=${params.id}`);
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
      <div className="min-h-screen bg-gradient-to-b from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 py-12">
        <div className="container max-w-xl mx-auto px-4">
          <LoadingSpinner />
        </div>
      </div>
    );
  }

  // Error state
  if (error && !intent) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 py-12">
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
      <div className="min-h-screen bg-gradient-to-b from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 py-12">
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
    <div className="min-h-screen bg-gradient-to-b from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 py-12">
      <div className="container max-w-xl mx-auto px-4">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-2">Secure Checkout</h1>
          <p className="text-muted-foreground">
            Complete your payment securely
          </p>
        </div>

        {/* Payment Summary */}
        {intent && (
          <PaymentSummary
            amount={intent.amount}
            currency={intent.currency}
            merchantName={process.env.NEXT_PUBLIC_MERCHANT_NAME}
          />
        )}

        {/* Payment Form */}
        <PaymentForm
          onSubmit={handlePayment}
          isSubmitting={submitting}
          error={error}
        />

        {/* Footer */}
        <div className="mt-6 text-center text-sm text-muted-foreground">
          <p>
            By completing this payment, you agree to our{" "}
            <a href="#" className="underline">
              Terms of Service
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}