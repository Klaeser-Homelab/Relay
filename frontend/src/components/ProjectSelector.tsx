import { useState } from 'react';
import { Search, Star, GitFork, Lock, Globe, Calendar } from 'lucide-react';
import type { GitHubRepository } from '../types/api';

interface ProjectSelectorProps {
  projects: GitHubRepository[];
  selectedProject: GitHubRepository | null;
  loading: boolean;
  error: string | null;
  onSelectProject: (project: GitHubRepository) => Promise<boolean>;
  onRefresh: () => void;
}

export function ProjectSelector({ 
  projects, 
  selectedProject, 
  loading, 
  error, 
  onSelectProject,
  onRefresh 
}: ProjectSelectorProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterLanguage, setFilterLanguage] = useState('');

  const filteredProjects = projects.filter(project => {
    const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.fullName.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesLanguage = !filterLanguage || project.language === filterLanguage;
    
    return matchesSearch && matchesLanguage;
  });

  const languages = Array.from(new Set(projects.map(p => p.language).filter(Boolean))) as string[];


  if (error) {
    return (
      <div className="card">
        <div className="text-center">
          <div className="text-red-600 mb-4">
            <h3 className="text-lg font-semibold">Error Loading Projects</h3>
            <p className="text-sm mt-2">{error}</p>
          </div>
          <button 
            onClick={onRefresh}
            className="btn-primary"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="card">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">
            Select GitHub Repository
          </h2>
          <button
            onClick={onRefresh}
            disabled={loading}
            className="btn-secondary"
          >
            {loading ? 'Loading...' : 'Refresh'}
          </button>
        </div>

        {/* Search and Filters */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search repositories..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          
          <select
            value={filterLanguage}
            onChange={(e) => setFilterLanguage(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Languages</option>
            {languages.map(lang => (
              <option key={lang} value={lang}>{lang}</option>
            ))}
          </select>
        </div>

        {/* Selected Project */}
        {selectedProject && (
          <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="text-sm font-medium text-blue-800 mb-1">
              Currently Selected
            </h3>
            <div className="text-blue-900 font-semibold">
              {selectedProject.fullName}
            </div>
          </div>
        )}

        {/* Project Grid */}
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="bg-gray-200 rounded-lg h-48"></div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredProjects.map((project) => (
              <ProjectCard
                key={project.fullName}
                project={project}
                isSelected={selectedProject?.fullName === project.fullName}
                onSelect={() => onSelectProject(project)}
              />
            ))}
          </div>
        )}

        {!loading && filteredProjects.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            {searchTerm || filterLanguage ? 'No repositories match your filters.' : 'No repositories found.'}
          </div>
        )}
      </div>
    </div>
  );
}

interface ProjectCardProps {
  project: GitHubRepository;
  isSelected: boolean;
  onSelect: () => void;
}

function ProjectCard({ project, isSelected, onSelect }: ProjectCardProps) {
  return (
    <div 
      className={`relative border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md ${
        isSelected 
          ? 'border-blue-500 bg-blue-50 shadow-md' 
          : 'border-gray-200 bg-white hover:border-gray-300'
      }`}
      onClick={onSelect}
    >
      {isSelected && (
        <div className="absolute top-2 right-2">
          <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
        </div>
      )}
      
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center space-x-2">
          {project.isPrivate ? (
            <Lock className="w-4 h-4 text-gray-500" />
          ) : (
            <Globe className="w-4 h-4 text-gray-500" />
          )}
          <h3 className="font-semibold text-gray-900 truncate">
            {project.name}
          </h3>
        </div>
      </div>

      <p className="text-sm text-gray-600 mb-3 line-clamp-2">
        {project.description || 'No description available'}
      </p>

      <div className="flex items-center justify-between text-xs text-gray-500 mb-3">
        <div className="flex items-center space-x-3">
          {project.language && (
            <span className="inline-flex items-center px-2 py-1 rounded-full bg-gray-100 text-gray-700">
              {project.language}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-1">
            <Star className="w-3 h-3" />
            <span>{project.stars}</span>
          </div>
          <div className="flex items-center space-x-1">
            <GitFork className="w-3 h-3" />
            <span>{project.forks}</span>
          </div>
        </div>
        
        <div className="flex items-center space-x-1">
          <Calendar className="w-3 h-3" />
          <span>{new Date(project.lastOpened).toLocaleDateString()}</span>
        </div>
      </div>
    </div>
  );
}