import { useState, useEffect } from 'react';
import type { GitHubRepository, RepositoryStatus } from '../types/api';

export function useGitHubProjects() {
  const [projects, setProjects] = useState<GitHubRepository[]>([]);
  const [selectedProject, setSelectedProject] = useState<GitHubRepository | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProjects = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/projects');
      if (!response.ok) {
        throw new Error('Failed to fetch projects');
      }
      
      const data = await response.json();
      setProjects(data.projects || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const selectProject = async (project: GitHubRepository) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/select`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error('Failed to select project');
      }
      
      setSelectedProject(project);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const getProjectStatus = async (projectName: string): Promise<RepositoryStatus | null> => {
    try {
      const response = await fetch(`/api/projects/${encodeURIComponent(projectName)}/status`);
      if (!response.ok) {
        throw new Error('Failed to get project status');
      }
      
      return await response.json();
    } catch (err) {
      console.error('Failed to get project status:', err);
      return null;
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  return {
    projects,
    selectedProject,
    loading,
    error,
    fetchProjects,
    selectProject,
    getProjectStatus
  };
}