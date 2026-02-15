function formatTimestamp(ts) {
  try {
    const d = new Date(ts)
    const h = String(d.getUTCHours()).padStart(2, '0')
    const m = String(d.getUTCMinutes()).padStart(2, '0')
    const s = String(d.getUTCSeconds()).padStart(2, '0')
    return `${h}:${m}:${s} UTC`
  } catch {
    return ts
  }
}

function InterfaceInfo({ name, timestamp, info }) {
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
        <div className="text-radar-100">{timestamp ? formatTimestamp(timestamp) : '—'}</div>
      </div>
    </div>
  )
}

export default InterfaceInfo
