import { describe, it, expect } from 'vitest'
import { render, screen, within, cleanup } from '@testing-library/react'
import SummaryStats from './SummaryStats'

const mockNeighbours = [
  {
    interface: 'eth0',
    mac: 'aa:bb:cc:dd:ee:01',
    ipv4: ['192.168.1.1'],
    ipv6: [],
    firstSeen: '2026-02-14T14:00:00Z',
    lastSeen: new Date().toISOString(), // recent
  },
  {
    interface: 'eth0',
    mac: 'aa:bb:cc:dd:ee:02',
    ipv4: [],
    ipv6: ['fe80::1'],
    firstSeen: '2020-01-01T00:00:00Z',
    lastSeen: '2020-01-01T00:00:00Z', // old
  },
  {
    interface: 'wlan0',
    mac: 'aa:bb:cc:dd:ee:03',
    ipv4: ['10.0.0.1'],
    ipv6: [],
    firstSeen: '2026-02-14T14:00:00Z',
    lastSeen: new Date().toISOString(), // recent
  },
]

function getStatValue(label) {
  const labelEl = screen.getByText(label)
  const card = labelEl.closest('.bg-radar-900')
  return within(card).getByText(/^\d+$/).textContent
}

describe('SummaryStats', () => {
  it('shows total neighbour count', () => {
    cleanup()
    render(<SummaryStats neighbours={mockNeighbours} />)
    expect(getStatValue('Total Neighbours')).toBe('3')
  })

  it('shows active count', () => {
    cleanup()
    render(<SummaryStats neighbours={mockNeighbours} />)
    expect(getStatValue('Active (5 min)')).toBe('2')
  })

  it('shows per-interface counts', () => {
    cleanup()
    render(<SummaryStats neighbours={mockNeighbours} />)
    expect(getStatValue('eth0')).toBe('2')
    expect(getStatValue('wlan0')).toBe('1')
  })

  it('renders with empty neighbours', () => {
    cleanup()
    render(<SummaryStats neighbours={[]} />)
    expect(getStatValue('Total Neighbours')).toBe('0')
    expect(getStatValue('Active (5 min)')).toBe('0')
  })
})
