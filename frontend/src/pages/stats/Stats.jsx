import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';
import SiteButton from '../../components/buttons/SiteButtons';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import SiteTabs from '../../components/tabs/SiteTabs';

// MetricCard Component
const MetricCard = ({
  title,
  value,
  formula,
  explanation,
  onToggleFormula,
  showFormula,
  colorClass = "text-white",
  isTotal = false,
  isStatus = false
}) => {
  const formatValue = (val) => {
    if (isStatus) return val;
    if (typeof val === 'number') {
      return val.toLocaleString();
    }
    return val;
  };

  return (
    <div className={`bg-gray-600 rounded-lg p-4 ${isTotal ? 'border-2 border-gray-500' : ''}`}>
      <div className="flex justify-between items-start mb-2">
        <h4 className="text-gray-300 text-sm font-medium">{title}</h4>
        {formula && (
          <button
            onClick={onToggleFormula}
            className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
            title="Toggle formula"
          >
            📐
          </button>
        )}
      </div>

      <p className={`${colorClass} ${isTotal ? 'text-3xl' : 'text-2xl'} font-bold mb-2`}>
        {formatValue(value)}
      </p>

      {explanation && (
        <p className="text-gray-400 text-xs mb-2">{explanation}</p>
      )}

      {formula && showFormula && (
        <div className="mt-3 p-3 bg-gray-800 rounded border border-gray-700">
          <p className="text-gray-300 text-xs font-mono">{formula}</p>
        </div>
      )}
    </div>
  );
};

