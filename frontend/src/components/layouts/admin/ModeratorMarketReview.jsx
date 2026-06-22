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
  updateMarketGroupTags,
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
  reviewGroupedMarketDescriptionAmendments,
  reviewMarketDescriptionAmendment,
  updateMarketGovernanceSettings,
} from '../../../api/marketDescriptionAmendmentsApi';
import {
  emptyPendingAdminReviewCounts,
  getPendingAdminReviewCounts,
} from '../../../api/adminReviewCountsApi';

const reviewTabs = [
  { label: 'Pending Review', status: 'proposed' },
  { label: 'Published', status: 'published' },
  { label: 'Rejected', status: 'rejected' },
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

const marketGroupAnswerPolicyOptions = [
  {
    value: 'auto',
    title: 'Auto-approve all answer options',
    description: 'Every active moderator answer option is immediately added to its grouped market.',
  },
  {
    value: 'moderator',
    title: 'Let the steward choose per market',
    description: 'Default. The grouped market steward controls auto-approval with the toggle on that market page.',
  },
  {
    value: 'admin',
    title: 'Only admins approve answer options',
    description: 'Answer options from other moderators stay pending for admin review, regardless of the steward toggle.',
  },
];

const maxMarketTagsPerMarket = 5;
const reviewPageSize = 20;

const reviewPaginationButtonClass = [
  'rounded-md border border-primary-pink px-3 py-2 text-xs font-semibold text-white transition',
  'hover:bg-primary-pink/20 focus:outline-none focus:ring-2 focus:ring-primary-pink/50',
  'disabled:cursor-not-allowed disabled:border-gray-700 disabled:text-gray-500 disabled:hover:bg-transparent',
].join(' ');

const adminReviewRetryDelayMs = 1200;

const isRateLimitError = (err) => err?.status === 429 || err?.reason === 'RATE_LIMITED';

const wait = (ms) => new Promise((resolve) => {
  window.setTimeout(resolve, ms);
});

const withAdminReviewRateLimitRetry = async (request, retries = 1) => {
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    try {
      return await request();
    } catch (err) {
      if (!isRateLimitError(err) || attempt === retries) {
        throw err;
      }
      await wait(adminReviewRetryDelayMs);
    }
  }
  return null;
};

const formatBadgeCount = (count) => {
  const numericCount = Number(count || 0);
  if (numericCount <= 0) return '';
  return numericCount > 99 ? '99+' : String(numericCount);
};

const ReviewSearchBox = ({
  id,
  label,
  value,
  onChange,
  placeholder,
  loading = false,
}) => (
  <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
    <label htmlFor={id} className="text-xs font-mono uppercase tracking-[0.16em] text-gray-400">
      {label}
    </label>
    <div className="relative">
      <input
        id={id}
        type="search"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 pr-10 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
      />
      {loading && (
        <div className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin rounded-full border-b-2 border-primary-pink" />
      )}
    </div>
  </div>
);

const ReviewPaginationControls = ({
  page,
  visibleCount,
  totalCount,
  hasPrevious,
  hasNext,
  onPrevious,
  onNext,
  loading = false,
}) => {
  const start = totalCount > 0 ? page * reviewPageSize + 1 : 0;
  const end = totalCount > 0 ? page * reviewPageSize + visibleCount : 0;
  return (
    <div className="flex flex-col gap-3 rounded-lg border border-gray-700 bg-gray-900/80 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="text-xs uppercase tracking-[0.18em] text-gray-400">
        Showing page {page + 1}
        <span className="ml-2 normal-case tracking-normal text-gray-300">
          {totalCount > 0 ? `(${start}-${end} of ${totalCount})` : '(0 results)'}
        </span>
      </div>
      <div className="flex gap-2">
        <button
          type="button"
          disabled={loading || !hasPrevious}
          onClick={onPrevious}
          className={reviewPaginationButtonClass}
        >
          Previous
        </button>
        <button
          type="button"
          disabled={loading || !hasNext}
          onClick={onNext}
          className={reviewPaginationButtonClass}
        >
          Next
        </button>
      </div>
    </div>
  );
};

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

