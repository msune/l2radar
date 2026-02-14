import { describe, it, expect } from 'vitest'
import { lookupOUI } from './ouiLookup'

describe('lookupOUI', () => {
  it('returns vendor for known OUI prefix', () => {
    // 28:6f:b9 = Nokia Shanghai Bell Co., Ltd.
    expect(lookupOUI('28:6f:b9:11:22:33')).toBe('Nokia Shanghai Bell Co., Ltd.')
  })

  it('returns vendor for another known prefix', () => {
    // 08:ea:44 = Extreme Networks Headquarters
    expect(lookupOUI('08:ea:44:aa:bb:cc')).toBe('Extreme Networks Headquarters')
  })

  it('is case-insensitive', () => {
    expect(lookupOUI('28:6F:B9:00:00:00')).toBe('Nokia Shanghai Bell Co., Ltd.')
  })

  it('returns empty string for unknown OUI', () => {
    expect(lookupOUI('ff:ff:ff:00:00:00')).toBe('')
  })

  it('returns empty string for empty input', () => {
    expect(lookupOUI('')).toBe('')
  })

  it('returns empty string for malformed MAC', () => {
    expect(lookupOUI('zz:zz:zz')).toBe('')
  })

  it('handles MAC with only 3 octets', () => {
    expect(lookupOUI('28:6f:b9')).toBe('Nokia Shanghai Bell Co., Ltd.')
  })
})
