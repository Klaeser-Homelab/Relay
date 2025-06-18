import { useState, useEffect } from 'react';
import { Settings, Save, CheckCircle, XCircle, Folder, User, Shield } from 'lucide-react';
import type { GitConfig } from '../types/api';

interface GitConfigProps {
  onConfigUpdated?: (config: GitConfig) => void;
}

export function GitConfig({ onConfigUpdated }: GitConfigProps) {
  const [config, setConfig] = useState<GitConfig>({
    baseDirectory: '/home/relay/projects',
    gitUsername: '',
    hasToken: false
  });
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const response = await fetch('/api/config/git');
      if (response.ok) {
        const gitConfig = await response.json();
        setConfig(gitConfig);
      } else {
        throw new Error('Failed to load configuration');
      }
    } catch (error) {
      console.error('Failed to load Git configuration:', error);
      setMessage({ type: 'error', text: 'Failed to load Git configuration' });
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    setSaving(true);
    setMessage(null);
    
    try {
      const response = await fetch('/api/config/git', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          baseDirectory: config.baseDirectory,
          gitUsername: config.gitUsername
        }),
      });

      if (response.ok) {
        const updatedConfig = await response.json();
        setConfig(updatedConfig);
        setMessage({ type: 'success', text: 'Configuration saved successfully' });
        onConfigUpdated?.(updatedConfig);
      } else {
        throw new Error('Failed to save configuration');
      }
    } catch (error) {
      console.error('Failed to save Git configuration:', error);
      setMessage({ type: 'error', text: 'Failed to save configuration' });
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field: keyof GitConfig, value: string) => {
    setConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  if (loading) {
    return (
      <div className="card">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-10 bg-gray-200 rounded mb-4"></div>
          <div className="h-10 bg-gray-200 rounded mb-4"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div 
        className="flex items-center justify-between cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <h3 className="text-lg font-semibold text-gray-900 flex items-center">
          <Settings className="w-5 h-5 mr-2" />
          Git Configuration
        </h3>
        <div className="flex items-center space-x-2">
          {config.hasToken ? (
            <div className="flex items-center text-green-600 text-sm">
              <CheckCircle className="w-4 h-4 mr-1" />
              Token configured
            </div>
          ) : (
            <div className="flex items-center text-red-600 text-sm">
              <XCircle className="w-4 h-4 mr-1" />
              No token
            </div>
          )}
          <span className="text-gray-400">
            {isExpanded ? '▼' : '▶'}
          </span>
        </div>
      </div>

      {isExpanded && (
        <div className="mt-6 space-y-4">
          {/* GitHub Token Status */}
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-start space-x-3">
              <Shield className="w-5 h-5 text-gray-500 mt-0.5" />
              <div>
                <h4 className="font-medium text-gray-900">GitHub Token</h4>
                <p className="text-sm text-gray-600 mt-1">
                  {config.hasToken 
                    ? 'GitHub token is configured via environment variables' 
                    : 'GitHub token is required for cloning repositories. Please set GH_TOKEN environment variable.'
                  }
                </p>
              </div>
            </div>
          </div>

          {/* Base Directory */}
          <div>
            <label htmlFor="baseDirectory" className="block text-sm font-medium text-gray-700 mb-2">
              <Folder className="w-4 h-4 inline mr-1" />
              Base Directory
            </label>
            <input
              type="text"
              id="baseDirectory"
              value={config.baseDirectory}
              onChange={(e) => handleInputChange('baseDirectory', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="/home/relay/projects"
            />
            <p className="text-xs text-gray-500 mt-1">
              Directory where repositories will be cloned on the server
            </p>
          </div>

          {/* Git Username */}
          <div>
            <label htmlFor="gitUsername" className="block text-sm font-medium text-gray-700 mb-2">
              <User className="w-4 h-4 inline mr-1" />
              GitHub Username
            </label>
            <input
              type="text"
              id="gitUsername"
              value={config.gitUsername}
              onChange={(e) => handleInputChange('gitUsername', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="your-github-username"
            />
            <p className="text-xs text-gray-500 mt-1">
              Your GitHub username for Git operations
            </p>
          </div>

          {/* Message */}
          {message && (
            <div className={`rounded-lg p-3 ${
              message.type === 'success' 
                ? 'bg-green-50 text-green-800 border border-green-200' 
                : 'bg-red-50 text-red-800 border border-red-200'
            }`}>
              <div className="flex items-center">
                {message.type === 'success' ? (
                  <CheckCircle className="w-4 h-4 mr-2" />
                ) : (
                  <XCircle className="w-4 h-4 mr-2" />
                )}
                {message.text}
              </div>
            </div>
          )}

          {/* Save Button */}
          <div className="flex justify-end">
            <button
              onClick={saveConfig}
              disabled={saving || !config.gitUsername.trim()}
              className="btn-primary flex items-center disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Save className="w-4 h-4 mr-2" />
              {saving ? 'Saving...' : 'Save Configuration'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}