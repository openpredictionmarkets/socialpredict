import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

const LeaderboardActivity = ({ marketId }) => {
    const [leaderboard, setLeaderboard] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchLeaderboard = async () => {
            try {
                setLoading(true);
                const response = await fetch(`${API_URL}/v0/markets/leaderboard/${marketId}`);
                if (response.ok) {
                    const data = await response.json();
                    setLeaderboard(data);
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

    return (
        <div className="p-4">
            {/* Header */}
            <div className="bg-gray-800 p-3 rounded-lg shadow">
                <div className="grid grid-cols-7 gap-2 text-white font-bold text-sm">
                    <div>Rank</div>
                    <div>User</div>
                    <div>Position</div>
                    <div>Profit</div>
                    <div>Current Value</div>
                    <div>Total Spent</div>
                    <div>Shares</div>
                </div>
            </div>
            
            {/* Leaderboard Rows */}
            {leaderboard.map((entry, index) => (
                <div key={entry.username} className="bg-gray-800 p-3 rounded-lg shadow mt-2 grid grid-cols-7 gap-2 items-center">
                    {/* Rank */}
                    <div className="text-white font-bold text-lg">
                        {getRankDisplay(entry.rank)}
                    </div>
                    
                    {/* Username */}
                    <div className="text-blue-500 font-bold">
                        <Link to={`/user/${entry.username}`} className="underline hover:text-blue-700">
                            {entry.username}
                        </Link>
                    </div>
                    
                    {/* Position */}
                    <div>
                        <span className={getPositionBadge(entry.position)}>
                            {entry.position}
                        </span>
                    </div>
                    
                    {/* Profit */}
                    <div className={`font-bold ${getProfitColor(entry.profit)}`}>
                        {entry.profit >= 0 ? '+' : ''}{formatCurrency(entry.profit)}
                    </div>
                    
                    {/* Current Value */}
                    <div className="text-gray-300">
                        {formatCurrency(entry.currentValue)}
                    </div>
                    
                    {/* Total Spent */}
                    <div className="text-gray-300">
                        {formatCurrency(entry.totalSpent)}
                    </div>
                    
                    {/* Shares */}
                    <div className="text-gray-300 text-xs">
                        <div>YES: {entry.yesSharesOwned}</div>
                        <div>NO: {entry.noSharesOwned}</div>
                    </div>
                </div>
            ))}
        </div>
    );
};

export default LeaderboardActivity;
