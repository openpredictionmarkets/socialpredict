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

const paginationButtonClass = [
  'rounded',
  'border',
  'border-transparent',
  'bg-neutral-btn',
  'px-3',
  'py-1.5',
  'text-xs',
  'font-semibold',
  'text-white',
  'transition-colors',
  'duration-200',
  'hover:bg-neutral-btn-hover',
  'disabled:cursor-not-allowed',
  'disabled:bg-custom-gray-light',
  'disabled:text-gray-400',
  'disabled:opacity-60',
].join(' ');

const PositionsActivityLayout = ({ marketId, market, refreshTrigger }) => {
  const pageSize = 20;
  const [positions, setPositions] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const { token } = useAuth();

  useEffect(() => {
    setPage(0);
  }, [marketId, refreshTrigger, token]);

  useEffect(() => {
    const fetchPositions = async () => {
      if (!token) {
        setPositions([]);
        setHasNextPage(false);
        return;
      }

      const offset = page * pageSize;
      const response = await fetch(`${API_URL}/v0/markets/positions/${marketId}?limit=${pageSize}&offset=${offset}`, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        setPositions([]);
        setHasNextPage(false);
        return;
      }

      const rawData = unwrapApiResponse(await response.json());
      const filteredSorted = rawData
        .filter(user => user.noSharesOwned > 0 || user.yesSharesOwned > 0)
        .sort((a, b) => (b.noSharesOwned + b.yesSharesOwned) - (a.noSharesOwned + a.yesSharesOwned));

      setPositions(filteredSorted);
      setHasNextPage(rawData.length === pageSize);
    };

    fetchPositions();
  }, [marketId, refreshTrigger, token, page]);

  const labels = market ? getMarketLabels(market) : { yes: "YES", no: "NO" };
  const pageStart = page * pageSize;
  const canPageBack = page > 0;
  const canPageForward = hasNextPage;

  return (
    <div className="p-4">
      <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="text-xs uppercase tracking-[0.16em] text-gray-400">
          Showing positions page {page + 1}{positions.length ? ` (${pageStart + 1}-${pageStart + positions.length})` : ''}
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setPage(current => Math.max(0, current - 1))}
            disabled={!canPageBack}
            className={paginationButtonClass}
          >
            Previous
          </button>
          <button
            type="button"
            onClick={() => setPage(current => current + 1)}
            disabled={!canPageForward}
            className={paginationButtonClass}
          >
            Next
          </button>
        </div>
      </div>
      <div className="flex flex-row gap-4">
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
    </div>
  );
};

export default PositionsActivityLayout;
