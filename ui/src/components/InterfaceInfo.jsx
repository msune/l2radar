import { useEffect, useState } from 'react'
import { formatAgo } from '../lib/timeago'

function parseDuration(s) {
  if (!s) return 0
  const match = s.match(/^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+(?:\.\d+)?)s)?$/)
  if (!match) return 0
  const h = parseInt(match[1] || '0', 10)
  const m = parseInt(match[2] || '0', 10)
  const sec = parseFloat(match[3] || '0')
  return (h * 3600 + m * 60 + sec) * 1000
}

function InterfaceInfo({ name, timestamp, info }) {
  const [now, setNow] = useState(Date.now())

  useEffect(() => {
    if (!timestamp) return
    const id = setInterval(() => setNow(Date.now()), 5000)
    return () => clearInterval(id)
  }, [timestamp])

  if (!info) return null

  const intervalMs = parseDuration(info.exportInterval)
  const elapsed = timestamp ? now - new Date(timestamp).getTime() : 0
  const overdue = intervalMs > 0 && elapsed > intervalMs * 2

  const intervalLabel = info.exportInterval
    ? `Export interval: ${info.exportInterval}`
    : ''
  const tooltip = [timestamp || '', intervalLabel].filter(Boolean).join(' | ')

  return (
    <div className="bg-radar-900 border border-radar-700 rounded px-4 py-3 mb-4 grid grid-cols-2 sm:grid-cols-5 gap-3 text-sm">
      <div>
        <span className="text-radar-500 text-xs">Interface</span>
        <div className="text-radar-100 font-semibold">{name || '—'}</div>
      </div>
      <div>
        <span className="text-radar-500 text-xs">MAC</span>
        <div className="text-radar-100 font-mono">{info.mac || '—'}</div>
      </div>
      <div>
        <span className="text-radar-500 text-xs">IPv4</span>
        <div className="text-radar-100 font-mono">
          {info.ipv4.length > 0
            ? info.ipv4.map((ip) => <div key={ip}>{ip}</div>)
            : '—'}
        </div>
      </div>
      <div>
        <span className="text-radar-500 text-xs">IPv6</span>
        <div className="text-radar-100 font-mono">
          {info.ipv6.length > 0
            ? info.ipv6.map((ip) => <div key={ip}>{ip}</div>)
            : '—'}
        </div>
      </div>
      <div>
        <span className="text-radar-500 text-xs">Last update</span>
        <div className={overdue ? 'text-red-400' : 'text-radar-100'} title={tooltip}>
          {timestamp ? formatAgo(timestamp) : '—'}
        </div>
      </div>
    </div>
  )
}

export default InterfaceInfo
