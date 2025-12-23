"use client"

import { useEffect, useState } from "react"
import { XCircle } from "lucide-react"
import { useRouter } from "next/navigation"

interface PaymentCancelProps {
  onRetry: () => void
}

export function PaymentCancel({ onRetry }: PaymentCancelProps) {
  const [showParticles, setShowParticles] = useState(false)
  const [particles, setParticles] = useState<Array<{ id: number; x: number; y: number }>>([])
  const [showIcon, setShowIcon] = useState(false)

  useEffect(() => {
    const playSound = () => {
      try {
        const AudioContext = window.AudioContext || (window as any).webkitAudioContext
        if (!AudioContext) return;

        const audioContext = new AudioContext()
        const oscillator = audioContext.createOscillator()
        const gainNode = audioContext.createGain()

        oscillator.connect(gainNode)
        gainNode.connect(audioContext.destination)

        oscillator.frequency.value = 150 // Lower frequency for error
        oscillator.type = 'sawtooth'
        gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5)

        oscillator.start(audioContext.currentTime)
        oscillator.stop(audioContext.currentTime + 0.5)
      } catch (e) {
        console.error("Audio play failed", e)
      }
    }

    playSound()
    setShowIcon(true)

    // Auto-retry or close after a delay? 
    // Usually cancel requires user action to try again. 
    // We'll leave it to the user to click or just show it briefly.
    // For this animation, we'll just show it and let the user click "Try Again" which will be rendered by the parent or we can provide a button here.
    // But the request was for "animation before redirection" (for success). For cancel, maybe just an overlay.
    // We'll provide a timer to close it if needed, but better to let user dismiss.
    // However, to match the success flow, maybe we just show it for 2.5s then return to form?
    const timer = setTimeout(() => {
      onRetry()
    }, 2500)
    
    return () => clearTimeout(timer)
  }, [onRetry])

  useEffect(() => {
    setShowParticles(true)
    const newParticles = Array.from({ length: 30 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
    }))
    setParticles(newParticles)
  }, [])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      {/* Animated background particles - red/orange for cancel */}
      {showParticles &&
        particles.map((particle) => (
          <div
            key={particle.id}
            className="absolute w-2 h-2 bg-red-500 rounded-full animate-pulse"
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
        .animate-cancel-check {
          animation: scaleIn 0.7s cubic-bezier(0.34, 1.56, 0.64, 1);
        }
      `}</style>

      {showIcon && (
        <div className="animate-cancel-check flex flex-col items-center">
          <XCircle className="w-32 h-32 text-red-600 drop-shadow-lg mb-4" />
          <h1 className="text-3xl font-bold text-foreground">Payment Cancelled</h1>
        </div>
      )}
    </div>
  )
}
