function SearchBar({ search, onSearchChange, interfaces, selectedInterface, onInterfaceChange }) {
  return (
    <div className="flex flex-col sm:flex-row gap-2 mb-4">
      <input
        type="text"
        placeholder="Search MAC or IP..."
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
        className="flex-1 bg-radar-900 border border-radar-700 rounded px-3 py-1.5 text-sm text-radar-100 placeholder-radar-500 focus:outline-none focus:border-accent-500"
      />
      <select
        value={selectedInterface}
        onChange={(e) => onInterfaceChange(e.target.value)}
        className="bg-radar-900 border border-radar-700 rounded px-3 py-1.5 text-sm text-radar-100 focus:outline-none focus:border-accent-500"
      >
        <option value="">All Interfaces</option>
        {interfaces.map((iface) => (
          <option key={iface} value={iface}>
            {iface}
          </option>
        ))}
      </select>
    </div>
  )
}

export default SearchBar
