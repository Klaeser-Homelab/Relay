import { useState, useMemo } from 'react';
import { Search, ArrowLeft, MessageSquare, Calendar, FolderOpen } from 'lucide-react';

interface Chat {
  id: string;
  title: string;
  repository_name: string;
  repository_full_name: string;
  updated_at: string;
  last_accessed_at: string;
}

interface AllChatsProps {
  chats: Chat[];
  currentChatId: string | null;
  onSelectChat: (chat: Chat) => void;
  onBack: () => void;
}

export function AllChats({ chats, currentChatId, onSelectChat, onBack }: AllChatsProps) {
  const [searchQuery, setSearchQuery] = useState('');

  // Filter chats based on search query
  const filteredChats = useMemo(() => {
    if (!searchQuery.trim()) return chats;
    
    const query = searchQuery.toLowerCase();
    return chats.filter(chat => 
      chat.title.toLowerCase().includes(query) ||
      chat.repository_name.toLowerCase().includes(query) ||
      chat.repository_full_name.toLowerCase().includes(query)
    );
  }, [chats, searchQuery]);

  // Group chats by date
  const groupedChats = useMemo(() => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000);
    const weekAgo = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
    const monthAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);

    const groups: { [key: string]: Chat[] } = {
      'Today': [],
      'Yesterday': [],
      'Previous 7 days': [],
      'Previous 30 days': [],
      'Older': []
    };

    filteredChats.forEach(chat => {
      const chatDate = new Date(chat.last_accessed_at || chat.updated_at);
      
      if (chatDate >= today) {
        groups['Today'].push(chat);
      } else if (chatDate >= yesterday) {
        groups['Yesterday'].push(chat);
      } else if (chatDate >= weekAgo) {
        groups['Previous 7 days'].push(chat);
      } else if (chatDate >= monthAgo) {
        groups['Previous 30 days'].push(chat);
      } else {
        groups['Older'].push(chat);
      }
    });

    // Remove empty groups
    Object.keys(groups).forEach(key => {
      if (groups[key].length === 0) {
        delete groups[key];
      }
    });

    return groups;
  }, [filteredChats]);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="h-full bg-gray-900 text-white">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-700">
        <div className="flex items-center space-x-3">
          <button
            onClick={onBack}
            className="p-2 text-gray-400 hover:text-gray-200 hover:bg-gray-800 rounded-md transition-colors"
            title="Back to main view"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <h1 className="text-xl font-semibold">All Chats</h1>
          <span className="text-sm text-gray-400">({chats.length})</span>
        </div>
      </div>

      {/* Search */}
      <div className="p-4 border-b border-gray-700">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
          <input
            type="text"
            placeholder="Search chats..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-gray-800 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Chat List */}
      <div className="flex-1 overflow-y-auto">
        {Object.keys(groupedChats).length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-gray-500">
            <MessageSquare className="w-12 h-12 mb-4 opacity-50" />
            <p className="text-lg font-medium">No chats found</p>
            <p className="text-sm">Try adjusting your search</p>
          </div>
        ) : (
          <div className="p-4 space-y-6">
            {Object.entries(groupedChats).map(([groupName, groupChats]) => (
              <div key={groupName}>
                {/* Group Header */}
                <div className="flex items-center space-x-2 mb-3">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <h2 className="text-sm font-medium text-gray-400 uppercase tracking-wide">
                    {groupName}
                  </h2>
                  <span className="text-xs text-gray-500">({groupChats.length})</span>
                </div>

                {/* Group Chats */}
                <div className="space-y-2">
                  {groupChats.map((chat) => (
                    <button
                      key={chat.id}
                      onClick={() => onSelectChat(chat)}
                      className={`w-full text-left p-3 rounded-lg transition-colors ${
                        currentChatId === chat.id
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 hover:text-white'
                      }`}
                    >
                      <div className="flex items-start justify-between space-x-3">
                        <div className="flex items-start space-x-3 flex-1 min-w-0">
                          <MessageSquare className="w-4 h-4 mt-1 flex-shrink-0" />
                          <div className="flex-1 min-w-0">
                            <div className="font-medium truncate mb-1">{chat.title}</div>
                            <div className="flex items-center space-x-2 text-xs text-gray-500">
                              <FolderOpen className="w-3 h-3" />
                              <span className="truncate">{chat.repository_name}</span>
                            </div>
                          </div>
                        </div>
                        <div className="text-xs text-gray-500 flex-shrink-0">
                          {formatDate(chat.last_accessed_at || chat.updated_at)}
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}