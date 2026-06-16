import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import LoadingSpinner from '../loaders/LoadingSpinner';
import GroupedMarketChart from '../charts/GroupedMarketChart';
import MarketTagChips from '../markets/MarketTagChips';
import SiteTabs from '../tabs/SiteTabs';
import TradeTabs from '../tabs/TradeTabs';
import TradeCTA from '../TradeCTA';
import MarkdownLite from '../markdown/MarkdownLite';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import StewardTag, { stewardUsernameFor } from '../markets/StewardTag';
import MarketGroupAnswerAdditionReviewQueue from '../marketGroups/MarketGroupAnswerAdditionReviewQueue';
import { resolveMarketGroup } from '../modals/resolution/ResolveUtils';
import { summarizeGroupedResolutions } from '../../helpers/marketGroups';
import {
  getMarketGroupDetails,
  proposeMarketGroupAnswerAddition,
  updateMarketGroupAnswerAdditionSettings,
} from '../../api/marketsApi';
import { proposeMarketDescriptionAmendment } from '../../api/marketDescriptionAmendmentsApi';
import { apiRequest } from '../../api/httpClient';
import useFrontendConfig from '../../hooks/useFrontendConfig';

const DEFAULT_CREATOR_EMOJI = '👤';

const toNumber = (value, fallback = 0) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const uniqueTagsBySlug = (answers = []) => {
  const seen = new Set();
  const tags = [];
  answers.forEach((answer) => {
    (answer?.market?.market?.tags || []).forEach((tag) => {
      const key = tag?.slug || tag?.id || tag?.displayName;
      if (!key || seen.has(key)) {
        return;
      }
      seen.add(key);
      tags.push(tag);
    });
  });
  return tags;
};

const probabilityDisplay = (answer) => {
  const probability = toNumber(
    answer?.probabilityChanges?.[answer.probabilityChanges.length - 1]?.probability
      ?? answer?.summary?.lastProbability
      ?? answer?.market?.lastProbability
      ?? answer?.market?.market?.initialProbability,
    0.5,
  );
  return probability.toFixed(2);
};

const groupedCloseStatusValue = (group, answers, closeDate) => {
  const lifecycle = String(group?.lifecycleStatus || '').toLowerCase();
  if (lifecycle === 'resolved' || answers.some((answer) => answer?.market?.market?.isResolved)) {
    return 'Resolved';
  }

  const close = closeDate ? new Date(closeDate) : null;
  if ((lifecycle === 'closed') || (close && !Number.isNaN(close.getTime()) && close <= new Date())) {
    return 'Closed';
  }

  return formatResolutionDate(closeDate);
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

const rateLimitRetryDelayMs = 1200;
const wait = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));
const isRateLimitError = (err) => err?.status === 429 || err?.reason === 'RATE_LIMITED';

const withRateLimitRetry = async (request) => {
  try {
    return await request();
  } catch (err) {
    if (!isRateLimitError(err)) {
      throw err;
    }
    await wait(rateLimitRetryDelayMs);
    return request();
  }
};

const groupedFetchJson = async (path, token = '') => {
  return withRateLimitRetry(() => apiRequest(path, {
    authenticated: Boolean(token),
    authToken: token,
    reasonMessages: {
      RATE_LIMITED: 'Grouped activity is loading. Wait a moment and try again.',
    },
    fallbackMessage: 'Failed to load grouped market activity.',
  }));
};

const answerLabelFor = (answer) => answer?.answerLabel || `Answer ${Number(answer?.displayOrder || 0) + 1}`;

const childDescriptionAmendments = (answer) => {
  const amendments = answer?.descriptionAmendments || answer?.DescriptionAmendments || [];
  return Array.isArray(amendments) ? amendments : [];
};

const uniqueGroupedAmendments = (answers = []) => {
  const grouped = new Map();
  const amendments = [];
  answers.forEach((answer) => {
    (answer.descriptionAmendments || []).forEach((amendment) => {
      const approvedAt = amendment.approvedAt || amendment.ApprovedAt || '';
      const approvedAtDate = approvedAt ? new Date(approvedAt) : null;
      const approvedAtKey = approvedAtDate && !Number.isNaN(approvedAtDate.getTime())
        ? String(Math.floor(approvedAtDate.getTime() / 1000))
        : String(approvedAt || '');
      const key = [
        String(amendment.body || amendment.Body || '').trim(),
        String(amendment.createdBy || amendment.CreatedBy || '').trim(),
        approvedAtKey,
      ].join('|');
      if (!key.trim()) {
        return;
      }
      const answerLabel = answerLabelFor(answer);
      const existing = grouped.get(key);
      if (existing) {
        if (!existing.answerLabels.includes(answerLabel)) {
          existing.answerLabels.push(answerLabel);
        }
        return;
      }
      const row = {
        ...amendment,
        groupKey: key,
        answerLabels: [answerLabel],
      };
      grouped.set(key, row);
      amendments.push(row);
    });
  });
  return amendments;
};

const canTradeMarket = (market, isLoggedIn) => {
  if (!isLoggedIn || market?.isResolved) {
    return false;
  }
  const lifecycle = String(market?.lifecycleStatus || '').toLowerCase();
  const status = String(market?.status || '').toLowerCase();
  if (lifecycle && lifecycle !== 'published') {
    return false;
  }
  if (status && status !== 'active') {
    return false;
  }
  const closeDate = market?.resolutionDateTime ? new Date(market.resolutionDateTime) : null;
  if (!closeDate || Number.isNaN(closeDate.getTime())) {
    return false;
  }
  return closeDate > new Date();
};

