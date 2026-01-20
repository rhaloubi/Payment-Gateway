"use client"

import { useEffect, useState } from "react"
import { Store } from "lucide-react"

export default function Home() {
  const [isAnimating, setIsAnimating] = useState(false)

  useEffect(() => {
    setIsAnimating(true)
  }, [])

  return (
    <main className="min-h-screen bg-background flex items-center justify-center p-4 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-20 left-20 w-72 h-72 bg-primary/5 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-20 right-20 w-96 h-96 bg-accent/5 rounded-full blur-3xl animate-pulse" />
      </div>

      {/* Content */}
      <div className="relative z-10 text-center max-w-lg">
        {/* Icon with scale animation */}
        <div
          className={`flex justify-center mb-8 transition-all duration-700 ${isAnimating ? "scale-100 opacity-100" : "scale-0 opacity-0"}`}
        >
          <div className="relative">
            <Store className="w-24 h-24 text-primary" strokeWidth={1.5} />
            <div className="absolute inset-0 bg-primary/20 rounded-full blur-xl animate-pulse" />
          </div>
        </div>

        {/* Title */}
        <div
          className={`mb-4 transition-all duration-700 delay-100 ${isAnimating ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`}
        >
          <h1 className="text-4xl md:text-5xl font-bold text-foreground mb-2">
            Checkout Service
          </h1>
        </div>

        {/* Message */}
        <div
          className={`mb-8 transition-all duration-700 delay-200 ${isAnimating ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`}
        >
          <p className="text-muted-foreground font-mono text-base leading-relaxed max-w-md mx-auto">
            We don&apos;t have a main page here. This is a dedicated checkout service.
            <br />
            <span className="block mt-2 font-semibold text-foreground">
              Please go to the merchant site.
            </span>
          </p>
        </div>

        {/* Decorative elements */}
        <div className="mt-16 flex justify-center gap-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="w-2 h-2 bg-primary/60 rounded-full animate-bounce"
              style={{
                animationDelay: `${i * 0.2}s`,
              }}
            />
          ))}
        </div>
      </div>
    </main>
  )
}
