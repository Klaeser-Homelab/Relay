import { useState } from 'react';
import { Search, Lock, Globe, Download, Loader, HardDrive } from 'lucide-react';
import type { GitHubRepository } from '../types/api';

interface ProjectSelectorProps {
  projects: GitHubRepository[];
  selectedProject: GitHubRepository | null;
  loading: boolean;
  error: string | null;
  onSelectProject: (project: GitHubRepository) => Promise<boolean>;
  onRefresh: () => void;
  onCloneRepository?: (project: GitHubRepository) => Promise<boolean>;
}

export function ProjectSelector({ 
  projects, 
  selectedProject, 
  loading, 
  error, 
  onSelectProject,
  onRefresh,
  onCloneRepository 
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
          <h2 className="text-xl font-semibold text-white">
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

        {/* Selected Project */}
        {selectedProject && (
          <div className="mb-6 p-4 bg-blue-900 border border-blue-700 rounded-lg">
            <h3 className="text-sm font-medium text-blue-200 mb-1">
              Currently Selected
            </h3>
            <div className="text-blue-100 font-semibold">
              {selectedProject.fullName}
            </div>
          </div>
        )}

        {/* Project Grid */}
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
            {filteredProjects.map((project) => (
              <ProjectCard
                key={project.fullName}
                project={project}
                isSelected={selectedProject?.fullName === project.fullName}
                onSelect={() => onSelectProject(project)}
                onClone={onCloneRepository ? () => onCloneRepository(project) : undefined}
              />
            ))}
          </div>
        )}

        {!loading && filteredProjects.length === 0 && (
          <div className="text-center py-12 text-gray-400">
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
  onClone?: () => void;
}

function ProjectCard({ project, isSelected, onSelect, onClone }: ProjectCardProps) {
  const [isCloning, setIsCloning] = useState(false);

  const handleClone = async (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent selecting the project
    if (!onClone || isCloning) return;
    
    setIsCloning(true);
    try {
      await onClone();
    } finally {
      setIsCloning(false);
    }
  };

  return (
    <div 
      className={`relative border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md ${
        isSelected 
          ? 'border-blue-400 bg-blue-900 shadow-md' 
          : 'border-gray-600 bg-gray-800 hover:border-gray-500'
      }`}
      onClick={onSelect}
    >
      {isSelected && (
        <div className="absolute top-2 right-2">
          <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
        </div>
      )}
      
      <div className="flex items-center space-x-2 mb-3">
        {project.isPrivate ? (
          <Lock className="w-4 h-4 text-gray-500" />
        ) : (
          <Globe className="w-4 h-4 text-gray-500" />
        )}
        <h3 className="font-semibold text-white truncate">
          {project.name}
        </h3>
      </div>

      <p className="text-sm text-gray-300 mb-3 line-clamp-2">
        {project.description || 'No description available'}
      </p>

      {/* Bottom line: Language tag + Clone/Local status */}
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center space-x-2">
          {project.language && (
            <span className="inline-flex items-center px-2 py-1 rounded-full bg-gray-700 text-gray-200">
              {project.language}
            </span>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          {project.isCloned ? (
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
  );
}