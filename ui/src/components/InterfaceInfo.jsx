import { useEffect, useState, useRef } from 'react'
import { formatAgo } from '../lib/timeago'

function InterfaceInfo({ name, timestamp, info }) {
  const [, setTick] = useState(0)
  const prevTimestamp = useRef(timestamp)
  const [highlight, setHighlight] = useState(false)

  useEffect(() => {
    if (!timestamp) return
    const id = setInterval(() => setTick((t) => t + 1), 5000)
    return () => clearInterval(id)
  }, [timestamp])

  useEffect(() => {
    if (prevTimestamp.current && timestamp && timestamp !== prevTimestamp.current) {
      setHighlight(true)
      const id = setTimeout(() => setHighlight(false), 5000)
      prevTimestamp.current = timestamp
      return () => clearTimeout(id)
    }
    prevTimestamp.current = timestamp
  }, [timestamp])

  if (!info) return null

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
      <div className={highlight ? 'highlight-fresh rounded' : ''}>
        <span className="text-radar-500 text-xs">Last update</span>
        <div className="text-radar-100" title={timestamp || ''}>
          {timestamp ? formatAgo(timestamp) : '—'}
        </div>
      </div>
    </div>
  )
}

export default InterfaceInfo
