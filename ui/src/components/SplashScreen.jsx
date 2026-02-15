import { useState, useEffect } from 'react'
import logoSplash from '../../../assets/img/logo_splash.png'

const SPLASH_DURATION = 3000
const FADE_DURATION = 500

function SplashScreen({ onDone }) {
  const [fading, setFading] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => setFading(true), SPLASH_DURATION)
    return () => clearTimeout(timer)
  }, [])

  useEffect(() => {
    if (!fading) return
    const timer = setTimeout(onDone, FADE_DURATION)
    return () => clearTimeout(timer)
  }, [fading, onDone])

  return (
    <div
      className="fixed inset-0 z-[9999] flex items-center justify-center bg-radar-950/90 backdrop-blur-sm cursor-pointer"
      style={{
        opacity: fading ? 0 : 1,
        transition: `opacity ${FADE_DURATION}ms ease-out`,
      }}
      onClick={() => setFading(true)}
    >
      <img
        src={logoSplash}
        alt="L2 Radar"
        className="max-w-[80vw] max-h-[80vh] w-auto h-auto object-contain"
      />
    </div>
  )
}

export default SplashScreen
