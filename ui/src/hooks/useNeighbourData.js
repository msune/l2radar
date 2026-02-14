import { useState, useEffect, useCallback, useRef } from 'react'
import { mergeNeighbours } from '../lib/parseNeighbours'

const DATA_BASE_URL = '/data/'
const DEFAULT_POLL_INTERVAL = 5000

/**
 * Hook that discovers and polls JSON files from the data endpoint.
 * Uses If-Modified-Since to avoid re-downloading unchanged files.
 */
export function useNeighbourData(pollInterval = DEFAULT_POLL_INTERVAL) {
  const [neighbours, setNeighbours] = useState([])
  const [timestamps, setTimestamps] = useState({})
  const [interfaceInfo, setInterfaceInfo] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const lastModifiedRef = useRef({})

  const fetchData = useCallback(async () => {
    try {
      // Discover available JSON files via nginx autoindex
      const listResp = await fetch(DATA_BASE_URL)
      if (!listResp.ok) {
        throw new Error(`Failed to list data files: ${listResp.status}`)
      }

      const listing = await listResp.json()
      const jsonFiles = listing
        .filter((entry) => entry.name.endsWith('.json'))
        .map((entry) => entry.name)

      if (jsonFiles.length === 0) {
        setNeighbours([])
        setTimestamps({})
        setInterfaceInfo({})
        setLoading(false)
        return
      }

      // Fetch each JSON file with If-Modified-Since
      const dataArray = []
      const newLastModified = { ...lastModifiedRef.current }
      let anyUpdated = false

      for (const file of jsonFiles) {
        const url = `${DATA_BASE_URL}${file}`
        const headers = {}
        if (lastModifiedRef.current[file]) {
          headers['If-Modified-Since'] = lastModifiedRef.current[file]
        }

        const resp = await fetch(url, { headers })

        if (resp.status === 304) {
          // Not modified — use cached data (skip)
          continue
        }

        if (!resp.ok) {
          console.warn(`Failed to fetch ${file}: ${resp.status}`)
          continue
        }

        const lastMod = resp.headers.get('Last-Modified')
        if (lastMod) {
          newLastModified[file] = lastMod
        }

        const data = await resp.json()
        dataArray.push(data)
        anyUpdated = true
      }

      lastModifiedRef.current = newLastModified

      if (anyUpdated || loading) {
        // Re-fetch all files to get complete picture when any file changed
        const allData = []
        for (const file of jsonFiles) {
          const url = `${DATA_BASE_URL}${file}`
          const resp = await fetch(url)
          if (resp.ok) {
            allData.push(await resp.json())
          }
        }
        const merged = mergeNeighbours(allData)
        setNeighbours(merged.neighbours)
        setTimestamps(merged.timestamps)
        setInterfaceInfo(merged.interfaceInfo)
      }

      setError(null)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [loading])

  useEffect(() => {
    fetchData()
    const id = setInterval(fetchData, pollInterval)
    return () => clearInterval(id)
  }, [fetchData, pollInterval])

  return { neighbours, timestamps, interfaceInfo, loading, error }
}
