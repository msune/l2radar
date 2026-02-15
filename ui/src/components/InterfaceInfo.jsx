import { useEffect, useState } from 'react'
import { format } from 'timeago.js'

function InterfaceInfo({ name, timestamp, info }) {
  const [, setTick] = useState(0)

  useEffect(() => {
    if (!timestamp) return
    const id = setInterval(() => setTick((t) => t + 1), 1000)
    return () => clearInterval(id)
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
      <div>
        <span className="text-radar-500 text-xs">Last update</span>
        <div className="text-radar-100" title={timestamp || ''}>
          {timestamp ? format(timestamp) : '—'}
        </div>
      </div>
    </div>
  )
}

export default InterfaceInfo
