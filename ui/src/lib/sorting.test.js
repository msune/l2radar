import { describe, it, expect } from 'vitest'
import { sortNeighbours } from './sorting'

const data = [
  {
    interface: 'wlan0',
    mac: 'cc:cc:cc:cc:cc:cc',
    ipv4: ['10.0.0.1'],
    ipv6: ['fe80::2'],
    firstSeen: '2026-02-14T14:00:00Z',
    lastSeen: '2026-02-14T14:30:00Z',
  },
  {
    interface: 'eth0',
    mac: 'aa:aa:aa:aa:aa:aa',
    ipv4: ['192.168.1.1'],
    ipv6: ['fe80::1'],
    firstSeen: '2026-02-14T13:00:00Z',
    lastSeen: '2026-02-14T14:32:00Z',
  },
  {
    interface: 'eth0',
    mac: 'bb:bb:bb:bb:bb:bb',
    ipv4: [],
    ipv6: [],
    firstSeen: '2026-02-14T14:10:00Z',
    lastSeen: '2026-02-14T14:10:00Z',
  },
]

describe('sortNeighbours', () => {
  it('sorts by lastSeen descending by default', () => {
    const sorted = sortNeighbours(data, 'lastSeen', 'desc')
    expect(sorted[0].mac).toBe('aa:aa:aa:aa:aa:aa')
    expect(sorted[1].mac).toBe('cc:cc:cc:cc:cc:cc')
    expect(sorted[2].mac).toBe('bb:bb:bb:bb:bb:bb')
  })

  it('sorts by lastSeen ascending', () => {
    const sorted = sortNeighbours(data, 'lastSeen', 'asc')
    expect(sorted[0].mac).toBe('bb:bb:bb:bb:bb:bb')
    expect(sorted[2].mac).toBe('aa:aa:aa:aa:aa:aa')
  })

  it('sorts by MAC ascending', () => {
    const sorted = sortNeighbours(data, 'mac', 'asc')
    expect(sorted[0].mac).toBe('aa:aa:aa:aa:aa:aa')
    expect(sorted[1].mac).toBe('bb:bb:bb:bb:bb:bb')
    expect(sorted[2].mac).toBe('cc:cc:cc:cc:cc:cc')
  })

  it('sorts by interface', () => {
    const sorted = sortNeighbours(data, 'interface', 'asc')
    expect(sorted[0].interface).toBe('eth0')
    expect(sorted[2].interface).toBe('wlan0')
  })

  it('sorts by firstSeen', () => {
    const sorted = sortNeighbours(data, 'firstSeen', 'asc')
    expect(sorted[0].mac).toBe('aa:aa:aa:aa:aa:aa')
    expect(sorted[2].mac).toBe('bb:bb:bb:bb:bb:bb')
  })

  it('sorts by ipv4', () => {
    const sorted = sortNeighbours(data, 'ipv4', 'asc')
    // Empty ipv4 sorts first
    expect(sorted[0].mac).toBe('bb:bb:bb:bb:bb:bb')
  })

  it('does not mutate the original array', () => {
    const original = [...data]
    sortNeighbours(data, 'mac', 'asc')
    expect(data[0].mac).toBe(original[0].mac)
  })

  it('handles empty array', () => {
    expect(sortNeighbours([], 'mac', 'asc')).toEqual([])
  })
})
