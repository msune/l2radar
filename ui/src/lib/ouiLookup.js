import ouiDb from './oui.json'

/**
 * Look up the OUI vendor name for a MAC address.
 * Returns empty string if not found or input is invalid.
 */
export function lookupOUI(mac) {
  if (!mac || typeof mac !== 'string') return ''

  // Extract first 3 octets, lowercase
  const prefix = mac.toLowerCase().split(':').slice(0, 3).join(':')
  if (!/^[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}$/.test(prefix)) return ''

  return ouiDb[prefix] || ''
}
