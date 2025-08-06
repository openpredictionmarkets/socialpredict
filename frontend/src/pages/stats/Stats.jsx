import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';

const Stats = () => {
  const [statsData, setStatsData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch(`${API_URL}/api/v0/stats`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch stats: ${response.status}`);
        }

        const data = await response.json();
        setStatsData(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  const setupConfigExplanations = {
    initialMarketProbability: 'Default probability percentage for new markets when first created',
    initialMarketSubsidization: 'Initial funding provided to new markets to bootstrap liquidity',
    initialMarketYes: 'Starting number of YES shares available in new markets',
    initialMarketNo: 'Starting number of NO shares available in new markets',
    createMarketCost: 'Cost in points for users to create a new prediction market',
    traderBonus: 'Bonus points awarded to users for participating in trading',
    initialAccountBalance: 'Starting balance given to new user accounts',
    maximumDebtAllowed: 'Maximum negative balance users can reach before restrictions',
    minimumBet: 'Smallest bet amount allowed on any market',
    maxDustPerSale: 'Maximum dust (small remainder) allowed when selling positions',
    initialBetFee: 'Fee charged when placing the first bet on a market',
    buySharesFee: 'Fee charged when purchasing shares in a market',
    sellSharesFee: 'Fee charged when selling shares back to the market'
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-white text-xl">Loading stats...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-red-400 text-xl">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-4">Platform Statistics</h1>
        <p className="text-gray-300 text-lg">
          System configuration and financial metrics for the SocialPredict platform.
        </p>
      </div>

      {/* Setup Configuration Section */}
      <div className="bg-gray-800 rounded-lg p-6 mb-8">
        <h2 className="text-2xl font-semibold text-white mb-6">Setup Configuration</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-600">
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Setup Variable</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Value</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Explanation</th>
              </tr>
            </thead>
            <tbody>
              {statsData?.setupConfiguration && Object.entries(statsData.setupConfiguration).map(([key, value]) => (
                <tr key={key} className="border-b border-gray-700 hover:bg-gray-700/50 transition-colors">
                  <td className="py-3 px-4 text-blue-400 font-mono text-sm">
                    {key}
                  </td>
                  <td className="py-3 px-4 text-white font-semibold">
                    {typeof value === 'number' ? value.toLocaleString() : value.toString()}
                  </td>
                  <td className="py-3 px-4 text-gray-300">
                    {setupConfigExplanations[key] || 'Configuration parameter for platform behavior'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Financial Statistics Section */}
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-2xl font-semibold text-white mb-6">Financial Statistics</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {statsData?.financialStats && Object.entries(statsData.financialStats).map(([key, value]) => (
            <div key={key} className="bg-gray-700 rounded-lg p-4">
              <h3 className="text-gray-300 text-sm font-medium mb-2">
                {key.replace(/([A-Z])/g, ' $1').replace(/^./, str => str.toUpperCase())}
              </h3>
              <p className="text-white text-2xl font-bold">
                {typeof value === 'number' ? value.toLocaleString() : value}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Stats;
