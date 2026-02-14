/**
 * Sort neighbours by a given column.
 * Returns a new sorted array.
 */
export function sortNeighbours(neighbours, sortKey, sortDir) {
  return [...neighbours].sort((a, b) => {
    let cmp = 0

    switch (sortKey) {
      case 'interface':
        cmp = a.interface.localeCompare(b.interface)
        break
      case 'mac':
        cmp = a.mac.localeCompare(b.mac)
        break
      case 'ipv4':
        cmp = (a.ipv4[0] || '').localeCompare(b.ipv4[0] || '')
        break
      case 'ipv6':
        cmp = (a.ipv6[0] || '').localeCompare(b.ipv6[0] || '')
        break
      case 'firstSeen':
        cmp = new Date(a.firstSeen) - new Date(b.firstSeen)
        break
      case 'lastSeen':
      default:
        cmp = new Date(a.lastSeen) - new Date(b.lastSeen)
        break
    }

    return sortDir === 'asc' ? cmp : -cmp
  })
}
