import { Routes, Route, Navigate } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { ComparePage } from './pages/compare'
import { UsagePage } from './pages/UsagePage'
import { ConversationPage } from './pages/ConversationPage'
import { ModelConfigProvider } from './contexts/ModelConfigContext.tsx'
import { ConversationProvider } from './contexts/ConversationContext.tsx'
import { ActionsProvider } from './contexts/ActionsContext.tsx'
import { SidebarProvider } from './contexts/SidebarContext.tsx'
import { SettingsProvider } from './contexts/SettingsContext.tsx'
import { Sidebar } from './components/Sidebar.tsx'

function App() {
  return (
    <SettingsProvider>
      <ModelConfigProvider>
        <ConversationProvider>
          <SidebarProvider>
            <ActionsProvider>
              <div className="relative">
                <Routes>
                  <Route path="/" element={<HomePage />} />
                  <Route path="/home" element={<HomePage />} />
                  <Route path="/conversation/:id" element={<ConversationPage />} />
                  <Route path="/compare" element={<ComparePage />} />
                  <Route path="/usage" element={<UsagePage />} />
                  <Route path="/*" element={<Navigate to="/" />} />
                </Routes>
                <Sidebar />
              </div>
            </ActionsProvider>
          </SidebarProvider>
        </ConversationProvider>
      </ModelConfigProvider>
    </SettingsProvider>
  )
}

export default App