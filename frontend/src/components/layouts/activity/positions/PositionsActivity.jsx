import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getMarketLabels } from '../../../../utils/labelMapping';
import { useAuth } from '../../../../helpers/AuthContent';

const unwrapApiResponse = (payload) => {
  if (payload && typeof payload === 'object' && 'ok' in payload) {
    if (payload.ok === false) {
      throw new Error(payload.reason || 'Request failed');
    }

    if (payload.ok === true && 'result' in payload) {
      return payload.result;
    }
  }

  return payload;
};

const PositionsActivityLayout = ({ marketId, market, refreshTrigger }) => {
  const [positions, setPositions] = useState([]);
  const { token } = useAuth();

  useEffect(() => {
    const fetchPositions = async () => {
      if (!token) {
        setPositions([]);
        return;
      }

      const response = await fetch(`${API_URL}/v0/markets/positions/${marketId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        setPositions([]);
        return;
      }

      const rawData = unwrapApiResponse(await response.json());
      const filteredSorted = rawData
        .filter(user => user.noSharesOwned > 0 || user.yesSharesOwned > 0)
        .sort((a, b) => (b.noSharesOwned + b.yesSharesOwned) - (a.noSharesOwned + a.yesSharesOwned));

      setPositions(filteredSorted);
    };

    fetchPositions();
  }, [marketId, refreshTrigger, token]);

  const labels = market ? getMarketLabels(market) : { yes: "YES", no: "NO" };

  return (
    <div className="flex flex-row gap-4 p-4">
      {/* NO Shares */}
      <div className="flex-1">
        <h2 className="text-center font-bold mb-2">Shares for: <span className="text-red-500">{labels.no}</span></h2>
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
        <h2 className="text-center font-bold mb-2">Shares for: <span className="text-green-500">{labels.yes}</span></h2>
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
