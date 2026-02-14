import { describe, it, expect } from 'vitest'
import { filterNeighbours, getInterfaces } from './filter'

const data = [
  {
    interface: 'eth0',
    mac: 'aa:bb:cc:dd:ee:01',
    ipv4: ['192.168.1.1', '10.0.0.5'],
    ipv6: ['fe80::1'],
  },
  {
    interface: 'eth0',
    mac: 'aa:bb:cc:dd:ee:02',
    ipv4: [],
    ipv6: [],
  },
  {
    interface: 'wlan0',
    mac: 'ff:ff:ff:00:00:01',
    ipv4: ['172.16.0.1'],
    ipv6: ['2001:db8::1'],
  },
]

describe('filterNeighbours', () => {
  it('returns all when no filters applied', () => {
    expect(filterNeighbours(data)).toHaveLength(3)
  })

  it('filters by interface', () => {
    const result = filterNeighbours(data, { iface: 'eth0' })
    expect(result).toHaveLength(2)
    result.forEach((n) => expect(n.interface).toBe('eth0'))
  })

  it('filters by MAC (partial match)', () => {
    const result = filterNeighbours(data, { search: 'ee:01' })
    expect(result).toHaveLength(1)
    expect(result[0].mac).toBe('aa:bb:cc:dd:ee:01')
  })

  it('filters by IPv4 (partial match)', () => {
    const result = filterNeighbours(data, { search: '192.168' })
    expect(result).toHaveLength(1)
    expect(result[0].mac).toBe('aa:bb:cc:dd:ee:01')
  })

  it('filters by IPv6 (partial match)', () => {
    const result = filterNeighbours(data, { search: '2001:db8' })
    expect(result).toHaveLength(1)
    expect(result[0].mac).toBe('ff:ff:ff:00:00:01')
  })

  it('search is case-insensitive', () => {
    const result = filterNeighbours(data, { search: 'AA:BB' })
    expect(result).toHaveLength(2)
  })

  it('combines search and interface filter', () => {
    const result = filterNeighbours(data, { search: 'ee:01', iface: 'eth0' })
    expect(result).toHaveLength(1)
  })

  it('combined filter with no match returns empty', () => {
    const result = filterNeighbours(data, { search: 'ee:01', iface: 'wlan0' })
    expect(result).toHaveLength(0)
  })

  it('handles empty neighbours', () => {
    expect(filterNeighbours([], { search: 'foo' })).toEqual([])
  })

  it('matches neighbour with no IPs by MAC only', () => {
    const result = filterNeighbours(data, { search: 'ee:02' })
    expect(result).toHaveLength(1)
    expect(result[0].mac).toBe('aa:bb:cc:dd:ee:02')
  })
})

describe('getInterfaces', () => {
  it('returns sorted unique interfaces', () => {
    expect(getInterfaces(data)).toEqual(['eth0', 'wlan0'])
  })

  it('returns empty for no neighbours', () => {
    expect(getInterfaces([])).toEqual([])
  })
})