const Stats = () => {
  const [statsData, setStatsData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // System metrics state
  const [systemMetrics, setSystemMetrics] = useState(null);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [metricsError, setMetricsError] = useState(null);
  const [showFormulas, setShowFormulas] = useState({});

  // Global leaderboard state
  const [globalLeaderboard, setGlobalLeaderboard] = useState(null);
  const [leaderboardLoading, setLeaderboardLoading] = useState(false);
  const [leaderboardError, setLeaderboardError] = useState(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch(`${API_URL}/v0/stats`, {
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

  const fetchSystemMetrics = async () => {
    setMetricsLoading(true);
    setMetricsError(null);
    try {
      const response = await fetch(`${API_URL}/v0/system/metrics`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch system metrics: ${response.status}`);
      }

      const data = await response.json();
      setSystemMetrics(data);
    } catch (err) {
      setMetricsError(err.message);
    } finally {
      setMetricsLoading(false);
    }
  };

  const fetchGlobalLeaderboard = async () => {
    setLeaderboardLoading(true);
    setLeaderboardError(null);
    try {
      const response = await fetch(`${API_URL}/v0/global/leaderboard`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch global leaderboard: ${response.status}`);
      }

      const data = await response.json();
      setGlobalLeaderboard(data);
    } catch (err) {
      setLeaderboardError(err.message);
    } finally {
      setLeaderboardLoading(false);
    }
  };

  const toggleFormula = (key) => {
    setShowFormulas(prev => ({
      ...prev,
      [key]: !prev[key]
    }));
  };

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

  // Setup Configuration Tab Content
  const setupConfigContent = (
    <div className="bg-gray-800 rounded-lg p-6">
      <h2 className="text-2xl font-semibold text-white mb-6">Setup Configuration</h2>

      {/* Mobile-responsive grid for setup configuration */}
      <div className="space-y-2">
        <div className="sp-grid-setup-header">
          <div>Setup Variable</div>
          <div>Value</div>
          <div>Explanation</div>
        </div>

        {statsData?.setupConfiguration && Object.entries(statsData.setupConfiguration).map(([key, value]) => (
          <div key={key} className="sp-grid-setup-row hover:bg-gray-700/50 transition-colors">
            {/* Variable Name */}
            <div className="sp-cell-username">
              <div className="sp-ellipsis text-xs sm:text-sm font-mono text-blue-400">
                {key}
              </div>
            </div>

            {/* Value */}
            <div className="sp-cell-num text-xs sm:text-sm text-white font-semibold">
              {typeof value === 'number' ? value.toLocaleString() : value.toString()}
            </div>

            {/* Explanation (desktop only on mobile, full width below on mobile) */}
            <div className="hidden sm:block text-gray-300 text-xs sm:text-sm">
              {setupConfigExplanations[key] || 'Configuration parameter for platform behavior'}
            </div>

            {/* Mobile explanation - spans full width */}
            <div className="col-span-2 sm:hidden sp-subline mt-1">
              {setupConfigExplanations[key] || 'Configuration parameter for platform behavior'}
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  // System Financial Metrics Tab Content
  const systemMetricsContent = (
    <div className="bg-gray-800 rounded-lg p-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-semibold text-white">
          System Financial Metrics <span className="text-warning-orange text-lg">(Beta)</span>
        </h2>
        <SiteButton
          onClick={fetchSystemMetrics}
          isSelected={false}
          disabled={metricsLoading}
          className="bg-info-blue hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
        >
          {metricsLoading ? 'Calculating...' : 'Calculate Metrics'}
        </SiteButton>
      </div>

      {/* Beta Disclaimer */}
      <div className="bg-warning-orange/20 border border-warning-orange/50 rounded-lg p-4 mb-6">
        <div className="flex items-start">
          <span className="text-warning-orange text-xl mr-3">⚠️</span>
          <div>
            <h4 className="text-warning-orange font-medium mb-2">Beta Feature Notice</h4>
            <p className="text-gray-300 text-sm">
              These financial metrics are currently in beta. Balance calculations may not perfectly align
              as we continue to refine the accounting logic. We are actively working to improve accuracy
              and ensure complete balance reconciliation in future versions.
            </p>
          </div>
        </div>
      </div>

      {metricsLoading && (
        <div className="flex justify-center items-center py-8">
          <LoadingSpinner />
          <span className="ml-3 text-gray-300">Computing system metrics...</span>
        </div>
      )}

      {metricsError && (
        <div className="bg-red-900/50 border border-red-600 rounded-lg p-4 mb-6">
          <p className="text-red-300">Error loading metrics: {metricsError}</p>
        </div>
      )}

      {!systemMetrics && !metricsLoading && !metricsError && (
        <div className="text-center py-8">
          <p className="text-gray-400 mb-4">Click "Calculate Metrics" to view detailed financial analysis</p>
          <p className="text-gray-500 text-sm">This will analyze all users, markets, and transactions to provide comprehensive financial metrics</p>
        </div>
      )}

      {systemMetrics && (
        <div className="space-y-8">
          {/* Money Created Section */}
          <div className="bg-gray-700 rounded-lg p-6">
            <h3 className="text-xl font-semibold text-white mb-4 flex items-center">
              💰 Money Created
              <span className="ml-2 text-sm text-gray-400">(Total System Capacity)</span>
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <MetricCard
                title="User Debt Capacity"
                value={systemMetrics.moneyCreated.userDebtCapacity.value}
                formula={systemMetrics.moneyCreated.userDebtCapacity.formula}
                explanation={systemMetrics.moneyCreated.userDebtCapacity.explanation}
                onToggleFormula={() => toggleFormula('userDebtCapacity')}
                showFormula={showFormulas.userDebtCapacity}
                colorClass="text-blue-400"
              />
              <MetricCard
                title="Number of Users"
                value={systemMetrics.moneyCreated.numUsers.value}
                explanation={systemMetrics.moneyCreated.numUsers.explanation}
                colorClass="text-white"
              />
            </div>
          </div>

          {/* Money Utilized Section */}
          <div className="bg-gray-700 rounded-lg p-6">
            <h3 className="text-xl font-semibold text-white mb-4 flex items-center">
              📊 Money Utilized
              <span className="ml-2 text-sm text-gray-400">(Where the money went)</span>
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <MetricCard
                title="Unused Debt Capacity"
                value={systemMetrics.moneyUtilized.unusedDebt.value}
                formula={systemMetrics.moneyUtilized.unusedDebt.formula}
                explanation={systemMetrics.moneyUtilized.unusedDebt.explanation}
                onToggleFormula={() => toggleFormula('unusedDebt')}
                showFormula={showFormulas.unusedDebt}
                colorClass="text-yellow-400"
              />
              <MetricCard
                title="Active Bet Volume"
                value={systemMetrics.moneyUtilized.activeBetVolume.value}
                formula={systemMetrics.moneyUtilized.activeBetVolume.formula}
                explanation={systemMetrics.moneyUtilized.activeBetVolume.explanation}
                onToggleFormula={() => toggleFormula('activeBetVolume')}
                showFormula={showFormulas.activeBetVolume}
                colorClass="text-purple-400"
              />
              <MetricCard
                title="Market Creation Fees"
                value={systemMetrics.moneyUtilized.marketCreationFees.value}
                formula={systemMetrics.moneyUtilized.marketCreationFees.formula}
                explanation={systemMetrics.moneyUtilized.marketCreationFees.explanation}
                onToggleFormula={() => toggleFormula('marketCreationFees')}
                showFormula={showFormulas.marketCreationFees}
                colorClass="text-orange-400"
              />
              <MetricCard
                title="Participation Fees"
                value={systemMetrics.moneyUtilized.participationFees.value}
                formula={systemMetrics.moneyUtilized.participationFees.formula}
                explanation={systemMetrics.moneyUtilized.participationFees.explanation}
                onToggleFormula={() => toggleFormula('participationFees')}
                showFormula={showFormulas.participationFees}
                colorClass="text-cyan-400"
              />
              <MetricCard
                title="Bonuses Paid"
                value={systemMetrics.moneyUtilized.bonusesPaid.value}
                explanation={systemMetrics.moneyUtilized.bonusesPaid.explanation}
                colorClass="text-pink-400"
              />
            </div>
            <div className="mt-4 pt-4 border-t border-gray-600">
              <MetricCard
                title="Total Utilized"
                value={systemMetrics.moneyUtilized.totalUtilized.value}
                formula={systemMetrics.moneyUtilized.totalUtilized.formula}
                explanation={systemMetrics.moneyUtilized.totalUtilized.explanation}
                onToggleFormula={() => toggleFormula('totalUtilized')}
                showFormula={showFormulas.totalUtilized}
                colorClass="text-white"
                isTotal={true}
              />
            </div>
          </div>

          {/* Verification Section */}
          <div className="bg-gray-700 rounded-lg p-6">
            <h3 className="text-xl font-semibold text-white mb-4 flex items-center">
              ✅ Accounting Verification
              <span className="ml-2 text-sm text-gray-400">(Balance Check)</span>
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <MetricCard
                title="System Balanced"
                value={systemMetrics.verification.balanced.value === true ? 'YES' : 'NO'}
                explanation={systemMetrics.verification.balanced.explanation}
                colorClass={systemMetrics.verification.balanced.value === true ? 'text-green-400' : 'text-red-400'}
                isStatus={true}
              />
              <MetricCard
                title="Surplus/Deficit"
                value={systemMetrics.verification.surplus.value}
                formula={systemMetrics.verification.surplus.formula}
                explanation={systemMetrics.verification.surplus.explanation}
                onToggleFormula={() => toggleFormula('surplus')}
                showFormula={showFormulas.surplus}
                colorClass={systemMetrics.verification.surplus.value === 0 ? 'text-green-400' :
                           systemMetrics.verification.surplus.value > 0 ? 'text-yellow-400' : 'text-red-400'}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );

  // Global Leaderboard Tab Content
  const globalLeaderboardContent = (
    <div className="bg-gray-800 rounded-lg p-6">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-4">
        <h2 className="text-2xl font-semibold text-white">
          Global Leaderboard <span className="text-warning-orange text-lg">(Beta)</span>
        </h2>
        <SiteButton
          onClick={fetchGlobalLeaderboard}
          isSelected={false}
          disabled={leaderboardLoading}
          className="bg-info-blue hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors w-full sm:w-auto"
        >
          {leaderboardLoading ? 'Calculating...' : 'Calculate Leaderboard'}
        </SiteButton>
      </div>

      {/* Beta Disclaimer */}
      <div className="bg-warning-orange/20 border border-warning-orange/50 rounded-lg p-4 mb-6">
        <div className="flex items-start">
          <span className="text-warning-orange text-xl mr-3">🏆</span>
          <div>
            <h4 className="text-warning-orange font-medium mb-2">Beta Feature Notice</h4>
            <p className="text-gray-300 text-sm">
              This global leaderboard aggregates profit calculations across all markets. Rankings are based on
              total profit (current position value minus total amount spent) across both resolved and active markets.
            </p>
          </div>
        </div>
      </div>

      {leaderboardLoading && (
        <div className="flex justify-center items-center py-8">
          <LoadingSpinner />
          <span className="ml-3 text-gray-300">Computing global leaderboard...</span>
        </div>
      )}

      {leaderboardError && (
        <div className="bg-red-900/50 border border-red-600 rounded-lg p-4 mb-6">
          <p className="text-red-300">Error loading leaderboard: {leaderboardError}</p>
        </div>
      )}

      {!globalLeaderboard && !leaderboardLoading && !leaderboardError && (
        <div className="text-center py-8">
          <p className="text-gray-400 mb-4">Click "Calculate Leaderboard" to view global profit rankings</p>
          <p className="text-gray-500 text-sm">This will analyze all user positions across all markets to create a comprehensive ranking</p>
        </div>
      )}

      {globalLeaderboard && globalLeaderboard.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-600">
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Rank</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">User</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Total Profit</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Current Value</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Total Spent</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Active Markets</th>
                <th className="text-left py-3 px-4 text-gray-300 font-medium">Resolved Markets</th>
              </tr>
            </thead>
            <tbody>
              {globalLeaderboard.map((user, index) => {
                const getRankDisplay = (rank) => {
                  if (rank === 1) return '🥇';
                  if (rank === 2) return '🥈';
                  if (rank === 3) return '🥉';
                  return `#${rank}`;
                };

                const getProfitColor = (profit) => {
                  if (profit > 0) return 'text-green-400';
                  if (profit < 0) return 'text-red-400';
                  return 'text-gray-300';
                };

                return (
                  <tr key={user.username} className="border-b border-gray-700 hover:bg-gray-700/50 transition-colors">
                    <td className="py-3 px-4 text-white font-semibold">
                      {getRankDisplay(user.rank)}
                    </td>
                    <td className="py-3 px-4">
                      <Link
                        to={`/user/${user.username}`}
                        className="text-blue-400 font-medium hover:text-blue-300 transition-colors"
                      >
                        {user.username}
                      </Link>
                    </td>
                    <td className={`py-3 px-4 font-semibold ${getProfitColor(user.totalProfit)}`}>
                      {user.totalProfit >= 0 ? '+' : ''}{user.totalProfit.toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-gray-300">
                      {user.totalCurrentValue.toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-gray-300">
                      {user.totalSpent.toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-gray-300 text-center">
                      {user.activeMarkets}
                    </td>
                    <td className="py-3 px-4 text-gray-300 text-center">
                      {user.resolvedMarkets}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {globalLeaderboard && globalLeaderboard.length === 0 && (
        <div className="text-center py-8">
          <p className="text-gray-400">No users with betting activity found.</p>
        </div>
      )}
    </div>
  );

  const tabs = [
    {
      label: 'Global Leaderboard (Beta)',
      content: globalLeaderboardContent
    },
    {
      label: 'System Financial Metrics (Beta)',
      content: systemMetricsContent
    },
    {
      label: 'Setup Configuration',
      content: setupConfigContent
    }
  ];

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-4">Platform Statistics</h1>
        <p className="text-gray-300 text-lg">
          System configuration and financial metrics for the SocialPredict platform.
        </p>
      </div>

      <SiteTabs tabs={tabs} defaultTab="Global Leaderboard (Beta)" />
    </div>
  );
};

export default Stats;