const normalizeAnswerAdditionApprovalPolicy = (value, legacyAutoApprove = false) => {
  const normalized = String(value || '').trim().toLowerCase();
  if (normalized === 'auto' || normalized === 'admin' || normalized === 'moderator') {
    return normalized;
  }
  return legacyAutoApprove ? 'auto' : 'moderator';
};

const reviewRowKey = (market) => (
  market?.rowKey || (market?.isMarketGroup ? `group:${market.marketGroup.id}` : `market:${market.id}`)
);

const reviewMarketTagTargetIds = (market) => {
  if (market?.isMarketGroup) {
    return Array.from(new Set(
      (market.childMarkets || [])
        .map((child) => child.id)
        .filter(Boolean),
    ));
  }
  return market?.id ? [market.id] : [];
};

const amendmentRowKey = (amendment) => (
  amendment?.rowKey || (amendment?.isMarketGroupAmendment
    ? `group:${amendment.marketGroup.id}:${amendment.body}:${amendment.createdBy}:${amendment.submitReason || ''}:${amendment.status}`
    : `amendment:${amendment.id}`)
);

const useReviewWorkCounts = () => {
  const { token } = useAuth();
  const [counts, setCounts] = useState(emptyPendingAdminReviewCounts);

  useEffect(() => {
    if (!token) {
      setCounts(emptyPendingAdminReviewCounts);
      return undefined;
    }

    let ignore = false;
    const loadCounts = async () => {
      try {
        const nextCounts = await getPendingAdminReviewCounts({ token });
        if (ignore) return;
        setCounts(nextCounts);
        window.dispatchEvent(new CustomEvent('socialpredict:admin-review-counts', { detail: nextCounts }));
      } catch {
        if (!ignore) {
          setCounts(emptyPendingAdminReviewCounts);
          window.dispatchEvent(new CustomEvent('socialpredict:admin-review-counts', { detail: emptyPendingAdminReviewCounts }));
        }
      }
    };

    const initialTimeoutId = window.setTimeout(loadCounts, 1500);
    const intervalId = window.setInterval(loadCounts, 60000);
    return () => {
      ignore = true;
      window.clearTimeout(initialTimeoutId);
      window.clearInterval(intervalId);
    };
  }, [token]);

  return counts;
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
    marketGroupAnswerAdditionApprovalPolicy: 'moderator',
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
          setSettings({
            ...data,
            marketGroupAnswerAdditionApprovalPolicy: normalizeAnswerAdditionApprovalPolicy(
              data.marketGroupAnswerAdditionApprovalPolicy,
              data.autoApproveMarketGroupAnswers,
            ),
          });
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
        marketGroupAnswerAdditionApprovalPolicy: normalizeAnswerAdditionApprovalPolicy(
          settings.marketGroupAnswerAdditionApprovalPolicy,
          settings.autoApproveMarketGroupAnswers,
        ),
        [settingKey]: draft,
      };
      const saved = await updateMarketGovernanceSettings({
        token,
        autoApproveDescriptionAmendments: nextSettings.autoApproveDescriptionAmendments,
        autoApproveMarketProposals: nextSettings.autoApproveMarketProposals,
        autoApproveMarketGroupAnswers: nextSettings.autoApproveMarketGroupAnswers,
        marketGroupAnswerAdditionApprovalPolicy: nextSettings.marketGroupAnswerAdditionApprovalPolicy,
        version: settings.version,
      });
      setSettings({
        ...saved,
        marketGroupAnswerAdditionApprovalPolicy: normalizeAnswerAdditionApprovalPolicy(
          saved.marketGroupAnswerAdditionApprovalPolicy,
          saved.autoApproveMarketGroupAnswers,
        ),
      });
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

