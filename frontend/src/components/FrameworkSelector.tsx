import React from 'react'
import { useSettings } from '../contexts/SettingsContext'

export function FrameworkSelector() {
  const { selectedFramework, availableFrameworks, frameworksLoading, setSelectedFramework } = useSettings()

  if (frameworksLoading) {
    return (
      <div className="framework-selector">
        <span className="text-gray-500">Loading frameworks...</span>
      </div>
    )
  }

  return (
    <div className="framework-selector flex items-center gap-2">
      <label htmlFor="framework-select" className="text-sm text-gray-600">
        Agent:
      </label>
      <select
        id="framework-select"
        value={selectedFramework}
        onChange={(e) => setSelectedFramework(e.target.value)}
        className="text-sm px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        {availableFrameworks.map((framework) => (
          <option key={framework.name} value={framework.name}>
            {framework.description}
          </option>
        ))}
      </select>
    </div>
  )
}