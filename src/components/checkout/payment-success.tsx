"use client"

import { useEffect, useState } from "react"
import { CheckCircle2 } from "lucide-react"
import { useRouter } from "next/navigation"

interface PaymentSuccessProps {
  cardType?: string
  paymentIntentId: string
}

export function PaymentSuccess({ cardType, paymentIntentId }: PaymentSuccessProps) {
  const [showConfetti, setShowConfetti] = useState(false)
  const [particles, setParticles] = useState<Array<{ id: number; x: number; y: number }>>([])
  const [showCheck, setShowCheck] = useState(false)
  const router = useRouter()

  useEffect(() => {
    const playSound = () => {
      // Audio context logic (wrapped in try/catch for safety)
      try {
        const AudioContext = window.AudioContext || (window as any).webkitAudioContext
        if (!AudioContext) return;
        
        const audioContext = new AudioContext()
        const oscillator = audioContext.createOscillator()
        const gainNode = audioContext.createGain()

        oscillator.connect(gainNode)
        gainNode.connect(audioContext.destination)

        oscillator.frequency.value = 800
        gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5)

        oscillator.start(audioContext.currentTime)
        oscillator.stop(audioContext.currentTime + 0.5)
      } catch (e) {
        console.error("Audio play failed", e)
      }
    }

    playSound()
    setShowCheck(true)

    const timer = setTimeout(() => {
      router.push(`/success?payment_intent=${paymentIntentId}`)
    }, 2500)

    return () => clearTimeout(timer)
  }, [router, paymentIntentId])

  useEffect(() => {
    setShowConfetti(true)
    const newParticles = Array.from({ length: 30 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
    }))
    setParticles(newParticles)
  }, [])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      {/* Animated background particles */}
      {showConfetti &&
        particles.map((particle) => (
          <div
            key={particle.id}
            className="absolute w-2 h-2 bg-indigo-500 rounded-full animate-pulse"
            style={{
              left: `${particle.x}%`,
              top: `${particle.y}%`,
              animation: `float ${2 + Math.random() * 2}s ease-in-out infinite`,
              opacity: Math.random() * 0.5 + 0.3,
            }}
          />
        ))}

      <style jsx global>{`
        @keyframes float {
          0%, 100% { transform: translateY(0px) translateX(0px); }
          50% { transform: translateY(-20px) translateX(10px); }
        }
        @keyframes scaleIn {
          0% { transform: scale(0); opacity: 0; }
          50% { transform: scale(1.15); }
          100% { transform: scale(1); opacity: 1; }
        }
        .animate-success-check {
          animation: scaleIn 0.7s cubic-bezier(0.34, 1.56, 0.64, 1);
        }
      `}</style>

      {showCheck && (
        <div className="animate-success-check flex flex-col items-center">
          <CheckCircle2 className="w-32 h-32 text-indigo-600 drop-shadow-lg mb-4" />
          <h1 className="text-3xl font-bold text-foreground">Payment Complete</h1>
        </div>
      )}
    </div>
  )
}
