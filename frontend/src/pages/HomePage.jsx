import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { RepositorySelector } from '../components/RepositorySelector'
import { Header } from '../components/Header'
import { api } from '../config/api'

export function HomePage() {
  const navigate = useNavigate()

  const handleSelectRepository = async (repository) => {
    try {
      // First select the repository
      const selectResponse = await api.post('/repositories/select', null, {
        params: { repository_name: repository.full_name }
      })
      
      if (selectResponse.data.success) {
        // Create a new conversation for this repository
        const conversationResponse = await api.post('/conversations', {
          title: `${repository.name} - ${new Date().toLocaleDateString()}`,
          repository_name: repository.full_name
        })
        
        if (conversationResponse.data && conversationResponse.data.id) {
          // Navigate to the conversation page
          navigate(`/conversation/${conversationResponse.data.id}`)
        }
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
      alert('Failed to create conversation. Please try again.')
    }
  }

  const handleCloneRepository = async (repository) => {
    try {
      const response = await api.post('/repositories/clone', {
        repository: repository
      })
      
      if (response.data.success) {
        alert('Repository cloned successfully!')
        return true
      } else {
        alert(`Failed to clone repository: ${response.data.message}`)
        return false
      }
    } catch (error) {
      console.error('Error cloning repository:', error)
      alert('Error cloning repository. Please try again.')
      return false
    }
  }

  return (
    <div className="min-h-screen bg-gray-800">
      <Header title="What next?" />

      <main className="pt-20 bg-gray-900 relative flex flex-col min-h-screen">
        <div className="flex-1 px-4 sm:px-6 lg:px-8 py-8">
          <div className="max-w-7xl mx-auto">
            <RepositorySelector
              onSelectRepository={handleSelectRepository}
              onCloneRepository={handleCloneRepository}
            />
          </div>
        </div>
      </main>
    </div>
  )
}