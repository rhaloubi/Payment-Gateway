import type {
  PaymentIntentResponse,
  ConfirmPaymentResponse,
  CardData,
} from "~/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

/**
 * Fetch payment intent details (browser-safe, no auth)
 */
export async function getPaymentIntent(
  intentId: string
): Promise<PaymentIntentResponse> {
  try {
    const response = await fetch(
      `${API_URL}/api/public/payment-intents/${intentId}`,
      {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        cache: "no-store",
      }
    );

    if (!response.ok) {
      const error = (await response.json().catch(() => ({
        error: "Failed to fetch payment intent",
      }))) as { error?: string };
      throw new Error(error.error ?? "Failed to fetch payment intent");
    }

    return (await response.json()) as PaymentIntentResponse;
  } catch (error) {
    console.error("Error fetching payment intent:", error);
    throw error;
  }
}

/**
 * Confirm payment intent (process payment)
 */
export async function confirmPaymentIntent(
  intentId: string,
  clientSecret: string,
  cardData: CardData,
  customerEmail?: string
): Promise<ConfirmPaymentResponse> {
  try {
    const response = await fetch(
      `${API_URL}/api/public/payment-intents/${intentId}/confirm`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Client-Secret": clientSecret,
        },
        body: JSON.stringify({
          card: {
            number: cardData.number.replace(/\s/g, ""),
            cardholder_name: cardData.cardholder_name,
            exp_month: cardData.exp_month,
            exp_year: cardData.exp_year,
            cvv: cardData.cvv,
          },
          customer_email: customerEmail,
        }),
      }
    );

    if (!response.ok) {
      const error = (await response.json().catch(() => ({
        error: "Payment failed",
      }))) as { error?: string };
      throw new Error(error.error ?? "Payment failed");
    }

    return (await response.json()) as ConfirmPaymentResponse;
  } catch (error) {
    console.error("Error confirming payment:", error);
    throw error;
  }
}

/**
 * Validate client secret format
 */
export function isValidClientSecret(clientSecret: string): boolean {
  return clientSecret.startsWith("pi_secret_") && clientSecret.length > 20;
}