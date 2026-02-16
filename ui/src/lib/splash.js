const COOKIE_NAME = 'l2radar_splash'
const TTL_HOURS = 2

export function hasSplashCookie() {
  return document.cookie.split(';').some((c) => c.trim().startsWith(`${COOKIE_NAME}=`))
}

export function setSplashCookie() {
  const expires = new Date(Date.now() + TTL_HOURS * 60 * 60 * 1000).toUTCString()
  document.cookie = `${COOKIE_NAME}=1; expires=${expires}; path=/; SameSite=Strict`
}
