import { Layers, Cable } from 'lucide-react'

function InterfaceTabs({ interfaces, selectedInterface, onInterfaceChange }) {
  const isActive = (iface) =>
    iface === '' ? selectedInterface === '' : selectedInterface === iface

  const baseClasses =
    'px-3 py-1.5 text-sm rounded-t font-medium transition-colors cursor-pointer'
  const activeClasses = 'bg-radar-800 text-accent-400 border border-b-0 border-accent-400'
  const inactiveClasses = 'text-radar-400 hover:text-radar-200 hover:bg-radar-800/50 border border-radar-700'

  return (
    <div className="flex gap-1 overflow-x-auto border-b border-radar-700 mb-4">
      <button
        className={`${baseClasses} ${isActive('') ? activeClasses : inactiveClasses} flex items-center gap-1.5`}
        onClick={() => onInterfaceChange('')}
      >
        <Layers size={14} />
        All
      </button>
      {interfaces.map((iface) => (
        <button
          key={iface}
          className={`${baseClasses} ${isActive(iface) ? activeClasses : inactiveClasses} flex items-center gap-1.5`}
          onClick={() => onInterfaceChange(iface)}
        >
          <Cable size={14} />
          {iface}
        </button>
      ))}
    </div>
  )
}

export default InterfaceTabs
