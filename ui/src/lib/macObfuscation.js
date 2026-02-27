/**
 * Build a mapping from real MAC addresses to obfuscated ones.
 * Preserves the OUI prefix (first 3 bytes), replaces last 3 with a counter.
 * @param {string[]} macs
 * @returns {Map<string, string>}
 */
export function buildMacMapping(macs) {
  const mapping = new Map()
  let counter = 0
  for (const mac of macs) {
    if (mapping.has(mac)) continue
    const oui = mac.slice(0, 8)
    const lo = counter & 0xff
    const mid = (counter >> 8) & 0xff
    const hi = (counter >> 16) & 0xff
    const suffix = [hi, mid, lo].map((b) => b.toString(16).padStart(2, '0')).join(':')
    mapping.set(mac, `${oui}:${suffix}`)
    counter++
  }
  return mapping
}

/**
 * Mask the last 3 bytes of IPv6 link-local addresses (fe80::/10).
 * These bytes leak the MAC host identifier via EUI-64 encoding.
 * Non-link-local addresses are returned unchanged.
 * @param {string[]} ipv6Addrs
 * @returns {string[]}
 */
export function obfuscateIPv6LinkLocal(ipv6Addrs) {
  if (!ipv6Addrs) return ipv6Addrs
  return ipv6Addrs.map((addr) => {
    if (!addr.toLowerCase().startsWith('fe80:')) return addr
    // Expand :: into full 8-group form
    const groups = expandIPv6(addr)
    // Zero the last 3 bytes: low byte of group[6] and all of group[7]
    groups[6] = groups[6] & 0xff00
    groups[7] = 0
    return compressIPv6(groups)
  })
}

/**
 * Expand an IPv6 address string into an array of 8 16-bit group values.
 * @param {string} addr
 * @returns {number[]}
 */
function expandIPv6(addr) {
  const halves = addr.split('::')
  const left = halves[0] ? halves[0].split(':').map((g) => parseInt(g, 16)) : []
  const right = halves.length > 1 && halves[1] ? halves[1].split(':').map((g) => parseInt(g, 16)) : []
  const missing = 8 - left.length - right.length
  return [...left, ...Array(missing).fill(0), ...right]
}

/**
 * Compress an array of 8 16-bit group values into shortest IPv6 string.
 * @param {number[]} groups
 * @returns {string}
 */
function compressIPv6(groups) {
  const hex = groups.map((g) => g.toString(16))
  // Find longest run of consecutive '0' groups for :: compression
  let bestStart = -1, bestLen = 0, curStart = -1, curLen = 0
  for (let i = 0; i < 8; i++) {
    if (hex[i] === '0') {
      if (curStart === -1) curStart = i
      curLen = i - curStart + 1
      if (curLen > bestLen) { bestStart = curStart; bestLen = curLen }
    } else {
      curStart = -1; curLen = 0
    }
  }
  if (bestLen < 2) return hex.join(':')
  const before = hex.slice(0, bestStart).join(':')
  const after = hex.slice(bestStart + bestLen).join(':')
  return `${before}::${after}`
}

/**
 * Apply MAC and IPv6 link-local obfuscation to neighbours and interfaceInfo.
 * Returns cloned data with MACs replaced according to the mapping and
 * link-local IPv6 addresses masked.
 * @param {object[]} neighbours
 * @param {object} interfaceInfo
 * @param {Map<string, string>} mapping
 * @returns {{ neighbours: object[], interfaceInfo: object }}
 */
export function obfuscateData(neighbours, interfaceInfo, mapping) {
  const newNeighbours = neighbours.map((n) => {
    const mapped = mapping.get(n.mac)
    return {
      ...n,
      ...(mapped && { mac: mapped }),
      ...(n.ipv6 && { ipv6: obfuscateIPv6LinkLocal(n.ipv6) }),
    }
  })

  const newInterfaceInfo = {}
  for (const [iface, info] of Object.entries(interfaceInfo)) {
    const mapped = mapping.get(info.mac)
    newInterfaceInfo[iface] = {
      ...info,
      ...(mapped && { mac: mapped }),
      ...(info.ipv6 && { ipv6: obfuscateIPv6LinkLocal(info.ipv6) }),
    }
  }

  return { neighbours: newNeighbours, interfaceInfo: newInterfaceInfo }
}
