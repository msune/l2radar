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

function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${Number.isInteger(val) ? val : val.toFixed(1)} ${units[i]}`
}

function InterfaceInfo({ name, timestamp, info }) {
  const [now, setNow] = useState(Date.now())
  const [statsOpen, setStatsOpen] = useState(false)

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

  const stats = info.stats

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
      {stats && (
        <div className="col-span-2 sm:col-span-5 border-t border-radar-700 pt-2 -mx-4 px-4">
          <button
            type="button"
            className="flex items-center gap-1 text-radar-500 text-xs hover:text-radar-300 cursor-pointer"
            onClick={() => setStatsOpen(!statsOpen)}
          >
            <span className={`inline-block transition-transform ${statsOpen ? 'rotate-90' : ''}`}>&#9654;</span>
            Interface Stats
          </button>
          {statsOpen && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-2">
              <div>
                <span className="text-radar-500 text-xs">TX Bytes</span>
                <div className="text-radar-100 font-mono">{formatBytes(stats.tx_bytes)}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">RX Bytes</span>
                <div className="text-radar-100 font-mono">{formatBytes(stats.rx_bytes)}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">TX Packets</span>
                <div className="text-radar-100 font-mono">{stats.tx_packets.toLocaleString()}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">RX Packets</span>
                <div className="text-radar-100 font-mono">{stats.rx_packets.toLocaleString()}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">TX Errors</span>
                <div className="text-radar-100 font-mono">{stats.tx_errors.toLocaleString()}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">RX Errors</span>
                <div className="text-radar-100 font-mono">{stats.rx_errors.toLocaleString()}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">TX Dropped</span>
                <div className="text-radar-100 font-mono">{stats.tx_dropped.toLocaleString()}</div>
              </div>
              <div>
                <span className="text-radar-500 text-xs">RX Dropped</span>
                <div className="text-radar-100 font-mono">{stats.rx_dropped.toLocaleString()}</div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default InterfaceInfo
