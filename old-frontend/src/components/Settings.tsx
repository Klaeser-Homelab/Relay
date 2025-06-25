import { X, Settings as SettingsIcon } from 'lucide-react';
import { GitConfig } from './GitConfig';

interface SettingsProps {
  isOpen: boolean;
  onClose: () => void;
  onConfigUpdated?: () => void;
}

export function Settings({ isOpen, onClose, onConfigUpdated }: SettingsProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-lg shadow-xl w-11/12 max-w-2xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-700">
          <div className="flex items-center space-x-2">
            <SettingsIcon className="w-5 h-5 text-gray-400" />
            <h2 className="text-xl font-semibold text-white">Settings</h2>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-200 text-xl font-bold"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 p-6 overflow-y-auto">
          <div className="space-y-6">
            {/* Git Configuration Section */}
            <div>
              <h3 className="text-lg font-medium text-white mb-4">Git Configuration</h3>
              <div className="bg-gray-700 rounded-lg p-4">
                <GitConfig onConfigUpdated={onConfigUpdated} />
              </div>
            </div>

            {/* Future settings sections can be added here */}
            <div className="border-t border-gray-600 pt-6">
              <h3 className="text-lg font-medium text-white mb-2">About</h3>
              <p className="text-sm text-gray-300">
                Relay Voice - Voice-controlled GitHub repository management powered by OpenAI Realtime API
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 bg-gray-700 border-t border-gray-600 rounded-b-lg">
          <div className="flex justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-200 bg-gray-600 border border-gray-500 rounded-md hover:bg-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}