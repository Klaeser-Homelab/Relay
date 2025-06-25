import { useEffect, useState } from 'react';

export function UsageStats({ model = null }) {
  console.log('UsageStats: Component rendering...');
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchStats = async () => {
    try {
      console.log('UsageStats: Fetching stats from API...');
      setLoading(true);
      const url = model 
        ? `http://localhost:8080/usage/stats?model=${model}`
        : 'http://localhost:8080/usage/stats';
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      console.log('UsageStats: Received data:', data);
      setStats(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch usage stats');
      console.error('Failed to fetch usage stats:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  if (loading) {
    return (
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-700 rounded w-1/4 mb-2"></div>
          <div className="grid grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-16 bg-gray-700 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-900 border border-red-700 rounded-lg p-4 mb-4">
        <div className="flex items-center space-x-2">
          <span className="text-red-300 text-sm">Usage stats unavailable: {error}</span>
        </div>
      </div>
    );
  }

  if (!stats) return null;

  const formatCost = (cost) => {
    return `$${Number(cost).toFixed(4)}`;
  };

  const formatTime = (seconds) => {
    return `${Number(seconds).toFixed(2)}s`;
  };

  const formatNumber = (num) => {
    return Number(num).toLocaleString();
  };

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-2">
          <span className="text-white text-sm font-medium">
            API Usage{model ? ` (${model})` : ''}
          </span>
        </div>
        <button
          onClick={fetchStats}
          className="text-gray-400 hover:text-gray-200 text-xs px-2 py-1 rounded hover:bg-gray-700 transition-colors"
        >
          Refresh
        </button>
      </div>
      
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Requests</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatNumber(stats.total_requests)}
          </div>
          <div className="text-gray-400 text-xs">
            {stats.successful_requests} success
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Tokens</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatNumber(stats.total_input_tokens + stats.total_output_tokens)}
          </div>
          <div className="text-gray-400 text-xs">
            {formatNumber(stats.total_input_tokens)} in / {formatNumber(stats.total_output_tokens)} out
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Cost</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatCost(stats.total_cost)}
          </div>
          <div className="text-gray-400 text-xs">
            Total usage cost
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Avg Time</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatTime(stats.average_processing_time)}
          </div>
          <div className="text-gray-400 text-xs">
            Per request
          </div>
        </div>
      </div>
    </div>
  );
}