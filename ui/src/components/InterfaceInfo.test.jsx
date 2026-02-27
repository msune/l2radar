import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent } from '@testing-library/react'
import InterfaceInfo from './InterfaceInfo'

afterEach(cleanup)

describe('InterfaceInfo', () => {
  it('renders interface name', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [] }}
      />
    )
    expect(screen.getByText('eth0')).toBeInTheDocument()
  })

  it('renders last update as relative time', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [] }}
      />
    )
    expect(screen.getByText(/Last update/)).toBeInTheDocument()
    // timeago.js renders relative time like "X hours ago", "X days ago", etc.
    expect(screen.getByText(/ago/)).toBeInTheDocument()
  })

  it('renders MAC address', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: 'de:ad:be:ef:00:01', ipv4: [], ipv6: [] }}
      />
    )
    expect(screen.getByText('de:ad:be:ef:00:01')).toBeInTheDocument()
  })

  it('renders IPv4 addresses', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: ['192.168.1.10', '10.0.0.1'], ipv6: [] }}
      />
    )
    expect(screen.getByText('192.168.1.10')).toBeInTheDocument()
    expect(screen.getByText('10.0.0.1')).toBeInTheDocument()
  })

  it('renders IPv6 addresses', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: ['fe80::1', '2001:db8::1'] }}
      />
    )
    expect(screen.getByText('fe80::1')).toBeInTheDocument()
    expect(screen.getByText('2001:db8::1')).toBeInTheDocument()
  })

  it('renders nothing when no info is provided', () => {
    const { container } = render(
      <InterfaceInfo name="eth0" timestamp="" info={null} />
    )
    expect(container.innerHTML).toBe('')
  })

  it('shows dash for empty MAC', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [] }}
      />
    )
    const macLabel = screen.getByText('MAC')
    expect(macLabel.closest('div').textContent).toContain('—')
  })

  it('renders fields in order: Interface, MAC, IPv4, IPv6, Last update', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: 'de:ad:be:ef:00:01', ipv4: ['10.0.0.1'], ipv6: ['fe80::1'] }}
      />
    )
    const labels = screen.getAllByText(/^(Interface|MAC|IPv4|IPv6|Last update)$/)
    const order = labels.map((el) => el.textContent)
    expect(order).toEqual(['Interface', 'MAC', 'IPv4', 'IPv6', 'Last update'])
  })

  it('shows stats toggle button when stats are present', () => {
    const stats = {
      tx_bytes: 1234567, rx_bytes: 7890123,
      tx_packets: 10000, rx_packets: 20000,
      tx_errors: 0, rx_errors: 1,
      tx_dropped: 0, rx_dropped: 0,
    }
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [], stats }}
      />
    )
    expect(screen.getByText(/Interface Stats/)).toBeInTheDocument()
  })

  it('stats are collapsed by default', () => {
    const stats = {
      tx_bytes: 1234567, rx_bytes: 7890123,
      tx_packets: 10000, rx_packets: 20000,
      tx_errors: 0, rx_errors: 0,
      tx_dropped: 0, rx_dropped: 0,
    }
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [], stats }}
      />
    )
    // Stats values should not be visible when collapsed
    expect(screen.queryByText('TX Bytes')).not.toBeInTheDocument()
  })

  it('expands stats on click', () => {
    const stats = {
      tx_bytes: 1234567, rx_bytes: 7890123,
      tx_packets: 10000, rx_packets: 20000,
      tx_errors: 0, rx_errors: 1,
      tx_dropped: 2, rx_dropped: 0,
    }
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [], stats }}
      />
    )
    fireEvent.click(screen.getByText(/Interface Stats/))

    // All stat labels should be visible
    expect(screen.getByText('TX Bytes')).toBeInTheDocument()
    expect(screen.getByText('RX Bytes')).toBeInTheDocument()
    expect(screen.getByText('TX Packets')).toBeInTheDocument()
    expect(screen.getByText('RX Packets')).toBeInTheDocument()
    expect(screen.getByText('TX Errors')).toBeInTheDocument()
    expect(screen.getByText('RX Errors')).toBeInTheDocument()
    expect(screen.getByText('TX Dropped')).toBeInTheDocument()
    expect(screen.getByText('RX Dropped')).toBeInTheDocument()
  })

  it('formats bytes in human-readable form', () => {
    const stats = {
      tx_bytes: 1536, rx_bytes: 2621440,
      tx_packets: 10, rx_packets: 20,
      tx_errors: 0, rx_errors: 0,
      tx_dropped: 0, rx_dropped: 0,
    }
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [], stats }}
      />
    )
    fireEvent.click(screen.getByText(/Interface Stats/))

    // 1536 bytes = 1.5 KB, 2621440 bytes = 2.5 MB
    expect(screen.getByText('1.5 KB')).toBeInTheDocument()
    expect(screen.getByText('2.5 MB')).toBeInTheDocument()
  })

  it('grays out last 3 MAC bytes in privacy mode', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: 'de:ad:be:ef:00:01', ipv4: [], ipv6: [] }}
        privacyMode={true}
      />
    )
    const macLabel = screen.getByText('MAC')
    const macContainer = macLabel.closest('div')
    const graySpan = macContainer.querySelector('.text-radar-600')
    expect(graySpan).toBeInTheDocument()
    expect(graySpan.textContent).toBe(':ef:00:01')
  })

  it('does not gray MAC bytes when privacy mode is off', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: 'de:ad:be:ef:00:01', ipv4: [], ipv6: [] }}
        privacyMode={false}
      />
    )
    const macLabel = screen.getByText('MAC')
    const macContainer = macLabel.closest('div')
    expect(macContainer.querySelector('.text-radar-600')).toBeNull()
  })

  it('grays out last 2 groups of IPv6 link-local in privacy mode', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: ['fe80::aabb:ccff:fe00:0'] }}
        privacyMode={true}
      />
    )
    const ipv6Label = screen.getByText('IPv6')
    const ipv6Container = ipv6Label.closest('div')
    const graySpan = ipv6Container.querySelector('.text-radar-600')
    expect(graySpan).toBeInTheDocument()
    expect(graySpan.textContent).toBe('fe00:0')
  })

  it('does not gray non-link-local IPv6 in privacy mode', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: ['2001:db8::1'] }}
        privacyMode={true}
      />
    )
    const ipv6Label = screen.getByText('IPv6')
    const ipv6Container = ipv6Label.closest('div')
    expect(ipv6Container.querySelector('.text-radar-600')).toBeNull()
  })

  it('does not show stats toggle when stats is null', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [], stats: null }}
      />
    )
    expect(screen.queryByText(/Interface Stats/)).not.toBeInTheDocument()
  })
})
