import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../../../config';
import LoadingSpinner from '../../../loaders/LoadingSpinner';
import { mapInternalToDisplay } from '../../../../utils/labelMapping';
import { useAuth } from '../../../../helpers/AuthContent';

const PAGE_SIZE = 20;
const paginationButtonClass = [
    'rounded',
    'border',
    'border-transparent',
    'bg-neutral-btn',
    'px-3',
    'py-1.5',
    'text-xs',
    'font-semibold',
    'text-white',
    'transition-colors',
    'duration-200',
    'hover:bg-neutral-btn-hover',
    'disabled:cursor-not-allowed',
    'disabled:bg-custom-gray-light',
    'disabled:text-gray-400',
    'disabled:opacity-60',
].join(' ');

const PortfolioTabContent = ({ username }) => {
    const [positions, setPositions] = useState([]);
    const [page, setPage] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { isLoggedIn, token } = useAuth();

    useEffect(() => {
        setPage(0);
    }, [username]);

    useEffect(() => {
        const fetchPositions = async () => {
            setLoading(true);
            setError(null);
            try {
                const response = await fetch(`${API_URL}/v0/portfolio/${username}`, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                });
                if (response.ok) {
                    const data = await response.json();
                    setPositions(data.portfolioItems || []);
                    setPage(0);
                } else {
                    throw new Error(`Error fetching portfolio: ${response.statusText}`);
                }
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        if (username && token) {
            fetchPositions();
        } else {
            setPositions([]);
            setError(null);
            setLoading(false);
        }
    }, [username, token]);

    if (loading) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="flex items-center justify-center">
                    <LoadingSpinner />
                    <span className="ml-2 text-gray-300">Loading portfolio...</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-center text-red-400">
                    Error loading portfolio: {error}
                </div>
            </div>
        );
    }

    if (!isLoggedIn) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-center">
                    <div className="text-xl font-semibold text-blue-100">
                        Log in to see portfolio
                    </div>
                    <div className="text-sm text-gray-400 mt-2">
                        User portfolio positions are visible to logged-in players.
                    </div>
                </div>
            </div>
        );
    }

    if (!positions || positions.length === 0) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-center text-gray-400">
                    No positions found for this user.
                </div>
            </div>
        );
    }

    const CompactSharesBadge = ({ type, count, market = null }) => {
        if (count === 0) return null;

        const bgColor = type === "YES" ? "bg-green-600" : "bg-red-600";
        const textColor = "text-white";
        
        // Use custom label if market is provided, otherwise use internal type
        const displayLabel = market ? mapInternalToDisplay(type, market) : type;
        return (
            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${bgColor} ${textColor}`}>
                {count} {displayLabel}
            </span>
        );
    };

    const start = page * PAGE_SIZE;
    const visiblePositions = positions.slice(start, start + PAGE_SIZE);
    const hasNextPage = start + PAGE_SIZE < positions.length;

    return (
        <div className="space-y-6">
            {/* Portfolio Summary */}
            <div className="bg-primary-background shadow-md rounded-lg border-2 border-gold-btn">
                <div className="p-6">
                    <h3 className="text-xl font-bold text-gold-btn mb-4 text-center">Portfolio Summary</h3>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-center">
                        <div>
                            <div className="text-custom-gray-light font-medium mb-1">Total Markets</div>
                            <div className="text-2xl font-bold text-white">{positions.length}</div>
                        </div>
                        <div>
                            <div className="text-custom-gray-light font-medium mb-1">Total YES Shares</div>
                            <div className="text-2xl font-bold text-green-400">
                                {positions.reduce((sum, pos) => sum + pos.yesSharesOwned, 0)} 🪙
                            </div>
                        </div>
                        <div>
                            <div className="text-custom-gray-light font-medium mb-1">Total NO Shares</div>
                            <div className="text-2xl font-bold text-red-400">
                                {positions.reduce((sum, pos) => sum + pos.noSharesOwned, 0)} 🪙
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Portfolio Table */}
            <div className="bg-primary-background shadow-md rounded-lg border border-custom-gray-dark overflow-hidden">
                <div className="flex flex-col gap-2 border-b border-custom-gray-dark bg-gray-900/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="text-xs uppercase tracking-[0.16em] text-gray-400">
                        Showing portfolio page {page + 1}{visiblePositions.length ? ` (${start + 1}-${start + visiblePositions.length} of ${positions.length})` : ''}
                    </div>
                    <div className="flex gap-2">
                        <button
                            type="button"
                            onClick={() => setPage((current) => Math.max(0, current - 1))}
                            disabled={page === 0}
                            className={paginationButtonClass}
                        >
                            Previous
                        </button>
                        <button
                            type="button"
                            onClick={() => setPage((current) => current + 1)}
                            disabled={!hasNextPage}
                            className={paginationButtonClass}
                        >
                            Next
                        </button>
                    </div>
                </div>
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-custom-gray-dark">
                        <thead className="bg-custom-gray-dark">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-custom-gray-light uppercase tracking-wider">
                                    Market
                                </th>
                                <th className="px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider">
                                    YES Shares
                                </th>
                                <th className="px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider">
                                    NO Shares
                                </th>
                                <th className="px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider">
                                    Total Shares
                                </th>
                                <th className="px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider">
                                    Last Bet
                                </th>
                            </tr>
                        </thead>
                        <tbody className="bg-primary-background divide-y divide-custom-gray-dark">
                            {visiblePositions.map((position, index) => (
                                <tr key={position.marketId || index} className="hover:bg-custom-gray-dark transition-colors">
                                    <td className="px-6 py-4">
                                        <Link
                                            to={`/markets/${position.marketId}`}
                                            className="text-sm font-medium text-custom-gray-verylight hover:text-gold-btn transition-colors duration-200"
                                        >
                                            {position.questionTitle || 'Unknown Market'}
                                        </Link>
                                        <div className="text-xs text-gray-400">
                                            ID: {position.marketId}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        {position.yesSharesOwned > 0 ? (
                                            <CompactSharesBadge type="YES" count={position.yesSharesOwned} />
                                        ) : (
                                            <span className="text-gray-500 text-sm">-</span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        {position.noSharesOwned > 0 ? (
                                            <CompactSharesBadge type="NO" count={position.noSharesOwned} />
                                        ) : (
                                            <span className="text-gray-500 text-sm">-</span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        <span className="text-sm font-medium text-white">
                                            {position.yesSharesOwned + position.noSharesOwned}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        <span className="text-sm text-gray-300">
                                            {new Date(position.lastBetPlaced).toLocaleDateString()}
                                        </span>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default PortfolioTabContent;
