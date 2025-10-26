import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom'; // Import Link

const BetsActivityLayout = ({ marketId, refreshTrigger }) => {
    const [bets, setBets] = useState([]);

    useEffect(() => {
        const fetchBets = async () => {
            const response = await fetch(`${API_URL}/v0/markets/bets/${marketId}`, {
            });
            if (response.ok) {
                const data = await response.json();
                setBets(data.sort((a, b) => new Date(b.placedAt) - new Date(a.placedAt)));
            } else {
                console.error('Error fetching bets:', response.statusText);
            }
        };
        fetchBets();
    }, [marketId, refreshTrigger]);

    return (
        <div className="p-4">
            <div className="sp-grid-bets-header">
                <div>Username</div>
                <div className="text-center">Outcome</div>
                <div className="text-right">Amount</div>
                <div className="text-right">After</div>
                <div className="text-right">Placed</div>
            </div>
            {bets.map((bet, index) => (
                <div key={index} className="sp-grid-bets-row mt-2">
                    {/* Username */}
                    <div className="sp-cell-username">
                        <div className="sp-ellipsis text-xs sm:text-sm font-medium">
                            <Link to={`/user/${bet.username}`} className="text-blue-500 hover:text-blue-400 transition-colors">
                                {bet.username}
                            </Link>
                        </div>
                    </div>

                    {/* Outcome */}
                    <div className="justify-self-start sm:justify-self-center">
                        <span className={`px-2 py-1 rounded text-xs font-bold ${bet.outcome === 'YES' ? 'bg-green-600' : 'bg-red-600'} text-white`}>
                            {bet.outcome}
                        </span>
                    </div>

                    {/* Amount */}
                    <div className="sp-cell-num text-xs sm:text-sm text-gray-300">{bet.amount}</div>

                    {/* After (sm+) */}
                    <div className="hidden sm:block sp-cell-num text-gray-300">{bet.probability.toFixed(2)}</div>

                    {/* Placed (stack full width on xs) */}
                    <div className="col-span-3 sm:col-span-1 text-right sp-subline">
                        {new Date(bet.placedAt).toLocaleString()}
                    </div>
                </div>
            ))}
        </div>
    );


};

export default BetsActivityLayout;
