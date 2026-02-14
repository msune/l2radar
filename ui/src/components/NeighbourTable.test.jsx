import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup, within } from '@testing-library/react'
import NeighbourTable from './NeighbourTable'

afterEach(cleanup)

const mockData = [
  {
    interface: 'eth0',
    mac: 'aa:bb:cc:dd:ee:01',
    ipv4: ['192.168.1.1'],
    ipv6: ['fe80::1'],
    firstSeen: '2026-02-14T14:00:00Z',
    lastSeen: '2026-02-14T14:30:00Z',
  },
  {
    interface: 'wlan0',
    mac: 'aa:bb:cc:dd:ee:02',
    ipv4: [],
    ipv6: [],
    firstSeen: '2026-02-14T14:10:00Z',
    lastSeen: '2026-02-14T14:32:00Z',
  },
]

describe('NeighbourTable', () => {
  it('renders all MAC addresses', () => {
    render(<NeighbourTable neighbours={mockData} />)
    expect(screen.getAllByText('aa:bb:cc:dd:ee:01').length).toBeGreaterThan(0)
    expect(screen.getAllByText('aa:bb:cc:dd:ee:02').length).toBeGreaterThan(0)
  })

  it('renders IPv4 addresses', () => {
    render(<NeighbourTable neighbours={mockData} />)
    expect(screen.getAllByText('192.168.1.1').length).toBeGreaterThan(0)
  })

  it('shows dash for empty IPs in desktop view', () => {
    render(<NeighbourTable neighbours={mockData} />)
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('shows empty state message', () => {
    render(<NeighbourTable neighbours={[]} />)
    expect(screen.getAllByText('No neighbours found').length).toBeGreaterThan(0)
  })

  it('sorts by MAC when header clicked', () => {
    render(<NeighbourTable neighbours={mockData} />)
    // Get the table element specifically
    const table = screen.getByRole('table')
    const macHeader = within(table).getByText('MAC')
    fireEvent.click(macHeader)
    // After clicking MAC, should sort ascending
    const rows = within(table).getAllByRole('row')
    // rows[0] is header, rows[1] is first data row
    expect(within(rows[1]).getByText('aa:bb:cc:dd:ee:01')).toBeInTheDocument()
    expect(within(rows[2]).getByText('aa:bb:cc:dd:ee:02')).toBeInTheDocument()
  })

  it('toggles sort direction on same column click', () => {
    render(<NeighbourTable neighbours={mockData} />)
    const table = screen.getByRole('table')
    const macHeader = within(table).getByText('MAC')
    // First click: asc
    fireEvent.click(macHeader)
    // Second click: desc
    fireEvent.click(macHeader)
    const rows = within(table).getAllByRole('row')
    expect(within(rows[1]).getByText('aa:bb:cc:dd:ee:02')).toBeInTheDocument()
    expect(within(rows[2]).getByText('aa:bb:cc:dd:ee:01')).toBeInTheDocument()
  })
})
