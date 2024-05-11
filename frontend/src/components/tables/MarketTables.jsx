import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';

function MarketsTable() {
    const [marketsData, setMarketsData] = useState([]);

    useEffect(() => {
        fetch(`${API_URL}/api/v0/markets`)
            .then((response) => response.json())
            .then((data) => setMarketsData(data.markets))
            .catch((error) => console.error('Error fetching market data:', error));
    }, []);


    return (
        <div className="overflow-auto">
        <h1 className="text-sm md:text-lg font-medium text-gray-500 uppercase tracking-wider p-3 md:p-6">Markets</h1>
            {marketsData.length > 0 ? (
                <table className="w-full divide-y divide-gray-200 bg-primary-background">
                    <thead className="bg-gray-50">
                        <tr>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Trade</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">🪙</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Question</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">📅 Closes</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Creator</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">👤 Users</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">📊 Size</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">💬</th>
                        </tr>
                    </thead>
                    <tbody className="bg-primary-background divide-y divide-gray-200">
                        {marketsData.map((market) => (
                            <tr key={market.market.id}>
                                <td className="px-6 py-4 text-white">
                                    <Link to={`/markets/${market.market.id}`}>⬆️⬇️</Link>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-500">{market.lastProbability.toFixed(3)}</td>
                                <td className="px-6 py-4 text-sm font-mono text-gray-500">
                                    <Link to={`/markets/${market.market.id}`} className="text-blue-600 hover:text-blue-800">{market.market.questionTitle}</Link>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-500">
                                    {new Date(market.market.resolutionDateTime).toLocaleDateString()}
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
            ) : (
                <div className="px-6 py-4">Loading markets... None may be available.</div>
            )}
        </div>
    );
}

export default MarketsTable;
