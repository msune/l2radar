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

  it('displays vendor name for known OUI prefix', () => {
    const data = [
      {
        interface: 'eth0',
        // 28:6f:b9 = Nokia Shanghai Bell Co., Ltd.
        mac: '28:6f:b9:11:22:33',
        ipv4: [],
        ipv6: [],
        firstSeen: '2026-02-14T14:00:00Z',
        lastSeen: '2026-02-14T14:30:00Z',
      },
    ]
    render(<NeighbourTable neighbours={data} />)
    expect(screen.getAllByText(/Nokia Shanghai Bell/).length).toBeGreaterThan(0)
  })

  it('hides Interface column when showInterface is false', () => {
    render(<NeighbourTable neighbours={mockData} showInterface={false} />)
    const table = screen.getByRole('table')
    expect(within(table).queryByText('Interface')).toBeNull()
  })

  it('shows Interface column by default', () => {
    render(<NeighbourTable neighbours={mockData} />)
    const table = screen.getByRole('table')
    expect(within(table).getByText('Interface')).toBeInTheDocument()
  })

  it('grays out last 3 MAC bytes in privacy mode', () => {
    const data = [{
      interface: 'eth0',
      mac: 'aa:bb:cc:dd:ee:01',
      ipv4: [],
      ipv6: [],
      firstSeen: '2026-02-14T14:00:00Z',
      lastSeen: '2026-02-14T14:30:00Z',
    }]
    render(<NeighbourTable neighbours={data} privacyMode={true} />)
    const table = screen.getByRole('table')
    const macCell = table.querySelector('td.font-mono.text-accent-300')
    expect(macCell.textContent).toContain('aa:bb:cc')
    const graySpan = macCell.querySelector('.text-radar-600')
    expect(graySpan).toBeInTheDocument()
    expect(graySpan.textContent).toBe(':dd:ee:01')
  })

  it('does not gray MAC bytes when privacy mode is off', () => {
    const data = [{
      interface: 'eth0',
      mac: 'aa:bb:cc:dd:ee:01',
      ipv4: [],
      ipv6: [],
      firstSeen: '2026-02-14T14:00:00Z',
      lastSeen: '2026-02-14T14:30:00Z',
    }]
    render(<NeighbourTable neighbours={data} privacyMode={false} />)
    const table = screen.getByRole('table')
    const macCell = table.querySelector('td.font-mono.text-accent-300')
    expect(macCell.querySelector('.text-radar-600')).toBeNull()
    expect(macCell.textContent).toContain('aa:bb:cc:dd:ee:01')
  })

  it('grays out last 2 groups of IPv6 link-local in privacy mode', () => {
    const data = [{
      interface: 'eth0',
      mac: 'aa:bb:cc:dd:ee:01',
      ipv4: [],
      ipv6: ['fe80::aabb:ccff:fe00:0', '2001:db8::1'],
      firstSeen: '2026-02-14T14:00:00Z',
      lastSeen: '2026-02-14T14:30:00Z',
    }]
    render(<NeighbourTable neighbours={data} privacyMode={true} />)
    const table = screen.getByRole('table')
    // The masked suffix of the link-local should be grayed
    const ipv6Cell = table.querySelectorAll('td.font-mono.text-xs')[0]
    const graySpan = ipv6Cell.querySelector('.text-radar-600')
    expect(graySpan).toBeInTheDocument()
    expect(graySpan.textContent).toBe('fe00:0')
    // Non-link-local address should have no graying
    expect(ipv6Cell.textContent).toContain('2001:db8::1')
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
