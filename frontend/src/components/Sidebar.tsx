import { useState } from 'react';
import { Settings, Code, X, Moon, Plus, MessageSquare } from 'lucide-react';

interface Chat {
  id: string;
  title: string;
  repository_name: string;
  repository_full_name: string;
  updated_at: string;
  last_accessed_at: string;
}

interface SidebarProps {
  developerMode: boolean;
  onToggleDeveloperMode: () => void;
  onOpenSettings: () => void;
  isOpen: boolean;
  onToggle: () => void;
  selectedProject: any;
  repositoryCount: number;
  quietMode: boolean;
  onToggleQuietMode: () => void;
  recentChats: Chat[];
  currentChatId: string | null;
  onSelectChat: (chat: Chat) => void;
  onNewChat: () => void;
}

export function Sidebar({ 
  developerMode, 
  onToggleDeveloperMode, 
  onOpenSettings, 
  isOpen, 
  onToggle, 
  selectedProject, 
  repositoryCount,
  quietMode,
  onToggleQuietMode,
  recentChats,
  currentChatId,
  onSelectChat,
  onNewChat
}: SidebarProps) {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="bg-gray-950 text-gray-200 w-64 min-h-screen flex flex-col border-r border-gray-700">
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center space-x-3">
        <div className="grid grid-cols-2 gap-1 w-4 h-4">
            <div className="w-1.5 h-1.5 bg-white rounded-full"></div>
            <div className="w-1.5 h-1.5 bg-green-500 rounded-full"></div>
            <div className="w-1.5 h-1.5 bg-orange-500 rounded-full"></div>
            <div className="w-1.5 h-1.5 bg-blue-500 rounded-full"></div>
          </div>
          <h1 className="text-2xl font-bold text-white">Relay</h1>
        </div>
      </div>
      
      {/* Recents Section */}
      <div className="flex-1 p-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wide">Recents</h3>
          <button
            onClick={onNewChat}
            className="p-1 text-gray-400 hover:text-gray-200 rounded hover:bg-gray-800 transition-colors"
            title="New chat"
          >
            <Plus className="w-4 h-4" />
          </button>
        </div>
        
        {/* Recent Chats List */}
        <div className="space-y-1 mb-6 max-h-64 overflow-y-auto">
          {recentChats.length > 0 ? (
            recentChats.map((chat) => (
              <button
                key={chat.id}
                onClick={() => onSelectChat(chat)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors group ${
                  currentChatId === chat.id
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-300 hover:text-white hover:bg-gray-800'
                }`}
              >
                <div className="flex items-start space-x-2">
                  <MessageSquare className="w-4 h-4 mt-0.5 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="font-medium truncate">{chat.title}</div>
                    <div className="text-xs text-gray-500 truncate">{chat.repository_name}</div>
                  </div>
                </div>
              </button>
            ))
          ) : (
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p className="text-xs">No recent chats</p>
              <p className="text-xs">Start a new conversation</p>
            </div>
          )}
        </div>
      </div>

      {/* Navigation */}
      <div className="px-4 pb-4">
        <nav className="space-y-2">
          <button
            onClick={onToggleDeveloperMode}
            className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md transition-colors ${
              developerMode
                ? 'bg-orange-600 text-white'
                : 'text-gray-300 hover:text-white hover:bg-gray-800'
            }`}
          >
            <Code className="w-5 h-5 flex-shrink-0" />
            <span>Dev Mode</span>
          </button>

          {/* Only show quiet mode when a project is selected */}
          {selectedProject && (
            <button
              onClick={onToggleQuietMode}
              className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md transition-colors ${
                quietMode
                  ? 'bg-purple-600 text-white'
                  : 'text-gray-300 hover:text-white hover:bg-gray-800'
              }`}
            >
              <Moon className="w-5 h-5 flex-shrink-0" />
              <span>Quiet Mode</span>
            </button>
          )}

         
        </nav>
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-gray-700">
      <button
            onClick={onOpenSettings}
            className="w-full flex items-center space-x-3 px-3 py-2 rounded-md text-gray-300 hover:text-white hover:bg-gray-800 transition-colors"
          >
            <Settings className="w-5 h-5 flex-shrink-0" />
            <span>Settings</span>
          </button>
      </div>
    </div>
  );
}