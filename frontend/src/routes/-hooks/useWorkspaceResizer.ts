import { useEffect, useRef, useState } from 'react'

export interface UseWorkspaceResizerOptions {
  onRatioChange: (ratio: number) => void
  minRatio?: number
  maxRatio?: number
}

export function useWorkspaceResizer({
  onRatioChange,
  minRatio = 0.2,
  maxRatio = 0.8,
}: UseWorkspaceResizerOptions) {
  const [isDragging, setIsDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement | null>(null)

  const handleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleResetRatio = () => {
    onRatioChange(0.5)
  }

  useEffect(() => {
    if (!isDragging) return

    const handleMouseMove = (e: MouseEvent) => {
      if (!containerRef.current) return
      const rect = containerRef.current.getBoundingClientRect()
      const newRatio = (e.clientX - rect.left) / rect.width
      const clamped = Math.max(minRatio, Math.min(maxRatio, newRatio))
      onRatioChange(clamped)
    }

    const handleMouseUp = () => {
      setIsDragging(false)
    }

    window.addEventListener('mousemove', handleMouseMove)
    window.addEventListener('mouseup', handleMouseUp)

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, minRatio, maxRatio, onRatioChange])

  return {
    containerRef,
    isDragging,
    handleMouseDown,
    handleResetRatio,
  }
}
