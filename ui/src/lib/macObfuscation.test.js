import { describe, it, expect } from 'vitest'
import { buildMacMapping, obfuscateData, obfuscateIPv6LinkLocal, splitMacForDisplay, splitIPv6ForDisplay } from './macObfuscation'

describe('buildMacMapping', () => {
  it('returns empty map for empty input', () => {
    const mapping = buildMacMapping([])
    expect(mapping.size).toBe(0)
  })

  it('preserves OUI prefix (first 3 bytes)', () => {
    const mapping = buildMacMapping(['aa:bb:cc:11:22:33'])
    expect(mapping.get('aa:bb:cc:11:22:33')).toBe('aa:bb:cc:00:00:00')
  })

  it('assigns sequential counter across all MACs', () => {
    const mapping = buildMacMapping([
      'aa:bb:cc:11:22:33',
      'dd:ee:ff:44:55:66',
      'aa:bb:cc:77:88:99',
    ])
    expect(mapping.get('aa:bb:cc:11:22:33')).toBe('aa:bb:cc:00:00:00')
    expect(mapping.get('dd:ee:ff:44:55:66')).toBe('dd:ee:ff:00:00:01')
    expect(mapping.get('aa:bb:cc:77:88:99')).toBe('aa:bb:cc:00:00:02')
  })

  it('produces stable mapping (same input same output)', () => {
    const macs = ['aa:bb:cc:11:22:33', 'dd:ee:ff:44:55:66']
    const m1 = buildMacMapping(macs)
    const m2 = buildMacMapping(macs)
    expect(m1.get('aa:bb:cc:11:22:33')).toBe(m2.get('aa:bb:cc:11:22:33'))
    expect(m1.get('dd:ee:ff:44:55:66')).toBe(m2.get('dd:ee:ff:44:55:66'))
  })

  it('deduplicates MACs', () => {
    const mapping = buildMacMapping([
      'aa:bb:cc:11:22:33',
      'aa:bb:cc:11:22:33',
      'dd:ee:ff:44:55:66',
    ])
    expect(mapping.size).toBe(2)
    expect(mapping.get('aa:bb:cc:11:22:33')).toBe('aa:bb:cc:00:00:00')
    expect(mapping.get('dd:ee:ff:44:55:66')).toBe('dd:ee:ff:00:00:01')
  })
})

describe('obfuscateIPv6LinkLocal', () => {
  it('masks last 3 bytes of link-local address', () => {
    const result = obfuscateIPv6LinkLocal(['fe80::c09e:74a6:4353:cd6a'])
    expect(result).toEqual(['fe80::c09e:74a6:4300:0'])
  })

  it('leaves non-link-local addresses unchanged', () => {
    const result = obfuscateIPv6LinkLocal(['2001:db8::1', 'fd00::1'])
    expect(result).toEqual(['2001:db8::1', 'fd00::1'])
  })

  it('handles mixed link-local and non-link-local', () => {
    const result = obfuscateIPv6LinkLocal(['fe80::aabb:ccff:fedd:eeff', '2001:db8::abcd'])
    expect(result).toEqual(['fe80::aabb:ccff:fe00:0', '2001:db8::abcd'])
  })

  it('returns null/undefined input unchanged', () => {
    expect(obfuscateIPv6LinkLocal(null)).toBe(null)
    expect(obfuscateIPv6LinkLocal(undefined)).toBe(undefined)
  })

  it('handles short link-local addresses', () => {
    const result = obfuscateIPv6LinkLocal(['fe80::1'])
    expect(result).toEqual(['fe80::'])
  })
})

describe('splitMacForDisplay', () => {
  it('splits OUI prefix from host identifier', () => {
    expect(splitMacForDisplay('aa:bb:cc:dd:ee:ff')).toEqual({
      prefix: 'aa:bb:cc',
      masked: ':dd:ee:ff',
    })
  })

  it('splits obfuscated MAC', () => {
    expect(splitMacForDisplay('aa:bb:cc:00:00:00')).toEqual({
      prefix: 'aa:bb:cc',
      masked: ':00:00:00',
    })
  })
})

describe('splitIPv6ForDisplay', () => {
  it('splits link-local address into prefix and last 2 groups', () => {
    const result = splitIPv6ForDisplay('fe80::aabb:ccff:fe00:0')
    expect(result).toEqual({ prefix: 'fe80::aabb:ccff:', masked: 'fe00:0' })
  })

  it('returns non-link-local address with empty masked', () => {
    const result = splitIPv6ForDisplay('2001:db8::1')
    expect(result).toEqual({ prefix: '2001:db8::1', masked: '' })
  })

  it('returns empty masked when last 2 groups are zero', () => {
    const result = splitIPv6ForDisplay('fe80::')
    expect(result).toEqual({ prefix: 'fe80::', masked: '' })
  })

  it('handles typical obfuscated address', () => {
    const result = splitIPv6ForDisplay('fe80::c09e:74a6:4300:0')
    expect(result).toEqual({ prefix: 'fe80::c09e:74a6:', masked: '4300:0' })
  })

  it('handles address with non-zero group 7 only', () => {
    const result = splitIPv6ForDisplay('fe80::1')
    // group 6 = 0, group 7 = 1 → both shown explicitly
    expect(result).toEqual({ prefix: 'fe80::', masked: '0:1' })
  })
})

