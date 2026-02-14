import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
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

  it('renders last update timestamp', () => {
    render(
      <InterfaceInfo
        name="eth0"
        timestamp="2026-02-14T14:30:00Z"
        info={{ mac: '', ipv4: [], ipv6: [] }}
      />
    )
    expect(screen.getByText(/Last update/)).toBeInTheDocument()
    expect(screen.getByText(/14:30:00 UTC/)).toBeInTheDocument()
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
})
