import { useState, useMemo } from 'react'
import { useNeighbourData } from './hooks/useNeighbourData'
import { useUsername } from './hooks/useUsername'
import { useConfig } from './hooks/useConfig'
import { filterNeighbours, getInterfaces } from './lib/filter'
import { buildMacMapping, obfuscateData } from './lib/macObfuscation'
import { hasSplashCookie, setSplashCookie } from './lib/splash'
import SummaryStats from './components/SummaryStats'
import NeighbourTable from './components/NeighbourTable'
import SearchBar from './components/SearchBar'
import InterfaceTabs from './components/InterfaceTabs'
import InterfaceInfo from './components/InterfaceInfo'
import HeaderMenu from './components/HeaderMenu'
import PrivacyToggle from './components/PrivacyToggle'
import SplashScreen from './components/SplashScreen'
import logoSmall from '../../assets/img/logo_small.png'

function App() {
  const { neighbours, timestamps, interfaceInfo, loading, error } = useNeighbourData()
  const username = useUsername()
  const config = useConfig()
  const [privacyMode, setPrivacyMode] = useState(null)
  const [search, setSearch] = useState('')
  const [selectedInterface, setSelectedInterface] = useState('')
  const [splashDone, setSplashDone] = useState(hasSplashCookie)

  // Initialize privacyMode from config once loaded
  const isPrivacyMode = privacyMode !== null ? privacyMode : config.privacyMode

  // Obfuscate data when privacy mode is active
  const displayData = useMemo(() => {
    if (!isPrivacyMode) return { neighbours, interfaceInfo }
    const allMacs = [
      ...neighbours.map((n) => n.mac),
      ...Object.values(interfaceInfo).map((i) => i.mac).filter(Boolean),
    ]
    const mapping = buildMacMapping(allMacs)
    return obfuscateData(neighbours, interfaceInfo, mapping)
  }, [isPrivacyMode, neighbours, interfaceInfo])

  const interfaces = getInterfaces(displayData.neighbours)
  const filtered = filterNeighbours(displayData.neighbours, {
    search,
    iface: selectedInterface,
  })

  return (
    <div className="min-h-screen bg-radar-950 text-radar-100 flex flex-col">
      {!splashDone && <SplashScreen onDone={() => { setSplashCookie(); setSplashDone(true) }} />}
      <header className="bg-radar-900 border-b border-radar-700 px-4 py-1 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <a
            href="/"
          >
              <img src={logoSmall} alt="L2 Radar" className="h-11" />
          </a>
          {error && (
            <span className="text-xs text-red-400">Connection error</span>
          )}
          {loading && (
            <span className="text-xs text-radar-500">Loading...</span>
          )}
        </div>
        <div className="flex items-center gap-6">
          <PrivacyToggle enabled={isPrivacyMode} onToggle={() => setPrivacyMode((d) => d !== null ? !d : !config.privacyMode)} />
          <HeaderMenu username={username} />
        </div>
      </header>
      <main className="flex-1 p-4">
        <div className="max-w-screen-2xl mx-auto">
        <SummaryStats neighbours={filtered} />
        <InterfaceTabs
          interfaces={interfaces}
          selectedInterface={selectedInterface}
          onInterfaceChange={setSelectedInterface}
        />
        {selectedInterface && (
          <InterfaceInfo
            name={selectedInterface}
            timestamp={timestamps[selectedInterface]}
            info={displayData.interfaceInfo[selectedInterface]}
          />
        )}
        <SearchBar
          search={search}
          onSearchChange={setSearch}
        />
        <NeighbourTable neighbours={filtered} showInterface={!selectedInterface} />
        </div>
      </main>
      <footer className="bg-radar-800 px-4 py-2 text-center text-xs text-radar-500">
        © {new Date().getFullYear()}{' '}
	<a
	  href="https://github.com/msune/l2radar"
	  target="_blank"
	  rel="noopener noreferrer"
          className="text-accent-400 hover:text-accent-300"
	>
	  L2 Radar
	</a>,
	powered by{' '}
	<a
	  href="https://ebpf.io/"
	  target="_blank"
	  rel="noopener noreferrer"
	>
	  🐝
	</a> 
	· Made with ❤️  from Barcelona ·{' '}
        <a
          href="https://github.com/msune"
          target="_blank"
          rel="noopener noreferrer"
          className="text-accent-400 hover:text-accent-300"
        >
          GitHub
        </a>
      </footer>
    </div>
  )
}

export default App
