import { useState } from 'react'
import { useNeighbourData } from './hooks/useNeighbourData'
import { filterNeighbours, getInterfaces } from './lib/filter'
import SummaryStats from './components/SummaryStats'
import NeighbourTable from './components/NeighbourTable'
import SearchBar from './components/SearchBar'
import InterfaceTabs from './components/InterfaceTabs'
import InterfaceInfo from './components/InterfaceInfo'

function App() {
  const { neighbours, timestamps, interfaceInfo, loading, error } = useNeighbourData()
  const [search, setSearch] = useState('')
  const [selectedInterface, setSelectedInterface] = useState('')

  const interfaces = getInterfaces(neighbours)
  const filtered = filterNeighbours(neighbours, {
    search,
    iface: selectedInterface,
  })

  return (
    <div className="min-h-screen bg-radar-950 text-radar-100">
      <header className="bg-radar-900 border-b border-radar-700 px-4 py-3 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-accent-400">L2 Radar</h1>
        {error && (
          <span className="text-xs text-red-400">Connection error</span>
        )}
        {loading && (
          <span className="text-xs text-radar-500">Loading...</span>
        )}
      </header>
      <main className="p-4 max-w-screen-2xl mx-auto">
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
            info={interfaceInfo[selectedInterface]}
          />
        )}
        <SearchBar
          search={search}
          onSearchChange={setSearch}
        />
        <NeighbourTable neighbours={filtered} showInterface={!selectedInterface} />
      </main>
    </div>
  )
}

export default App
