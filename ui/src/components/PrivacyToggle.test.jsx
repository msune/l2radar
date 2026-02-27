import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import PrivacyToggle from './PrivacyToggle'

afterEach(cleanup)

describe('PrivacyToggle', () => {
  it('renders as a toggle switch', () => {
    render(<PrivacyToggle enabled={false} onToggle={() => {}} />)
    expect(screen.getByRole('switch')).toBeInTheDocument()
  })

  it('shows Eye icon when disabled (MACs visible)', () => {
    const { container } = render(<PrivacyToggle enabled={false} onToggle={() => {}} />)
    expect(container.querySelector('[data-testid="icon-eye"]')).toBeInTheDocument()
  })

  it('shows VenetianMask icon when enabled (MACs hidden)', () => {
    const { container } = render(<PrivacyToggle enabled={true} onToggle={() => {}} />)
    expect(container.querySelector('[data-testid="icon-mask"]')).toBeInTheDocument()
  })

  it('reflects enabled state via aria-checked', () => {
    render(<PrivacyToggle enabled={true} onToggle={() => {}} />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  })

  it('reflects disabled state via aria-checked', () => {
    render(<PrivacyToggle enabled={false} onToggle={() => {}} />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')
  })

  it('calls onToggle on click', () => {
    const onToggle = vi.fn()
    render(<PrivacyToggle enabled={false} onToggle={onToggle} />)
    fireEvent.click(screen.getByRole('switch'))
    expect(onToggle).toHaveBeenCalledTimes(1)
  })
})
