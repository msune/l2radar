function SearchBar({ search, onSearchChange }) {
  return (
    <div className="mb-4">
      <input
        type="text"
        placeholder="Search MAC or IP..."
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
        className="w-full bg-radar-900 border border-radar-700 rounded px-3 py-1.5 text-sm text-radar-100 placeholder-radar-500 focus:outline-none focus:border-accent-500"
      />
    </div>
  )
}

export default SearchBar
