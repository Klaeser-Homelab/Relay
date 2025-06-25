import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

export function UsagePage() {
  const [stats, setStats] = useState(null);
  const [history, setHistory] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const navigate = useNavigate();

  const fetchTotalStats = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/usage/stats');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      setStats(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch usage stats');
      console.error('Failed to fetch usage stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchHistory = async (page = 1) => {
    try {
      const response = await fetch(`http://localhost:8080/usage/history?page=${page}&per_page=10`);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      setHistory(data);
    } catch (err) {
      console.error('Failed to fetch usage history:', err);
    }
  };

  useEffect(() => {
    fetchTotalStats();
    fetchHistory();
  }, []);

  const formatCost = (cost) => {
    return `$${Number(cost).toFixed(6)}`;
  };

  const formatTime = (seconds) => {
    return `${Number(seconds).toFixed(2)}s`;
  };

  const formatNumber = (num) => {
    return Number(num).toLocaleString();
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const handleBack = () => {
    navigate(-1); // Go back to previous page
  };

  if (loading) {
    return (
      <div className="flex flex-col h-screen bg-gray-400">
        <div className="p-4 bg-gray-800 border-b border-gray-700">
          <div className="flex items-center justify-between">
            <h1 className="text-white text-xl font-bold">Total API Usage</h1>
            <button
              onClick={handleBack}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              Back
            </button>
          </div>
        </div>
        <div className="flex-1 p-8 flex items-center justify-center">
          <div className="text-white text-lg">Loading usage data...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col h-screen bg-gray-400">
        <div className="p-4 bg-gray-800 border-b border-gray-700">
          <div className="flex items-center justify-between">
            <h1 className="text-white text-xl font-bold">Total API Usage</h1>
            <button
              onClick={handleBack}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              Back
            </button>
          </div>
        </div>
        <div className="flex-1 p-8 flex items-center justify-center">
          <div className="text-red-300 text-lg">Error: {error}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-gray-400">
      <div className="p-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h1 className="text-white text-xl font-bold">Total API Usage</h1>
          <button
            onClick={handleBack}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
          >
            Back
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {/* Total Stats */}
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
          <h2 className="text-white text-lg font-semibold mb-4">Overall Statistics</h2>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-gray-700 rounded-lg p-4">
              <div className="text-gray-300 text-sm mb-1">Total Requests</div>
              <div className="text-white text-2xl font-bold">{formatNumber(stats?.total_requests || 0)}</div>
              <div className="text-gray-400 text-xs">{stats?.successful_requests || 0} successful</div>
            </div>

            <div className="bg-gray-700 rounded-lg p-4">
              <div className="text-gray-300 text-sm mb-1">Total Tokens</div>
              <div className="text-white text-2xl font-bold">
                {formatNumber((stats?.total_input_tokens || 0) + (stats?.total_output_tokens || 0))}
              </div>
              <div className="text-gray-400 text-xs">
                {formatNumber(stats?.total_input_tokens || 0)} in / {formatNumber(stats?.total_output_tokens || 0)} out
              </div>
            </div>

            <div className="bg-gray-700 rounded-lg p-4">
              <div className="text-gray-300 text-sm mb-1">Total Cost</div>
              <div className="text-white text-2xl font-bold">{formatCost(stats?.total_cost || 0)}</div>
              <div className="text-gray-400 text-xs">All-time usage cost</div>
            </div>

            <div className="bg-gray-700 rounded-lg p-4">
              <div className="text-gray-300 text-sm mb-1">Avg Time</div>
              <div className="text-white text-2xl font-bold">{formatTime(stats?.average_processing_time || 0)}</div>
              <div className="text-gray-400 text-xs">Per request</div>
            </div>
          </div>
        </div>

        {/* Usage History */}
        {history && (
          <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-white text-lg font-semibold">Recent Usage History</h2>
              <button
                onClick={() => {
                  fetchTotalStats();
                  fetchHistory(currentPage);
                }}
                className="text-gray-400 hover:text-gray-200 text-sm px-3 py-1 rounded hover:bg-gray-700 transition-colors"
              >
                Refresh
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-600">
                    <th className="text-left py-2 text-gray-300">Time</th>
                    <th className="text-left py-2 text-gray-300">Model</th>
                    <th className="text-right py-2 text-gray-300">Tokens</th>
                    <th className="text-right py-2 text-gray-300">Cost</th>
                    <th className="text-right py-2 text-gray-300">Time</th>
                    <th className="text-left py-2 text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {history.records?.map((record) => (
                    <tr key={record.id} className="border-b border-gray-700 hover:bg-gray-750">
                      <td className="py-2 text-gray-200">
                        {formatDate(record.timestamp)}
                      </td>
                      <td className="py-2 text-gray-200">{record.model}</td>
                      <td className="py-2 text-right text-gray-200">
                        {record.input_tokens + record.output_tokens}
                      </td>
                      <td className="py-2 text-right text-gray-200">
                        {formatCost(record.total_cost)}
                      </td>
                      <td className="py-2 text-right text-gray-200">
                        {formatTime(record.processing_time)}
                      </td>
                      <td className="py-2">
                        <span className={`px-2 py-1 rounded text-xs ${
                          record.success 
                            ? 'bg-green-900 text-green-300' 
                            : 'bg-red-900 text-red-300'
                        }`}>
                          {record.success ? 'Success' : 'Failed'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {history.total > history.per_page && (
              <div className="flex items-center justify-between mt-4">
                <div className="text-gray-400 text-sm">
                  Showing {(history.page - 1) * history.per_page + 1} to {Math.min(history.page * history.per_page, history.total)} of {history.total} records
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => {
                      setCurrentPage(currentPage - 1);
                      fetchHistory(currentPage - 1);
                    }}
                    disabled={currentPage <= 1}
                    className="px-3 py-1 bg-gray-700 text-white rounded disabled:bg-gray-600 disabled:cursor-not-allowed hover:bg-gray-600"
                  >
                    Previous
                  </button>
                  <span className="px-3 py-1 text-gray-300">
                    Page {history.page}
                  </span>
                  <button
                    onClick={() => {
                      setCurrentPage(currentPage + 1);
                      fetchHistory(currentPage + 1);
                    }}
                    disabled={currentPage * history.per_page >= history.total}
                    className="px-3 py-1 bg-gray-700 text-white rounded disabled:bg-gray-600 disabled:cursor-not-allowed hover:bg-gray-600"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}