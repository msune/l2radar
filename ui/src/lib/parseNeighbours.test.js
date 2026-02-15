import { describe, it, expect } from 'vitest'
import { parseInterfaceData, mergeNeighbours } from './parseNeighbours'
import eth0Data from '../../../testdata/neigh-eth0.json'
import wlan0Data from '../../../testdata/neigh-wlan0.json'

describe('parseInterfaceData', () => {
  it('parses eth0 golden file correctly', () => {
    const neighbours = parseInterfaceData(eth0Data)
    expect(neighbours).toHaveLength(4)

    // All should have interface set
    neighbours.forEach((n) => {
      expect(n.interface).toBe('eth0')
    })
  })

  it('parses wlan0 golden file correctly', () => {
    const neighbours = parseInterfaceData(wlan0Data)
    expect(neighbours).toHaveLength(3)

    neighbours.forEach((n) => {
      expect(n.interface).toBe('wlan0')
    })
  })

  it('preserves MAC format from golden files', () => {
    const neighbours = parseInterfaceData(eth0Data)
    // MAC addresses should be colon-separated hex
    neighbours.forEach((n) => {
      expect(n.mac).toMatch(/^[0-9a-f]{2}(:[0-9a-f]{2}){5}$/)
    })
  })

  it('handles multi-IPv4 and multi-IPv6 neighbour', () => {
    const neighbours = parseInterfaceData(eth0Data)
    const multi = neighbours.find((n) => n.mac === 'dc:4b:a1:69:38:16')
    expect(multi).toBeDefined()
    expect(multi.ipv4).toEqual(['192.168.1.33', '10.0.0.5'])
    expect(multi.ipv6).toEqual(['fe80::c09e:74a6:4353:cd6a', '2001:db8::1'])
  })

  it('handles IPv4-only neighbour', () => {
    const neighbours = parseInterfaceData(eth0Data)
    const v4only = neighbours.find((n) => n.mac === '02:42:ac:11:00:02')
    expect(v4only).toBeDefined()
    expect(v4only.ipv4).toEqual(['192.168.1.100'])
    expect(v4only.ipv6).toEqual([])
  })

  it('handles IPv6-only neighbour', () => {
    const neighbours = parseInterfaceData(eth0Data)
    const v6only = neighbours.find((n) => n.mac === 'aa:bb:cc:dd:ee:01')
    expect(v6only).toBeDefined()
    expect(v6only.ipv4).toEqual([])
    expect(v6only.ipv6).toEqual(['fe80::1'])
  })

  it('handles neighbour with no IPs', () => {
    const neighbours = parseInterfaceData(eth0Data)
    const noIp = neighbours.find((n) => n.mac === 'aa:bb:cc:dd:ee:02')
    expect(noIp).toBeDefined()
    expect(noIp.ipv4).toEqual([])
    expect(noIp.ipv6).toEqual([])
  })

  it('handles neighbour with 3 IPv4 and 3 IPv6', () => {
    const neighbours = parseInterfaceData(wlan0Data)
    const multi = neighbours.find((n) => n.mac === 'f0:de:f1:23:45:67')
    expect(multi).toBeDefined()
    expect(multi.ipv4).toHaveLength(3)
    expect(multi.ipv6).toHaveLength(3)
  })

  it('parses timestamps as valid ISO 8601 strings', () => {
    const neighbours = parseInterfaceData(eth0Data)
    neighbours.forEach((n) => {
      expect(new Date(n.firstSeen).toISOString()).toBeTruthy()
      expect(new Date(n.lastSeen).toISOString()).toBeTruthy()
      // Should not be "Invalid Date"
      expect(isNaN(new Date(n.firstSeen))).toBe(false)
      expect(isNaN(new Date(n.lastSeen))).toBe(false)
    })
  })

  it('rejects data without interface field', () => {
    expect(() => parseInterfaceData({ neighbours: [] })).toThrow()
  })

  it('rejects data without neighbours field', () => {
    expect(() => parseInterfaceData({ interface: 'eth0' })).toThrow()
  })
})

describe('mergeNeighbours', () => {
  it('merges both golden files', () => {
    const { neighbours } = mergeNeighbours([eth0Data, wlan0Data])
    // eth0: 4, wlan0: 3
    expect(neighbours).toHaveLength(7)
  })

  it('preserves interface tags after merge', () => {
    const { neighbours } = mergeNeighbours([eth0Data, wlan0Data])
    const eth0Count = neighbours.filter((n) => n.interface === 'eth0').length
    const wlan0Count = neighbours.filter((n) => n.interface === 'wlan0').length
    expect(eth0Count).toBe(4)
    expect(wlan0Count).toBe(3)
  })

  it('same MAC on different interfaces appears as separate entries', () => {
    const { neighbours } = mergeNeighbours([eth0Data, wlan0Data])
    const sameMac = neighbours.filter((n) => n.mac === 'dc:4b:a1:69:38:16')
    expect(sameMac).toHaveLength(2)
    expect(sameMac.map((n) => n.interface).sort()).toEqual(['eth0', 'wlan0'])
  })

  it('handles empty array', () => {
    const { neighbours, timestamps, interfaceInfo } = mergeNeighbours([])
    expect(neighbours).toEqual([])
    expect(timestamps).toEqual({})
    expect(interfaceInfo).toEqual({})
  })

  it('returns per-interface timestamps', () => {
    const { timestamps } = mergeNeighbours([eth0Data, wlan0Data])
    expect(timestamps).toEqual({
      eth0: eth0Data.timestamp,
      wlan0: wlan0Data.timestamp,
    })
  })

  it('returns timestamps even for interfaces with no neighbours', () => {
    const emptyIface = { interface: 'br0', timestamp: '2026-02-14T15:00:00Z', neighbours: [] }
    const { neighbours, timestamps } = mergeNeighbours([emptyIface])
    expect(neighbours).toEqual([])
    expect(timestamps).toEqual({ br0: '2026-02-14T15:00:00Z' })
  })

  it('returns per-interface info with mac, addresses, and stats', () => {
    const { interfaceInfo } = mergeNeighbours([eth0Data, wlan0Data])
    expect(interfaceInfo.eth0).toEqual({
      mac: eth0Data.mac,
      ipv4: eth0Data.ipv4,
      ipv6: eth0Data.ipv6,
      exportInterval: eth0Data.export_interval || '',
      stats: eth0Data.stats || null,
    })
    expect(interfaceInfo.wlan0).toEqual({
      mac: wlan0Data.mac,
      ipv4: wlan0Data.ipv4,
      ipv6: wlan0Data.ipv6,
      exportInterval: wlan0Data.export_interval || '',
      stats: wlan0Data.stats || null,
    })
  })

  it('includes stats counters from golden files', () => {
    const { interfaceInfo } = mergeNeighbours([eth0Data, wlan0Data])
    expect(interfaceInfo.eth0.stats).toBeDefined()
    expect(interfaceInfo.eth0.stats.tx_bytes).toBe(1234567)
    expect(interfaceInfo.eth0.stats.rx_bytes).toBe(7890123)
    expect(interfaceInfo.wlan0.stats).toBeDefined()
    expect(interfaceInfo.wlan0.stats.rx_errors).toBe(1)
  })

  it('defaults interface info arrays and stats to empty/null when missing', () => {
    const minimal = { interface: 'br0', timestamp: '2026-02-14T15:00:00Z', neighbours: [] }
    const { interfaceInfo } = mergeNeighbours([minimal])
    expect(interfaceInfo.br0).toEqual({
      mac: '',
      ipv4: [],
      ipv6: [],
      exportInterval: '',
      stats: null,
    })
  })
})
