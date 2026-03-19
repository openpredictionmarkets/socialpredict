import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../../../config';
import LoadingSpinner from '../../../loaders/LoadingSpinner';

const BetHistoryTabContent = ({ username }) => {
    const [bets, setBets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchBets = async () => {
            try {
                const response = await fetch(`${API_URL}/v0/users/${username}/bets`);
                if (response.ok) {
                    const data = await response.json();
                    setBets(data || []);
                } else {
                    throw new Error(`Error fetching bet history: ${response.statusText}`);
                }
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        if (username) fetchBets();
    }, [username]);

    if (loading) {
        return (
            <div className='bg-primary-background shadow-md rounded-lg p-6'>
                <div className='flex items-center justify-center'>
                    <LoadingSpinner />
                    <span className='ml-2 text-gray-300'>Loading bet history…</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className='bg-primary-background shadow-md rounded-lg p-6'>
                <div className='text-center text-red-400'>Error loading bet history: {error}</div>
            </div>
        );
    }

    if (!bets || bets.length === 0) {
        return (
            <div className='bg-primary-background shadow-md rounded-lg p-6'>
                <div className='text-center text-gray-400'>No bets placed yet.</div>
            </div>
        );
    }

    return (
        <div className='bg-primary-background shadow-md rounded-lg border border-custom-gray-dark overflow-hidden'>
            <div className='overflow-x-auto'>
                <table className='min-w-full divide-y divide-custom-gray-dark'>
                    <thead className='bg-custom-gray-dark'>
                        <tr>
                            <th className='px-6 py-3 text-left text-xs font-medium text-custom-gray-light uppercase tracking-wider'>
                                Market
                            </th>
                            <th className='px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider'>
                                Action
                            </th>
                            <th className='px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider'>
                                Outcome
                            </th>
                            <th className='px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider'>
                                Amount
                            </th>
                            <th className='px-6 py-3 text-center text-xs font-medium text-custom-gray-light uppercase tracking-wider'>
                                Date
                            </th>
                        </tr>
                    </thead>
                    <tbody className='bg-primary-background divide-y divide-custom-gray-dark'>
                        {bets.map((bet, index) => {
                            const isSell = bet.action === 'SELL';
                            return (
                                <tr
                                    key={bet.id || index}
                                    className='hover:bg-custom-gray-dark transition-colors'
                                >
                                    <td className='px-6 py-4'>
                                        <Link
                                            to={`/markets/${bet.marketId}`}
                                            className='text-sm font-medium text-custom-gray-verylight hover:text-gold-btn transition-colors duration-200'
                                        >
                                            {bet.questionTitle || `Market #${bet.marketId}`}
                                        </Link>
                                    </td>
                                    <td className='px-6 py-4 text-center'>
                                        <span
                                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                                isSell
                                                    ? 'bg-yellow-700 text-yellow-100'
                                                    : 'bg-blue-700 text-blue-100'
                                            }`}
                                        >
                                            {bet.action}
                                        </span>
                                    </td>
                                    <td className='px-6 py-4 text-center'>
                                        <span
                                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                                bet.outcome === 'YES'
                                                    ? 'bg-green-700 text-green-100'
                                                    : 'bg-red-700 text-red-100'
                                            }`}
                                        >
                                            {bet.outcome}
                                        </span>
                                    </td>
                                    <td className='px-6 py-4 text-center'>
                                        <span className='text-sm font-medium text-white'>
                                            {Math.abs(bet.amount)} 🪙
                                        </span>
                                    </td>
                                    <td className='px-6 py-4 text-center'>
                                        <span className='text-sm text-gray-300'>
                                            {new Date(bet.placedAt).toLocaleString()}
                                        </span>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default BetHistoryTabContent;
