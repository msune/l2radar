import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import SearchBar from './SearchBar'

afterEach(cleanup)

describe('SearchBar', () => {
  it('renders search input with placeholder', () => {
    render(<SearchBar search="" onSearchChange={() => {}} />)
    expect(screen.getByPlaceholderText('Search MAC or IP...')).toBeInTheDocument()
  })

  it('calls onSearchChange when typing', () => {
    const onChange = vi.fn()
    render(<SearchBar search="" onSearchChange={onChange} />)
    fireEvent.change(screen.getByPlaceholderText('Search MAC or IP...'), {
      target: { value: '192.168' },
    })
    expect(onChange).toHaveBeenCalledWith('192.168')
  })

  it('does not render interface dropdown', () => {
    render(<SearchBar search="" onSearchChange={() => {}} />)
    expect(screen.queryByRole('combobox')).toBeNull()
  })
})
