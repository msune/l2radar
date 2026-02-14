function SummaryStats({ neighbours }) {
  const total = neighbours.length

  // Count per interface
  const perInterface = {}
  for (const n of neighbours) {
    perInterface[n.interface] = (perInterface[n.interface] || 0) + 1
  }

  // Active in last 5 minutes
  const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000)
  const recentCount = neighbours.filter(
    (n) => new Date(n.lastSeen) >= fiveMinAgo
  ).length

  const interfaces = Object.entries(perInterface).sort(([a], [b]) =>
    a.localeCompare(b)
  )

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2 mb-4">
      <StatCard label="Total Neighbours" value={total} />
      <StatCard
        label="Active (5 min)"
        value={recentCount}
        accent={recentCount > 0}
      />
      {interfaces.map(([iface, count]) => (
        <StatCard key={iface} label={iface} value={count} />
      ))}
    </div>
  )
}

function StatCard({ label, value, accent = false }) {
  return (
    <div className="bg-radar-900 border border-radar-700 rounded px-3 py-2">
      <div className="text-xs text-radar-400 uppercase tracking-wide">
        {label}
      </div>
      <div
        className={`text-xl font-semibold ${accent ? 'text-accent-400' : 'text-radar-100'}`}
      >
        {value}
      </div>
    </div>
  )
}

export default SummaryStats
