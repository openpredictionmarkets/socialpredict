import React, { useEffect, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import SiteTabs from '../../tabs/SiteTabs';
import MarketLifecycleTable from '../profile/MarketLifecycleTable';
import LoadingSpinner from '../../loaders/LoadingSpinner';
import {
  approveProposedMarket,
  rejectProposedMarket,
  reassignMarketSteward,
} from '../../../api/moderationApi';
import { listAdminLifecycleMarkets } from '../../../api/lifecycleMarketsApi';
import { listAdminUsers } from '../../../api/adminUsersApi';

const reviewTabs = [
  { label: 'Proposed Markets', status: 'proposed' },
  { label: 'Approved Markets', status: 'published' },
  { label: 'Rejected Markets', status: 'rejected' },
];

const isActiveModerator = (user) => (
  user?.usertype === 'MODERATOR' && (user.moderatorStatus || 'active') === 'active'
);

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
        showSteward
        actions={status === 'proposed' ? renderActions : null}
      />
    </div>
  );
};

const MarketStewardshipQueue = () => {
  const { token } = useAuth();
  const [marketsByStatus, setMarketsByStatus] = useState({});
  const [moderators, setModerators] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyMarketId, setBusyMarketId] = useState(null);
  const [stewardForms, setStewardForms] = useState({});

  const markets = reviewTabs.flatMap((tab) => marketsByStatus[tab.status] || []);

  const loadStewardshipData = async () => {
    setLoading(true);
    setError('');
    try {
      const [usersResult, ...marketResults] = await Promise.all([
        listAdminUsers({ token, limit: 250 }),
        ...reviewTabs.map((tab) => listAdminLifecycleMarkets({ token, status: tab.status, limit: 100 })),
      ]);

      setModerators((usersResult.users || []).filter(isActiveModerator));
      setMarketsByStatus(reviewTabs.reduce((acc, tab, index) => {
        acc[tab.status] = marketResults[index]?.markets || [];
        return acc;
      }, {}));
    } catch (err) {
      setError(err.message || 'Unable to load stewardship data.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStewardshipData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const updateStewardForm = (marketId, updates) => {
    setStewardForms((current) => ({
      ...current,
      [marketId]: {
        ...(current[marketId] || {}),
        ...updates,
      },
    }));
  };

  const stewardFormFor = (market) => {
    const currentSteward = market.stewardUsername || market.creatorUsername || '';
    return {
      stewardUsername: currentSteward,
      reason: '',
      ...(stewardForms[market.id] || {}),
    };
  };

  const updateMarketInQueues = (updatedMarket) => {
    setMarketsByStatus((current) => Object.fromEntries(
      Object.entries(current).map(([status, statusMarkets]) => [
        status,
        statusMarkets.map((market) => (
          market.id === updatedMarket.id ? { ...market, ...updatedMarket } : market
        )),
      ]),
    ));
  };

  const reassignSteward = async (market) => {
    const form = stewardFormFor(market);
    setBusyMarketId(market.id);
    setError('');
    setSuccessMessage('');
    try {
      const updatedMarket = await reassignMarketSteward({
        marketId: market.id,
        token,
        stewardUsername: form.stewardUsername,
        reason: form.reason,
      });
      updateMarketInQueues(updatedMarket);
      setStewardForms((current) => ({
        ...current,
        [market.id]: {
          stewardUsername: updatedMarket.stewardUsername || form.stewardUsername,
          reason: '',
        },
      }));
      setSuccessMessage(`Market ${updatedMarket.id} steward reassigned to ${updatedMarket.stewardUsername}.`);
    } catch (err) {
      setError(err.message || 'Unable to reassign market steward.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const renderStewardshipActions = (market) => {
    const form = stewardFormFor(market);
    const currentSteward = market.stewardUsername || market.creatorUsername || '';
    const selectedSteward = String(form.stewardUsername || '').trim();
    const reason = String(form.reason || '').trim();
    const canSubmit = selectedSteward && reason && selectedSteward !== currentSteward;

    return (
      <div className="grid min-w-[260px] gap-3">
        <label className="grid gap-1 text-xs text-gray-300">
          <span className="font-mono uppercase tracking-[0.14em] text-gray-400">New steward</span>
          <select
            value={form.stewardUsername}
            onChange={(event) => updateStewardForm(market.id, { stewardUsername: event.target.value })}
            className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          >
            <option value="">Select active moderator</option>
            {moderators.map((moderator) => (
              <option key={moderator.username} value={moderator.username}>
                {moderator.username}
              </option>
            ))}
          </select>
        </label>
        <textarea
          value={form.reason}
          onChange={(event) => updateStewardForm(market.id, { reason: event.target.value })}
          rows={3}
          placeholder="Reason for stewardship reassignment"
          className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
        />
        <button
          type="button"
          disabled={busyMarketId === market.id || !canSubmit}
          onClick={() => reassignSteward(market)}
          className="rounded-md bg-sky-700 px-3 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Reassign Steward
        </button>
        {selectedSteward === currentSteward && (
          <p className="text-xs text-gray-500">Choose a different active moderator to enable reassignment.</p>
        )}
      </div>
    );
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="grid gap-4">
      <div className="rounded-lg border border-sky-800/70 bg-sky-950/30 p-4 text-sm text-sky-100">
        <p className="text-sky-100/80">
          Creator stays immutable. Reassign a market steward when a moderator is suspended,
          unavailable, conflicted, or no longer responsible for resolving the market.
        </p>
      </div>
      {error && (
        <div className="rounded-md bg-red-700 p-3 text-sm text-white">
          {error}
        </div>
      )}
      {successMessage && (
        <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">
          {successMessage}
        </div>
      )}
      {moderators.length === 0 && (
        <div className="rounded-md bg-amber-700 p-3 text-sm text-white">
          No active moderators are available for stewardship reassignment.
        </div>
      )}
      <MarketLifecycleTable
        markets={markets}
        emptyMessage="No markets found for stewardship governance."
        showCreator
        showSteward
        actions={renderStewardshipActions}
      />
    </div>
  );
};

function ModeratorMarketReview() {
  const tabsData = [
    ...reviewTabs.map((tab) => ({
      label: tab.label,
      content: <AdminMarketQueue status={tab.status} />,
    })),
    {
      label: 'Stewardship',
      content: <MarketStewardshipQueue />,
    },
  ];

  return (
    <section className="p-6 bg-primary-background shadow-md rounded-lg text-white">
      <div className="mb-6">
        <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">
          Moderator mode
        </p>
        <h1 className="text-2xl font-bold mt-2">Market Review Queue</h1>
        <p className="text-sm text-gray-300 mt-2 max-w-3xl">
          Review proposed markets without manually entering IDs. Approved markets become published and tradable; rejected markets record the reason and refund the proposal cost. Stewardship controls which active moderator is responsible for operational market governance.
        </p>
      </div>

      <SiteTabs tabs={tabsData} defaultTab="Proposed Markets" />
    </section>
  );
}

export default ModeratorMarketReview;