const AnswerAdditionApprovalPolicySetting = () => {
  const { token } = useAuth();
  const [settings, setSettings] = useState({
    autoApproveDescriptionAmendments: false,
    autoApproveMarketProposals: false,
    autoApproveMarketGroupAnswers: false,
    marketGroupAnswerAdditionApprovalPolicy: 'moderator',
    version: 0,
  });
  const [draft, setDraft] = useState('moderator');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) return undefined;
    let ignore = false;
    const loadSettings = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await getMarketGovernanceSettings({ token });
        if (!ignore) {
          const policy = normalizeAnswerAdditionApprovalPolicy(
            data.marketGroupAnswerAdditionApprovalPolicy,
            data.autoApproveMarketGroupAnswers,
          );
          setSettings({
            ...data,
            marketGroupAnswerAdditionApprovalPolicy: policy,
          });
          setDraft(policy);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load answer option approval settings.');
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
  }, [token]);

  const saveSettings = async () => {
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const saved = await updateMarketGovernanceSettings({
        token,
        autoApproveDescriptionAmendments: Boolean(settings.autoApproveDescriptionAmendments),
        autoApproveMarketProposals: Boolean(settings.autoApproveMarketProposals),
        autoApproveMarketGroupAnswers: draft === 'auto',
        marketGroupAnswerAdditionApprovalPolicy: draft,
        version: settings.version,
      });
      const policy = normalizeAnswerAdditionApprovalPolicy(
        saved.marketGroupAnswerAdditionApprovalPolicy,
        saved.autoApproveMarketGroupAnswers,
      );
      setSettings({
        ...saved,
        marketGroupAnswerAdditionApprovalPolicy: policy,
      });
      setDraft(policy);
      setMessage('Answer option approval policy saved.');
    } catch (err) {
      setError(err.message || 'Unable to save answer option approval policy.');
    } finally {
      setSaving(false);
    }
  };

  const changed = draft !== normalizeAnswerAdditionApprovalPolicy(
    settings.marketGroupAnswerAdditionApprovalPolicy,
    settings.autoApproveMarketGroupAnswers,
  );

  return (
    <div className="grid gap-3 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
      <div>
        <h3 className="text-sm font-semibold uppercase tracking-wide text-white">Answer Option Approval Policy</h3>
        <p className="mt-1 text-sm text-gray-400">
          Controls what happens when active moderators add answer options to grouped markets.
        </p>
      </div>
      <div className="grid gap-2">
        {marketGroupAnswerPolicyOptions.map((option) => (
          <label
            key={option.value}
            className={`flex cursor-pointer gap-3 rounded-md border p-3 text-sm transition ${
              draft === option.value
                ? 'border-sky-500 bg-sky-950/50 text-white'
                : 'border-gray-700 bg-gray-950/50 text-gray-300 hover:border-gray-500'
            }`}
          >
            <input
              type="radio"
              name="marketGroupAnswerAdditionApprovalPolicy"
              value={option.value}
              checked={draft === option.value}
              disabled={loading || saving}
              onChange={(event) => setDraft(event.target.value)}
              className="mt-1 h-4 w-4 border-gray-600 bg-gray-800 text-primary-pink focus:ring-primary-pink"
            />
            <span>
              <span className="block font-semibold">{option.title}</span>
              <span className="mt-1 block text-xs text-gray-400">{option.description}</span>
            </span>
          </label>
        ))}
      </div>
      <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <span className="text-xs text-gray-500">Version {settings.version || 1}</span>
        <button
          type="button"
          disabled={loading || saving || !changed}
          onClick={saveSettings}
          className="rounded-md bg-sky-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {saving ? 'Saving...' : 'Save Policy'}
        </button>
      </div>
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
  const [page, setPage] = useState(0);
  const [totalMarkets, setTotalMarkets] = useState(0);

  const tagEditingEnabled = status === 'proposed' || status === 'published';

  const loadMarkets = async ({ query = searchQuery, pageNumber = page } = {}) => {
    setLoading(true);
    setError('');
    try {
      const data = await withAdminReviewRateLimitRetry(() => listAdminLifecycleMarkets({
        token,
        status,
        query,
        limit: reviewPageSize,
        offset: pageNumber * reviewPageSize,
      }));
      setMarkets(data.markets || []);
      setTotalMarkets(Number(data.total || 0));
    } catch (err) {
      setError(err.message || 'Unable to load market queue.');
      setMarkets([]);
      setTotalMarkets(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setPage(0);
  }, [status, searchQuery]);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      loadMarkets({ query: searchQuery, pageNumber: page });
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, token, searchQuery, page]);

  useEffect(() => {
    if (!token || !tagEditingEnabled) {
      return;
    }

    let ignore = false;
    const loadActiveTags = async () => {
      try {
        const result = await withAdminReviewRateLimitRetry(() => listAdminMarketTags({ token, includeInactive: false }));
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
      await loadMarkets({ query: searchQuery, pageNumber: page });
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
      await loadMarkets({ query: searchQuery, pageNumber: page });
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

  const tagSlugsForForm = (market) => tagForms[reviewRowKey(market)] || marketTagSlugs(market);

  const toggleMarketTag = (market, slug) => {
    const key = reviewRowKey(market);
    setTagForms((current) => {
      const selected = current[key] || marketTagSlugs(market);
      const next = selected.includes(slug)
        ? selected.filter((item) => item !== slug)
        : [...selected, slug].sort();
      return { ...current, [key]: next };
    });
  };

  const saveMarketTags = async (market) => {
    const key = reviewRowKey(market);
    const targetIds = reviewMarketTagTargetIds(market);
    const tagSlugs = tagSlugsForForm(market);
    setBusyMarketId(key);
    setError('');
    setSuccessMessage('');
    try {
      if (!market.isMarketGroup && !targetIds.length) {
        throw new Error('No market IDs were found for tag assignment.');
      }
      if (market.isMarketGroup) {
        const groupId = market.marketGroup?.id || market.id;
        if (!groupId) {
          throw new Error('No market group ID was found for tag assignment.');
        }
        const updatedGroup = await updateMarketGroupTags({ groupId, token, tagSlugs });
        setMarkets((current) => current.map((row) => (
          reviewRowKey(row) === key
            ? { ...row, ...updatedGroup }
            : row
        )));
      } else {
        const updatedMarket = await updateMarketTags({ marketId: market.id, token, tagSlugs });
        updateMarketInQueue(updatedMarket);
      }
      setTagForms((current) => {
        const next = { ...current };
        delete next[key];
        return next;
      });
      setSuccessMessage(
        market.isMarketGroup
          ? `Updated tags for grouped market ${market.marketGroup?.id || market.id}.`
          : `Updated tags for market ${market.id}.`,
      );
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
    const key = reviewRowKey(market);
    const targetCount = reviewMarketTagTargetIds(market).length;

    return (
      <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/80 p-3">
        <div>
          <div className="font-mono text-xs uppercase tracking-[0.14em] text-gray-400">
            Admin tag adjustment
          </div>
          <p className="mt-1 text-xs text-gray-500">
            {market.isMarketGroup
              ? `Add or remove active tags across all ${targetCount || 'grouped'} answer markets.`
              : 'Add or remove active tags before or after publication.'}
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
                  disabled={disabled || busyMarketId === key}
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
          disabled={busyMarketId === key || !changed}
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
        {renderTagEditor(market)}
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
      <ReviewSearchBox
        id={`market-review-search-${status}`}
        label="Search markets"
        value={searchQuery}
        onChange={setSearchQuery}
        placeholder={`Search ${status} markets by title or description`}
        loading={loading}
      />
      <ReviewPaginationControls
        page={page}
        visibleCount={markets.length}
        totalCount={totalMarkets}
        hasPrevious={page > 0}
        hasNext={(page + 1) * reviewPageSize < totalMarkets}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
        loading={loading}
      />
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
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);
  const [totalAmendments, setTotalAmendments] = useState(0);
  const canReview = status === 'pending';

  const loadAmendments = async ({ query = searchQuery, pageNumber = page } = {}) => {
    setLoading(true);
    setError('');
    try {
      const data = await withAdminReviewRateLimitRetry(() => listAdminMarketDescriptionAmendments({
        token,
        status,
        query,
        limit: reviewPageSize,
        offset: pageNumber * reviewPageSize,
      }));
      setAmendments(data.amendments || []);
      setTotalAmendments(Number(data.total || 0));
    } catch (err) {
      setError(err.message || 'Unable to load description amendments.');
      setAmendments([]);
      setTotalAmendments(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    const timeoutId = window.setTimeout(() => {
      loadAmendments({ query: searchQuery, pageNumber: page });
    }, 300);
    return () => {
      window.clearTimeout(timeoutId);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, status, searchQuery, page]);

  useEffect(() => {
    setPage(0);
  }, [status, searchQuery]);

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
      if (amendment.isMarketGroupAmendment) {
        await reviewGroupedMarketDescriptionAmendments({
          token,
          amendmentIds: childAmendments.map((child) => child.id),
          status: nextStatus,
          reason,
        });
      } else {
        await reviewMarketDescriptionAmendment({
          token,
          amendmentId: amendment.id,
          status: nextStatus,
          reason,
        });
      }
      setReason(key, '');
      setSuccessMessage(`${amendment.isMarketGroupAmendment ? 'Grouped amendment' : `Amendment v${amendment.version}`} ${nextStatus}.`);
      await loadAmendments({ query: searchQuery, pageNumber: page });
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
      <ReviewSearchBox
        id={`description-amendment-search-${status}`}
        label="Search amendments"
        value={searchQuery}
        onChange={setSearchQuery}
        placeholder="Search market title, description, amendment text, submitter, or reason"
        loading={loading}
      />
      <ReviewPaginationControls
        page={page}
        visibleCount={amendments.length}
        totalCount={totalAmendments}
        hasPrevious={page > 0}
        hasNext={(page + 1) * reviewPageSize < totalAmendments}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
        loading={loading}
      />
      {totalAmendments === 0 && (
        <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300">
          No {status} description amendments found.
        </div>
      )}
      {amendments.map((amendment) => {
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
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);
  const [totalAdditions, setTotalAdditions] = useState(0);
  const canReview = status === 'pending';

  const loadAdditions = async ({ query = searchQuery, pageNumber = page } = {}) => {
    setLoading(true);
    setError('');
    try {
      const data = await withAdminReviewRateLimitRetry(() => listAdminMarketGroupAnswerAdditions({
        token,
        status,
        query,
        limit: reviewPageSize,
        offset: pageNumber * reviewPageSize,
      }));
      setAdditions(data.additions || []);
      setTotalAdditions(Number(data.total || 0));
    } catch (err) {
      setError(err.message || 'Unable to load grouped answer additions.');
      setAdditions([]);
      setTotalAdditions(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    const timeoutId = window.setTimeout(() => {
      loadAdditions({ query: searchQuery, pageNumber: page });
    }, 300);
    return () => {
      window.clearTimeout(timeoutId);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, status, searchQuery, page]);

  useEffect(() => {
    setPage(0);
  }, [status, searchQuery]);

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
      await loadAdditions({ query: searchQuery, pageNumber: page });
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
      <ReviewSearchBox
        id={`answer-addition-search-${status}`}
        label="Search answer options"
        value={searchQuery}
        onChange={setSearchQuery}
        placeholder="Search group title, answer label, proposer, reviewer, or reason"
        loading={loading}
      />
      <ReviewPaginationControls
        page={page}
        visibleCount={additions.length}
        totalCount={totalAdditions}
        hasPrevious={page > 0}
        hasNext={(page + 1) * reviewPageSize < totalAdditions}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
        loading={loading}
      />
      {totalAdditions === 0 && (
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
      <AnswerAdditionApprovalPolicySetting />
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
  const [page, setPage] = useState(0);
  const [totalMarkets, setTotalMarkets] = useState(0);

  useEffect(() => {
    setPage(0);
  }, [searchQuery]);

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
          limit: reviewPageSize,
          offset: page * reviewPageSize,
        });

        if (!ignore) {
          setMarkets(result.markets || []);
          setTotalMarkets(Number(result.total || 0));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load stewardship markets.');
          setMarkets([]);
          setTotalMarkets(0);
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
  }, [token, searchQuery, page]);

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
      <div className="grid gap-2">
        <ReviewSearchBox
          id="stewardship-market-search"
          label="Search stewardship markets"
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search title or description across proposed, published, closed, and resolved markets"
          loading={loading}
        />
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
      <ReviewPaginationControls
        page={page}
        visibleCount={markets.length}
        totalCount={totalMarkets}
        hasPrevious={page > 0}
        hasNext={(page + 1) * reviewPageSize < totalMarkets}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
        loading={loading}
      />
      <MarketLifecycleTable
        markets={markets}
        emptyMessage="No markets found for stewardship governance."
        showCreator
        showSteward
        actions={renderStewardshipActions}
      />
      <ReviewPaginationControls
        page={page}
        visibleCount={markets.length}
        totalCount={totalMarkets}
        hasPrevious={page > 0}
        hasNext={(page + 1) * reviewPageSize < totalMarkets}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
        loading={loading}
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

const ReviewShortcutCard = ({ eyebrow, title, description, count, active, onClick }) => (
  <button
    type="button"
    onClick={onClick}
    className={`grid gap-2 rounded-xl border p-4 text-left transition ${
      active
        ? 'border-primary-pink bg-primary-pink/15 shadow-lg shadow-primary-pink/10'
        : 'border-gray-700 bg-gray-900/70 hover:border-sky-500/70 hover:bg-gray-900'
    }`}
  >
    <div className="flex items-start justify-between gap-3">
      <span className="text-xs font-semibold uppercase tracking-[0.18em] text-gray-400">{eyebrow}</span>
      {count !== null && count !== undefined ? (
        <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${
          count > 0 ? 'bg-primary-pink text-white' : 'bg-gray-800 text-gray-400'
        }`}
        >
          {count > 99 ? '99+' : count}
        </span>
      ) : null}
    </div>
    <div className="text-lg font-semibold text-white">{title}</div>
    <p className="text-sm leading-5 text-gray-400">{description}</p>
  </button>
);

const ReviewQueueShortcuts = ({ activeTab, counts, onSelect }) => {
  const cards = [
    {
      tab: 'Pending Review',
      eyebrow: 'Market Queue',
      title: 'Pending Markets',
      description: 'Approve, reject, and tag proposed binary or grouped markets.',
      count: counts.pendingMarkets,
    },
    {
      tab: 'Description Amendments',
      eyebrow: 'Contract Changes',
      title: 'Pending Amendments',
      description: 'Review append-only market description changes before publication.',
      count: counts.pendingAmendments,
    },
    {
      tab: 'Answer Additions',
      eyebrow: 'Grouped Markets',
      title: 'Pending Answer Options',
      description: 'Review added answer options for multiple-choice binary markets.',
      count: counts.pendingAnswers,
    },
    {
      tab: 'Stewardship',
      eyebrow: 'Operations',
      title: 'Stewardship',
      description: 'Reassign operational responsibility when markets need a new steward.',
      count: null,
    },
  ];

  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      {cards.map((card) => (
        <ReviewShortcutCard
          key={card.tab}
          {...card}
          active={activeTab === card.tab}
          onClick={() => onSelect(card.tab)}
        />
      ))}
    </div>
  );
};

function ModeratorMarketReview({ defaultTab = 'Pending Review' }) {
  const counts = useReviewWorkCounts();
  const [activeTab, setActiveTab] = useState(defaultTab);

  useEffect(() => {
    setActiveTab(defaultTab || 'Pending Review');
  }, [defaultTab]);

  const badgeForTab = (label) => {
    if (label === 'Pending Review') return formatBadgeCount(counts.pendingMarkets);
    if (label === 'Description Amendments') return formatBadgeCount(counts.pendingAmendments);
    if (label === 'Answer Additions') return formatBadgeCount(counts.pendingAnswers);
    return '';
  };

  const tabsData = [
    ...reviewTabs.map((tab) => ({
      label: tab.label,
      badge: badgeForTab(tab.label),
      content: <AdminMarketQueue status={tab.status} />,
    })),
    {
      label: 'Stewardship',
      content: <MarketStewardshipQueue />,
    },
    {
      label: 'Description Amendments',
      badge: badgeForTab('Description Amendments'),
      content: <DescriptionAmendmentQueue />,
    },
    {
      label: 'Answer Additions',
      badge: badgeForTab('Answer Additions'),
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
        <h1 className="text-2xl font-bold mt-2">Review Markets</h1>
        <p className="mt-2 max-w-3xl text-sm text-gray-300">
          Review markets, amendments, answer options, stewardship, and tags from one operational queue.
        </p>
      </div>

      <ReviewQueueShortcuts
        activeTab={activeTab}
        counts={counts}
        onSelect={setActiveTab}
      />

      <div className="mt-6">
        <SiteTabs
          tabs={tabsData}
          activeTab={activeTab}
          onTabChange={setActiveTab}
        />
      </div>
    </section>
  );
}

export default ModeratorMarketReview;
