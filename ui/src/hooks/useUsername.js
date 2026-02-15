import { useState, useEffect } from 'react'

export function useUsername() {
  const [username, setUsername] = useState('')

  useEffect(() => {
    fetch('/api/whoami')
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data?.username) setUsername(data.username)
      })
      .catch(() => {})
  }, [])

  return username
}
