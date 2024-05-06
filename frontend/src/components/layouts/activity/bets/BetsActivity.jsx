import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom'; // Import Link

const BetsActivityLayout = ({ marketId }) => {
    const [bets, setBets] = useState([]);

    useEffect(() => {
        const fetchBets = async () => {
            const response = await fetch(`${API_URL}/api/v0/markets/bets/${marketId}`, {
            });
            if (response.ok) {
                const data = await response.json();
                setBets(data.sort((a, b) => new Date(b.placedAt) - new Date(a.placedAt)));
            } else {
                console.error('Error fetching bets:', response.statusText);
            }
        };
        fetchBets();
    }, [marketId]);

    return (
        <div className="flex flex-col gap-2 p-4">
            {bets.map((bet, index) => (
                <div key={index} className="bg-gray-800 p-3 rounded-lg shadow flex items-center">
                    <div className="flex-1 font-bold text-blue-500">
                        <Link to={`/user/${bet.username}`} className="underline hover:text-blue-700">
                            {bet.username}
                        </Link>
                    </div>
                    <div className="flex-1 text-gray-300">Outcome: {bet.outcome}</div>
                    <div className="flex-1 text-gray-300">Amount: {bet.amount}</div>
                    <div className="flex-1 text-gray-300">Probability: {bet.probability.toFixed(3)}</div>
                    <div className="flex-1 text-gray-400 text-xs">Placed: {new Date(bet.placedAt).toLocaleString()}</div>
                </div>
            ))}
        </div>
    );

};

export default BetsActivityLayout;