const GroupedBetsActivity = ({ groupId, refreshTrigger }) => {
  const pageSize = 20;
  const [bets, setBets] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setPage((current) => (current === 0 ? current : 0));
  }, [groupId, refreshTrigger]);

  useEffect(() => {
    let ignore = false;
    const loadBets = async () => {
      if (refreshTrigger && page !== 0) {
        return;
      }
      const offset = page * pageSize;
      setError('');
      try {
        const payload = await groupedFetchJson(`/v0/market-groups/${groupId}/bets?limit=${pageSize + 1}&offset=${offset}`);
        const rows = Array.isArray(payload?.bets) ? payload.bets : [];
        if (!ignore) {
          setBets(rows.slice(0, pageSize));
          setHasNextPage(toNumber(payload?.total) > offset + pageSize || rows.length > pageSize);
        }
      } catch (err) {
        if (!ignore) {
          setBets([]);
          setHasNextPage(false);
          setError(err.message || 'Failed to load grouped bets.');
        }
      }
    };
    if (groupId) {
      loadBets();
    }
    return () => {
      ignore = true;
    };
  }, [groupId, page, refreshTrigger]);

  const pageStart = page * pageSize;

  return (
    <div className='p-4'>
      <GroupedActivityHeader
        label='bets'
        page={page}
        pageStart={pageStart}
        count={bets.length}
        hasNextPage={hasNextPage}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
      />
      {error && <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{error}</div>}
      <div className='grid grid-cols-[minmax(0,1fr),minmax(0,1fr),auto,auto] gap-2 rounded-t bg-gray-800 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gray-300 sm:grid-cols-[minmax(0,1fr),minmax(0,1fr),auto,auto,auto,auto]'>
        <div>User</div>
        <div>Answer</div>
        <div>Outcome</div>
        <div className='text-right'>Amount</div>
        <div className='hidden text-right sm:block'>After</div>
        <div className='hidden text-right sm:block'>Placed</div>
      </div>
      {bets.length === 0 && !error && (
        <div className='rounded-b bg-gray-800/60 p-4 text-center text-sm text-gray-400'>No bets yet</div>
      )}
      {bets.map((bet, index) => (
        <div key={`${bet.answerMarketId}-${bet.placedAt}-${index}`} className='grid grid-cols-[minmax(0,1fr),minmax(0,1fr),auto,auto] gap-2 border-t border-gray-800 bg-gray-900 px-3 py-3 text-sm sm:grid-cols-[minmax(0,1fr),minmax(0,1fr),auto,auto,auto,auto]'>
          <Link to={`/user/${bet.username}`} className='truncate text-blue-400 hover:text-blue-300'>
            {bet.username}
          </Link>
          <Link to={`/markets/${bet.answerMarketId}`} className='truncate text-gray-200 hover:text-primary-pink'>
            {bet.answerLabel}
          </Link>
          <span className={`justify-self-start rounded px-2 py-1 text-xs font-bold text-white ${bet.outcome === 'YES' ? 'bg-green-600' : 'bg-red-600'}`}>
            {bet.outcome}
          </span>
          <div className='text-right text-gray-300'>{bet.amount}</div>
          <div className='hidden text-right text-gray-300 sm:block'>{toNumber(bet.probability, 0).toFixed(2)}</div>
          <div className='col-span-4 text-right text-xs text-gray-500 sm:col-span-1'>
            {bet.placedAt ? new Date(bet.placedAt).toLocaleString() : ''}
          </div>
        </div>
      ))}
    </div>
  );
};

const GroupedPositionEntry = ({ entry }) => (
  <div className='flex flex-col rounded-lg bg-gray-800 p-3 shadow'>
    <Link
      to={`/user/${entry.username}`}
      className='font-bold text-blue-400 underline hover:text-blue-300'
    >
      {entry.username}
    </Link>
    <Link
      to={`/markets/${entry.answerMarketId}`}
      className='mt-1 text-xs font-semibold uppercase tracking-[0.12em] text-gray-400 hover:text-primary-pink'
    >
      {entry.answerLabel}
    </Link>
    <div className='mt-1 text-sm text-gray-300'>Shares: {entry.shares}</div>
    <div className='text-sm text-green-400'>Value: {entry.value}</div>
  </div>
);

