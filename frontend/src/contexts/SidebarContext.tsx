import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../config/api'

interface Conversation {
  id: string
  title: string
  repository_name?: string
  repository_full_name?: string
  updated_at: string
  last_accessed_at?: string
}

interface SidebarContextType {
  // Sidebar state
  isOpen: boolean
  toggleSidebar: () => void
  closeSidebar: () => void
  openSidebar: () => void
  
  // Feature toggles
  developerMode: boolean
  quietMode: boolean
  toggleDeveloperMode: () => void
  toggleQuietMode: () => void
  
  // Conversations
  recentConversations: Conversation[]
  loadingConversations: boolean
  fetchRecentConversations: () => Promise<void>
  
  // Settings
  openSettings: () => void
  
  // Project
  selectedProject: any | null
  setSelectedProject: (project: any) => void
}

const SidebarContext = createContext<SidebarContextType | undefined>(undefined)

export function useSidebar(): SidebarContextType {
  const context = useContext(SidebarContext)
  if (!context) {
    throw new Error('useSidebar must be used within a SidebarProvider')
  }
  return context
}

interface SidebarProviderProps {
  children: ReactNode
}

export function SidebarProvider({ children }: SidebarProviderProps) {
  const navigate = useNavigate()
  
  // Sidebar state
  const [isOpen, setIsOpen] = useState(false)
  
  // Feature toggles
  const [developerMode, setDeveloperMode] = useState(() => {
    const saved = localStorage.getItem('developerMode')
    return saved === 'true'
  })
  
  const [quietMode, setQuietMode] = useState(() => {
    const saved = localStorage.getItem('quietMode')
    return saved === 'true'
  })
  
  // Conversations
  const [recentConversations, setRecentConversations] = useState<Conversation[]>([])
  const [loadingConversations, setLoadingConversations] = useState(false)
  
  // Project
  const [selectedProject, setSelectedProject] = useState<any | null>(null)
  
  // Persist developer mode
  useEffect(() => {
    localStorage.setItem('developerMode', developerMode.toString())
  }, [developerMode])
  
  // Persist quiet mode
  useEffect(() => {
    localStorage.setItem('quietMode', quietMode.toString())
  }, [quietMode])
  
  // Sidebar functions
  const toggleSidebar = () => setIsOpen(!isOpen)
  const closeSidebar = () => setIsOpen(false)
  const openSidebar = () => setIsOpen(true)
  
  // Feature toggle functions
  const toggleDeveloperMode = () => setDeveloperMode(!developerMode)
  const toggleQuietMode = () => setQuietMode(!quietMode)
  
  // Fetch recent conversations
  const fetchRecentConversations = async () => {
    try {
      setLoadingConversations(true)
      const response = await api.get('/conversations/')
      
      if (response.status !== 200) {
        throw new Error('Failed to fetch conversations')
      }
      
      const data = await response.data
      
      // Handle different response structures
      let conversations = data
      if (data && data.conversations) {
        conversations = data.conversations
      } else if (data && data.data) {
        conversations = data.data
      }
      
      // Ensure we have an array
      if (!Array.isArray(conversations)) {
        console.warn('API response is not an array:', data)
        conversations = []
      }
      
      // Sort by last_accessed_at or updated_at
      const sorted = conversations.sort((a: Conversation, b: Conversation) => {
        const dateA = new Date(a.last_accessed_at || a.updated_at).getTime()
        const dateB = new Date(b.last_accessed_at || b.updated_at).getTime()
        return dateB - dateA
      })
      
      setRecentConversations(sorted)
    } catch (error) {
      console.error('Error fetching recent conversations:', error)
    } finally {
      setLoadingConversations(false)
    }
  }
  
  // Fetch conversations on mount
  useEffect(() => {
    fetchRecentConversations()
  }, [])
  
  // Settings function
  const openSettings = () => {
    closeSidebar()
    // You can implement settings navigation here
    console.log('Open settings')
  }
  
  const value: SidebarContextType = {
    // Sidebar state
    isOpen,
    toggleSidebar,
    closeSidebar,
    openSidebar,
    
    // Feature toggles
    developerMode,
    quietMode,
    toggleDeveloperMode,
    toggleQuietMode,
    
    // Conversations
    recentConversations,
    loadingConversations,
    fetchRecentConversations,
    
    // Settings
    openSettings,
    
    // Project
    selectedProject,
    setSelectedProject
  }
  
  return (
    <SidebarContext.Provider value={value}>
      {children}
    </SidebarContext.Provider>
  )
}