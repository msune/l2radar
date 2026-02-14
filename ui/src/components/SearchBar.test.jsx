import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import SearchBar from './SearchBar'

afterEach(cleanup)

describe('SearchBar', () => {
  it('renders search input with placeholder', () => {
    render(
      <SearchBar
        search=""
        onSearchChange={() => {}}
        interfaces={[]}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    expect(screen.getByPlaceholderText('Search MAC or IP...')).toBeInTheDocument()
  })

  it('calls onSearchChange when typing', () => {
    const onChange = vi.fn()
    render(
      <SearchBar
        search=""
        onSearchChange={onChange}
        interfaces={[]}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    fireEvent.change(screen.getByPlaceholderText('Search MAC or IP...'), {
      target: { value: '192.168' },
    })
    expect(onChange).toHaveBeenCalledWith('192.168')
  })

  it('renders interface options', () => {
    render(
      <SearchBar
        search=""
        onSearchChange={() => {}}
        interfaces={['eth0', 'wlan0']}
        selectedInterface=""
        onInterfaceChange={() => {}}
      />
    )
    expect(screen.getByText('All Interfaces')).toBeInTheDocument()
    expect(screen.getByText('eth0')).toBeInTheDocument()
    expect(screen.getByText('wlan0')).toBeInTheDocument()
  })

  it('calls onInterfaceChange when selecting', () => {
    const onChange = vi.fn()
    render(
      <SearchBar
        search=""
        onSearchChange={() => {}}
        interfaces={['eth0', 'wlan0']}
        selectedInterface=""
        onInterfaceChange={onChange}
      />
    )
    fireEvent.change(screen.getByDisplayValue('All Interfaces'), {
      target: { value: 'eth0' },
    })
    expect(onChange).toHaveBeenCalledWith('eth0')
  })
})
