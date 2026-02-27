import { useState, useEffect } from 'react'

const defaults = { privacyMode: false }

export function useConfig() {
  const [config, setConfig] = useState(defaults)

  useEffect(() => {
    fetch('/config.json')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data) setConfig({ ...defaults, ...data })
      })
      .catch(() => {})
  }, [])

  return config
}