const GroupedPositionsActivity = ({ groupId, token, refreshTrigger }) => {
  const pageSize = 20;
  const [positions, setPositions] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [freshnessLabel, setFreshnessLabel] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    setPage((current) => (current === 0 ? current : 0));
  }, [groupId, refreshTrigger, token]);

  useEffect(() => {
    let ignore = false;
    const loadPositions = async () => {
      if (refreshTrigger && page !== 0) {
        return;
      }
      if (!token) {
        setPositions([]);
        setHasNextPage(false);
        setFreshnessLabel('');
        setError('');
        return;
      }
      setError('');
      try {
        const offset = page * pageSize;
        const payload = await groupedFetchJson(`/v0/market-groups/${groupId}/positions?limit=${pageSize + 1}&offset=${offset}`, token);
        const rows = Array.isArray(payload?.positions) ? payload.positions : [];
        if (!ignore) {
          setPositions(rows.slice(0, pageSize));
          setHasNextPage(toNumber(payload?.total) > offset + pageSize || rows.length > pageSize);
          setFreshnessLabel(payload?.freshness?.generatedAt ? new Date(payload.freshness.generatedAt).toLocaleTimeString() : '');
        }
      } catch (err) {
        if (!ignore) {
          setPositions([]);
          setHasNextPage(false);
          setFreshnessLabel('');
          setError(err.message || 'Failed to load grouped positions.');
        }
      }
    };
    if (groupId) {
      loadPositions();
    }
    return () => {
      ignore = true;
    };
  }, [groupId, page, refreshTrigger, token]);

  const noPositionRows = positions.flatMap((position) => (
    position.answers
      .filter((answer) => toNumber(answer.noSharesOwned) > 0)
      .map((answer) => ({
        username: position.username,
        answerLabel: answer.answerLabel,
        answerMarketId: answer.answerMarketId,
        shares: toNumber(answer.noSharesOwned),
        value: toNumber(answer.value),
      }))
  ));
  const yesPositionRows = positions.flatMap((position) => (
    position.answers
      .filter((answer) => toNumber(answer.yesSharesOwned) > 0)
      .map((answer) => ({
        username: position.username,
        answerLabel: answer.answerLabel,
        answerMarketId: answer.answerMarketId,
        shares: toNumber(answer.yesSharesOwned),
        value: toNumber(answer.value),
      }))
  ));

  return (
    <div className='p-4'>
      <GroupedActivityHeader
        label='positions'
        page={page}
        pageStart={page * pageSize}
        count={positions.length}
        hasNextPage={hasNextPage}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
      />
      {freshnessLabel && (
        <div className='mb-3 text-xs text-gray-500'>
          Grouped positions combine 10-minute child-market display snapshots generated as early as {freshnessLabel}. Trade confirmations remain authoritative.
        </div>
      )}
      {!token && <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>Log in to see grouped positions.</div>}
      {error && <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{error}</div>}
      {token && positions.length === 0 && !error && (
        <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>No positions yet</div>
      )}
      {positions.length > 0 && (
        <div className='flex flex-row gap-4'>
          <div className='flex-1'>
            <h2 className='mb-2 text-center font-bold'>Shares for: <span className='text-red-500'>NO</span></h2>
            <div className='flex flex-col gap-2'>
              {noPositionRows.length === 0 && (
                <div className='rounded-lg bg-gray-800/60 p-3 text-center text-sm text-gray-400'>No NO positions</div>
              )}
              {noPositionRows.map((entry) => (
                <GroupedPositionEntry key={`${entry.username}-${entry.answerMarketId}-no`} entry={entry} />
              ))}
            </div>
          </div>
          <div className='flex-1'>
            <h2 className='mb-2 text-center font-bold'>Shares for: <span className='text-green-500'>YES</span></h2>
            <div className='flex flex-col gap-2'>
              {yesPositionRows.length === 0 && (
                <div className='rounded-lg bg-gray-800/60 p-3 text-center text-sm text-gray-400'>No YES positions</div>
              )}
              {yesPositionRows.map((entry) => (
                <GroupedPositionEntry key={`${entry.username}-${entry.answerMarketId}-yes`} entry={entry} />
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const GroupedLeaderboardActivity = ({ groupId, refreshTrigger }) => {
  const pageSize = 20;
  const [leaderboard, setLeaderboard] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [freshnessLabel, setFreshnessLabel] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    setPage((current) => (current === 0 ? current : 0));
  }, [groupId, refreshTrigger]);

  useEffect(() => {
    let ignore = false;
    const loadLeaderboard = async () => {
      if (refreshTrigger && page !== 0) {
        return;
      }
      setError('');
      try {
        const offset = page * pageSize;
        const payload = await groupedFetchJson(`/v0/market-groups/${groupId}/leaderboard?limit=${pageSize + 1}&offset=${offset}`);
        const rows = Array.isArray(payload?.leaderboard) ? payload.leaderboard : [];
        if (!ignore) {
          setLeaderboard(rows.slice(0, pageSize));
          setHasNextPage(toNumber(payload?.total) > offset + pageSize || rows.length > pageSize);
          setFreshnessLabel(payload?.freshness?.generatedAt ? new Date(payload.freshness.generatedAt).toLocaleTimeString() : '');
        }
      } catch (err) {
        if (!ignore) {
          setLeaderboard([]);
          setHasNextPage(false);
          setFreshnessLabel('');
          setError(err.message || 'Failed to load grouped leaderboard.');
        }
      }
    };
    if (groupId) {
      loadLeaderboard();
    }
    return () => {
      ignore = true;
    };
  }, [groupId, page, refreshTrigger]);

  const formatCurrency = (amount) => toNumber(amount).toLocaleString();
  const getProfitColor = (profit) => {
    const numericProfit = toNumber(profit);
    if (numericProfit > 0) return 'text-green-400';
    if (numericProfit < 0) return 'text-red-400';
    return 'text-gray-300';
  };
  const getPositionBadge = (position) => {
    const baseClasses = 'rounded px-2 py-1 text-xs font-bold';
    switch (position) {
      case 'YES':
        return `${baseClasses} bg-green-600 text-white`;
      case 'NO':
        return `${baseClasses} bg-red-600 text-white`;
      case 'MIXED':
        return `${baseClasses} bg-blue-600 text-white`;
      case 'NEUTRAL':
        return `${baseClasses} bg-yellow-600 text-white`;
      default:
        return `${baseClasses} bg-gray-600 text-white`;
    }
  };
  const getRankDisplay = (rank) => {
    if (rank === 1) return '🥇';
    if (rank === 2) return '🥈';
    if (rank === 3) return '🥉';
    return `#${rank}`;
  };
  const getGroupedPosition = (entry) => {
    const yesShares = toNumber(entry.yesSharesOwned);
    const noShares = toNumber(entry.noSharesOwned);
    if (yesShares > 0 && noShares > 0) return 'MIXED';
    if (yesShares > 0) return 'YES';
    if (noShares > 0) return 'NO';
    return 'NEUTRAL';
  };

  return (
    <div className='p-4'>
      <GroupedActivityHeader
        label='leaderboard'
        page={page}
        pageStart={page * pageSize}
        count={leaderboard.length}
        hasNextPage={hasNextPage}
        onPrevious={() => setPage((current) => Math.max(0, current - 1))}
        onNext={() => setPage((current) => current + 1)}
      />
      {freshnessLabel && (
        <div className='mb-3 text-xs text-gray-500'>
          Grouped leaderboard combines 10-minute child-market display snapshots generated as early as {freshnessLabel}. Trade confirmations remain authoritative.
        </div>
      )}
      {error && <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{error}</div>}
      {leaderboard.length === 0 && !error && (
        <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>No participants yet</div>
      )}
      {leaderboard.length > 0 && (
        <div className='sp-grid-leaderboard-header'>
          <div>Rank</div>
          <div>User</div>
          <div>Position</div>
          <div className='text-right'>Profit</div>
          <div className='text-right'>Current Value</div>
          <div className='text-right'>Total Spent</div>
          <div>Shares</div>
        </div>
      )}
      {leaderboard.map((entry) => (
        <div key={entry.username} className='sp-grid-leaderboard-row mt-2'>
          <div className='flex items-center justify-start'>
            <div className='mr-2 text-lg font-bold text-white'>
              {getRankDisplay(entry.rank)}
            </div>
            <div className='sp-cell-username sm:hidden'>
              <div className='sp-ellipsis text-xs font-medium'>
                <Link to={`/user/${entry.username}`} className='text-blue-500 transition-colors hover:text-blue-400'>
                  {entry.username}
                </Link>
              </div>
            </div>
          </div>

          <div className='sp-cell-username hidden sm:block'>
            <div className='sp-ellipsis font-medium'>
              <Link to={`/user/${entry.username}`} className='text-blue-500 transition-colors hover:text-blue-400'>
                {entry.username}
              </Link>
            </div>
          </div>

          <div className='hidden sm:block'>
            <span className={getPositionBadge(getGroupedPosition(entry))}>
              {getGroupedPosition(entry)}
            </span>
          </div>

          <div className='text-right'>
            <div className={`text-sm font-bold ${getProfitColor(entry.profit)}`}>
              {toNumber(entry.profit) >= 0 ? '+' : ''}{formatCurrency(entry.profit)}
            </div>
            <div className='sp-subline sm:hidden'>
              Pos {getGroupedPosition(entry)} · {entry.yesSharesOwned}Y {entry.noSharesOwned}N
            </div>
          </div>

          <div className='sp-cell-num hidden text-gray-300 sm:block'>
            {formatCurrency(entry.currentValue)}
          </div>

          <div className='sp-cell-num hidden text-gray-300 sm:block'>
            {formatCurrency(entry.totalSpent)}
          </div>

          <div className='hidden text-xs text-gray-300 sm:block'>
            <div>YES: {entry.yesSharesOwned}</div>
            <div>NO: {entry.noSharesOwned}</div>
          </div>

          <div className='col-span-2 flex flex-wrap gap-1 border-t border-slate-700/60 pt-2 sm:col-span-7'>
            {entry.answers.map((answer) => (
              <span key={`${entry.username}-${answer.answerMarketId}`} className='rounded-full border border-gray-700 bg-slate-900 px-2 py-0.5 text-[10px] text-gray-200 sm:text-xs'>
                {answer.answerLabel}: {toNumber(answer.profit) >= 0 ? '+' : ''}{formatCurrency(answer.profit)}
              </span>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
};

const GroupedActivityHeader = ({
  label,
  page,
  pageStart,
  count,
  hasNextPage,
  onPrevious,
  onNext,
}) => (
  <div className='mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
    <div className='text-xs uppercase tracking-[0.16em] text-gray-400'>
      Showing grouped {label} page {page + 1}{count ? ` (${pageStart + 1}-${pageStart + count})` : ''}
    </div>
    <div className='flex gap-2'>
      <button
        type='button'
        onClick={onPrevious}
        disabled={page <= 0}
        className={paginationButtonClass}
      >
        Previous
      </button>
      <button
        type='button'
        onClick={onNext}
        disabled={!hasNextPage}
        className={paginationButtonClass}
      >
        Next
      </button>
    </div>
  </div>
);

const GroupedActivityTabs = ({ groupId, token, refreshTrigger }) => {
  const tabsData = [
    { label: 'Bets', content: <GroupedBetsActivity groupId={groupId} refreshTrigger={refreshTrigger} /> },
    { label: 'Positions', content: <GroupedPositionsActivity groupId={groupId} token={token} refreshTrigger={refreshTrigger} /> },
    { label: 'Leaderboard', content: <GroupedLeaderboardActivity groupId={groupId} refreshTrigger={refreshTrigger} /> },
    { label: 'Comments', content: <div className='p-4 text-gray-400'>Comments Go here...</div> },
  ];

  return <SiteTabs tabs={tabsData} />;
};

export default function GroupedMarketDetailsLayout({
  marketGroup,
  fallbackMarket,
  creator,
  isLoggedIn,
  token,
  username,
  usertype,
  moderatorStatus,
  refetchData,
}) {
  const { frontendConfig } = useFrontendConfig();
  const [groupData, setGroupData] = useState(null);
  const [answers, setAnswers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showFullDescription, setShowFullDescription] = useState(false);
  const [showTradeModal, setShowTradeModal] = useState(false);
  const [selectedTradeMarketId, setSelectedTradeMarketId] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);
  const [amendmentBody, setAmendmentBody] = useState('');
  const [amendmentReason, setAmendmentReason] = useState('');
  const [amendmentMessage, setAmendmentMessage] = useState('');
  const [amendmentError, setAmendmentError] = useState('');
  const [submittingAmendment, setSubmittingAmendment] = useState(false);
  const [answerAdditionLabel, setAnswerAdditionLabel] = useState('');
  const [answerAdditionMessage, setAnswerAdditionMessage] = useState('');
  const [answerAdditionError, setAnswerAdditionError] = useState('');
  const [submittingAnswerAddition, setSubmittingAnswerAddition] = useState(false);
  const [autoApproveAnswerAdditions, setAutoApproveAnswerAdditions] = useState(false);
  const [answerPolicyMessage, setAnswerPolicyMessage] = useState('');
  const [answerPolicyError, setAnswerPolicyError] = useState('');
  const [savingAnswerPolicy, setSavingAnswerPolicy] = useState(false);
  const [showResolveModal, setShowResolveModal] = useState(false);
  const [resolveMode, setResolveMode] = useState('exclusive_yes');
  const [winningMarketId, setWinningMarketId] = useState(0);
  const [manualResolutions, setManualResolutions] = useState({});
  const [resolveError, setResolveError] = useState('');
  const [resolvingGroup, setResolvingGroup] = useState(false);
  const hasLoadedGroupRef = useRef(false);
  const currentGroupIdRef = useRef(null);

  useEffect(() => {
    let ignore = false;
    const currentGroupId = marketGroup?.id;

    if (currentGroupIdRef.current !== currentGroupId) {
      currentGroupIdRef.current = currentGroupId;
      hasLoadedGroupRef.current = false;
    }

    const loadGroup = async () => {
      const isInitialLoad = !hasLoadedGroupRef.current;
      if (isInitialLoad) {
        setLoading(true);
        setError('');
      }
      try {
        const data = await getMarketGroupDetails(marketGroup.id);
        const rawAnswers = [...(data?.answers || [])].sort((left, right) => (
          Number(left.displayOrder || 0) - Number(right.displayOrder || 0)
        ));
        if (!ignore) {
          setGroupData(data);
          setAnswers(rawAnswers.map((answer) => ({
            ...answer,
            descriptionAmendments: childDescriptionAmendments(answer),
          })));
          hasLoadedGroupRef.current = true;
          setError('');
        }
      } catch (err) {
        if (!ignore) {
          if (!hasLoadedGroupRef.current) {
            setError(err.message || 'Failed to load grouped market.');
          } else {
            console.warn('[grouped-market] background refresh failed', err);
          }
        }
      } finally {
        if (!ignore && !hasLoadedGroupRef.current) {
          setLoading(false);
        } else if (!ignore) {
          setLoading(false);
        }
      }
    };

    if (marketGroup?.id) {
      loadGroup();
    }

    return () => {
      ignore = true;
    };
  }, [marketGroup?.id, refreshKey]);

  useEffect(() => {
    if (!answers.length) {
      setWinningMarketId(0);
      setManualResolutions({});
      setSelectedTradeMarketId(0);
      return;
    }
    setWinningMarketId((current) => current || answers[0]?.marketId || 0);
    setSelectedTradeMarketId((current) => {
      const stillExists = answers.some((answer) => Number(answer.marketId) === Number(current));
      if (stillExists) {
        return current;
      }
      const firstTradable = answers.find((answer) => canTradeMarket(answer?.market?.market || {}, isLoggedIn));
      return firstTradable?.marketId || answers[0]?.marketId || 0;
    });
    setManualResolutions((current) => {
      const next = {};
      answers.forEach((answer) => {
        next[answer.marketId] = current[answer.marketId] || 'NO';
      });
      return next;
    });
  }, [answers]);

  const group = groupData?.group || marketGroup || {};
  const groupCreator = groupData?.creator || creator || {};
  const creatorUsername = group.creatorUsername || groupCreator.username || fallbackMarket.creatorUsername || 'unknown';
  const creatorEmoji = groupCreator.personalEmoji || DEFAULT_CREATOR_EMOJI;
  const tags = useMemo(() => uniqueTagsBySlug(answers), [answers]);
  const aggregate = useMemo(() => ({
    users: Math.max(0, ...answers.map((answer) => toNumber(answer?.summary?.numUsers ?? answer?.market?.numUsers))),
    volume: answers.reduce((sum, answer) => sum + toNumber(answer?.summary?.totalVolume ?? answer?.market?.totalVolume), 0),
    dust: answers.reduce((sum, answer) => sum + toNumber(answer?.summary?.marketDust ?? answer?.market?.marketDust), 0),
  }), [answers]);

  const closeDate = fallbackMarket?.resolutionDateTime || group?.resolutionDateTime;
  const sortedAnswers = useMemo(() => [...answers].sort((left, right) => (
    Number(left.displayOrder || 0) - Number(right.displayOrder || 0)
  )), [answers]);
  const closeStatusValue = useMemo(
    () => groupedCloseStatusValue(group, sortedAnswers, closeDate),
    [group, sortedAnswers, closeDate],
  );
  const descriptionAmendments = useMemo(() => uniqueGroupedAmendments(sortedAnswers), [sortedAnswers]);
  const groupedResolutionSummary = useMemo(() => summarizeGroupedResolutions(sortedAnswers.map((answer) => ({
    marketId: answer.marketId,
    answerLabel: answerLabelFor(answer),
    isResolved: Boolean(answer?.market?.market?.isResolved),
    resolutionResult: answer?.market?.market?.resolutionResult,
  }))), [sortedAnswers]);
  const anyTradableAnswer = sortedAnswers.some((answer) => canTradeMarket(answer?.market?.market || {}, isLoggedIn));
  const groupStewardUsername = stewardUsernameFor({
    stewardUsername: group.stewardUsername || fallbackMarket?.stewardUsername,
    creatorUsername,
  }, creatorUsername);
  const isCurrentGroupSteward = String(username || '').trim() === String(groupStewardUsername || '').trim();
  const isActiveModerator = usertype === 'MODERATOR' && moderatorStatus === 'active';
  const canProposeDescriptionAmendment =
    isLoggedIn &&
    token &&
    isCurrentGroupSteward &&
    sortedAnswers.length > 0 &&
    closeDate &&
    new Date(closeDate) > new Date() &&
    !['rejected', 'resolved', 'cancelled'].includes(String(group.lifecycleStatus || '').toLowerCase());
  const canProposeAnswerAddition =
    isLoggedIn &&
    token &&
    isActiveModerator &&
    sortedAnswers.length > 0 &&
    closeDate &&
    new Date(closeDate) > new Date() &&
    String(group.lifecycleStatus || '').toLowerCase() === 'published' &&
    sortedAnswers.every((answer) => !answer?.market?.market?.isResolved);
  const canManageAnswerAdditions = canProposeAnswerAddition && isCurrentGroupSteward;
  const canResolveGroup =
    isLoggedIn &&
    token &&
    sortedAnswers.length > 0 &&
    isCurrentGroupSteward &&
    String(group.lifecycleStatus || '').toLowerCase() === 'published' &&
    sortedAnswers.every((answer) => !answer?.market?.market?.isResolved);
  const tradeButtonLabel = (() => {
    if (!isLoggedIn) {
      return 'LOG IN TO TRADE';
    }
    if (anyTradableAnswer) {
      return 'TRADE';
    }
    const hasProposed = sortedAnswers.some((answer) => (
      String(answer?.market?.market?.lifecycleStatus || '').toLowerCase() === 'proposed'
    ));
    return hasProposed ? 'AWAITING APPROVAL' : 'TRADING CLOSED';
  })();
  const addAnswerCost = toNumber(
    frontendConfig?.marketGroups?.multipleChoiceBinary?.addAnswerCost,
    0,
  );

  useEffect(() => {
    setAutoApproveAnswerAdditions(Boolean(group.autoApproveAnswerAdditions));
  }, [group.id, group.autoApproveAnswerAdditions]);

  const handleTransactionSuccess = () => {
    setShowTradeModal(false);
    setRefreshKey((current) => current + 1);
  };

  const submitGroupResolution = async () => {
    setResolveError('');
    if (!canResolveGroup) {
      setResolveError('You are not allowed to resolve this grouped market.');
      return;
    }
    const payload = resolveMode === 'exclusive_yes'
      ? {
          mode: 'exclusive_yes',
          winningMarketId: Number(winningMarketId),
        }
      : {
          mode: 'manual',
          resolutions: sortedAnswers.map((answer) => ({
            marketId: answer.marketId,
            resolution: manualResolutions[answer.marketId] || 'NO',
          })),
        };
    setResolvingGroup(true);
    try {
      await resolveMarketGroup(group.id || marketGroup.id, token, payload);
      setShowResolveModal(false);
      setRefreshKey((current) => current + 1);
      if (refetchData) {
        refetchData();
      }
      alert('Grouped market resolved successfully.');
    } catch (err) {
      setResolveError(err.message || 'Unable to resolve grouped market.');
    } finally {
      setResolvingGroup(false);
    }
  };

  const submitDescriptionAmendment = async (event) => {
    event.preventDefault();
    setAmendmentMessage('');
    setAmendmentError('');
    setSubmittingAmendment(true);
    try {
      await Promise.all(sortedAnswers.map((answer) => (
        proposeMarketDescriptionAmendment({
          token,
          marketId: answer.marketId,
          body: amendmentBody,
          submitReason: amendmentReason,
        })
      )));
      setAmendmentBody('');
      setAmendmentReason('');
      setAmendmentMessage('Grouped description amendment submitted for admin review.');
    } catch (err) {
      setAmendmentError(err.message || 'Unable to submit grouped description amendment.');
    } finally {
      setSubmittingAmendment(false);
    }
  };

  const submitAnswerAddition = async (event) => {
    event.preventDefault();
    const label = answerAdditionLabel.trim();
    if (!label) {
      return;
    }
    setAnswerAdditionMessage('');
    setAnswerAdditionError('');
    setSubmittingAnswerAddition(true);
    try {
      const addition = await proposeMarketGroupAnswerAddition({
        groupId: group.id || marketGroup.id,
        token,
        answerLabel: label,
      });
      setAnswerAdditionLabel('');
      const status = String(addition?.status || '').toLowerCase();
      setAnswerAdditionMessage(
        status === 'approved'
          ? `Answer option "${addition?.answerLabel || label}" added and published.`
          : `Answer option "${addition?.answerLabel || label}" submitted for steward review.`,
      );
      if (status === 'approved' || canManageAnswerAdditions) {
        setRefreshKey((current) => current + 1);
        if (refetchData) {
          refetchData();
        }
      }
    } catch (err) {
      setAnswerAdditionError(err.message || 'Unable to propose answer option.');
    } finally {
      setSubmittingAnswerAddition(false);
    }
  };

  const toggleAnswerAdditionPolicy = async () => {
    const nextValue = !autoApproveAnswerAdditions;
    setAnswerPolicyMessage('');
    setAnswerPolicyError('');
    setAutoApproveAnswerAdditions(nextValue);
    setSavingAnswerPolicy(true);
    try {
      await updateMarketGroupAnswerAdditionSettings({
        groupId: group.id || marketGroup.id,
        token,
        autoApproveAnswerAdditions: nextValue,
      });
      setAnswerPolicyMessage(nextValue
        ? 'Incoming answer options will auto-approve.'
        : 'Incoming answer options now require your approval.');
      setRefreshKey((current) => current + 1);
    } catch (err) {
      setAutoApproveAnswerAdditions(!nextValue);
      setAnswerPolicyError(err.message || 'Unable to update answer option policy.');
    } finally {
      setSavingAnswerPolicy(false);
    }
  };

  const selectedTradeAnswer = sortedAnswers.find((answer) => (
    Number(answer.marketId) === Number(selectedTradeMarketId)
  )) || sortedAnswers[0];
  const selectedTradeMarket = selectedTradeAnswer?.market?.market || {};
  const selectedTradeIsTradable = canTradeMarket(selectedTradeMarket, isLoggedIn);

  if (loading) {
    return (
      <div className='rounded-lg bg-gray-900 p-6 text-gray-300'>
        <LoadingSpinner />
        Loading grouped market...
      </div>
    );
  }

  if (error) {
    return (
      <div className='rounded-lg border border-red-500 bg-red-950/50 p-4 text-red-100'>
        {error}
      </div>
    );
  }

  return (
    <div className='bg-gray-900 text-gray-300 p-4 rounded-lg shadow-lg w-full'>
      <section className='mb-4'>
        <h1 className='text-xl font-semibold text-white mb-2 break-words line-clamp-2'>
          {group.questionTitle || marketGroup.questionTitle || fallbackMarket.questionTitle}
        </h1>
        <div className='flex flex-wrap items-center gap-2 text-sm text-gray-400'>
          <Link
            to={`/user/${creatorUsername}`}
            className='hover:text-blue-400 transition-colors duration-200'
          >
            <span role='img' aria-label='Creator'>
              {creatorEmoji}
            </span>
            @{creatorUsername}
          </Link>
          <span className='text-gray-600' aria-hidden='true'>•</span>
          <span>{answers.length} answers</span>
          <StewardTag username={groupStewardUsername} creatorUsername={creatorUsername} />
        </div>
        <MarketTagChips tags={tags} className='mt-3' />
      </section>

      <div className='mb-4'>
        <GroupedMarketChart answers={answers} title='Probability Changes' />
      </div>

      {groupedResolutionSummary && (
        <section className='mb-4 rounded-lg border border-emerald-800/70 bg-emerald-950/25 p-4'>
          <p className='text-xs font-semibold uppercase tracking-[0.16em] text-emerald-200'>Resolution</p>
          <p className={`mt-2 text-lg font-semibold ${groupedResolutionSummary.className}`}>
            {groupedResolutionSummary.label}
          </p>
          <p className='mt-1 text-sm text-emerald-100/70'>
            This grouped market has resolved. Child answer markets retain their own YES/NO outcomes.
          </p>
        </section>
      )}

      {(group.description || descriptionAmendments.length > 0) && (
        <>
          <div className='mb-4'>
            <button
              type='button'
              onClick={() => setShowFullDescription(!showFullDescription)}
              className='w-full py-2 bg-gray-700 hover:bg-gray-600 transition-colors duration-200 rounded-lg text-center text-sm'
            >
              {showFullDescription ? 'Hide Contract Text' : 'Show Full Contract Text'}
            </button>
          </div>
          {showFullDescription && (
            <div className='mb-4 rounded-lg bg-gray-800 p-4 text-sm'>
              <div className='grid gap-4'>
                <section>
                  <h2 className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400'>
                    Description
                  </h2>
                  {group.description ? (
                    <p className='whitespace-pre-wrap'>{group.description}</p>
                  ) : (
                    <p className='text-gray-500 italic'>No description provided.</p>
                  )}
                </section>
                {descriptionAmendments.length > 0 && (
                  <section className='grid gap-3'>
                    <h2 className='text-sm font-semibold uppercase tracking-[0.14em] text-sky-200'>Amendments</h2>
                    {descriptionAmendments.map((amendment, index) => (
                      <article key={amendment.groupKey || `${amendment.body || amendment.Body}-${index}`} className='rounded-md border border-sky-900/70 bg-sky-950/30 p-3'>
                        <div className='mb-2 flex flex-wrap gap-2 text-xs text-sky-100/80'>
                          <span>Amendment {index + 1}</span>
                          <span>Submitted by @{amendment.createdBy || amendment.CreatedBy}</span>
                          {(amendment.approvedAt || amendment.ApprovedAt) && (
                            <span>Approved {new Date(amendment.approvedAt || amendment.ApprovedAt).toLocaleString()}</span>
                          )}
                          {amendment.answerLabels?.length > 1 && (
                            <span>
                              Applies to {amendment.answerLabels.length} answers: {amendment.answerLabels.join(', ')}
                            </span>
                          )}
                        </div>
                        <MarkdownLite className='text-gray-200'>{amendment.body || amendment.Body}</MarkdownLite>
                      </article>
                    ))}
                  </section>
                )}
              </div>
            </div>
          )}
        </>
      )}

      {canManageAnswerAdditions && (
        <section className='mb-4 grid gap-4 rounded-lg border border-emerald-800/70 bg-emerald-950/20 p-4'>
          <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
            <div>
              <p className='text-sm font-semibold text-emerald-100'>Answer Option Review</p>
              <p className='mt-1 text-xs text-emerald-100/70'>
                When this is on, active moderators can add answer options immediately. When it is off, their options wait for your approval.
              </p>
            </div>
            <button
              type='button'
              aria-pressed={autoApproveAnswerAdditions}
              onClick={toggleAnswerAdditionPolicy}
              disabled={savingAnswerPolicy}
              className={`relative inline-flex h-8 w-16 shrink-0 items-center rounded-full border transition ${
                autoApproveAnswerAdditions
                  ? 'border-emerald-300 bg-emerald-700'
                  : 'border-gray-600 bg-gray-800'
              } disabled:cursor-not-allowed disabled:opacity-60`}
            >
              <span
                className={`inline-block h-6 w-6 transform rounded-full bg-white transition ${
                  autoApproveAnswerAdditions ? 'translate-x-8' : 'translate-x-1'
                }`}
              />
              <span className='sr-only'>Toggle answer option auto-approval</span>
            </button>
          </div>
          <div className='text-xs text-emerald-100/80'>
            Auto-approval is <span className='font-semibold text-white'>{autoApproveAnswerAdditions ? 'on' : 'off'}</span>.
          </div>
          {answerPolicyMessage && (
            <div className='rounded-md bg-emerald-700 p-3 text-sm text-white'>{answerPolicyMessage}</div>
          )}
          {answerPolicyError && (
            <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{answerPolicyError}</div>
          )}
          <div className='grid gap-3 border-t border-emerald-800/60 pt-4'>
            <p className='text-sm font-semibold text-emerald-100'>Incoming Answer Options</p>
            <MarketGroupAnswerAdditionReviewQueue
              token={token}
              groupId={group.id || marketGroup.id}
              status='pending'
              emptyMessage='No pending answer options for this grouped market.'
              onReviewed={() => {
                setRefreshKey((current) => current + 1);
                if (refetchData) {
                  refetchData();
                }
              }}
            />
          </div>
        </section>
      )}

      {canProposeAnswerAddition && (
        <form onSubmit={submitAnswerAddition} className='mb-4 grid gap-3 rounded-lg border border-emerald-800/60 bg-emerald-950/20 p-4'>
          <div>
            <p className='text-sm font-semibold text-emerald-100'>
              {canManageAnswerAdditions ? 'Add Answer Option' : 'Propose Answer Option'}
            </p>
            <p className='mt-1 text-xs text-emerald-100/70'>
              {canManageAnswerAdditions
                ? `Adds a new YES/NO answer market to this group immediately. You will be charged ${addAnswerCost} credits.`
                : autoApproveAnswerAdditions
                  ? `This market auto-approves incoming options. If approved by policy, you are charged ${addAnswerCost} credits.`
                  : `Submits a new YES/NO answer market to the steward for review. If approved, you are charged ${addAnswerCost} credits.`}
            </p>
          </div>
          {answerAdditionMessage && (
            <div className='rounded-md bg-emerald-700 p-3 text-sm text-white'>{answerAdditionMessage}</div>
          )}
          {answerAdditionError && (
            <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{answerAdditionError}</div>
          )}
          <div className='flex flex-col gap-2 sm:flex-row'>
            <input
              type='text'
              value={answerAdditionLabel}
              onChange={(event) => setAnswerAdditionLabel(event.target.value)}
              maxLength={160}
              placeholder='New answer label'
              className='min-w-0 flex-1 rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
              required
            />
            <button
              type='submit'
              disabled={submittingAnswerAddition || !answerAdditionLabel.trim()}
              className='rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50'
            >
              {submittingAnswerAddition ? 'Submitting...' : canManageAnswerAdditions ? 'Add Answer' : 'Submit Answer'}
            </button>
          </div>
        </form>
      )}

      {canProposeDescriptionAmendment && (
        <form onSubmit={submitDescriptionAmendment} className='mb-4 grid gap-3 rounded-lg border border-sky-800/60 bg-sky-950/20 p-4'>
          <div>
            <p className='text-sm font-semibold text-sky-100'>Propose Description Amendment</p>
            <p className='mt-1 text-xs text-sky-100/70'>
              Applies the same append-only clarification to every answer in this grouped market.
            </p>
          </div>
          {amendmentMessage && (
            <div className='rounded-md bg-emerald-700 p-3 text-sm text-white'>{amendmentMessage}</div>
          )}
          {amendmentError && (
            <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{amendmentError}</div>
          )}
          <textarea
            value={amendmentBody}
            onChange={(event) => setAmendmentBody(event.target.value)}
            rows={5}
            maxLength={2000}
            placeholder='Append clarification using markdown-lite. Raw HTML is not allowed.'
            className='rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
            required
          />
          {amendmentBody && (
            <div className='rounded-md border border-gray-700 bg-gray-950 p-3'>
              <p className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400'>Preview</p>
              <MarkdownLite>{amendmentBody}</MarkdownLite>
            </div>
          )}
          <input
            type='text'
            value={amendmentReason}
            onChange={(event) => setAmendmentReason(event.target.value)}
            maxLength={500}
            placeholder='Optional submit reason for admin review'
            className='rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
          />
          <button
            type='submit'
            disabled={submittingAmendment || !amendmentBody.trim()}
            className='justify-self-start rounded-md bg-sky-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50'
          >
            {submittingAmendment ? 'Submitting...' : 'Submit Amendment for Review'}
          </button>
        </form>
      )}

      <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
        {[
          { label: 'Answers', value: answers.length, icon: '▦' },
          { label: 'Users', value: aggregate.users, icon: '👤' },
          { label: 'Volume', value: Math.round(aggregate.volume), icon: '📊' },
          { label: 'Closes', value: closeStatusValue, icon: '📅' },
        ].map((item) => (
          <div key={item.label} className='rounded-lg bg-gray-800 p-2'>
            <div className='text-xs text-gray-400'>{item.label}</div>
            <div className='truncate text-sm font-semibold'>
              {item.icon} {item.value}
            </div>
          </div>
        ))}
      </div>

      {aggregate.dust > 0 && (
        <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
          <div className='bg-gray-800 p-2 rounded-lg'>
            <div className='text-xs text-gray-400'>Dust</div>
            <div className='text-sm font-semibold truncate'>✨ {aggregate.dust}</div>
          </div>
        </div>
      )}

      <div className='mb-4 grid gap-2 sm:grid-cols-3'>
        {sortedAnswers.map((answer) => {
          const childMarket = answer?.market?.market || {};
          const childResolved = Boolean(childMarket.isResolved);
          const childResolution = String(childMarket.resolutionResult || '').toUpperCase();
          const resolutionTone = childResolution === 'YES'
            ? 'border-emerald-600/70 bg-emerald-950/30 text-emerald-100'
            : 'border-rose-700/70 bg-rose-950/25 text-rose-100';
          return (
            <div key={answer.id || answer.marketId} className='rounded-lg bg-gray-800 p-3 text-center'>
              <div className='truncate text-sm font-semibold text-white'>{answer.answerLabel}</div>
              <div className='mt-1 text-xs text-gray-400'>YES {probabilityDisplay(answer)}</div>
              {childResolved && (
                <div className={`mt-2 rounded-full border px-2 py-1 text-xs font-semibold ${resolutionTone}`}>
                  Resolved {childResolution || 'N/A'}
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className='flex items-center justify-center mb-4 space-x-4 py-4'>
        {canResolveGroup && (
          <button
            type='button'
            onClick={() => setShowResolveModal(true)}
            className='min-w-32 rounded border bg-custom-gray-light px-4 py-2 text-xs text-white transition hover:bg-neutral-btn sm:text-sm md:text-base'
          >
            RESOLVE
          </button>
        )}
        <button
          type='button'
          disabled={!anyTradableAnswer}
          onClick={() => setShowTradeModal(true)}
          className='min-w-32 rounded border bg-custom-gray-light px-4 py-2 text-xs text-white transition hover:bg-neutral-btn disabled:cursor-not-allowed disabled:opacity-50 sm:text-sm md:text-base'
        >
          {tradeButtonLabel}
        </button>
      </div>

      {anyTradableAnswer && (
        <TradeCTA onClick={() => setShowTradeModal(true)} disabled={!token} />
      )}

      <div className='mx-auto mb-4 w-full'>
        <GroupedActivityTabs groupId={group.id || marketGroup.id} token={token} refreshTrigger={refreshKey} />
      </div>

      <div className='h-32 md:hidden' />

      {showTradeModal && (
        <div className='fixed inset-0 z-50 flex items-center justify-center bg-gray-600 bg-opacity-50'>
          <div className='bet-modal relative m-4 mx-auto max-h-[90vh] w-[min(94vw,780px)] overflow-y-auto rounded-lg bg-blue-900 p-6 text-white shadow-xl'>
            <div className='mb-4 pr-8'>
              <h2 className='text-xl'>Trade</h2>
              <p className='mt-1 text-sm text-gray-300'>
                Select an answer, then buy or sell shares.
              </p>
            </div>
            <div className='grid gap-4 md:grid-cols-[230px,minmax(0,1fr)]'>
              <div className='rounded-lg border border-gray-200/30 bg-blue-950/25 p-3'>
                <h3 className='mb-3 text-sm font-semibold text-gray-200'>Answer</h3>
                <div className='grid max-h-72 gap-2 overflow-y-auto pr-1 md:max-h-[520px]'>
                  {sortedAnswers.map((answer) => {
                    const childMarket = answer?.market?.market || {};
                    const tradable = canTradeMarket(childMarket, isLoggedIn);
                    const active = Number(answer.marketId) === Number(selectedTradeAnswer?.marketId);
                    return (
                      <button
                        key={answer.marketId}
                        type='button'
                        onClick={() => setSelectedTradeMarketId(answer.marketId)}
                        className={`w-full rounded border px-4 py-2 text-left text-white transition focus:outline-none ${
                          active
                            ? 'border-white/70 bg-neutral-btn'
                            : 'border-white/20 bg-custom-gray-light hover:bg-neutral-btn'
                        }`}
                      >
                        <span className='block truncate text-sm font-semibold'>{answerLabelFor(answer)}</span>
                        <span className='mt-1 block text-xs text-gray-300'>
                          YES {probabilityDisplay(answer)}
                          {!tradable ? ' · Closed' : ''}
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>
              <div className='min-w-0'>
                <div className='mb-3 border-b border-gray-200 pb-3'>
                  <p className='text-sm text-gray-300'>Selected Answer</p>
                  <p className='mt-1 text-lg text-white'>
                    {selectedTradeAnswer ? answerLabelFor(selectedTradeAnswer) : 'No answer selected'}
                  </p>
                </div>
                {selectedTradeIsTradable ? (
                  <TradeTabs
                    marketId={selectedTradeAnswer.marketId}
                    market={selectedTradeMarket}
                    token={token}
                    onTransactionSuccess={handleTransactionSuccess}
                  />
                ) : (
                  <div className='rounded-lg bg-blue-900 p-6 text-sm text-white'>
                    This answer is not open for trading.
                  </div>
                )}
              </div>
            </div>
            <button
              type='button'
              onClick={() => setShowTradeModal(false)}
              className='absolute right-0 top-0 mr-4 mt-4 text-gray-400 hover:text-white'
            >
              x
            </button>
          </div>
        </div>
      )}

      {showResolveModal && (
        <div className='fixed inset-0 z-50 flex items-center justify-center bg-gray-600 bg-opacity-50'>
          <div className='relative m-6 mx-auto max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-lg bg-blue-950 p-6 text-white shadow-xl'>
            <h2 className='mb-2 text-xl font-semibold'>Resolve Grouped Market</h2>
            <p className='mb-4 text-sm text-blue-100/80'>
              Resolution applies to every answer market in this group. Child markets use the normal binary payout path.
            </p>
            {resolveError && (
              <div className='mb-4 rounded-md bg-red-700 p-3 text-sm text-white'>{resolveError}</div>
            )}

            <div className='mb-4 grid gap-2 sm:grid-cols-2'>
              <button
                type='button'
                onClick={() => setResolveMode('exclusive_yes')}
                className={`rounded-md border px-3 py-2 text-left text-sm transition ${
                  resolveMode === 'exclusive_yes'
                    ? 'border-sky-300 bg-sky-800 text-white'
                    : 'border-blue-800 bg-blue-900 text-blue-100 hover:bg-blue-800'
                }`}
              >
                <span className='block font-semibold'>One answer resolves YES</span>
                <span className='mt-1 block text-xs opacity-80'>Selected answer becomes YES; all others become NO.</span>
              </button>
              <button
                type='button'
                onClick={() => setResolveMode('manual')}
                className={`rounded-md border px-3 py-2 text-left text-sm transition ${
                  resolveMode === 'manual'
                    ? 'border-sky-300 bg-sky-800 text-white'
                    : 'border-blue-800 bg-blue-900 text-blue-100 hover:bg-blue-800'
                }`}
              >
                <span className='block font-semibold'>Resolve each answer manually</span>
                <span className='mt-1 block text-xs opacity-80'>Assign YES or NO independently for every answer.</span>
              </button>
            </div>

            {resolveMode === 'exclusive_yes' ? (
              <div className='mb-4 grid gap-2'>
                {sortedAnswers.map((answer) => (
                  <label
                    key={answer.marketId}
                    className='flex cursor-pointer items-center gap-3 rounded-md border border-blue-800 bg-blue-900 px-3 py-2 text-sm hover:bg-blue-800'
                  >
                    <input
                      type='radio'
                      name='winningAnswer'
                      checked={Number(winningMarketId) === Number(answer.marketId)}
                      onChange={() => setWinningMarketId(answer.marketId)}
                    />
                    <span className='font-semibold'>{answer.answerLabel}</span>
                    <span className='text-xs text-blue-100/70'>Market #{answer.marketId}</span>
                  </label>
                ))}
              </div>
            ) : (
              <div className='mb-4 grid gap-2'>
                {sortedAnswers.map((answer) => (
                  <div key={answer.marketId} className='rounded-md border border-blue-800 bg-blue-900 p-3'>
                    <div className='mb-2 flex flex-wrap items-center justify-between gap-2'>
                      <span className='font-semibold'>{answer.answerLabel}</span>
                      <span className='text-xs text-blue-100/70'>Market #{answer.marketId}</span>
                    </div>
                    <div className='grid grid-cols-2 gap-2'>
                      {['YES', 'NO'].map((resolution) => (
                        <button
                          key={resolution}
                          type='button'
                          onClick={() => setManualResolutions((current) => ({
                            ...current,
                            [answer.marketId]: resolution,
                          }))}
                          className={`rounded-md px-3 py-2 text-sm font-semibold transition ${
                            manualResolutions[answer.marketId] === resolution
                              ? 'bg-sky-700 text-white'
                              : 'bg-blue-800 text-blue-100 hover:bg-blue-700'
                          }`}
                        >
                          {resolution}
                        </button>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className='flex flex-wrap justify-end gap-2 border-t border-blue-800 pt-4'>
              <button
                type='button'
                onClick={() => setShowResolveModal(false)}
                className='rounded-md border border-blue-700 px-4 py-2 text-sm text-blue-100 transition hover:bg-blue-900'
              >
                Cancel
              </button>
              <button
                type='button'
                onClick={submitGroupResolution}
                disabled={resolvingGroup || (resolveMode === 'exclusive_yes' && !winningMarketId)}
                className='rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50'
              >
                {resolvingGroup ? 'Resolving...' : 'Confirm Resolution'}
              </button>
            </div>
            <button
              type='button'
              onClick={() => setShowResolveModal(false)}
              className='absolute right-0 top-0 mr-4 mt-4 text-gray-400 hover:text-white'
            >
              x
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
