import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useConfig } from './useConfig'

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('useConfig', () => {
  it('returns config when file exists', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ privacyMode: true }),
    })

    const { result } = renderHook(() => useConfig())

    await waitFor(() => {
      expect(result.current.privacyMode).toBe(true)
    })
  })

  it('returns defaults on 404', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
    })

    const { result } = renderHook(() => useConfig())

    await waitFor(() => {
      expect(result.current.privacyMode).toBe(false)
    })
  })

  it('returns defaults on fetch error', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new Error('network'))

    const { result } = renderHook(() => useConfig())

    await waitFor(() => {
      expect(result.current.privacyMode).toBe(false)
    })
  })
})
