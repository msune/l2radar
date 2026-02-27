import { useState, useEffect, useRef } from 'react'
import { formatAgo } from '../lib/timeago'
import { sortNeighbours } from '../lib/sorting'
import { lookupOUI } from '../lib/ouiLookup'
import { splitMacForDisplay, splitIPv6ForDisplay } from '../lib/macObfuscation'

const COLUMNS = [
  { key: 'interface', label: 'Interface' },
  { key: 'mac', label: 'MAC' },
  { key: 'ipv4', label: 'IPv4' },
  { key: 'ipv6', label: 'IPv6' },
  { key: 'firstSeen', label: 'First Seen' },
  { key: 'lastSeen', label: 'Last Seen' },
]

const STALE_MS = 5 * 60 * 1000

function SortIndicator({ active, dir }) {
  if (!active) return <span className="text-radar-600 ml-1">&#x2195;</span>
  return (
    <span className="text-accent-400 ml-1">{dir === 'asc' ? '▲' : '▼'}</span>
  )
}

function rowKey(n) {
  return `${n.interface}-${n.mac}`
}

function renderMasked(text, splitFn) {
  const { prefix, masked } = splitFn(text)
  if (!masked) return text
  return <>{prefix}<span className="text-radar-600">{masked}</span></>
}

function NeighbourTable({ neighbours, showInterface = true, privacyMode = false }) {
  const columns = showInterface
    ? COLUMNS
    : COLUMNS.filter((c) => c.key !== 'interface')
  const [sortKey, setSortKey] = useState('lastSeen')
  const [sortDir, setSortDir] = useState('desc')
  const [, setTick] = useState(0)
  const prevLastSeen = useRef({})
  const [freshKeys, setFreshKeys] = useState(new Set())

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 5000)
    return () => clearInterval(id)
  }, [])

  useEffect(() => {
    const prev = prevLastSeen.current
    const newFresh = new Set()
    const next = {}

    for (const n of neighbours) {
      const k = rowKey(n)
      next[k] = n.lastSeen
      if (!prev[k] || prev[k] !== n.lastSeen) {
        newFresh.add(k)
      }
    }

    prevLastSeen.current = next

    if (newFresh.size > 0) {
      setFreshKeys(newFresh)
      const id = setTimeout(() => setFreshKeys(new Set()), 5000)
      return () => clearTimeout(id)
    }
  }, [neighbours])

  const handleSort = (key) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(key)
      setSortDir(key === 'lastSeen' || key === 'firstSeen' ? 'desc' : 'asc')
    }
  }

  const sorted = sortNeighbours(neighbours, sortKey, sortDir)
  const now = Date.now()

  function isStale(n) {
    if (!n.lastSeen) return false
    return now - new Date(n.lastSeen).getTime() > STALE_MS
  }

  return (
    <>
      {/* Desktop table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-radar-700">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className="text-left px-2 py-2 text-radar-400 font-medium cursor-pointer select-none hover:text-accent-400"
                  onClick={() => handleSort(col.key)}
                >
                  {col.label}
                  <SortIndicator
                    active={sortKey === col.key}
                    dir={sortDir}
                  />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.map((n, i) => {
              const k = rowKey(n)
              const stale = isStale(n)
              const fresh = freshKeys.has(k)
              return (
                <tr
                  key={k}
                  className={`border-b border-radar-800 hover:bg-radar-900 ${
                    i % 2 === 0 ? 'bg-radar-950' : 'bg-radar-900/30'
                  } ${stale ? 'opacity-40' : ''} ${fresh ? 'highlight-fresh' : ''}`}
                >
                  {showInterface && (
                    <td className="px-2 py-1.5 text-radar-400">{n.interface}</td>
                  )}
                  <td className="px-2 py-1.5 font-mono text-accent-300">
                    {privacyMode ? renderMasked(n.mac, splitMacForDisplay) : n.mac}
                    {lookupOUI(n.mac) && (
                      <span className="text-radar-500 text-xs ml-1">
                        ({lookupOUI(n.mac)})
                      </span>
                    )}
                  </td>
                  <td className="px-2 py-1.5 font-mono">
                    {n.ipv4.join(', ') || '—'}
                  </td>
                  <td className="px-2 py-1.5 font-mono text-xs">
                    {privacyMode
                      ? (n.ipv6.length > 0
                        ? n.ipv6.map((ip, j) => (
                          <span key={j}>{j > 0 && ', '}{renderMasked(ip, splitIPv6ForDisplay)}</span>
                        ))
                        : '—')
                      : (n.ipv6.join(', ') || '—')}
                  </td>
                  <td className="px-2 py-1.5 text-radar-300 whitespace-nowrap" title={n.firstSeen}>
                    {n.firstSeen ? formatAgo(n.firstSeen) : ''}
                  </td>
                  <td className="px-2 py-1.5 text-radar-300 whitespace-nowrap" title={n.lastSeen}>
                    {n.lastSeen ? formatAgo(n.lastSeen) : ''}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {sorted.length === 0 && (
          <p className="text-center text-radar-500 py-8">
            No neighbours found
          </p>
        )}
      </div>

      {/* Mobile cards */}
      <div className="md:hidden space-y-2">
        {sorted.map((n) => {
          const k = rowKey(n)
          const stale = isStale(n)
          const fresh = freshKeys.has(k)
          return (
            <div
              key={`${k}-mobile`}
              className={`bg-radar-900 border border-radar-700 rounded p-3 ${
                stale ? 'opacity-40' : ''
              } ${fresh ? 'highlight-fresh' : ''}`}
            >
              <div className="flex justify-between items-start mb-1">
                <div>
                  <span className="font-mono text-accent-300 text-sm">
                    {privacyMode ? renderMasked(n.mac, splitMacForDisplay) : n.mac}
                  </span>
                  {lookupOUI(n.mac) && (
                    <div className="text-radar-500 text-xs">
                      {lookupOUI(n.mac)}
                    </div>
                  )}
                </div>
                {showInterface && (
                  <span className="text-xs text-radar-500">{n.interface}</span>
                )}
              </div>
              {n.ipv4.length > 0 && (
                <div className="text-xs font-mono text-radar-200">
                  {n.ipv4.join(', ')}
                </div>
              )}
              {n.ipv6.length > 0 && (
                <div className="text-xs font-mono text-radar-300 break-all">
                  {privacyMode
                    ? n.ipv6.map((ip, j) => (
                      <span key={j}>{j > 0 && ', '}{renderMasked(ip, splitIPv6ForDisplay)}</span>
                    ))
                    : n.ipv6.join(', ')}
                </div>
              )}
              <div className="flex justify-between text-xs text-radar-500 mt-2">
                <span title={n.firstSeen}>First: {n.firstSeen ? formatAgo(n.firstSeen) : ''}</span>
                <span title={n.lastSeen}>Last: {n.lastSeen ? formatAgo(n.lastSeen) : ''}</span>
              </div>
            </div>
          )
        })}
        {sorted.length === 0 && (
          <p className="text-center text-radar-500 py-8">
            No neighbours found
          </p>
        )}
      </div>
    </>
  )
}

export default NeighbourTable
