import { useState, useRef, useEffect } from 'react'
import { User, Info, LogOut } from 'lucide-react'

function HeaderMenu({ username }) {
  const [open, setOpen] = useState(false)
  const menuRef = useRef(null)

  useEffect(() => {
    function handleClick(e) {
      if (menuRef.current && !menuRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  return (
    <div className="relative" ref={menuRef}>
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-1.5 text-sm text-radar-300 hover:text-radar-100 cursor-pointer"
      >
        <User size={16} />
        {username || '—'}
        <span className="text-xs">▾</span>
      </button>
      {open && (
        <div className="absolute right-0 top-full mt-1 bg-radar-800 border border-radar-700 rounded shadow-lg min-w-44 z-50 text-sm">
          <div className="flex items-center gap-2 px-3 py-2 text-radar-500 border-b border-radar-700">
            <Info size={14} />
            {__APP_VERSION__}
          </div>
          <a
            href="https://github.com/msune/l2radar"
            className="flex items-center gap-2 px-3 py-2 text-radar-300 hover:bg-radar-700 hover:text-radar-100"
          >
            <LogOut size={14} />
            Logout
          </a>
        </div>
      )}
    </div>
  )
}

export default HeaderMenu
