import React, { useEffect, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import SiteTabs from '../../tabs/SiteTabs';
import MarketLifecycleTable from '../profile/MarketLifecycleTable';
import LoadingSpinner from '../../loaders/LoadingSpinner';
import {
  approveProposedMarket,
  rejectProposedMarket,
  reassignMarketSteward,
  updateMarketTags,
} from '../../../api/moderationApi';
import { listAdminLifecycleMarkets } from '../../../api/lifecycleMarketsApi';
import { listAdminUsers } from '../../../api/adminUsersApi';
import {
  createAdminMarketTag,
  listAdminMarketTags,
  updateAdminMarketTag,
} from '../../../api/marketTagsApi';
import MarketTagChips, { MARKET_TAG_COLOR_OPTIONS } from '../../markets/MarketTagChips';
import MarkdownLite from '../../markdown/MarkdownLite';
import {
  getMarketGovernanceSettings,
  listAdminMarketDescriptionAmendments,
  reviewMarketDescriptionAmendment,
  updateMarketGovernanceSettings,
} from '../../../api/marketDescriptionAmendmentsApi';

const reviewTabs = [
  { label: 'Proposed Markets', status: 'proposed' },
  { label: 'Approved Markets', status: 'published' },
  { label: 'Rejected Markets', status: 'rejected' },
];

const amendmentReviewTabs = [
  { label: 'Pending Amendments', status: 'pending' },
  { label: 'Approved Amendments', status: 'approved' },
  { label: 'Rejected Amendments', status: 'rejected' },
];

const maxMarketTagsPerMarket = 5;

const isActiveModerator = (user) => (
  user?.usertype === 'MODERATOR' && (user.moderatorStatus || 'active') === 'active'
);

const marketTagSlugs = (market) => (market.tags || [])
  .map((tag) => tag.slug)
  .filter(Boolean)
  .sort();

const sameTagSlugs = (left, right) => (
  JSON.stringify([...left].sort()) === JSON.stringify([...right].sort())
);

const AdminMarketQueue = ({ status }) => {
  const { token } = useAuth();
  const [markets, setMarkets] = useState([]);
  const [activeTags, setActiveTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [rejectionReasons, setRejectionReasons] = useState({});
  const [tagForms, setTagForms] = useState({});
  const [busyMarketId, setBusyMarketId] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  const tagEditingEnabled = status === 'proposed' || status === 'published';

  const loadMarkets = async (query = searchQuery) => {
    setLoading(true);
    setError('');
    try {
      const data = await listAdminLifecycleMarkets({
        token,
        status,
        query,
        limit: 100,
      });
      setMarkets(data.markets || []);
    } catch (err) {
      setError(err.message || 'Unable to load market queue.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      loadMarkets(searchQuery);
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, token, searchQuery]);

  useEffect(() => {
    if (!token || !tagEditingEnabled) {
      return;
    }

    let ignore = false;
    const loadActiveTags = async () => {
      try {
        const result = await listAdminMarketTags({ token, includeInactive: false });
        if (!ignore) {
          setActiveTags(result.tags || []);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load active market tags.');
        }
      }
    };

    loadActiveTags();

    return () => {
      ignore = true;
    };
  }, [token, tagEditingEnabled]);

  const approveMarket = async (marketId) => {
    setBusyMarketId(marketId);
    setError('');
    setSuccessMessage('');
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
    setSuccessMessage('');
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

  const updateMarketInQueue = (updatedMarket) => {
    setMarkets((current) => current.map((market) => (
      market.id === updatedMarket.id ? { ...market, ...updatedMarket } : market
    )));
  };

  const tagSlugsForForm = (market) => tagForms[market.id] || marketTagSlugs(market);

  const toggleMarketTag = (market, slug) => {
    setTagForms((current) => {
      const selected = current[market.id] || marketTagSlugs(market);
      const next = selected.includes(slug)
        ? selected.filter((item) => item !== slug)
        : [...selected, slug].sort();
      return { ...current, [market.id]: next };
    });
  };

  const saveMarketTags = async (market) => {
    const tagSlugs = tagSlugsForForm(market);
    setBusyMarketId(market.id);
    setError('');
    setSuccessMessage('');
    try {
      const updatedMarket = await updateMarketTags({ marketId: market.id, token, tagSlugs });
      updateMarketInQueue(updatedMarket);
      setTagForms((current) => {
        const next = { ...current };
        delete next[market.id];
        return next;
      });
      setSuccessMessage(`Updated tags for market ${updatedMarket.id}.`);
    } catch (err) {
      setError(err.message || 'Unable to update market tags.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const renderTagEditor = (market) => {
    if (!tagEditingEnabled) {
      return null;
    }
    const activeBySlug = new Map(activeTags.map((tag) => [tag.slug, tag]));
    const tagChoices = [
      ...activeTags,
      ...(market.tags || []).filter((tag) => tag.slug && !activeBySlug.has(tag.slug)),
    ];
    const selectedSlugs = tagSlugsForForm(market);
    const originalSlugs = marketTagSlugs(market);
    const changed = !sameTagSlugs(selectedSlugs, originalSlugs);
    const atLimit = selectedSlugs.length >= maxMarketTagsPerMarket;

    return (
      <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/80 p-3">
        <div>
          <div className="font-mono text-xs uppercase tracking-[0.14em] text-gray-400">
            Admin tag adjustment
          </div>
          <p className="mt-1 text-xs text-gray-500">
            Add or remove active tags before or after publication.
          </p>
        </div>
        {tagChoices.length ? (
          <div className="flex max-w-[280px] flex-wrap gap-2">
            {tagChoices.map((tag) => {
              const selected = selectedSlugs.includes(tag.slug);
              const disabled = !selected && (atLimit || !tag.isActive);
              return (
                <button
                  key={tag.slug}
                  type="button"
                  disabled={disabled || busyMarketId === market.id}
                  onClick={() => toggleMarketTag(market, tag.slug)}
                  className={`rounded-full border px-2.5 py-1 text-xs font-semibold transition disabled:cursor-not-allowed disabled:opacity-50 ${
                    selected
                      ? 'border-primary-pink bg-primary-pink/20 text-white'
                      : 'border-gray-600 bg-gray-800 text-gray-300 hover:border-gray-400'
                  }`}
                  title={disabled ? `A market can have up to ${maxMarketTagsPerMarket} active tags.` : tag.description || tag.displayName}
                >
                  {selected ? '✓ ' : ''}
                  {tag.displayName || tag.slug}
                  {!tag.isActive ? ' (inactive)' : ''}
                </button>
              );
            })}
          </div>
        ) : (
          <p className="text-xs text-amber-200">Create active tags in the Tags tab before assigning them to markets.</p>
        )}
        <button
          type="button"
          disabled={busyMarketId === market.id || !changed}
          onClick={() => saveMarketTags(market)}
          className="justify-self-start rounded-md bg-sky-700 px-3 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Save Tags
        </button>
      </div>
    );
  };

  const renderActions = (market) => {
    if (!tagEditingEnabled) {
      return null;
    }

    return (
      <div className="grid min-w-[220px] gap-3">
        {renderTagEditor(market)}
        {status === 'proposed' && (
          <>
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
          </>
        )}
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
      {successMessage && (
        <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">
          {successMessage}
        </div>
      )}
      <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <label htmlFor={`market-review-search-${status}`} className="text-xs font-mono uppercase tracking-[0.16em] text-gray-400">
          Search markets
        </label>
        <div className="relative">
          <input
            id={`market-review-search-${status}`}
            type="search"
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            placeholder={`Search ${status} markets by title or description`}
            className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 pr-10 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
          {loading && (
            <div className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin rounded-full border-b-2 border-primary-pink" />
          )}
        </div>
      </div>
      <MarketLifecycleTable
        markets={markets}
        emptyMessage={`No ${status} markets found.`}
        showCreator
        showSteward
        actions={tagEditingEnabled ? renderActions : null}
      />
    </div>
  );
};

const DescriptionAmendmentStatusQueue = ({ status }) => {
  const { token } = useAuth();
  const [amendments, setAmendments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyAmendmentId, setBusyAmendmentId] = useState(null);
  const [reasonById, setReasonById] = useState({});
  const canReview = status === 'pending';

  const loadAmendments = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listAdminMarketDescriptionAmendments({
        token,
        status,
        limit: 100,
      });
      setAmendments(data.amendments || []);
    } catch (err) {
      setError(err.message || 'Unable to load description amendments.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    loadAmendments();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, status]);

  const setReason = (amendmentId, reason) => {
    setReasonById((current) => ({
      ...current,
      [amendmentId]: reason,
    }));
  };

  const reviewAmendment = async (amendment, nextStatus) => {
    const reason = String(reasonById[amendment.id] || '').trim();
    setBusyAmendmentId(amendment.id);
    setError('');
    setSuccessMessage('');
    try {
      await reviewMarketDescriptionAmendment({
        token,
        amendmentId: amendment.id,
        status: nextStatus,
        reason,
      });
      setReason(amendment.id, '');
      setSuccessMessage(`Amendment v${amendment.version} ${nextStatus}.`);
      await loadAmendments();
    } catch (err) {
      setError(err.message || 'Unable to review description amendment.');
    } finally {
      setBusyAmendmentId(null);
    }
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
      {successMessage && (
        <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">
          {successMessage}
        </div>
      )}
      {amendments.length === 0 && (
        <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300">
          No {status} description amendments found.
        </div>
      )}
      {amendments.map((amendment) => {
        const reason = reasonById[amendment.id] || '';
        const marketTitle = amendment.marketTitle || `Market #${amendment.marketId}`;
        const previousAmendments = Array.isArray(amendment.previousApprovedAmendments)
          ? amendment.previousApprovedAmendments
          : [];
        return (
          <article key={amendment.id} className="grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
            <div className="grid gap-2">
              <div className="flex flex-wrap items-center gap-2 text-sm text-gray-300">
                <span className="rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100">
                  Market #{amendment.marketId}
                </span>
                <span className="rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200">
                  Amendment {amendment.version}
                </span>
                <span>Submitted by @{amendment.createdBy}</span>
                <span>{amendment.createdAt ? new Date(amendment.createdAt).toLocaleString() : ''}</span>
              </div>
              <a
                href={`/markets/${amendment.marketId}`}
                className="text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200"
              >
                {marketTitle}
              </a>
            </div>
            {amendment.submitReason && (
              <div className="rounded-md border border-gray-700 bg-gray-950 p-3 text-sm text-gray-300">
                <span className="font-semibold text-gray-100">Submit reason:</span> {amendment.submitReason}
              </div>
            )}
            <div className="grid gap-3 rounded-md border border-gray-700 bg-gray-950 p-4">
              <div>
                <p className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">Description</p>
                <p className="whitespace-pre-wrap text-sm leading-6 text-gray-200">
                  {amendment.marketDescription || 'No market description was returned.'}
                </p>
              </div>
              {previousAmendments.length > 0 && (
                <div className="grid gap-2">
                  {previousAmendments.map((previous) => (
                    <article key={previous.id || previous.version} className="rounded-md border border-gray-700 bg-gray-900 p-3">
                      <div className="mb-2 flex flex-wrap gap-2 text-xs text-gray-400">
                        <span>Amendment {previous.version}</span>
                        <span>Approved by @{previous.approvedBy || 'admin'}</span>
                        {previous.approvedAt && <span>{new Date(previous.approvedAt).toLocaleString()}</span>}
                      </div>
                      <MarkdownLite>{previous.body}</MarkdownLite>
                    </article>
                  ))}
                </div>
              )}
            </div>
            <div className="rounded-md border border-sky-900/70 bg-sky-950/20 p-4">
              <p className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-sky-200">
                Proposed Amendment {amendment.version}
              </p>
              <MarkdownLite>{amendment.body}</MarkdownLite>
            </div>
            {status === 'rejected' && amendment.rejectionReason && (
              <div className="rounded-md border border-rose-800/70 bg-rose-950/30 p-3 text-sm text-rose-100">
                Rejection reason: {amendment.rejectionReason}
              </div>
            )}
            {canReview && (
              <div className="grid gap-3 md:grid-cols-[minmax(0,1fr),auto,auto] md:items-start">
                <textarea
                  value={reason}
                  onChange={(event) => setReason(amendment.id, event.target.value)}
                  rows={3}
                  placeholder="Decision reason required"
                  className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                />
                <button
                  type="button"
                  disabled={busyAmendmentId === amendment.id || !reason.trim()}
                  onClick={() => reviewAmendment(amendment, 'approved')}
                  className="rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Approve
                </button>
                <button
                  type="button"
                  disabled={busyAmendmentId === amendment.id || !reason.trim()}
                  onClick={() => reviewAmendment(amendment, 'rejected')}
                  className="rounded-md bg-rose-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Reject
                </button>
              </div>
            )}
          </article>
        );
      })}
    </div>
  );
};

const DescriptionAmendmentQueue = () => {
  const { token } = useAuth();
  const [settings, setSettings] = useState({
    autoApproveDescriptionAmendments: false,
    version: 0,
  });
  const [settingsDraft, setSettingsDraft] = useState(false);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [settingsError, setSettingsError] = useState('');
  const [settingsMessage, setSettingsMessage] = useState('');
  const tabsData = amendmentReviewTabs.map((tab) => ({
    label: tab.label,
    content: <DescriptionAmendmentStatusQueue status={tab.status} />,
  }));

  useEffect(() => {
    if (!token) return;
    let ignore = false;
    const loadSettings = async () => {
      setSettingsLoading(true);
      setSettingsError('');
      try {
        const data = await getMarketGovernanceSettings({ token });
        if (!ignore) {
          setSettings(data);
          setSettingsDraft(Boolean(data.autoApproveDescriptionAmendments));
        }
      } catch (err) {
        if (!ignore) {
          setSettingsError(err.message || 'Unable to load governance settings.');
        }
      } finally {
        if (!ignore) {
          setSettingsLoading(false);
        }
      }
    };
    loadSettings();
    return () => {
      ignore = true;
    };
  }, [token]);

  const saveSettings = async () => {
    setSettingsSaving(true);
    setSettingsError('');
    setSettingsMessage('');
    try {
      const saved = await updateMarketGovernanceSettings({
        token,
        autoApproveDescriptionAmendments: settingsDraft,
        version: settings.version,
      });
      setSettings(saved);
      setSettingsDraft(Boolean(saved.autoApproveDescriptionAmendments));
      setSettingsMessage('Amendment auto-approval setting saved.');
    } catch (err) {
      setSettingsError(err.message || 'Unable to save governance settings.');
    } finally {
      setSettingsSaving(false);
    }
  };

  const settingsChanged = settingsDraft !== Boolean(settings.autoApproveDescriptionAmendments);

  return (
    <div className="grid gap-4">
      <div className="rounded-lg border border-sky-800/70 bg-sky-950/30 p-4 text-sm text-sky-100">
        Description amendments are append-only contract clarifications. Approving one makes it visible on the public market page.
      </div>
      <div className="grid gap-3 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="font-semibold text-white">Auto-approve new amendments</p>
            <p className="mt-1 text-sm text-gray-400">
              When enabled, newly proposed steward amendments are immediately accepted. Turn it off and save to restore manual admin review.
            </p>
          </div>
          <label className="inline-flex cursor-pointer items-center gap-3 text-sm text-gray-200">
            <input
              type="checkbox"
              checked={settingsDraft}
              disabled={settingsLoading || settingsSaving}
              onChange={(event) => setSettingsDraft(event.target.checked)}
              className="h-5 w-5 rounded border-gray-600 bg-gray-800 text-primary-pink focus:ring-primary-pink"
            />
            <span>{settingsDraft ? 'Auto-accept on' : 'Auto-accept off'}</span>
          </label>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            disabled={settingsLoading || settingsSaving || !settingsChanged}
            onClick={saveSettings}
            className="rounded-md bg-sky-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {settingsSaving ? 'Saving...' : 'Save Auto-Approval Setting'}
          </button>
          <span className="text-xs text-gray-500">Version {settings.version || 1}</span>
        </div>
        {settingsError && <div className="rounded-md bg-red-700 p-3 text-sm text-white">{settingsError}</div>}
        {settingsMessage && <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">{settingsMessage}</div>}
      </div>
      <SiteTabs tabs={tabsData} defaultTab="Pending Amendments" />
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

const ColorKeyPicker = ({ value, onChange }) => {
  const [open, setOpen] = useState(false);
  const selectedOption = MARKET_TAG_COLOR_OPTIONS.find((option) => option.key === value)
    || MARKET_TAG_COLOR_OPTIONS[0];
  const previewTag = (option) => ({
    slug: option.key,
    displayName: option.label,
    colorKey: option.key,
    description: option.label,
  });

  const chooseColor = (colorKey) => {
    onChange(colorKey);
    setOpen(false);
  };

  return (
    <div className="relative">
      <button
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => setOpen((current) => !current)}
        className="flex w-full items-center justify-between gap-3 rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-left text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
      >
        <MarketTagChips tags={[previewTag(selectedOption)]} />
        <span className="text-xs text-gray-400">▾</span>
      </button>
      {open && (
        <div
          role="listbox"
          aria-label="Market tag color key"
          className="absolute z-30 mt-2 max-h-72 w-full overflow-y-auto rounded-lg border border-gray-700 bg-gray-950 p-2 shadow-xl"
        >
          {MARKET_TAG_COLOR_OPTIONS.map((option) => {
            const selected = option.key === selectedOption.key;
            return (
              <button
                key={option.key}
                type="button"
                role="option"
                aria-selected={selected}
                onClick={() => chooseColor(option.key)}
                className={`grid w-full gap-1 rounded-md px-3 py-2 text-left transition hover:bg-gray-800 ${
                  selected ? 'bg-gray-800 ring-1 ring-primary-pink/60' : ''
                }`}
              >
                <MarketTagChips tags={[previewTag(option)]} />
                <span className="font-mono text-xs text-gray-500">{option.key}</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
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
    if (!isActive) {
      const confirmed = window.confirm(
        `Deactivate "${tag.displayName}"?\n\nIt will stay visible on markets and existing layouts, but moderators/admins cannot newly assign it until reactivated.`,
      );
      if (!confirmed) {
        return;
      }
    }

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
        <p className="mt-2 text-xs text-amber-200">
          Deactivating a tag does not remove it from existing markets or topic pins. It only prevents new assignments until the tag is reactivated.
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
            <ColorKeyPicker
              value={form.colorKey}
              onChange={(colorKey) => updateForm({ colorKey })}
            />
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
              {!tag.isActive && <p className="mt-2 text-xs text-amber-300">Inactive tags stay visible on historical markets and layouts but cannot be newly assigned.</p>}
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
      label: 'Description Amendments',
      content: <DescriptionAmendmentQueue />,
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
      </div>

      <SiteTabs tabs={tabsData} defaultTab="Proposed Markets" />
    </section>
  );
}

export default ModeratorMarketReview;
