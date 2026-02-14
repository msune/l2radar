import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import InterfaceTabs from './InterfaceTabs'

afterEach(cleanup)

describe('InterfaceTabs', () => {
  it('renders All tab and one tab per interface', () => {
    render(
      <InterfaceTabs
        interfaces={['eth0', 'wlan0']}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    expect(screen.getByText('All')).toBeInTheDocument()
    expect(screen.getByText('eth0')).toBeInTheDocument()
    expect(screen.getByText('wlan0')).toBeInTheDocument()
  })

  it('highlights the active tab', () => {
    render(
      <InterfaceTabs
        interfaces={['eth0', 'wlan0']}
        selectedInterface="eth0"
        onInterfaceChange={() => {}}
      />
    )
    const eth0Button = screen.getByRole('button', { name: /eth0/i })
    const allButton = screen.getByRole('button', { name: /^All$/i })
    expect(eth0Button.className).toMatch(/accent/)
    expect(allButton.className).not.toMatch(/accent/)
  })

  it('inactive tabs have a border', () => {
    render(
      <InterfaceTabs
        interfaces={['eth0']}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    const eth0Button = screen.getByRole('button', { name: /eth0/i })
    expect(eth0Button.className).toMatch(/border/)
  })

  it('calls onInterfaceChange when clicking a tab', () => {
    const onChange = vi.fn()
    render(
      <InterfaceTabs
        interfaces={['eth0', 'wlan0']}
        selectedInterface=""
        onInterfaceChange={onChange}
      />
    )
    fireEvent.click(screen.getByText('eth0'))
    expect(onChange).toHaveBeenCalledWith('eth0')
  })

  it('calls onInterfaceChange with empty string when clicking All', () => {
    const onChange = vi.fn()
    render(
      <InterfaceTabs
        interfaces={['eth0']}
        selectedInterface="eth0"
        onInterfaceChange={onChange}
      />
    )
    fireEvent.click(screen.getByText('All'))
    expect(onChange).toHaveBeenCalledWith('')
  })

  it('does not show timestamps in tabs', () => {
    render(
      <InterfaceTabs
        interfaces={['eth0']}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    const eth0Button = screen.getByRole('button', { name: /eth0/i })
    expect(eth0Button.textContent).toBe('eth0')
  })

  it('renders only All tab when no interfaces exist', () => {
    render(
      <InterfaceTabs
        interfaces={[]}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    expect(screen.getByText('All')).toBeInTheDocument()
    expect(screen.getAllByRole('button')).toHaveLength(1)
  })
})
