/**
 * Filter neighbours by search query (matches MAC or IP, case-insensitive)
 * and by interface.
 */
export function filterNeighbours(neighbours, { search = '', iface = '' } = {}) {
  let filtered = neighbours

  if (iface) {
    filtered = filtered.filter((n) => n.interface === iface)
  }

  if (search) {
    const q = search.toLowerCase()
    filtered = filtered.filter((n) => {
      if (n.mac.toLowerCase().includes(q)) return true
      if (n.ipv4.some((ip) => ip.toLowerCase().includes(q))) return true
      if (n.ipv6.some((ip) => ip.toLowerCase().includes(q))) return true
      return false
    })
  }

  return filtered
}

/**
 * Extract unique interface names from a list of neighbours, sorted.
 */
export function getInterfaces(neighbours) {
  const set = new Set(neighbours.map((n) => n.interface))
  return [...set].sort()
}
