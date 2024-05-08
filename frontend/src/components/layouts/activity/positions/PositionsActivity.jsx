import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

const PositionsActivityLayout = ({ marketId }) => {
    const [positions, setPositions] = useState([]);

    useEffect(() => {
        const fetchPositions = async () => {
            const response = await fetch(`${API_URL}/api/v0/markets/positions/${marketId}`);
            if (response.ok) {
                const data = await response.json();
                console.log("API Data:", data);
                const sortedAndFiltered = data.filter(user => user.NoSharesOwned > 0 || user.YesSharesOwned > 0)
                                            .sort((a, b) => (b.NoSharesOwned + b.YesSharesOwned) - (a.NoSharesOwned + a.YesSharesOwned));
                console.log("Filtered and Sorted Data:", sortedAndFiltered);
                setPositions(sortedAndFiltered);
            } else {
                console.error('Error fetching positions:', response.statusText);
            }
        };
        fetchPositions();
    }, [marketId]);

    return (
        <div className="flex flex-row gap-4 p-4">
            <div className="flex-1">
                <h2 className="text-center font-bold">NO Shares</h2>
                <div className="flex flex-col gap-2">
                    {positions.filter(pos => pos.NoSharesOwned > 0).map((pos, index) => (
                        <div key={index} className="bg-gray-800 p-3 rounded-lg shadow flex items-center">
                            <div className="flex-1 font-bold text-blue-500">
                                <Link to={`/user/${pos.Username}`} className="underline hover:text-blue-700">
                                    {pos.Username}
                                </Link>
                            </div>
                            <div className="text-gray-300">{pos.NoSharesOwned}</div>
                        </div>
                    ))}
                </div>
            </div>
            <div className="flex-1">
                <h2 className="text-center font-bold">YES Shares</h2>
                <div className="flex flex-col gap-2">
                    {positions.filter(pos => pos.YesSharesOwned > 0).map((pos, index) => (
                        <div key={index} className="bg-gray-800 p-3 rounded-lg shadow flex items-center">
                            <div className="flex-1 font-bold text-blue-500">
                                <Link to={`/user/${pos.Username}`} className="underline hover:text-blue-700">
                                    {pos.Username}
                                </Link>
                            </div>
                            <div className="text-gray-300">{pos.YesSharesOwned}</div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default PositionsActivityLayout;
