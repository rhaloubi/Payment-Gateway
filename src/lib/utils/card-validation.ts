import type { CardBrand } from "~/types";

/**
 * Luhn algorithm for card number validation
 */
export function validateCardNumber(cardNumber: string): boolean {
  const digits = cardNumber.replace(/\D/g, "");
  
  // Check for Visa (starts with 4) or Mastercard (starts with 51-55)
  // We strictly allow only these two brands
  if (!/^4/.test(digits) && !/^5[1-5]/.test(digits)) {
    return false;
  }

  if (digits.length < 13 || digits.length > 19) {
    return false;
  }

  let sum = 0;
  let isEven = false;

  for (let i = digits.length - 1; i >= 0; i--) {
    let digit = parseInt(digits[i]!, 10);

    if (isEven) {
      digit *= 2;
      if (digit > 9) {
        digit -= 9;
      }
    }

    sum += digit;
    isEven = !isEven;
  }

  return sum % 10 === 0;
}

/**
 * Detect card brand from card number
 */
export function detectCardBrand(cardNumber: string): CardBrand {
  const digits = cardNumber.replace(/\D/g, "");

  if (/^4/.test(digits)) return "visa";
  if (/^5[1-5]/.test(digits)) return "mastercard";
  return "unknown";
}

/**
 * Validate expiry date (MM/YY format)
 */
export function validateExpiryDate(expiryDate: string): boolean {
  const cleaned = expiryDate.replace(/\D/g, "");
  
  if (cleaned.length !== 4) {
    return false;
  }

  const month = parseInt(cleaned.substring(0, 2), 10);
  const year = parseInt(cleaned.substring(2, 4), 10);

  if (month < 1 || month > 12) {
    return false;
  }

  const now = new Date();
  const currentYear = now.getFullYear() % 100;
  const currentMonth = now.getMonth() + 1;

  if (year < currentYear) {
    return false;
  }

  if (year === currentYear && month < currentMonth) {
    return false;
  }

  return true;
}

/**
 * Validate CVV
 */
export function validateCVV(cvv: string, cardBrand: CardBrand): boolean {
  const digits = cvv.replace(/\D/g, "");
  
  if (cardBrand === "amex") {
    return digits.length === 4;
  }
  
  return digits.length === 3;
}

/**
 * Format card number with spaces (4 digits groups)
 */
export function formatCardNumber(cardNumber: string): string {
  const digits = cardNumber.replace(/\D/g, "");
  const groups = digits.match(/.{1,4}/g) || [];
  return groups.join(" ");
}

/**
 * Format expiry date (MM/YY)
 */
export function formatExpiryDate(value: string): string {
  const digits = value.replace(/\D/g, "");
  
  if (digits.length >= 2) {
    return `${digits.substring(0, 2)}/${digits.substring(2, 4)}`;
  }
  
  return digits;
}

/**
 * Parse expiry date to month and year
 */
export function parseExpiryDate(expiryDate: string): { month: number; year: number } {
  const cleaned = expiryDate.replace(/\D/g, "");
  const month = parseInt(cleaned.substring(0, 2), 10);
  const year = parseInt(`20${cleaned.substring(2, 4)}`, 10);
  
  return { month, year };
}

/**
 * Mask card number (show only last 4 digits)
 */
export function maskCardNumber(cardNumber: string): string {
  const digits = cardNumber.replace(/\D/g, "");
  const last4 = digits.slice(-4);
  return `•••• •••• •••• ${last4}`;
}