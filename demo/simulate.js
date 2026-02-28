/**
 * L2Radar demo simulation.
 *
 * Writes neigh-eth0.json and neigh-wlan0.json to a target directory,
 * following the deterministic host timeline defined in hosts.js.
 *
 * Usage (module):
 *   import { runSimulation } from './simulate.js'
 *   const stop = runSimulation('/tmp/demo-data')
 *   // later:
 *   stop()
 *
 * Usage (CLI):
 *   node simulate.js [dataDir]
 */

import { writeFileSync, mkdirSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'
import { HOSTS, ETH0_IF, WLAN0_IF, TX_PER_STEP, RX_PER_STEP } from './hosts.js'

const WRITE_INTERVAL_MS = 2000   // write every 2 s (faster than the 5-s UI poll)
const SIMULATION_DURATION_MS = 42000  // 42 s covers all firstOffset values (max 28 s)

/**
 * Start the simulation.  Returns a stop() function.
 *
 * @param {string} dataDir  Directory to write JSON files into.
 * @returns {() => void}    Call to stop the simulation early.
 */
export function runSimulation(dataDir) {
  mkdirSync(dataDir, { recursive: true })

  // Simulation clock anchored to the real current time so that displayed
  // timestamps match the viewer's clock (avoids "N hours in the future").
  const SIM_START_REAL = Date.now()
  const SIM_BASE_TIME   = SIM_START_REAL

  // Per-host mutable state (bytes, packets, timestamps).
  const hostState = {}
  for (const h of HOSTS) {
    const firstSeenMs = SIM_BASE_TIME + h.firstOffset * 1000
    const firstSeenISO = new Date(firstSeenMs).toISOString()
    hostState[h.id] = {
      firstSeen:  firstSeenISO,
      lastSeen:   firstSeenISO,   // silent hosts never advance this
      txBytes:    100_000 + h.id.charCodeAt(0) * 1_000 + h.id.charCodeAt(1) * 100,
      rxBytes:    200_000 + h.id.charCodeAt(0) * 2_000 + h.id.charCodeAt(1) * 200,
      txPackets:  1_000,
      rxPackets:  2_000,
    }
  }

  // Per-interface stats (incremented on every write step).
  const ifStats = {
    eth0:  { txBytes: 5_000_000, rxBytes: 10_000_000, txPackets: 50_000, rxPackets: 100_000 },
    wlan0: { txBytes: 2_000_000, rxBytes:  4_000_000, txPackets: 20_000, rxPackets:  40_000 },
  }

  let running = true

  function step() {
    if (!running) return

    const elapsed = Date.now() - SIM_START_REAL
    const nowMs   = SIM_BASE_TIME + elapsed
    const now     = new Date(nowMs).toISOString()

    // Visible hosts at this point in time.
    const visible = HOSTS.filter(h => elapsed >= h.firstOffset * 1000)

    // Update active-host state deterministically.
    for (const h of visible) {
      if (!h.active) continue
      const s  = hostState[h.id]
      const tx = TX_PER_STEP[h.id] ?? 1_024
      const rx = RX_PER_STEP[h.id] ?? 2_048
      s.txBytes   += tx
      s.rxBytes   += rx
      s.txPackets += Math.floor(tx / 512)
      s.rxPackets += Math.floor(rx / 512)
      s.lastSeen   = now
    }

    // Advance interface counters.
    ifStats.eth0.txBytes   += 50_000
    ifStats.eth0.rxBytes   += 100_000
    ifStats.eth0.txPackets += 500
    ifStats.eth0.rxPackets += 1_000
    ifStats.wlan0.txBytes   += 20_000
    ifStats.wlan0.rxBytes   += 40_000
    ifStats.wlan0.txPackets += 200
    ifStats.wlan0.rxPackets += 400

    // Write JSON files.
    const eth0Visible  = visible.filter(h => h.iface === 'eth0')
    const wlan0Visible = visible.filter(h => h.iface === 'wlan0')
    writeIfFile(dataDir, ETH0_IF,  eth0Visible,  hostState, ifStats.eth0,  now)
    writeIfFile(dataDir, WLAN0_IF, wlan0Visible, hostState, ifStats.wlan0, now)

    if (elapsed < SIMULATION_DURATION_MS) {
      setTimeout(step, WRITE_INTERVAL_MS)
    }
  }

  step()
  return () => { running = false }
}

function writeIfFile(dataDir, ifCfg, hosts, hostState, stats, now) {
  const payload = {
    interface:       ifCfg.name,
    timestamp:       now,
    export_interval: '5s',
    mac:             ifCfg.mac,
    ipv4:            ifCfg.ipv4,
    ipv6:            ifCfg.ipv6,
    stats: {
      tx_bytes:   stats.txBytes,
      rx_bytes:   stats.rxBytes,
      tx_packets: stats.txPackets,
      rx_packets: stats.rxPackets,
      tx_errors:  0,
      rx_errors:  0,
      tx_dropped: 0,
      rx_dropped: 0,
    },
    neighbours: hosts.map(h => {
      const s = hostState[h.id]
      return {
        mac:        h.mac,
        ipv4:       h.ipv4,
        ipv6:       h.ipv6,
        first_seen: s.firstSeen,
        last_seen:  s.lastSeen,
      }
    }),
  }
  writeFileSync(join(dataDir, `neigh-${ifCfg.name}.json`), JSON.stringify(payload, null, 2))
}

// ── CLI entry point ────────────────────────────────────────────────────────────
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const dataDir = process.argv[2] ?? '/tmp/l2radar-demo-data'
  console.log(`Simulating to ${dataDir} (Ctrl+C to stop)`)
  const stop = runSimulation(dataDir)
  process.on('SIGINT', () => { stop(); process.exit(0) })
}
