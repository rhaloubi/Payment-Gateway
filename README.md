# Payment Checkout Application

A modern, secure payment checkout application built with Next.js 15, React 19, and TypeScript. This application provides a hosted checkout experience for merchants using the Payment Gateway Microservices.

## Features

- **Secure Payment Processing**: PCI-compliant card handling with client-side validation
- **Responsive Design**: Mobile-first design with Tailwind CSS
- **Real-time Validation**: Instant card validation with proper error messages
- **Payment Intent Integration**: Seamless integration with Payment Intents API
- **Success & Error Handling**: Comprehensive payment status management
- **404 Page**: Custom animated 404 page with smooth transitions
- **Analytics**: Vercel Analytics integration for usage tracking

## Tech Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS with custom components
- **UI Components**: Radix UI primitives with custom styling
- **Icons**: Lucide React
- **Validation**: Zod for environment variables
- **Build Tool**: Bun (compatible with npm/yarn)

## Project Structure

```
payment-checkout/
├── public/                 # Static assets
│   ├── images/            # Payment method icons
│   ├── favicon.ico        # Favicon
│   └── icon.svg           # App icon
├── src/
│   ├── app/               # Next.js app directory
│   │   ├── checkout/      # Checkout pages
│   │   │   └── [id]/      # Dynamic checkout page
│   │   ├── layout.tsx     # Root layout
│   │   └── not-found.tsx  # Custom 404 page
│   ├── components/        # React components
│   │   ├── checkout/      # Checkout-specific components
│   │   └── ui/            # Reusable UI components
│   ├── lib/               # Utilities and API clients
│   │   ├── api/           # API clients
│   │   └── utils/         # Utility functions
│   ├── styles/            # Global styles
│   └── types/             # TypeScript type definitions
└── configuration files
```

## Key Components

### Layout (`src/app/layout.tsx`)
- Root layout with Google Fonts integration
- Vercel Analytics for tracking
- Global CSS styles with Tailwind

### Checkout Page (`src/app/checkout/[id]/page.tsx`)
- Dynamic route for payment intents
- Payment intent loading and validation
- Payment form handling with error states
- Success animation and redirect management
- Comprehensive status handling (created, authorized, captured, failed, canceled, expired)

### 404 Page (`src/app/not-found.tsx`)
- Animated 404 page with smooth transitions
- Gradient text effects and pulse animations
- Home navigation button

### Payment Components
- `PaymentForm`: Secure card input form with validation
- `PaymentSuccess`: Success animation and redirect component
- `PaymentSummary`: Order summary display
- `LoadingSpinner`: Animated loading component

## Environment Variables

Create a `.env` file based on `.env.example`:

```env
# Server Environment
NODE_ENV=development

# Client Environment
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_CHECKOUT_URL=http://localhost:3000
NEXT_PUBLIC_ENV=development
NEXT_PUBLIC_MERCHANT_NAME="Your Merchant Name"
NEXT_PUBLIC_SUPPORT_EMAIL=support@merchant.com
```

## Setup Instructions

### Prerequisites
- Node.js 18+ or Bun
- Payment API Service running locally

### Installation

1. **Install dependencies:**
   ```bash
   bun install
   # or
   npm install
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run the development server:**
   ```bash
   bun dev
   # or
   npm run dev
   ```

4. **Open your browser:**
   Navigate to [http://localhost:3000](http://localhost:3000)

### Building for Production

```bash
# Build the application
bun build
# or
npm run build

# Start production server
bun start
# or
npm run start
```

## Development Scripts

- `bun dev` - Start development server with turbo mode
- `bun build` - Build for production
- `bun start` - Start production server
- `bun lint` - Run ESLint
- `bun typecheck` - Run TypeScript compiler
- `bun format:check` - Check code formatting
- `bun format:write` - Format code

## API Integration

The checkout application integrates with the Payment Intents API:

### Required Endpoints
- `GET /payment-intents/:id` - Retrieve payment intent details
- `POST /payment-intents/:id/confirm` - Confirm payment with card details

### Authentication
- Uses `client_secret` query parameter for browser authentication
- Validates client secrets before processing payments

### Payment Flow
1. Merchant creates payment intent via server API
2. Customer is redirected to checkout page with `client_secret`
3. Checkout page validates client secret and loads intent details
4. Customer enters card details and submits payment
5. Payment is confirmed via API
6. Customer is redirected to success/cancel URL based on result

## Security Features

- **Client-Side Validation**: Card number, expiry date, and CVC validation
- **HTTPS Enforcement**: Security headers in production
- **Input Sanitization**: Proper handling of user input
- **Error Handling**: Comprehensive error messages without exposing sensitive data
- **Expiration Handling**: Automatic redirect for expired payment sessions

## Browser Support

- Chrome 88+
- Firefox 85+
- Safari 14+
- Edge 88+

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is proprietary software. All rights reserved.

## Support

For support, please contact:
- Email: redahaloubi8@gmail.com
- Documentation: [API Documentation](./../payment-api-service/README.md)