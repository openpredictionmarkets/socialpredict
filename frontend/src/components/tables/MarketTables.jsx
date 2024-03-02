import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';

function MarketsTable() {
    const [marketsData, setMarketsData] = useState([]);

    useEffect(() => {
        fetch(`${API_URL}/api/v0/markets`)
            .then((response) => response.json())
            .then((data) => setMarketsData(data))
            .catch((error) => console.error('Error fetching market data:', error));
    }, []);

    return (
        <div className="overflow-auto">
            <h1 className="text-xs font-medium text-gray-500 uppercase tracking-wider p-6">Markets</h1>
            {marketsData.length > 0 ? (
                <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
                    <thead className="bg-gray-50">
                        <tr>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Trend</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Probability</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Question</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Participants</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Market Size</th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Comments</th>
                        </tr>
                    </thead>
                    <tbody className="bg-primary-background divide-y divide-gray-200">
                        {marketsData.map((market) => (
                            <tr key={market.id}>
                                <td className="px-6 py-4 text-white">⬆️⬇️</td>
                                <td className="px-6 py-4 text-sm text-gray-500">{market.initialProbability}</td>
                                <td className="px-6 py-4 text-sm font-mono text-gray-500">
                                    <Link to={`/markets/${market.id}`} className="text-blue-600 hover:text-blue-800">{market.questionTitle}</Link>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-500">📅 {new Date(market.resolutionDateTime).toLocaleDateString()}</td>
                                <td className="px-6 py-4 text-sm text-gray-500">😀 admin</td>
                                <td className="px-6 py-4 text-sm text-gray-500">👤 20</td>
                                <td className="px-6 py-4 text-sm text-gray-500">📊 1.5k</td>
                                <td className="px-6 py-4 text-sm text-gray-500">💬 12</td>
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
