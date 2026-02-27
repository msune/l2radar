import { Eye, VenetianMask } from 'lucide-react'

function PrivacyToggle({ enabled, onToggle }) {
  return (
    <button
      role="switch"
      aria-checked={enabled}
      aria-label="Privacy mode"
      onClick={onToggle}
      className="flex items-center gap-2 text-sm text-radar-300 hover:text-radar-100 cursor-pointer"
    >
      {enabled
        ? <VenetianMask size={16} data-testid="icon-mask" />
        : <Eye size={16} data-testid="icon-eye" />}
      <span
        className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
          enabled ? 'bg-accent-500' : 'bg-radar-700'
        }`}
      >
        <span
          className={`inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform ${
            enabled ? 'translate-x-4.5' : 'translate-x-0.5'
          }`}
        />
      </span>
    </button>
  )
}

export default PrivacyToggle
