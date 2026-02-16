import { describe, it, expect, beforeEach } from 'vitest'
import { hasSplashCookie, setSplashCookie } from './splash'

describe('splash cookie', () => {
  beforeEach(() => {
    // Clear all cookies
    document.cookie.split(';').forEach((c) => {
      const name = c.split('=')[0].trim()
      if (name) document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`
    })
  })

  it('returns false when no cookie is set', () => {
    expect(hasSplashCookie()).toBe(false)
  })

  it('returns true after setting the cookie', () => {
    setSplashCookie()
    expect(hasSplashCookie()).toBe(true)
  })

  it('sets cookie with l2radar_splash name', () => {
    setSplashCookie()
    expect(document.cookie).toContain('l2radar_splash=1')
  })
})
