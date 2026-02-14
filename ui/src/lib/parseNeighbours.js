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
 * Returns { neighbours, timestamps, interfaceInfo } where:
 * - timestamps maps interface name to the probe's export timestamp
 * - interfaceInfo maps interface name to { mac, ipv4, ipv6 }
 */
export function mergeNeighbours(interfaceDataArray) {
  const neighbours = []
  const timestamps = {}
  const interfaceInfo = {}
  for (const data of interfaceDataArray) {
    neighbours.push(...parseInterfaceData(data))
    if (data.timestamp) {
      timestamps[data.interface] = data.timestamp
    }
    interfaceInfo[data.interface] = {
      mac: data.mac || '',
      ipv4: data.ipv4 || [],
      ipv6: data.ipv6 || [],
    }
  }
  return { neighbours, timestamps, interfaceInfo }
}