describe('obfuscateData', () => {
  const mapping = new Map([
    ['aa:bb:cc:11:22:33', 'aa:bb:cc:00:00:00'],
    ['dd:ee:ff:44:55:66', 'dd:ee:ff:00:00:01'],
  ])

  it('replaces neighbour MACs', () => {
    const neighbours = [
      { mac: 'aa:bb:cc:11:22:33', interface: 'eth0', ipv4: ['10.0.0.1'] },
      { mac: 'dd:ee:ff:44:55:66', interface: 'eth0', ipv4: ['10.0.0.2'] },
    ]
    const result = obfuscateData(neighbours, {}, mapping)
    expect(result.neighbours[0].mac).toBe('aa:bb:cc:00:00:00')
    expect(result.neighbours[1].mac).toBe('dd:ee:ff:00:00:01')
  })

  it('masks IPv6 link-local addresses in neighbours', () => {
    const neighbours = [
      { mac: 'aa:bb:cc:11:22:33', interface: 'eth0', ipv6: ['fe80::aabb:ccff:fe11:2233', '2001:db8::1'] },
    ]
    const result = obfuscateData(neighbours, {}, mapping)
    expect(result.neighbours[0].ipv6).toEqual(['fe80::aabb:ccff:fe00:0', '2001:db8::1'])
  })

  it('masks IPv6 link-local addresses in interfaceInfo', () => {
    const interfaceInfo = {
      eth0: { mac: 'aa:bb:cc:11:22:33', ipv6: ['fe80::dcad:beff:feef:1'] },
    }
    const result = obfuscateData([], interfaceInfo, mapping)
    expect(result.interfaceInfo.eth0.ipv6).toEqual(['fe80::dcad:beff:fe00:0'])
  })

  it('does not mutate original neighbours', () => {
    const neighbours = [
      { mac: 'aa:bb:cc:11:22:33', interface: 'eth0' },
    ]
    obfuscateData(neighbours, {}, mapping)
    expect(neighbours[0].mac).toBe('aa:bb:cc:11:22:33')
  })

  it('replaces interface info MACs', () => {
    const interfaceInfo = {
      eth0: { mac: 'aa:bb:cc:11:22:33', ipv4: ['10.0.0.1'] },
    }
    const result = obfuscateData([], interfaceInfo, mapping)
    expect(result.interfaceInfo.eth0.mac).toBe('aa:bb:cc:00:00:00')
  })

  it('does not mutate original interfaceInfo', () => {
    const interfaceInfo = {
      eth0: { mac: 'aa:bb:cc:11:22:33' },
    }
    obfuscateData([], interfaceInfo, mapping)
    expect(interfaceInfo.eth0.mac).toBe('aa:bb:cc:11:22:33')
  })

  it('preserves other fields', () => {
    const neighbours = [
      { mac: 'aa:bb:cc:11:22:33', interface: 'eth0', ipv4: ['10.0.0.1'], vendor: 'Test' },
    ]
    const interfaceInfo = {
      eth0: { mac: 'dd:ee:ff:44:55:66', ipv4: ['10.0.0.254'], name: 'eth0' },
    }
    const result = obfuscateData(neighbours, interfaceInfo, mapping)
    expect(result.neighbours[0].interface).toBe('eth0')
    expect(result.neighbours[0].ipv4).toEqual(['10.0.0.1'])
    expect(result.neighbours[0].vendor).toBe('Test')
    expect(result.interfaceInfo.eth0.ipv4).toEqual(['10.0.0.254'])
    expect(result.interfaceInfo.eth0.name).toBe('eth0')
  })

  it('preserves non-link-local IPv6 addresses in neighbours', () => {
    const neighbours = [
      { mac: 'aa:bb:cc:11:22:33', interface: 'eth0', ipv6: ['2001:db8::1', 'fd00::1'] },
    ]
    const result = obfuscateData(neighbours, {}, mapping)
    expect(result.neighbours[0].ipv6).toEqual(['2001:db8::1', 'fd00::1'])
  })

  it('leaves MAC unchanged if not in mapping', () => {
    const neighbours = [
      { mac: '11:22:33:44:55:66', interface: 'eth0' },
    ]
    const result = obfuscateData(neighbours, {}, mapping)
    expect(result.neighbours[0].mac).toBe('11:22:33:44:55:66')
  })
})
