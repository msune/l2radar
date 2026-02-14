/**
 * Parse a single interface JSON file into an array of neighbour objects,
 * each tagged with the interface name.
 */
export function parseInterfaceData(data) {
  if (!data || !data.interface || !Array.isArray(data.neighbours)) {
    throw new Error('Invalid interface data: missing required fields')
  }

  return data.neighbours.map((n) => ({
    interface: data.interface,
    mac: n.mac,
    ipv4: n.ipv4 || [],
    ipv6: n.ipv6 || [],
    firstSeen: n.first_seen,
    lastSeen: n.last_seen,
  }))
}

/**
 * Merge neighbour arrays from multiple interfaces into a single array.
 */
export function mergeNeighbours(interfaceDataArray) {
  const result = []
  for (const data of interfaceDataArray) {
    result.push(...parseInterfaceData(data))
  }
  return result
}
