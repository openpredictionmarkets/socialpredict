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
        <div className="p-4">
            <div className="bg-gray-800 p-3 rounded-lg shadow">
                <div className="grid grid-cols-5 gap-2 text-white font-bold">
                    <div>Username</div>
                    <div>Outcome</div>
                    <div>Amount</div>
                    <div>🪙 After</div>
                    <div>Placed</div>
                </div>
            </div>
            {bets.map((bet, index) => (
                <div key={index} className="bg-gray-800 p-3 rounded-lg shadow mt-2 grid grid-cols-5 gap-2 items-center">
                    <div className="text-blue-500 font-bold">
                        <Link to={`/user/${bet.username}`} className="underline hover:text-blue-700">
                            {bet.username}
                        </Link>
                    </div>
                    <div className="text-gray-300">{bet.outcome}</div>
                    <div className="text-gray-300">{bet.amount}</div>
                    <div className="text-gray-300">{bet.probability.toFixed(3)}</div>
                    <div className="text-gray-400 text-xs">{new Date(bet.placedAt).toLocaleString()}</div>
                </div>
            ))}
        </div>
    );


};

export default BetsActivityLayout;
