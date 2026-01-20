"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { AlertCircle } from "lucide-react"
import { Button } from "~/components/ui/button"

export default function NotFound() {
  const [isAnimating, setIsAnimating] = useState(false)

  useEffect(() => {
    setIsAnimating(true)
  }, [])

  return (
    <main className="min-h-screen bg-background flex items-center justify-center p-4 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-20 right-20 w-72 h-72 bg-primary/5 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-20 left-20 w-96 h-96 bg-accent/5 rounded-full blur-3xl animate-pulse" />
      </div>

      {/* Content */}
      <div className="relative z-10 text-center max-w-md">
        {/* Icon with scale animation */}
        <div
          className={`flex justify-center mb-8 transition-all duration-700 ${isAnimating ? "scale-100 opacity-100" : "scale-0 opacity-0"}`}
        >
          <div className="relative">
            <AlertCircle className="w-24 h-24 text-primary" strokeWidth={1.5} />
            <div className="absolute inset-0 bg-primary/20 rounded-full blur-xl animate-pulse" />
          </div>
        </div>

        {/* 404 Text */}
        <div
          className={`mb-4 transition-all duration-700 delay-100 ${isAnimating ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`}
        >
          <h1 className="text-6xl md:text-7xl font-bold text-foreground mb-2">
            <span className="bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">404</span>
          </h1>
        </div>

        {/* Error message */}
        <div
          className={`mb-8 transition-all duration-700 delay-200 ${isAnimating ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`}
        >
          <h2 className="text-2xl font-semibold text-foreground mb-3">Page not found</h2>
          <p className="text-muted-foreground font-mono text-sm leading-relaxed">
            The page you&apos;re looking for doesn&apos;t exist or has been moved to another location.
          </p>
        </div>

        {/* Button */}
        <div
          className={`transition-all duration-700 delay-300 ${isAnimating ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`}
        >
          <Link href="/">
            <Button className="bg-primary hover:bg-primary/90 text-primary-foreground font-mono font-semibold px-8 py-6 text-base rounded-lg transition-all hover:scale-105 active:scale-95">
              Go back home
            </Button>
          </Link>
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
