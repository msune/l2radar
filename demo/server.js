/**
 * L2Radar demo HTTP server.
 *
 * Mimics the subset of the nginx config that the React UI depends on:
 *   GET /data/          nginx-style autoindex JSON (file listing)
 *   GET /data/<file>    JSON file with Last-Modified / 304 support
 *   GET /api/whoami     { "username": "demo" }
 *   GET /*              static files from ui/dist/ (SPA fallback)
 *
 * Usage (module):
 *   import { createServer } from './server.js'
 *   const server = await createServer('/tmp/demo-data', 3737)
 *   // later:
 *   server.close()
 *
 * Usage (CLI):
 *   node server.js [dataDir] [port]
 */

import express from 'express'
import { existsSync, statSync, readdirSync, readFileSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const UI_DIST   = join(__dirname, '..', 'ui', 'dist')

/**
 * Start the HTTP server.
 *
 * @param {string} dataDir  Directory that simulate.js writes JSON files into.
 * @param {number} port
 * @returns {Promise<import('http').Server>}
 */
export function createServer(dataDir, port = 3737) {
  const app = express()

  // ── GET /data/ ─── nginx autoindex JSON ──────────────────────────────────
  app.get('/data/', (_req, res) => {
    const files = existsSync(dataDir)
      ? readdirSync(dataDir).filter(f => f.endsWith('.json'))
      : []

    const listing = files.map(name => {
      const s = statSync(join(dataDir, name))
      return { name, type: 'file', mtime: s.mtime.toUTCString(), size: s.size }
    })

    res.setHeader('Content-Type', 'application/json')
    res.setHeader('Cache-Control', 'no-cache')
    res.json(listing)
  })

  // ── GET /data/:file ─── conditional JSON file serving ────────────────────
  app.get('/data/:file', (req, res) => {
    // Prevent path traversal.
    const name = req.params.file
    if (name.includes('/') || name.includes('..')) return res.status(400).end()

    const filePath = join(dataDir, name)
    if (!existsSync(filePath)) return res.status(404).end()

    const stat         = statSync(filePath)
    const lastModified = stat.mtime.toUTCString()

    const ifModSince = req.headers['if-modified-since']
    if (ifModSince) {
      const sinceMs = new Date(ifModSince).getTime()
      if (!isNaN(sinceMs) && stat.mtimeMs <= sinceMs) {
        return res.status(304).end()
      }
    }

    res.setHeader('Last-Modified', lastModified)
    res.setHeader('Content-Type', 'application/json')
    res.setHeader('Cache-Control', 'no-cache')
    res.send(readFileSync(filePath))
  })

  // ── GET /api/whoami ───────────────────────────────────────────────────────
  app.get('/api/whoami', (_req, res) => {
    res.json({ username: 'demo' })
  })

  // ── Static React app ──────────────────────────────────────────────────────
  app.use(express.static(UI_DIST))

  // SPA fallback: all unknown routes return index.html.
  app.get('*', (_req, res) => {
    res.sendFile(join(UI_DIST, 'index.html'))
  })

  return new Promise((resolve, reject) => {
    const server = app.listen(port, (err) => {
      if (err) reject(err)
      else resolve(server)
    })
  })
}

// ── CLI entry point ────────────────────────────────────────────────────────────
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const dataDir = process.argv[2] ?? '/tmp/l2radar-demo-data'
  const port    = parseInt(process.argv[3] ?? '3737', 10)
  createServer(dataDir, port).then(server => {
    console.log(`Demo server running on http://localhost:${port}`)
    console.log(`Serving data from: ${dataDir}`)
    console.log(`Serving UI from:   ${join(dirname(fileURLToPath(import.meta.url)), '..', 'ui', 'dist')}`)
    process.on('SIGINT', () => { server.close(); process.exit(0) })
  })
}
