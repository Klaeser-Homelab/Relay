import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Menu } from 'lucide-react';
import { useSidebar } from '../contexts/SidebarContext';
import { Sidebar } from './Sidebar';

interface HeaderProps {
  title?: string;
  conversationId?: string;
  onNewConversation?: () => void;
}

export const Header: React.FC<HeaderProps> = ({ 
  title, 
  conversationId, 
  onNewConversation 
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { toggleSidebar } = useSidebar();

  const getTitle = () => {
    if (title) return title;
    return 'Relay';
  };

  return (
    <>
      <div className="p-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              onClick={toggleSidebar}
              className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
              title="Toggle sidebar"
            >
              <Menu className="w-5 h-5" />
            </button>
            <div>
              <h1 className="text-white text-xl font-bold">{getTitle()}</h1>
              {conversationId && (
                <p className="text-gray-400 text-xs">
                  ID: {conversationId.slice(0, 20)}...
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            {onNewConversation && (
              <button
                onClick={onNewConversation}
                className="px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 flex items-center gap-1"
                title="Start new conversation"
              >
                <span className="text-lg">+</span>
                New
              </button>
            )}
          </div>
        </div>
      </div>
      <Sidebar />
    </>
  );
};