import React, { useState } from 'react';
import { Link } from 'react-router-dom';

const formatDate = (value) => {
  if (!value) return 'n/a';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'n/a';
  return date.toLocaleString();
};

const statusLabel = (market) => market.lifecycleStatus || market.status || 'unknown';

const marketLabel = (value, fallback) => {
  const label = String(value || '').trim();
  return label || fallback;
};

const MarketLabels = ({ market }) => (
  <div className="mt-3 grid gap-2 text-xs sm:grid-cols-2">
    <div className="rounded-md border border-emerald-700/70 bg-emerald-950/40 px-3 py-2">
      <div className="font-mono uppercase tracking-[0.16em] text-emerald-300">YES label</div>
      <div className="mt-1 break-words font-semibold text-emerald-50">{marketLabel(market.yesLabel, 'YES')}</div>
    </div>
    <div className="rounded-md border border-rose-700/70 bg-rose-950/40 px-3 py-2">
      <div className="font-mono uppercase tracking-[0.16em] text-rose-300">NO label</div>
      <div className="mt-1 break-words font-semibold text-rose-50">{marketLabel(market.noLabel, 'NO')}</div>
    </div>
  </div>
);

const MarketLifecycleTable = ({ markets = [], emptyMessage, showCreator = false, showSteward = false, actions }) => {
  const [expandedDescriptions, setExpandedDescriptions] = useState({});

  const toggleDescription = (marketId) => {
    setExpandedDescriptions((current) => ({
      ...current,
      [marketId]: !current[marketId],
    }));
  };

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
            {showSteward && <th className="px-4 py-3 text-left">Steward</th>}
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
                  <div className="mt-2 max-w-xl text-xs text-gray-300">
                    <div className={expandedDescriptions[market.id] ? 'whitespace-pre-wrap break-words' : 'line-clamp-2 break-words'}>
                      {market.description}
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleDescription(market.id)}
                      className="mt-2 text-primary-pink hover:underline"
                    >
                      {expandedDescriptions[market.id] ? 'Show less description' : 'Show full description'}
                    </button>
                  </div>
                )}
                <MarketLabels market={market} />
                {statusLabel(market) === 'published' && (
                  <Link className="mt-2 inline-block text-primary-pink hover:underline" to={`/markets/${market.id}`}>
                    View market
                  </Link>
                )}
              </td>
              {showCreator && (
                <td className="px-4 py-4 font-mono text-gray-300">{market.creatorUsername || 'n/a'}</td>
              )}
              {showSteward && (
                <td className="px-4 py-4 font-mono text-gray-300">{market.stewardUsername || market.creatorUsername || 'n/a'}</td>
              )}
              <td className="px-4 py-4">
                <span className="rounded-full bg-gray-700 px-3 py-1 font-mono text-xs text-white">
                  {statusLabel(market)}
                </span>
              </td>
              <td className="px-4 py-4 text-gray-300">{formatDate(market.createdAt)}</td>
              <td className="px-4 py-4 text-xs text-gray-300">
                {market.proposalCost > 0 && <div>Proposal cost: {market.proposalCost} credits</div>}
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
