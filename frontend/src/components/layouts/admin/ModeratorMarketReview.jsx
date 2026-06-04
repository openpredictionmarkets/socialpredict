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
import {
  createAdminMarketTag,
  listAdminMarketTags,
  updateAdminMarketTag,
} from '../../../api/marketTagsApi';
import MarketTagChips from '../../markets/MarketTagChips';

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
  const [markets, setMarkets] = useState([]);
  const [moderators, setModerators] = useState([]);
  const [loading, setLoading] = useState(true);
  const [moderatorsLoading, setModeratorsLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyMarketId, setBusyMarketId] = useState(null);
  const [stewardForms, setStewardForms] = useState({});
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    if (!token) {
      return;
    }

    let ignore = false;

    const loadModerators = async () => {
      setModeratorsLoading(true);
      try {
        const usersResult = await listAdminUsers({ token, limit: 250 });
        if (!ignore) {
          setModerators((usersResult.users || []).filter(isActiveModerator));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load active moderators.');
        }
      } finally {
        if (!ignore) {
          setModeratorsLoading(false);
        }
      }
    };

    loadModerators();

    return () => {
      ignore = true;
    };
  }, [token]);

  useEffect(() => {
    if (!token) {
      return;
    }

    let ignore = false;
    const timeoutId = window.setTimeout(async () => {
      setLoading(true);
      setError('');
      try {
        const result = await listAdminLifecycleMarkets({
          token,
          status: 'all',
          query: searchQuery,
          limit: 100,
        });

        if (!ignore) {
          setMarkets(result.markets || []);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load stewardship markets.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    }, 300);

    return () => {
      ignore = true;
      window.clearTimeout(timeoutId);
    };
  }, [token, searchQuery]);

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
    setMarkets((current) => current.map((market) => (
      market.id === updatedMarket.id ? { ...market, ...updatedMarket } : market
    )));
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
    const stewardListId = `steward-options-${market.id}`;

    return (
      <div className="grid min-w-[260px] gap-3">
        <label className="grid gap-1 text-xs text-gray-300">
          <span className="font-mono uppercase tracking-[0.14em] text-gray-400">New steward</span>
          <input
            list={stewardListId}
            value={form.stewardUsername}
            onChange={(event) => updateStewardForm(market.id, { stewardUsername: event.target.value })}
            placeholder="Search active moderators by username"
            className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
          <datalist id={stewardListId}>
            {moderators.map((moderator) => (
              <option
                key={moderator.username}
                value={moderator.username}
                label={moderator.displayName || moderator.username}
              />
            ))}
          </datalist>
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

  if (loading && !markets.length) {
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
      <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <label htmlFor="stewardship-market-search" className="text-xs font-mono uppercase tracking-[0.16em] text-gray-400">
          Search stewardship markets
        </label>
        <div className="relative">
          <input
            id="stewardship-market-search"
            type="search"
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            placeholder="Search title or description across proposed, published, closed, and resolved markets"
            className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 pr-10 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
          {loading && (
            <div className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin rounded-full border-b-2 border-primary-pink" />
          )}
        </div>
        <p className="text-xs text-gray-500">
          Rejected and cancelled markets are excluded from stewardship governance.
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
      {!moderatorsLoading && moderators.length === 0 && (
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

const emptyTagForm = {
  slug: '',
  displayName: '',
  description: '',
  colorKey: 'slate',
  sortOrder: 0,
};

const MarketTaxonomyAdmin = () => {
  const { token } = useAuth();
  const [tags, setTags] = useState([]);
  const [form, setForm] = useState(emptyTagForm);
  const [loading, setLoading] = useState(true);
  const [busySlug, setBusySlug] = useState('');
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const loadTags = async () => {
    if (!token) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      const result = await listAdminMarketTags({ token, includeInactive: true });
      setTags(result.tags || []);
    } catch (err) {
      setError(err.message || 'Unable to load market tags.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTags();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const updateForm = (updates) => {
    setForm((current) => ({ ...current, ...updates }));
  };

  const createTag = async (event) => {
    event.preventDefault();
    setError('');
    setSuccessMessage('');
    try {
      const created = await createAdminMarketTag({ token, tag: form });
      setTags((current) => [...current, created].sort((left, right) => (
        (left.sortOrder - right.sortOrder) || String(left.displayName).localeCompare(String(right.displayName))
      )));
      setForm(emptyTagForm);
      setSuccessMessage(`Created tag ${created.displayName}.`);
    } catch (err) {
      setError(err.message || 'Unable to create market tag.');
    }
  };

  const setTagActive = async (tag, isActive) => {
    setBusySlug(tag.slug);
    setError('');
    setSuccessMessage('');
    try {
      const updated = await updateAdminMarketTag({
        token,
        slug: tag.slug,
        tag: {
          displayName: tag.displayName,
          description: tag.description || '',
          colorKey: tag.colorKey || 'slate',
          sortOrder: tag.sortOrder || 0,
          isActive,
          confirmDeactivate: !isActive,
        },
      });
      setTags((current) => current.map((item) => (item.slug === updated.slug ? updated : item)));
      setSuccessMessage(`${updated.displayName} is now ${updated.isActive ? 'active' : 'inactive'}.`);
    } catch (err) {
      setError(err.message || 'Unable to update market tag.');
    } finally {
      setBusySlug('');
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="grid gap-6">
      <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <h2 className="text-lg font-semibold text-white">Market Tags</h2>
        <p className="mt-2 text-sm text-gray-300">
          Admins define the tag vocabulary. Moderators can attach active tags during market creation; admins can review those tags before publication.
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

      <form onSubmit={createTag} className="grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <div className="grid gap-4 md:grid-cols-2">
          <label className="grid gap-1 text-sm text-gray-300">
            Display name
            <input
              value={form.displayName}
              onChange={(event) => updateForm({ displayName: event.target.value })}
              required
              maxLength={120}
              className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            />
          </label>
          <label className="grid gap-1 text-sm text-gray-300">
            Slug (optional)
            <input
              value={form.slug}
              onChange={(event) => updateForm({ slug: event.target.value })}
              placeholder="auto-generated from display name"
              maxLength={64}
              className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            />
          </label>
        </div>
        <label className="grid gap-1 text-sm text-gray-300">
          Description
          <textarea
            value={form.description}
            onChange={(event) => updateForm({ description: event.target.value })}
            rows={3}
            maxLength={500}
            className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
        </label>
        <div className="grid gap-4 md:grid-cols-2">
          <label className="grid gap-1 text-sm text-gray-300">
            Color key
            <select
              value={form.colorKey}
              onChange={(event) => updateForm({ colorKey: event.target.value })}
              className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            >
              <option value="slate">Slate</option>
              <option value="sky">Sky</option>
              <option value="emerald">Emerald</option>
              <option value="amber">Amber</option>
              <option value="rose">Rose</option>
            </select>
          </label>
          <label className="grid gap-1 text-sm text-gray-300">
            Sort order
            <input
              type="number"
              value={form.sortOrder}
              onChange={(event) => updateForm({ sortOrder: Number(event.target.value || 0) })}
              className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            />
          </label>
        </div>
        <button
          type="submit"
          className="justify-self-start rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-pink/80"
        >
          Create Tag
        </button>
      </form>

      <div className="grid gap-3">
        {tags.map((tag) => (
          <div key={tag.slug} className="grid gap-3 rounded-lg border border-gray-700 bg-gray-900/70 p-4 md:grid-cols-[1fr,auto] md:items-center">
            <div>
              <MarketTagChips tags={[tag]} />
              <div className="mt-2 font-mono text-xs text-gray-500">{tag.slug}</div>
              {tag.description && <p className="mt-2 text-sm text-gray-300">{tag.description}</p>}
              {!tag.isActive && <p className="mt-2 text-xs text-amber-300">Inactive tags stay visible on historical markets but cannot be newly assigned.</p>}
            </div>
            <button
              type="button"
              disabled={busySlug === tag.slug}
              onClick={() => setTagActive(tag, !tag.isActive)}
              className={`rounded-md px-3 py-2 text-sm font-semibold text-white transition disabled:cursor-not-allowed disabled:opacity-50 ${
                tag.isActive ? 'bg-amber-700 hover:bg-amber-600' : 'bg-emerald-700 hover:bg-emerald-600'
              }`}
            >
              {tag.isActive ? 'Deactivate' : 'Reactivate'}
            </button>
          </div>
        ))}
        {!tags.length && (
          <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300">
            No tags have been created yet.
          </div>
        )}
      </div>
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
    {
      label: 'Tags',
      content: <MarketTaxonomyAdmin />,
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
