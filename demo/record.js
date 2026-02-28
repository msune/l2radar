/**
 * L2Radar demo recorder.
 *
 * Orchestrates the full demo pipeline:
 *   1. Start demo HTTP server (server.js).
 *   2. Start deterministic simulation (simulate.js).
 *   3. Launch headless Chromium with video recording.
 *   4. Pre-set the splash cookie so the splash screen is skipped.
 *   5. Navigate through the UI with realistic gestures.
 *   6. Save the recorded video as $DEMO_OUTPUT_DIR/output.webm.
 *
 * Usage:
 *   node demo/record.js
 *
 * Environment:
 *   DEMO_OUTPUT_DIR   Directory where output.webm is written.
 *                     Defaults to a system temp dir created automatically.
 *
 * The caller (Makefile / CI) is responsible for converting output.webm → assets/demo.gif.
 *
 * Gesture timeline (relative to browser navigation):
 *   t≈0s   Navigate; wait for table with initial 5 hosts.
 *   t≈3s   View "All" tab (5 hosts).
 *   t≈5s   Click "eth0" tab — InterfaceInfo + eth0 hosts.
 *   t≈9s   Click "wlan0" tab — InterfaceInfo + wlan0 hosts.
 *   t≈13s  Click "All" tab — watch host count grow.
 *   t≈15s  Type "192.168" in search bar.
 *   t≈17.5s Clear search.
 *   t≈19s  Passive observation — hosts grow from ~7 to 15.
 *   t≈37s  Click "eth0" briefly (9 eth0 hosts visible).
 *   t≈39s  Click "All" — final state.
 *   t≈42s  End recording.
 */

import { chromium }           from 'playwright'
import { mkdtempSync, renameSync } from 'fs'
import { join, dirname }      from 'path'
import { fileURLToPath }      from 'url'
import { tmpdir }             from 'os'
import { createServer }       from './server.js'
import { runSimulation }      from './simulate.js'

const __dirname   = dirname(fileURLToPath(import.meta.url))
const OUTPUT_DIR  = process.env.DEMO_OUTPUT_DIR ?? mkdtempSync(join(tmpdir(), 'l2radar-rec-'))
const PORT        = 3737

const sleep = ms => new Promise(r => setTimeout(r, ms))

async function main() {
  // ── 1. Temp data dir ──────────────────────────────────────────────────────
  const dataDir = mkdtempSync(join(tmpdir(), 'l2radar-demo-'))
  console.log(`[demo] data dir: ${dataDir}`)

  // ── 2. Start HTTP server ───────────────────────────────────────────────────
  const server = await createServer(dataDir, PORT)
  console.log(`[demo] server on http://localhost:${PORT}`)

  // ── 3. Start simulation ────────────────────────────────────────────────────
  const simStop = runSimulation(dataDir)
  console.log('[demo] simulation started')

  // Give the simulation one write cycle to produce the initial JSON files.
  await sleep(2_000)

  // ── 4. Launch browser ─────────────────────────────────────────────────────
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({
    recordVideo: {
      dir:  OUTPUT_DIR,
      size: { width: 1280, height: 720 },
    },
    // Pre-set the splash cookie so the splash screen doesn't block the demo.
    storageState: {
      cookies: [{
        name:     'l2radar_splash',
        value:    '1',
        domain:   'localhost',
        path:     '/',
        expires:  Math.floor(Date.now() / 1000) + 7_200,
        httpOnly: false,
        secure:   false,
        sameSite: 'Strict',
      }],
      origins: [],
    },
  })

  const page = await context.newPage()

  // ── 5. Navigate ───────────────────────────────────────────────────────────
  console.log('[demo] navigating...')
  await page.goto(`http://localhost:${PORT}`)

  // Wait for the neighbour table to render at least one row.
  await page.waitForSelector('table tbody tr', { timeout: 20_000 })
  console.log('[demo] table visible')

  // Let the initial 5 hosts settle on screen.
  await sleep(2_000)

  // ── 6. Gesture sequence ───────────────────────────────────────────────────

  // View "All" tab with initial 5 hosts.
  await sleep(2_000)

  // Click "eth0" tab.
  console.log('[demo] clicking eth0 tab')
  await page.click('button:has-text("eth0")')
  await sleep(4_000)

  // Click "wlan0" tab.
  console.log('[demo] clicking wlan0 tab')
  await page.click('button:has-text("wlan0")')
  await sleep(4_000)

  // Back to "All" tab.
  console.log('[demo] clicking All tab')
  await page.click('button:has-text("All")')
  await sleep(2_000)

  // Use the search bar — filter by IP prefix, then clear.
  console.log('[demo] using search bar')
  const searchInput = page.locator('input[placeholder="Search MAC or IP..."]')
  await searchInput.fill('192.168')
  await sleep(2_500)
  await searchInput.fill('')
  await sleep(1_500)

  // Passive observation: watch hosts grow from ~7 → 15.
  // Simulation milestones during this window:
  //   t≈20s: H7, H8 appear  → 11 hosts
  //   t≈24s: W5, W6 appear  → 13 hosts
  //   t≈28s: H9 appears     → 15 hosts (all visible)
  console.log('[demo] watching simulation...')
  await sleep(18_000)

  // Brief eth0 tour — now shows all 9 eth0 hosts.
  console.log('[demo] final eth0 tab')
  await page.click('button:has-text("eth0")')
  await sleep(2_000)

  // Return to "All" for the closing shot.
  await page.click('button:has-text("All")')
  await sleep(3_000)

  // ── 7. Finalise recording ─────────────────────────────────────────────────
  console.log('[demo] closing context...')
  const videoHandle = page.video()
  await context.close()
  const videoPath = await videoHandle.path()

  const outputPath = join(OUTPUT_DIR, 'output.webm')
  renameSync(videoPath, outputPath)
  console.log(`[demo] video saved to ${outputPath}`)
  // Print for callers that need to locate the file.
  console.log(`DEMO_WEBM=${outputPath}`)

  // ── 8. Cleanup ────────────────────────────────────────────────────────────
  await browser.close()
  simStop()
  server.close()
  console.log('[demo] done')
}

main().catch(err => {
  console.error('[demo] fatal error:', err)
  process.exit(1)
})
