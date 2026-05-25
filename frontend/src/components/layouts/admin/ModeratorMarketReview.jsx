import React, { useEffect, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import SiteTabs from '../../tabs/SiteTabs';
import MarketLifecycleTable from '../profile/MarketLifecycleTable';
import LoadingSpinner from '../../loaders/LoadingSpinner';
import {
  approveProposedMarket,
  rejectProposedMarket,
} from '../../../api/moderationApi';
import { listAdminLifecycleMarkets } from '../../../api/lifecycleMarketsApi';

const reviewTabs = [
  { label: 'Proposed Markets', status: 'proposed' },
  { label: 'Approved Markets', status: 'published' },
  { label: 'Rejected Markets', status: 'rejected' },
];

const AdminMarketQueue = ({ status }) => {
  const { token } = useAuth();
  const [markets, setMarkets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [rejectionReasons, setRejectionReasons] = useState({});
  const [busyMarketId, setBusyMarketId] = useState(null);

  const loadMarkets = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listAdminLifecycleMarkets({ token, status });
      setMarkets(data.markets || []);
    } catch (err) {
      setError(err.message || 'Unable to load market queue.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadMarkets();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, token]);

  const approveMarket = async (marketId) => {
    setBusyMarketId(marketId);
    setError('');
    try {
      await approveProposedMarket({ marketId, token });
      await loadMarkets();
    } catch (err) {
      setError(err.message || 'Unable to approve market.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const rejectMarket = async (marketId) => {
    const reason = rejectionReasons[marketId];
    setBusyMarketId(marketId);
    setError('');
    try {
      await rejectProposedMarket({ marketId, token, reason });
      setRejectionReasons((current) => ({ ...current, [marketId]: '' }));
      await loadMarkets();
    } catch (err) {
      setError(err.message || 'Unable to reject market.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const renderActions = (market) => {
    if (status !== 'proposed') {
      return null;
    }

    return (
      <div className="grid min-w-[220px] gap-3">
        <button
          type="button"
          disabled={busyMarketId === market.id}
          onClick={() => approveMarket(market.id)}
          className="rounded-md bg-emerald-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Approve
        </button>
        <textarea
          value={rejectionReasons[market.id] || ''}
          onChange={(event) => setRejectionReasons((current) => ({
            ...current,
            [market.id]: event.target.value,
          }))}
          rows={3}
          placeholder="Reason for rejection"
          className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
        />
        <button
          type="button"
          disabled={busyMarketId === market.id || !(rejectionReasons[market.id] || '').trim()}
          onClick={() => rejectMarket(market.id)}
          className="rounded-md bg-rose-700 px-3 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Reject and Refund
        </button>
      </div>
    );
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="grid gap-4">
      {error && (
        <div className="rounded-md bg-red-700 p-3 text-sm text-white">
          {error}
        </div>
      )}
      <MarketLifecycleTable
        markets={markets}
        emptyMessage={`No ${status} markets found.`}
        showCreator
        actions={status === 'proposed' ? renderActions : null}
      />
    </div>
  );
};

function ModeratorMarketReview() {
  const tabsData = reviewTabs.map((tab) => ({
    label: tab.label,
    content: <AdminMarketQueue status={tab.status} />,
  }));

  return (
    <section className="p-6 bg-primary-background shadow-md rounded-lg text-white">
      <div className="mb-6">
        <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">
          Moderator mode
        </p>
        <h1 className="text-2xl font-bold mt-2">Market Review Queue</h1>
        <p className="text-sm text-gray-300 mt-2 max-w-3xl">
          Review proposed markets without manually entering IDs. Approved markets become published and tradable; rejected markets record the reason and refund the proposal cost.
        </p>
      </div>

      <SiteTabs tabs={tabsData} defaultTab="Proposed Markets" />
    </section>
  );
}

export default ModeratorMarketReview;
