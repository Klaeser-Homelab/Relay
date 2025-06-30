import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import api from '../config/api'

interface Framework {
  name: string
  description: string
}

interface SettingsContextType {
  // Agent framework settings
  selectedFramework: string
  availableFrameworks: Framework[]
  frameworksLoading: boolean
  
  // Actions
  setSelectedFramework: (framework: string) => void
  refreshFrameworks: () => Promise<void>
}

const SettingsContext = createContext<SettingsContextType | undefined>(undefined)

export function useSettings(): SettingsContextType {
  const context = useContext(SettingsContext)
  if (!context) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}

interface SettingsProviderProps {
  children: ReactNode
}

export function SettingsProvider({ children }: SettingsProviderProps) {
  const [selectedFramework, setSelectedFramework] = useState<string>('openai_agents')
  const [availableFrameworks, setAvailableFrameworks] = useState<Framework[]>([])
  const [frameworksLoading, setFrameworksLoading] = useState(false)

  // Load saved framework preference from localStorage
  useEffect(() => {
    const savedFramework = localStorage.getItem('selectedFramework')
    if (savedFramework) {
      setSelectedFramework(savedFramework)
    }
  }, [])

  // Save framework preference to localStorage when it changes
  useEffect(() => {
    localStorage.setItem('selectedFramework', selectedFramework)
  }, [selectedFramework])

  // Fetch available frameworks from API
  const refreshFrameworks = async () => {
    try {
      setFrameworksLoading(true)
      const response = await api.get('/frameworks')
      if (response.status === 200) {
        setAvailableFrameworks(response.data)
      }
    } catch (error) {
      console.error('Failed to fetch frameworks:', error)
      // Set default frameworks as fallback
      setAvailableFrameworks([
        { name: 'openai_agents', description: 'OpenAI Agents SDK' },
        { name: 'langgraph', description: 'LangGraph Framework' }
      ])
    } finally {
      setFrameworksLoading(false)
    }
  }

  // Load frameworks on mount
  useEffect(() => {
    refreshFrameworks()
  }, [])

  const value: SettingsContextType = {
    selectedFramework,
    availableFrameworks,
    frameworksLoading,
    setSelectedFramework,
    refreshFrameworks
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}