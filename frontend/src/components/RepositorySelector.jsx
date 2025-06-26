import { useState, useEffect } from 'react'
import { Search, Lock, Globe, Download, Loader, HardDrive, RefreshCw } from 'lucide-react'
import { api } from '../config/api'

export function RepositorySelector({ 
  onSelectRepository,
  onCloneRepository 
}) {
  const [repositories, setRepositories] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterLanguage, setFilterLanguage] = useState('')

  useEffect(() => {
    fetchRepositories()
  }, [])

  const fetchRepositories = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.get('/repositories/')
      setRepositories(response.data.repositories || [])
    } catch (err) {
      setError(err.message || 'Failed to load repositories')
    } finally {
      setLoading(false)
    }
  }

  const filteredRepositories = repositories.filter(repository => {
    const matchesSearch = repository.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         repository.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         repository.full_name.toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesLanguage = !filterLanguage || repository.language === filterLanguage
    
    return matchesSearch && matchesLanguage
  })

  const languages = Array.from(new Set(repositories.map(r => r.language).filter(Boolean)))

  if (error) {
    return (
      <div className="card">
        <div className="text-center">
          <div className="text-red-600 mb-4">
            <h3 className="text-lg font-semibold">Error Loading Repositories</h3>
            <p className="text-sm mt-2">{error}</p>
          </div>
          <button 
            onClick={fetchRepositories}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="bg-gray-900 rounded-lg p-6 shadow-xl">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-white">
            Select GitHub Repository
          </h2>
          <button
            onClick={fetchRepositories}
            disabled={loading}
            className="flex items-center space-x-2 px-4 py-2 bg-gray-700 text-gray-200 rounded-lg hover:bg-gray-600 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            <span>{loading ? 'Loading...' : 'Refresh'}</span>
          </button>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search repositories..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-gray-800 text-white border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder-gray-400"
            />
          </div>
          
          <select
            value={filterLanguage}
            onChange={(e) => setFilterLanguage(e.target.value)}
            className="px-4 py-2 bg-gray-800 text-white border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Languages</option>
            {languages.map(lang => (
              <option key={lang} value={lang}>{lang}</option>
            ))}
          </select>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="bg-gray-700 rounded-lg h-48"></div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredRepositories.map((repository) => (
              <RepositoryCard
                key={repository.full_name}
                repository={repository}
                onSelect={() => onSelectRepository(repository)}
                onClone={onCloneRepository ? () => onCloneRepository(repository) : undefined}
              />
            ))}
          </div>
        )}

        {!loading && filteredRepositories.length === 0 && (
          <div className="text-center py-12 text-gray-400">
            {searchTerm || filterLanguage ? 'No repositories match your filters.' : 'No repositories found.'}
          </div>
        )}
      </div>
    </div>
  )
}

function RepositoryCard({ repository, onSelect, onClone }) {
  const [isCloning, setIsCloning] = useState(false)

  const handleClone = async (e) => {
    e.stopPropagation()
    if (!onClone || isCloning) return
    
    setIsCloning(true)
    try {
      await onClone()
    } finally {
      setIsCloning(false)
    }
  }

  return (
    <div 
      className="relative border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md border-gray-600 bg-gray-800 hover:border-gray-500"
      onClick={onSelect}
    >
      <div className="flex items-center space-x-2 mb-3">
        {repository.private ? (
          <Lock className="w-4 h-4 text-gray-500" />
        ) : (
          <Globe className="w-4 h-4 text-gray-500" />
        )}
        <h3 className="font-semibold text-white truncate">
          {repository.name}
        </h3>
      </div>

      <p className="text-sm text-gray-300 mb-3 line-clamp-2">
        {repository.description || 'No description available'}
      </p>

      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center space-x-2">
          {repository.language && (
            <span className="inline-flex items-center px-2 py-1 rounded-full bg-gray-700 text-gray-200">
              {repository.language}
            </span>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          {repository.is_cloned ? (
            <div className="flex items-center text-green-600">
              <HardDrive className="w-3 h-3 mr-1" />
              Local
            </div>
          ) : onClone ? (
            <button
              onClick={handleClone}
              disabled={isCloning}
              className="flex items-center text-blue-300 hover:text-blue-100 px-2 py-1 rounded bg-blue-800 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Clone repository locally"
            >
              {isCloning ? (
                <Loader className="w-3 h-3 mr-1 animate-spin" />
              ) : (
                <Download className="w-3 h-3 mr-1" />
              )}
              {isCloning ? 'Cloning...' : 'Clone'}
            </button>
          ) : null}
        </div>
      </div>
    </div>
  )
}