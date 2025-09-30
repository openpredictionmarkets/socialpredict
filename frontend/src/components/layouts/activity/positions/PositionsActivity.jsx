import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

const PositionsActivityLayout = ({ marketId }) => {
  const [positions, setPositions] = useState([]);

  useEffect(() => {
    const fetchPositions = async () => {
      const response = await fetch(`${API_URL}/v0/markets/positions/${marketId}`);
      if (response.ok) {
        const rawData = await response.json();
        console.log("API Data:", rawData);

        const filteredSorted = rawData
          .filter(user => user.noSharesOwned > 0 || user.yesSharesOwned > 0)
          .sort((a, b) => (b.noSharesOwned + b.yesSharesOwned) - (a.noSharesOwned + a.yesSharesOwned));

        console.log("Filtered and Sorted Data:", filteredSorted);
        setPositions(filteredSorted);
      } else {
        console.error('Error fetching positions:', response.statusText);
      }
    };
    fetchPositions();
  }, [marketId]);

  return (
    <div className="flex flex-row gap-4 p-4">
      {/* NO Shares */}
      <div className="flex-1">
        <h2 className="text-center font-bold text-red-500 mb-2">NO Shares</h2>
        <div className="flex flex-col gap-2">
          {positions.filter(pos => pos.noSharesOwned > 0).map((pos, index) => (
            <div key={index} className="bg-gray-800 p-3 rounded-lg shadow flex flex-col">
              <Link
                to={`/user/${pos.username}`}
                className="text-blue-400 font-bold underline hover:text-blue-600"
              >
                {pos.username}
              </Link>
              <div className="text-sm text-gray-300">Shares: {pos.noSharesOwned}</div>
              <div className="text-sm text-green-400">Value: {pos.value}</div>
            </div>
          ))}
        </div>
      </div>

      {/* YES Shares */}
      <div className="flex-1">
        <h2 className="text-center font-bold text-green-500 mb-2">YES Shares</h2>
        <div className="flex flex-col gap-2">
          {positions.filter(pos => pos.yesSharesOwned > 0).map((pos, index) => (
            <div key={index} className="bg-gray-800 p-3 rounded-lg shadow flex flex-col">
              <Link
                to={`/user/${pos.username}`}
                className="text-blue-400 font-bold underline hover:text-blue-600"
              >
                {pos.username}
              </Link>
              <div className="text-sm text-gray-300">Shares: {pos.yesSharesOwned}</div>
              <div className="text-sm text-green-400">Value: {pos.value}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default PositionsActivityLayout;
