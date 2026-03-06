import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getMarketLabels } from '../../../../utils/labelMapping';

const LeaderboardActivity = ({ marketId, market }) => {
    const [leaderboard, setLeaderboard] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchLeaderboard = async () => {
            try {
                setLoading(true);
                const response = await fetch(`${API_URL}/v0/markets/${marketId}/leaderboard`);
                if (response.ok) {
                    const data = await response.json();
                    const rows = Array.isArray(data?.leaderboard) ? data.leaderboard : [];
                    setLeaderboard(rows);
                } else {
                    console.error('Error fetching leaderboard:', response.statusText);
                    setError('Failed to load leaderboard data');
                }
            } catch (err) {
                console.error('Error fetching leaderboard:', err);
                setError('Failed to load leaderboard data');
            } finally {
                setLoading(false);
            }
        };

        if (marketId) {
            fetchLeaderboard();
        }
    }, [marketId]);

    const formatCurrency = (amount) => {
        return amount.toLocaleString();
    };

    const getProfitColor = (profit) => {
        if (profit > 0) return 'text-green-400';
        if (profit < 0) return 'text-red-400';
        return 'text-gray-300';
    };

    const getPositionBadge = (position) => {
        const baseClasses = "px-2 py-1 rounded text-xs font-bold";
        switch (position) {
            case 'YES':
                return `${baseClasses} bg-green-600 text-white`;
            case 'NO':
                return `${baseClasses} bg-red-600 text-white`;
            case 'NEUTRAL':
                return `${baseClasses} bg-yellow-600 text-white`;
            default:
                return `${baseClasses} bg-gray-600 text-white`;
        }
    };

    const getRankDisplay = (rank) => {
        if (rank === 1) return "🥇";
        if (rank === 2) return "🥈";
        if (rank === 3) return "🥉";
        return `#${rank}`;
    };

    if (loading) {
        return (
            <div className="p-4 text-center">
                <div className="text-gray-400">Loading leaderboard...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-4 text-center">
                <div className="text-red-400">{error}</div>
            </div>
        );
    }

    if (leaderboard.length === 0) {
        return (
            <div className="p-4 text-center">
                <div className="text-gray-400">No participants yet</div>
            </div>
        );
    }

    const labels = market ? getMarketLabels(market) : { yes: "YES", no: "NO" };

    return (
        <div className="p-4">
            {/* Header */}
            <div className="sp-grid-leaderboard-header">
                <div>Rank</div>
                <div>User</div>
                <div>Position</div>
                <div className="text-right">Profit</div>
                <div className="text-right">Current Value</div>
                <div className="text-right">Total Spent</div>
                <div>Shares</div>
            </div>

            {/* Leaderboard Rows */}
            {leaderboard.map((entry, index) => (
                <div key={entry.username} className="sp-grid-leaderboard-row mt-2">
                    {/* Rank + Username (xs) / Rank (sm+) */}
                    <div className="flex items-center justify-start">
                        <div className="text-white font-bold text-lg mr-2">
                            {getRankDisplay(entry.rank)}
                        </div>
                        <div className="sm:hidden sp-cell-username">
                            <div className="sp-ellipsis text-xs font-medium">
                                <Link to={`/user/${entry.username}`} className="text-blue-500 hover:text-blue-400 transition-colors">
                                    {entry.username}
                                </Link>
                            </div>
                        </div>
                    </div>

                    {/* Username (sm+) */}
                    <div className="hidden sm:block sp-cell-username">
                        <div className="sp-ellipsis font-medium">
                            <Link to={`/user/${entry.username}`} className="text-blue-500 hover:text-blue-400 transition-colors">
                                {entry.username}
                            </Link>
                        </div>
                    </div>

                    {/* Position (sm+) */}
                    <div className="hidden sm:block">
                        <span className={getPositionBadge(entry.position)}>
                            {entry.position}
                        </span>
                    </div>

                    {/* P&L + Subline (xs) / Profit (sm+) */}
                    <div className="text-right">
                        <div className={`font-bold text-sm ${getProfitColor(entry.profit)}`}>
                            {entry.profit >= 0 ? '+' : ''}{formatCurrency(entry.profit)}
                        </div>
                        <div className="sm:hidden sp-subline">
                            Pos {entry.position} • {entry.yesSharesOwned}Y {entry.noSharesOwned}N
                        </div>
                    </div>

                    {/* Current Value (sm+) */}
                    <div className="hidden sm:block sp-cell-num text-gray-300">
                        {formatCurrency(entry.currentValue)}
                    </div>

                    {/* Total Spent (sm+) */}
                    <div className="hidden sm:block sp-cell-num text-gray-300">
                        {formatCurrency(entry.totalSpent)}
                    </div>

                    {/* Shares (sm+) */}
                    <div className="hidden sm:block text-gray-300 text-xs">
                        <div>{labels.yes}: {entry.yesSharesOwned}</div>
                        <div>{labels.no}: {entry.noSharesOwned}</div>
                    </div>
                </div>
            ))}
        </div>
    );
};

export default LeaderboardActivity;
