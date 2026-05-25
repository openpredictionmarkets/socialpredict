import React from 'react';
import { Link } from 'react-router-dom';

const formatDate = (value) => {
  if (!value) return 'n/a';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'n/a';
  return date.toLocaleString();
};

const statusLabel = (market) => market.lifecycleStatus || market.status || 'unknown';

const MarketLifecycleTable = ({ markets = [], emptyMessage, showCreator = false, actions }) => {
  if (!markets.length) {
    return (
      <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300">
        {emptyMessage || 'No markets found.'}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-gray-700 bg-gray-900/60">
      <table className="min-w-full divide-y divide-gray-700 text-sm">
        <thead className="bg-gray-800 text-xs uppercase tracking-wide text-gray-300">
          <tr>
            <th className="px-4 py-3 text-left">Market</th>
            {showCreator && <th className="px-4 py-3 text-left">Creator</th>}
            <th className="px-4 py-3 text-left">Status</th>
            <th className="px-4 py-3 text-left">Created</th>
            <th className="px-4 py-3 text-left">Review Trail</th>
            {actions && <th className="px-4 py-3 text-left">Actions</th>}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-800 text-gray-100">
          {markets.map((market) => (
            <tr key={market.id} className="align-top">
              <td className="px-4 py-4">
                <div className="font-semibold text-white">{market.questionTitle}</div>
                <div className="mt-1 font-mono text-xs text-gray-400">ID: {market.id}</div>
                {market.description && (
                  <div className="mt-2 max-w-xl text-xs text-gray-400 line-clamp-2">{market.description}</div>
                )}
                {statusLabel(market) === 'published' && (
                  <Link className="mt-2 inline-block text-primary-pink hover:underline" to={`/markets/${market.id}`}>
                    View market
                  </Link>
                )}
              </td>
              {showCreator && (
                <td className="px-4 py-4 font-mono text-gray-300">{market.creatorUsername || 'n/a'}</td>
              )}
              <td className="px-4 py-4">
                <span className="rounded-full bg-gray-700 px-3 py-1 font-mono text-xs text-white">
                  {statusLabel(market)}
                </span>
              </td>
              <td className="px-4 py-4 text-gray-300">{formatDate(market.createdAt)}</td>
              <td className="px-4 py-4 text-xs text-gray-300">
                {market.approvedBy && <div>Approved by {market.approvedBy} at {formatDate(market.approvedAt)}</div>}
                {market.rejectedBy && <div>Rejected by {market.rejectedBy} at {formatDate(market.rejectedAt)}</div>}
                {market.rejectionReason && <div className="mt-1 text-rose-200">Reason: {market.rejectionReason}</div>}
                {!market.approvedBy && !market.rejectedBy && <span className="text-gray-500">Awaiting admin review</span>}
              </td>
              {actions && <td className="px-4 py-4">{actions(market)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default MarketLifecycleTable;
