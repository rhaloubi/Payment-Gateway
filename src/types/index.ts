// Payment Intent Types
export type PaymentIntentStatus =
  | "created"
  | "awaiting_payment_method"
  | "authorized"
  | "captured"
  | "failed"
  | "canceled"
  | "expired";

export interface PaymentIntent {
  success_url: string;
  cancel_url:string;
  id: string;
  status: PaymentIntentStatus;
  amount: number;
  currency: string;
  expires_at: string;
  created_at: string;
}

export interface PaymentIntentResponse {
  success: boolean;
  data: PaymentIntent;
  error?: string;
}

// Card Types
export interface CardData {
  number: string;
  cardholder_name: string;
  exp_month: number;
  exp_year: number;
  cvv: string;
}

export interface CardFormData {
  cardNumber: string;
  cardholderName: string;
  expiryDate: string; // MM/YY format
  cvv: string;
}

export type CardBrand = "visa" | "mastercard" | "amex" | "discover" | "unknown";

// Payment Response
export interface PaymentResponse {
  id: string;
  status: "authorized" | "captured" | "failed" | "canceled" | "voided" | "refunded";
  amount: number;
  currency: string;
  card_brand: string;
  card_last4: string;
  auth_code?: string;
  fraud_score?: number;
  transaction_id?: string;
}

export interface ConfirmPaymentResponse {
  success: boolean;
  data?: PaymentResponse;
  error?: string;
}

// API Error
export interface ApiError {
  success: false;
  error: string;
  code?: string;
}

// Form Validation
export interface ValidationError {
  field: string;
  message: string;
}

// Checkout State
export interface CheckoutState {
  loading: boolean;
  submitting: boolean;
  error: string | null;
  paymentIntent: PaymentIntent | null;
}