import React, { useEffect, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import SiteTabs from '../../tabs/SiteTabs';
import MarketLifecycleTable from '../profile/MarketLifecycleTable';
import LoadingSpinner from '../../loaders/LoadingSpinner';
import {
  approveProposedMarket,
  approveProposedMarketGroup,
  listAdminMarketGroupAnswerAdditions,
  rejectProposedMarket,
  rejectProposedMarketGroup,
  reassignMarketGroupSteward,
  reassignMarketSteward,
  reviewMarketGroupAnswerAddition,
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

const answerAdditionReviewTabs = [
  { label: 'Pending Answers', status: 'pending' },
  { label: 'Approved Answers', status: 'approved' },
  { label: 'Rejected Answers', status: 'rejected' },
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

const uniqueTagsBySlug = (markets = []) => {
  const seen = new Set();
  const tags = [];
  markets.forEach((market) => {
    (market.tags || []).forEach((tag) => {
      const key = tag.slug || tag.id || tag.displayName;
      if (!key || seen.has(key)) {
        return;
      }
      seen.add(key);
      tags.push(tag);
    });
  });
  return tags;
};

const reviewRowKey = (market) => (
  market?.isMarketGroup ? `group:${market.marketGroup.id}` : `market:${market.id}`
);

const groupAdminMarketRows = (markets = []) => {
  const rows = [];
  const groups = new Map();

  markets.forEach((market) => {
    const group = market.marketGroup;
    if (!group?.id) {
      rows.push({ ...market, rowKey: `market:${market.id}` });
      return;
    }

    const existing = groups.get(group.id);
    if (!existing) {
      const row = {
        ...market,
        id: group.id,
        rowKey: `group:${group.id}`,
        isMarketGroup: true,
        questionTitle: group.questionTitle || market.questionTitle,
        description: group.description || '',
        creatorUsername: group.creatorUsername || market.creatorUsername,
        stewardUsername: group.stewardUsername || market.stewardUsername || market.creatorUsername,
        lifecycleStatus: group.lifecycleStatus || market.lifecycleStatus,
        status: group.status || market.status,
        proposalCost: group.proposalCost ?? market.proposalCost,
        marketGroup: group,
        childMarkets: [market],
        tags: uniqueTagsBySlug([market]),
      };
      groups.set(group.id, row);
      rows.push(row);
      return;
    }

    existing.childMarkets.push(market);
    existing.tags = uniqueTagsBySlug(existing.childMarkets);
  });

  return rows;
};

const amendmentRowKey = (amendment) => (
  amendment?.isMarketGroupAmendment
    ? `group:${amendment.marketGroup.id}:${amendment.body}:${amendment.createdBy}:${amendment.status}`
    : `amendment:${amendment.id}`
);

const groupDescriptionAmendmentRows = (amendments = []) => {
  const rows = [];
  const groups = new Map();

  amendments.forEach((amendment) => {
    const group = amendment.marketGroup;
    if (!group?.id) {
      rows.push(amendment);
      return;
    }

    const key = [
      group.id,
      amendment.status,
      amendment.body,
      amendment.createdBy,
      amendment.submitReason || '',
    ].join('|');
    const existing = groups.get(key);
    if (!existing) {
      const row = {
        ...amendment,
        isMarketGroupAmendment: true,
        marketTitle: group.questionTitle || amendment.marketTitle,
        marketDescription: group.description || amendment.marketDescription,
        childAmendments: [amendment],
      };
      groups.set(key, row);
      rows.push(row);
      return;
    }

    existing.childAmendments.push(amendment);
    existing.childAmendments.sort((left, right) => Number(left.marketId || 0) - Number(right.marketId || 0));
  });

  return rows;
};

const GovernanceAutoApprovalSetting = ({
  settingKey,
  title,
  description,
  savedMessage,
}) => {
  const { token } = useAuth();
  const [settings, setSettings] = useState({
    autoApproveDescriptionAmendments: false,
    autoApproveMarketProposals: false,
    autoApproveMarketGroupAnswers: false,
    version: 0,
  });
  const [draft, setDraft] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) return;
    let ignore = false;
    const loadSettings = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await getMarketGovernanceSettings({ token });
        if (!ignore) {
          setSettings(data);
          setDraft(Boolean(data[settingKey]));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load governance settings.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };
    loadSettings();
    return () => {
      ignore = true;
    };
  }, [token, settingKey]);

  const saveSettings = async () => {
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const nextSettings = {
        autoApproveDescriptionAmendments: Boolean(settings.autoApproveDescriptionAmendments),
        autoApproveMarketProposals: Boolean(settings.autoApproveMarketProposals),
        autoApproveMarketGroupAnswers: Boolean(settings.autoApproveMarketGroupAnswers),
        [settingKey]: draft,
      };
      const saved = await updateMarketGovernanceSettings({
        token,
        autoApproveDescriptionAmendments: nextSettings.autoApproveDescriptionAmendments,
        autoApproveMarketProposals: nextSettings.autoApproveMarketProposals,
        autoApproveMarketGroupAnswers: nextSettings.autoApproveMarketGroupAnswers,
        version: settings.version,
      });
      setSettings(saved);
      setDraft(Boolean(saved[settingKey]));
      setMessage(savedMessage);
    } catch (err) {
      setError(err.message || 'Unable to save governance settings.');
    } finally {
      setSaving(false);
    }
  };

  const changed = draft !== Boolean(settings[settingKey]);

  return (
    <div className="grid gap-3 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <label className="flex cursor-pointer items-start gap-3 text-sm text-gray-200">
          <input
            type="checkbox"
            checked={draft}
            disabled={loading || saving}
            onChange={(event) => setDraft(event.target.checked)}
            className="mt-1 h-5 w-5 rounded border-gray-600 bg-gray-800 text-primary-pink focus:ring-primary-pink"
          />
          <span>
            <span className="block font-semibold text-white">{title}</span>
            <span className="mt-1 block text-gray-400">{description}</span>
          </span>
        </label>
        <button
          type="button"
          disabled={loading || saving || !changed}
          onClick={saveSettings}
          className="rounded-md bg-sky-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {saving ? 'Saving...' : 'Save Setting'}
        </button>
      </div>
      <span className="text-xs text-gray-500">Version {settings.version || 1}</span>
      {error && <div className="rounded-md bg-red-700 p-3 text-sm text-white">{error}</div>}
      {message && <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">{message}</div>}
    </div>
  );
};

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

  const approveMarket = async (market) => {
    const key = reviewRowKey(market);
    setBusyMarketId(key);
    setError('');
    setSuccessMessage('');
    try {
      if (market?.isMarketGroup) {
        await approveProposedMarketGroup({ groupId: market.marketGroup.id, token });
      } else {
        await approveProposedMarket({ marketId: market.id, token });
      }
      await loadMarkets();
    } catch (err) {
      setError(err.message || 'Unable to approve market.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const rejectMarket = async (market) => {
    const key = reviewRowKey(market);
    const reason = rejectionReasons[key];
    setBusyMarketId(key);
    setError('');
    setSuccessMessage('');
    try {
      if (market?.isMarketGroup) {
        await rejectProposedMarketGroup({ groupId: market.marketGroup.id, token, reason });
      } else {
        await rejectProposedMarket({ marketId: market.id, token, reason });
      }
      setRejectionReasons((current) => ({ ...current, [key]: '' }));
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
    const key = reviewRowKey(market);

    return (
      <div className="grid min-w-[220px] gap-3">
        {!market.isMarketGroup && renderTagEditor(market)}
        {status === 'proposed' && (
          <>
            <button
              type="button"
              disabled={busyMarketId === key}
              onClick={() => approveMarket(market)}
              className="rounded-md bg-emerald-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {market.isMarketGroup ? 'Approve Group' : 'Approve'}
            </button>
            <textarea
              value={rejectionReasons[key] || ''}
              onChange={(event) => setRejectionReasons((current) => ({
                ...current,
                [key]: event.target.value,
              }))}
              rows={3}
              placeholder="Reason for rejection"
              className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            />
            <button
              type="button"
              disabled={busyMarketId === key || !(rejectionReasons[key] || '').trim()}
              onClick={() => rejectMarket(market)}
              className="rounded-md bg-rose-700 px-3 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {market.isMarketGroup ? 'Reject Group and Refund' : 'Reject and Refund'}
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
      {status === 'proposed' && (
        <GovernanceAutoApprovalSetting
          settingKey="autoApproveMarketProposals"
          title="Auto-approve new market proposals"
          description="When enabled, new moderator-created proposals become published and tradable immediately."
          savedMessage="Market proposal auto-approval setting saved."
        />
      )}
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
        markets={groupAdminMarketRows(markets)}
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

  const setReason = (amendmentKey, reason) => {
    setReasonById((current) => ({
      ...current,
      [amendmentKey]: reason,
    }));
  };

  const reviewAmendment = async (amendment, nextStatus) => {
    const key = amendmentRowKey(amendment);
    const reason = String(reasonById[key] || '').trim();
    const childAmendments = amendment.isMarketGroupAmendment
      ? amendment.childAmendments || []
      : [amendment];
    setBusyAmendmentId(key);
    setError('');
    setSuccessMessage('');
    try {
      await Promise.all(childAmendments.map((child) => (
        reviewMarketDescriptionAmendment({
          token,
          amendmentId: child.id,
          status: nextStatus,
          reason,
        })
      )));
      setReason(key, '');
      setSuccessMessage(`${amendment.isMarketGroupAmendment ? 'Grouped amendment' : `Amendment v${amendment.version}`} ${nextStatus}.`);
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
      {groupDescriptionAmendmentRows(amendments).map((amendment) => {
        const key = amendmentRowKey(amendment);
        const childAmendments = amendment.isMarketGroupAmendment
          ? amendment.childAmendments || []
          : [amendment];
        const reason = reasonById[key] || '';
        const marketTitle = amendment.marketTitle || `Market #${amendment.marketId}`;
        const previousAmendments = Array.isArray(amendment.previousApprovedAmendments)
          ? amendment.previousApprovedAmendments
          : [];
        const primaryMarketID = childAmendments[0]?.marketId || amendment.marketId;
        return (
          <article key={key} className="grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
            <div className="grid gap-2">
              <div className="flex flex-wrap items-center gap-2 text-sm text-gray-300">
                <span className="rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100">
                  {amendment.isMarketGroupAmendment ? `Group #${amendment.marketGroup.id}` : `Market #${amendment.marketId}`}
                </span>
                <span className="rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200">
                  {amendment.isMarketGroupAmendment ? 'Grouped Amendment' : `Amendment ${amendment.version}`}
                </span>
                <span>Submitted by @{amendment.createdBy}</span>
                <span>{amendment.createdAt ? new Date(amendment.createdAt).toLocaleString() : ''}</span>
              </div>
              <a
                href={`/markets/${primaryMarketID}`}
                className="text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200"
              >
                {marketTitle}
              </a>
              {amendment.isMarketGroupAmendment && (
                <div className="flex flex-wrap gap-2">
                  {childAmendments.map((child) => (
                    <span key={child.id} className="rounded-full border border-sky-800/70 bg-sky-900/40 px-2.5 py-1 text-xs text-sky-100">
                      {child.marketGroup?.answerLabel || `Market #${child.marketId}`} · Amendment {child.version}
                    </span>
                  ))}
                </div>
              )}
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
                  onChange={(event) => setReason(key, event.target.value)}
                  rows={3}
                  placeholder="Decision reason required"
                  className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                />
                <button
                  type="button"
                  disabled={busyAmendmentId === key || !reason.trim()}
                  onClick={() => reviewAmendment(amendment, 'approved')}
                  className="rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {amendment.isMarketGroupAmendment ? 'Approve Group' : 'Approve'}
                </button>
                <button
                  type="button"
                  disabled={busyAmendmentId === key || !reason.trim()}
                  onClick={() => reviewAmendment(amendment, 'rejected')}
                  className="rounded-md bg-rose-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {amendment.isMarketGroupAmendment ? 'Reject Group' : 'Reject'}
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
  const tabsData = amendmentReviewTabs.map((tab) => ({
    label: tab.label,
    content: <DescriptionAmendmentStatusQueue status={tab.status} />,
  }));

  return (
    <div className="grid gap-4">
      <div className="rounded-lg border border-sky-800/70 bg-sky-950/30 p-4 text-sm text-sky-100">
        Description amendments are append-only contract clarifications. Approving one makes it visible on the public market page.
      </div>
      <GovernanceAutoApprovalSetting
        settingKey="autoApproveDescriptionAmendments"
        title="Auto-approve new amendments"
        description="When enabled, newly proposed steward amendments are immediately accepted."
        savedMessage="Amendment auto-approval setting saved."
      />
      <SiteTabs tabs={tabsData} defaultTab="Pending Amendments" />
    </div>
  );
};

const MarketGroupAnswerAdditionStatusQueue = ({ status }) => {
  const { token } = useAuth();
  const [additions, setAdditions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyAdditionId, setBusyAdditionId] = useState(null);
  const [reasonById, setReasonById] = useState({});
  const canReview = status === 'pending';

  const loadAdditions = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listAdminMarketGroupAnswerAdditions({
        token,
        status,
        limit: 100,
      });
      setAdditions(data.additions || []);
    } catch (err) {
      setError(err.message || 'Unable to load grouped answer additions.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    loadAdditions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, status]);

  const updateReason = (additionId, reason) => {
    setReasonById((current) => ({
      ...current,
      [additionId]: reason,
    }));
  };

  const reviewAddition = async (addition, nextStatus) => {
    const reason = String(reasonById[addition.id] || '').trim();
    if (nextStatus === 'rejected' && !reason) {
      setError('A rejection reason is required.');
      return;
    }
    setBusyAdditionId(addition.id);
    setError('');
    setSuccessMessage('');
    try {
      await reviewMarketGroupAnswerAddition({
        token,
        additionId: addition.id,
        status: nextStatus,
        reason,
        confirm: nextStatus === 'approved',
      });
      updateReason(addition.id, '');
      setSuccessMessage(`Answer option "${addition.answerLabel}" ${nextStatus}.`);
      await loadAdditions();
    } catch (err) {
      setError(err.message || 'Unable to review grouped answer addition.');
    } finally {
      setBusyAdditionId(null);
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
      {additions.length === 0 && (
        <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300">
          No {status} answer additions found.
        </div>
      )}
      {additions.map((addition) => {
        const group = addition.marketGroup || {};
        const marketHref = addition.marketId
          ? `/markets/${addition.marketId}`
          : (group.id ? `/markets/group/${group.id}` : '#');
        const reason = reasonById[addition.id] || '';
        return (
          <article key={addition.id} className="grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
            <div className="flex flex-wrap items-center gap-2 text-sm text-gray-300">
              <span className="rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100">
                Group #{addition.groupId}
              </span>
              <span className="rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200">
                {addition.status}
              </span>
              <span>Proposed by <a href={`/user/${addition.proposedBy}`} className="text-sky-300 hover:text-sky-200">@{addition.proposedBy}</a></span>
              <span>{addition.createdAt ? new Date(addition.createdAt).toLocaleString() : ''}</span>
            </div>
            <div className="grid gap-2">
              <a
                href={marketHref}
                className="text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200"
              >
                {group.questionTitle || addition.groupTitle || `Grouped market #${addition.groupId}`}
              </a>
              <div className="rounded-md border border-sky-900/70 bg-sky-950/20 p-4">
                <p className="text-xs font-semibold uppercase tracking-[0.14em] text-sky-200">Answer Option</p>
                <p className="mt-1 text-xl font-semibold text-white">{addition.answerLabel}</p>
                <p className="mt-2 text-sm text-sky-100/80">Configured add-answer cost: {addition.additionCost} credits</p>
              </div>
            </div>
            {addition.status === 'rejected' && addition.rejectionReason && (
              <div className="rounded-md border border-rose-800/70 bg-rose-950/30 p-3 text-sm text-rose-100">
                Rejection reason: {addition.rejectionReason}
              </div>
            )}
            {addition.status === 'approved' && (
              <div className="rounded-md border border-emerald-800/70 bg-emerald-950/30 p-3 text-sm text-emerald-100">
                Approved by @{addition.reviewedBy || 'admin'}{addition.reviewedAt ? ` at ${new Date(addition.reviewedAt).toLocaleString()}` : ''}.
              </div>
            )}
            {canReview && (
              <div className="grid gap-3 md:grid-cols-[minmax(0,1fr),auto,auto] md:items-start">
                <textarea
                  value={reason}
                  onChange={(event) => updateReason(addition.id, event.target.value)}
                  rows={3}
                  placeholder="Decision reason required for rejection"
                  className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                />
                <button
                  type="button"
                  disabled={busyAdditionId === addition.id}
                  onClick={() => reviewAddition(addition, 'approved')}
                  className="rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Approve Answer
                </button>
                <button
                  type="button"
                  disabled={busyAdditionId === addition.id || !reason.trim()}
                  onClick={() => reviewAddition(addition, 'rejected')}
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

const MarketGroupAnswerAdditionQueue = () => {
  const tabsData = answerAdditionReviewTabs.map((tab) => ({
    label: tab.label,
    content: <MarketGroupAnswerAdditionStatusQueue status={tab.status} />,
  }));

  return (
    <div className="grid gap-4">
      <div className="rounded-lg border border-sky-800/70 bg-sky-950/30 p-4 text-sm text-sky-100">
        Added answers create new binary child markets without changing existing child market history.
      </div>
      <GovernanceAutoApprovalSetting
        settingKey="autoApproveMarketGroupAnswers"
        title="Auto-approve added answers"
        description="When enabled, active moderator answer additions are immediately published as new child markets."
        savedMessage="Grouped answer auto-approval setting saved."
      />
      <SiteTabs tabs={tabsData} defaultTab="Pending Answers" />
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

  const updateStewardForm = (marketKey, updates) => {
    setStewardForms((current) => ({
      ...current,
      [marketKey]: {
        ...(current[marketKey] || {}),
        ...updates,
      },
    }));
  };

  const stewardFormFor = (market) => {
    const key = reviewRowKey(market);
    const currentSteward = market.stewardUsername || market.creatorUsername || '';
    return {
      stewardUsername: currentSteward,
      reason: '',
      ...(stewardForms[key] || {}),
    };
  };

  const updateMarketInQueues = (updatedMarket, originalMarket) => {
    if (originalMarket?.isMarketGroup) {
      const groupID = originalMarket.marketGroup?.id || updatedMarket.id;
      setMarkets((current) => current.map((market) => (
        market.marketGroup?.id === groupID
          ? {
            ...market,
            stewardUsername: updatedMarket.stewardUsername,
            marketGroup: {
              ...(market.marketGroup || {}),
              stewardUsername: updatedMarket.stewardUsername,
            },
          }
          : market
      )));
      return;
    }

    setMarkets((current) => current.map((market) => (
      market.id === updatedMarket.id ? { ...market, ...updatedMarket } : market
    )));
  };

  const reassignSteward = async (market) => {
    const key = reviewRowKey(market);
    const form = stewardFormFor(market);
    setBusyMarketId(key);
    setError('');
    setSuccessMessage('');
    try {
      const updatedMarket = market.isMarketGroup
        ? await reassignMarketGroupSteward({
          groupId: market.marketGroup.id,
          token,
          stewardUsername: form.stewardUsername,
          reason: form.reason,
        })
        : await reassignMarketSteward({
          marketId: market.id,
          token,
          stewardUsername: form.stewardUsername,
          reason: form.reason,
        });
      updateMarketInQueues(updatedMarket, market);
      setStewardForms((current) => ({
        ...current,
        [key]: {
          stewardUsername: updatedMarket.stewardUsername || form.stewardUsername,
          reason: '',
        },
      }));
      setSuccessMessage(`${market.isMarketGroup ? 'Market group' : 'Market'} ${updatedMarket.id} steward reassigned to ${updatedMarket.stewardUsername}.`);
    } catch (err) {
      setError(err.message || 'Unable to reassign market steward.');
    } finally {
      setBusyMarketId(null);
    }
  };

  const renderStewardshipActions = (market) => {
    const key = reviewRowKey(market);
    const form = stewardFormFor(market);
    const currentSteward = market.stewardUsername || market.creatorUsername || '';
    const selectedSteward = String(form.stewardUsername || '').trim();
    const reason = String(form.reason || '').trim();
    const canSubmit = selectedSteward && reason && selectedSteward !== currentSteward;
    const stewardListId = `steward-options-${key}`;

    return (
      <div className="grid min-w-[260px] gap-3">
        <label className="grid gap-1 text-xs text-gray-300">
          <span className="font-mono uppercase tracking-[0.14em] text-gray-400">New steward</span>
          <input
            list={stewardListId}
            value={form.stewardUsername}
            onChange={(event) => updateStewardForm(key, { stewardUsername: event.target.value })}
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
          onChange={(event) => updateStewardForm(key, { reason: event.target.value })}
          rows={3}
          placeholder="Reason for stewardship reassignment"
          className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
        />
        <button
          type="button"
          disabled={busyMarketId === key || !canSubmit}
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
        markets={groupAdminMarketRows(markets)}
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
      label: 'Answer Additions',
      content: <MarketGroupAnswerAdditionQueue />,
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
