import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';

function MarketsTable() {
    const [marketsData, setMarketsData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        fetch(`${API_URL}/api/v0/markets`)
            .then((response) => {
                if (response.ok) {
                    return response.json();
                } else {
                    throw new Error('Failed to fetch markets');
                }
            })
            .then((data) => {
                setMarketsData(data.markets);
                setLoading(false);
            })
            .catch((error) => {
                console.error('Error fetching market data:', error);
                setError(error.toString());
                setLoading(false);
            });
    }, []);

    if (loading) {
        return <div className="p-6 text-center">Loading markets...</div>;
    }

    if (error) {
        return <div className="p-6 text-center text-red-500">Error loading markets: {error}</div>;
    }

    if (!marketsData || marketsData.length === 0) {
        return <div className="p-6 text-center">No markets found. None may have been created yet.</div>;
    }

    return (
        <div className="overflow-auto">
            <h1 className="text-sm md:text-lg font-medium text-gray-500 uppercase tracking-wider p-3 md:p-6">Markets</h1>
            <table className="w-full divide-y divide-gray-200 bg-primary-background">
                <thead className="bg-gray-50">
                    <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Trade</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">🪙</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Question</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">📅 Closes</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Creator</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">👤 Users</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">📊 Size</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">💬</th>
                    </tr>
                </thead>
                <tbody className="bg-primary-background divide-y divide-gray-200">
                    {marketsData.map((market, index) => (
                        <tr key={index}>
                            <td className="px-6 py-4 text-white">
                                <Link to={`/markets/${market.marketId}`}>⬆️⬇️</Link>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-500">{market.lastProbability.toFixed(3)}</td>
                            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                                <Link to={`/markets/${market.marketId}`} className="text-blue-600 hover:text-blue-800">{market.questionTitle}</Link>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-500">
                                {new Date(market.resolutionDateTime).toLocaleDateString()}
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-500">
                                <Link
                                    to={`/user/${market.creator.username}`}
                                    className="text-blue-600 hover:text-blue-800 flex items-center"
                                >
                                    <span role='img' aria-label='Creator' className="mr-1">{market.creator.personalEmoji}</span>
                                    @{market.creator.username}
                                </Link>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-500">{market.numUsers}</td>
                            <td className="px-6 py-4 text-sm text-gray-500">{market.totalVolume}</td>
                            <td className="px-6 py-4 text-sm text-gray-500">0</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}

export default MarketsTable;
