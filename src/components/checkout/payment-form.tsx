"use client";

import { useState, type FormEvent } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Button } from "~/components/ui/button";
import { Alert, AlertDescription } from "~/components/ui/alert";
import {
  validateCardNumber,
  validateExpiryDate,
  validateCVV,
  detectCardBrand,
  formatCardNumber,
  formatExpiryDate,
  parseExpiryDate,
} from "~/lib/utils/card-validation";
import { CreditCard, Lock, AlertCircle } from "lucide-react";
import type { CardData, CardFormData } from "~/types";
import Image from "next/image";

export const VisaIcon = () => (
  <Image src="/images/visa.svg" alt="Visa" width={20} height={20} />
);

export const MastercardIcon = () => (
  <Image src="/images/mastercard.svg" alt="Mastercard" width={20} height={20} />
);

const CardDetection = ({ cardNumber }: { cardNumber: string }) => {
  const brand = detectCardBrand(cardNumber);
  
  if (brand === 'visa') return <VisaIcon />;
  if (brand === 'mastercard') return <MastercardIcon />;
  
  return <CreditCard className="h-6 w-6 text-muted-foreground" />;
};

interface PaymentFormProps {
  onSubmit: (cardData: CardData) => Promise<void>;
  isSubmitting: boolean;
  error: string | null;
}

export function PaymentForm({ onSubmit, isSubmitting, error }: PaymentFormProps) {
  const [formData, setFormData] = useState<CardFormData>({
    cardNumber: "",
    cardholderName: "",
    expiryDate: "",
    cvv: "",
  });

  const [errors, setErrors] = useState<Partial<CardFormData>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof CardFormData, boolean>>>({});

  const cardBrand = detectCardBrand(formData.cardNumber);

  // Validate field
  const validateField = (name: keyof CardFormData, value: string): string | null => {
    switch (name) {
      case "cardNumber":
        if (!value) return "Card number is required";
        if (!validateCardNumber(value)) return "Invalid card number";
        return null;

      case "cardholderName":
        if (!value) return "Cardholder name is required";
        if (value.length < 3) return "Name too short";
        return null;

      case "expiryDate":
        if (!value) return "Expiry date is required";
        if (!validateExpiryDate(value)) return "Invalid or expired date";
        return null;

      case "cvv":
        if (!value) return "CVV is required";
        if (!validateCVV(value, cardBrand)) {
          return cardBrand === "amex" ? "CVV must be 4 digits" : "CVV must be 3 digits";
        }
        return null;

      default:
        return null;
    }
  };

  // Handle input change
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    let formattedValue = value;

    // Format as user types
    if (name === "cardNumber") {
      formattedValue = formatCardNumber(value);
    } else if (name === "expiryDate") {
      formattedValue = formatExpiryDate(value);
    } else if (name === "cvv") {
      formattedValue = value.replace(/\D/g, "").slice(0, cardBrand === "amex" ? 4 : 3);
    }

    setFormData((prev) => ({ ...prev, [name]: formattedValue }));

    // Clear error when user starts typing
    if (errors[name as keyof CardFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  // Handle blur (validation)
  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setTouched((prev) => ({ ...prev, [name]: true }));

    const error = validateField(name as keyof CardFormData, value);
    if (error) {
      setErrors((prev) => ({ ...prev, [name]: error }));
    }
  };

  // Handle form submission
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    // Mark all fields as touched
    setTouched({
      cardNumber: true,
      cardholderName: true,
      expiryDate: true,
      cvv: true,
    });

    // Validate all fields
    const newErrors: Partial<CardFormData> = {};
    (Object.keys(formData) as Array<keyof CardFormData>).forEach((key) => {
      const error = validateField(key, formData[key]);
      if (error) {
        newErrors[key] = error;
      }
    });

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    // Parse expiry date
    const { month, year } = parseExpiryDate(formData.expiryDate);

    // Build card data
    const cardData: CardData = {
      number: formData.cardNumber.replace(/\s/g, ""),
      cardholder_name: formData.cardholderName,
      exp_month: month,
      exp_year: year,
      cvv: formData.cvv,
    };

    await onSubmit(cardData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Error Alert */}
      {error && (
        <Alert variant="destructive" className="animate-in slide-in-from-top-2 fade-in duration-300">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Payment Information */}
      <Card className="border-border/40 bg-card/50 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="text-xl">Payment Information</CardTitle>
          <CardDescription>Your payment details are secure and encrypted</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="cardNumber">Card Number</Label>
            <div className="relative">
              <Input
                id="cardNumber"
                name="cardNumber"
                placeholder="1234 5678 9012 3456"
                value={formData.cardNumber}
                onChange={handleChange}
                onBlur={handleBlur}
                required
                className={`bg-background/50 border-border/60 font-mono text-base tracking-wide pr-12 ${touched.cardNumber && errors.cardNumber ? "border-red-500 focus-visible:ring-red-500" : ""}`}
              />
              <div className="absolute right-3 top-1/2 -translate-y-1/2">
                <CardDetection cardNumber={formData.cardNumber} />
              </div>
            </div>
            {touched.cardNumber && errors.cardNumber && (
              <p className="text-sm text-red-500 animate-in slide-in-from-left-1 fade-in">{errors.cardNumber}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="cardholderName">Cardholder Name</Label>
             <Input
                id="cardholderName"
                name="cardholderName"
                placeholder="JOHN DOE"
                value={formData.cardholderName}
                onChange={handleChange}
                onBlur={handleBlur}
                required
                className={`bg-background/50 border-border/60 font-mono text-base ${touched.cardholderName && errors.cardholderName ? "border-red-500 focus-visible:ring-red-500" : ""}`}
              />
               {touched.cardholderName && errors.cardholderName && (
                  <p className="text-sm text-red-500 animate-in slide-in-from-left-1 fade-in">{errors.cardholderName}</p>
               )}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="expiry">Expiry Date</Label>
              <Input
                id="expiry"
                name="expiryDate"
                placeholder="MM/YY"
                value={formData.expiryDate}
                onChange={handleChange}
                onBlur={handleBlur}
                required
                className={`bg-background/50 border-border/60 font-mono text-base ${touched.expiryDate && errors.expiryDate ? "border-red-500 focus-visible:ring-red-500" : ""}`}
              />
              {touched.expiryDate && errors.expiryDate && (
                <p className="text-sm text-red-500 animate-in slide-in-from-left-1 fade-in">{errors.expiryDate}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="cvc">CVC</Label>
              <Input
                id="cvc"
                name="cvv"
                placeholder={cardBrand === "amex" ? "1234" : "123"}
                value={formData.cvv}
                onChange={handleChange}
                onBlur={handleBlur}
                required
                className={`bg-background/50 border-border/60 font-mono text-base ${touched.cvv && errors.cvv ? "border-red-500 focus-visible:ring-red-500" : ""}`}
              />
              {touched.cvv && errors.cvv && (
                <p className="text-sm text-red-500 animate-in slide-in-from-left-1 fade-in">{errors.cvv}</p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Button 
            type="submit" 
            className="w-full relative overflow-hidden transition-all duration-300" 
            size="lg" 
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <div className="flex items-center justify-center gap-2 animate-in fade-in duration-300">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                <span>Processing Payment...</span>
              </div>
            ) : (
              <div className="flex items-center justify-center gap-2 animate-in fade-in duration-300">
                <Lock className="h-4 w-4" />
                <span>Pay Now</span>
              </div>
            )}
          </Button>
    </form>
  );
}
