import React, { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import LoadingSpinner from '../loaders/LoadingSpinner';
import GroupedMarketChart from '../charts/GroupedMarketChart';
import MarketTagChips from '../markets/MarketTagChips';
import SiteTabs from '../tabs/SiteTabs';
import TradeTabs from '../tabs/TradeTabs';
import TradeCTA from '../TradeCTA';
import MarkdownLite from '../markdown/MarkdownLite';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import { getMarketGroupDetails } from '../../api/marketsApi';
import { proposeMarketDescriptionAmendment } from '../../api/marketDescriptionAmendmentsApi';
import { apiRequest } from '../../api/httpClient';

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

const groupedFetchJson = async (path, token = '') => {
  return apiRequest(path, {
    authenticated: Boolean(token),
    authToken: token,
    reasonMessages: {
      RATE_LIMITED: 'Grouped activity is loading several answer markets. Wait a moment and try again.',
    },
    fallbackMessage: 'Failed to load grouped market activity.',
  });
};

const answerLabelFor = (answer) => answer?.answerLabel || `Answer ${Number(answer?.displayOrder || 0) + 1}`;

const childMarketFor = (answer) => answer?.market?.market || {};

const groupedAnswerMeta = (answers = []) => answers.map((answer) => ({
  answerId: answer.id,
  answerLabel: answerLabelFor(answer),
  marketId: answer.marketId,
  market: childMarketFor(answer),
}));

const childDescriptionAmendments = (answer) => {
  const amendments = answer?.descriptionAmendments || answer?.DescriptionAmendments || [];
  return Array.isArray(amendments) ? amendments : [];
};

const fetchSequentially = async (items, fetcher) => {
  const results = [];
  for (const item of items) {
    // Avoid bursting one request per answer at the same instant in dev/staging.
    // The grouped page is display-only, so sequential reads are acceptable here.
    // eslint-disable-next-line no-await-in-loop
    results.push(await fetcher(item));
  }
  return results;
};

const uniqueGroupedAmendments = (answers = []) => {
  const seen = new Set();
  const amendments = [];
  answers.forEach((answer) => {
    (answer.descriptionAmendments || []).forEach((amendment) => {
      const key = [
        amendment.body || amendment.Body || '',
        amendment.createdBy || amendment.CreatedBy || '',
        amendment.approvedAt || amendment.ApprovedAt || '',
      ].join('|');
      if (!key.trim() || seen.has(key)) {
        return;
      }
      seen.add(key);
      amendments.push({
        ...amendment,
        answerLabel: answerLabelFor(answer),
      });
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

const GroupedBetsActivity = ({ answers, refreshTrigger }) => {
  const pageSize = 20;
  const [bets, setBets] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [error, setError] = useState('');
  const answerMeta = useMemo(() => groupedAnswerMeta(answers), [answers]);

  useEffect(() => {
    setPage(0);
  }, [answers, refreshTrigger]);

  useEffect(() => {
    let ignore = false;
    const loadBets = async () => {
      const offset = page * pageSize;
      const fetchLimit = offset + pageSize + 1;
      setError('');
      try {
        const results = await fetchSequentially(answerMeta, async (answer) => {
          const rows = await groupedFetchJson(`/v0/markets/bets/${answer.marketId}?limit=${fetchLimit}&offset=0`);
          return (Array.isArray(rows) ? rows : []).map((bet) => ({
            ...bet,
            answerLabel: answer.answerLabel,
            answerMarketId: answer.marketId,
          }));
        });
        const merged = results.flat().sort((left, right) => (
          new Date(right.placedAt).getTime() - new Date(left.placedAt).getTime()
        ));
        if (!ignore) {
          setBets(merged.slice(offset, offset + pageSize));
          setHasNextPage(merged.length > offset + pageSize);
        }
      } catch (err) {
        if (!ignore) {
          setBets([]);
          setHasNextPage(false);
          setError(err.message || 'Failed to load grouped bets.');
        }
      }
    };
    if (answerMeta.length) {
      loadBets();
    }
    return () => {
      ignore = true;
    };
  }, [answerMeta, page, refreshTrigger]);

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

const GroupedPositionsActivity = ({ answers, token, refreshTrigger }) => {
  const pageSize = 20;
  const [positions, setPositions] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [freshnessLabel, setFreshnessLabel] = useState('');
  const [error, setError] = useState('');
  const answerMeta = useMemo(() => groupedAnswerMeta(answers), [answers]);

  useEffect(() => {
    setPage(0);
  }, [answers, refreshTrigger, token]);

  useEffect(() => {
    let ignore = false;
    const loadPositions = async () => {
      if (!token) {
        setPositions([]);
        setHasNextPage(false);
        setFreshnessLabel('');
        setError('');
        return;
      }
      setError('');
      try {
        const results = await fetchSequentially(answerMeta, async (answer) => {
          const payload = await groupedFetchJson(`/v0/markets/positions/${answer.marketId}?limit=200&offset=0`, token);
          const rows = Array.isArray(payload?.positions)
            ? payload.positions
            : Array.isArray(payload)
              ? payload
              : [];
          return {
            answer,
            freshness: payload?.freshness || null,
            rows: rows.map((position) => ({ ...position, answerLabel: answer.answerLabel, answerMarketId: answer.marketId })),
          };
        });
        const byUser = new Map();
        results.forEach(({ rows }) => {
          rows.forEach((position) => {
            if (!position.username || (toNumber(position.yesSharesOwned) <= 0 && toNumber(position.noSharesOwned) <= 0)) {
              return;
            }
            const current = byUser.get(position.username) || {
              username: position.username,
              yesSharesOwned: 0,
              noSharesOwned: 0,
              value: 0,
              totalSpent: 0,
              totalSpentInPlay: 0,
              answers: [],
            };
            current.yesSharesOwned += toNumber(position.yesSharesOwned);
            current.noSharesOwned += toNumber(position.noSharesOwned);
            current.value += toNumber(position.value);
            current.totalSpent += toNumber(position.totalSpent);
            current.totalSpentInPlay += toNumber(position.totalSpentInPlay);
            current.answers.push(position);
            byUser.set(position.username, current);
          });
        });
        const merged = [...byUser.values()].sort((left, right) => (
          (right.yesSharesOwned + right.noSharesOwned) - (left.yesSharesOwned + left.noSharesOwned)
        ));
        const generatedTimes = results
          .map((item) => item.freshness?.generatedAt)
          .filter(Boolean)
          .map((value) => new Date(value))
          .filter((date) => !Number.isNaN(date.getTime()));
        if (!ignore) {
          const offset = page * pageSize;
          setPositions(merged.slice(offset, offset + pageSize));
          setHasNextPage(merged.length > offset + pageSize);
          setFreshnessLabel(generatedTimes.length ? new Date(Math.min(...generatedTimes.map((date) => date.getTime()))).toLocaleTimeString() : '');
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
    if (answerMeta.length) {
      loadPositions();
    }
    return () => {
      ignore = true;
    };
  }, [answerMeta, page, refreshTrigger, token]);

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
          Grouped positions combine child-market snapshots generated as early as {freshnessLabel}. Trade confirmations remain authoritative.
        </div>
      )}
      {!token && <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>Log in to see grouped positions.</div>}
      {error && <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{error}</div>}
      {token && positions.length === 0 && !error && (
        <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>No positions yet</div>
      )}
      {positions.map((position) => (
        <article key={position.username} className='mb-3 rounded-lg border border-gray-800 bg-gray-900 p-4'>
          <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
            <Link to={`/user/${position.username}`} className='font-semibold text-blue-400 hover:text-blue-300'>
              {position.username}
            </Link>
            <div className='text-sm text-gray-300'>
              YES {position.yesSharesOwned} · NO {position.noSharesOwned} · Value {position.value}
            </div>
          </div>
          <div className='mt-3 flex flex-wrap gap-2'>
            {position.answers.map((answer) => (
              <span key={`${position.username}-${answer.answerMarketId}`} className='rounded-full border border-gray-700 bg-gray-800 px-3 py-1 text-xs text-gray-200'>
                {answer.answerLabel}: {answer.yesSharesOwned}Y/{answer.noSharesOwned}N
              </span>
            ))}
          </div>
        </article>
      ))}
    </div>
  );
};

const GroupedLeaderboardActivity = ({ answers, refreshTrigger }) => {
  const pageSize = 20;
  const [leaderboard, setLeaderboard] = useState([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [freshnessLabel, setFreshnessLabel] = useState('');
  const [error, setError] = useState('');
  const answerMeta = useMemo(() => groupedAnswerMeta(answers), [answers]);

  useEffect(() => {
    setPage(0);
  }, [answers, refreshTrigger]);

  useEffect(() => {
    let ignore = false;
    const loadLeaderboard = async () => {
      setError('');
      try {
        const results = await fetchSequentially(answerMeta, async (answer) => {
          const payload = await groupedFetchJson(`/v0/markets/${answer.marketId}/leaderboard?limit=200&offset=0`);
          return {
            answer,
            freshness: payload?.freshness || null,
            rows: (Array.isArray(payload?.leaderboard) ? payload.leaderboard : []).map((row) => ({
              ...row,
              answerLabel: answer.answerLabel,
              answerMarketId: answer.marketId,
            })),
          };
        });
        const byUser = new Map();
        results.forEach(({ rows }) => {
          rows.forEach((row) => {
            if (!row.username) {
              return;
            }
            const current = byUser.get(row.username) || {
              username: row.username,
              profit: 0,
              currentValue: 0,
              totalSpent: 0,
              yesSharesOwned: 0,
              noSharesOwned: 0,
              answers: [],
            };
            current.profit += toNumber(row.profit);
            current.currentValue += toNumber(row.currentValue);
            current.totalSpent += toNumber(row.totalSpent);
            current.yesSharesOwned += toNumber(row.yesSharesOwned);
            current.noSharesOwned += toNumber(row.noSharesOwned);
            current.answers.push(row);
            byUser.set(row.username, current);
          });
        });
        const merged = [...byUser.values()]
          .sort((left, right) => right.profit - left.profit)
          .map((row, index) => ({ ...row, rank: index + 1 }));
        const generatedTimes = results
          .map((item) => item.freshness?.generatedAt)
          .filter(Boolean)
          .map((value) => new Date(value))
          .filter((date) => !Number.isNaN(date.getTime()));
        if (!ignore) {
          const offset = page * pageSize;
          setLeaderboard(merged.slice(offset, offset + pageSize));
          setHasNextPage(merged.length > offset + pageSize);
          setFreshnessLabel(generatedTimes.length ? new Date(Math.min(...generatedTimes.map((date) => date.getTime()))).toLocaleTimeString() : '');
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
    if (answerMeta.length) {
      loadLeaderboard();
    }
    return () => {
      ignore = true;
    };
  }, [answerMeta, page, refreshTrigger]);

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
          Grouped leaderboard combines child-market snapshots generated as early as {freshnessLabel}. Trade confirmations remain authoritative.
        </div>
      )}
      {error && <div className='rounded-md bg-red-700 p-3 text-sm text-white'>{error}</div>}
      {leaderboard.length === 0 && !error && (
        <div className='rounded-md bg-gray-800 p-4 text-center text-sm text-gray-400'>No participants yet</div>
      )}
      {leaderboard.map((entry) => (
        <article key={entry.username} className='mb-3 rounded-lg border border-gray-800 bg-gray-900 p-4'>
          <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
            <div className='flex items-center gap-3'>
              <span className='rounded bg-gray-800 px-2 py-1 text-sm font-semibold text-white'>#{entry.rank}</span>
              <Link to={`/user/${entry.username}`} className='font-semibold text-blue-400 hover:text-blue-300'>
                {entry.username}
              </Link>
            </div>
            <div className={entry.profit >= 0 ? 'text-green-400' : 'text-red-400'}>
              {entry.profit >= 0 ? '+' : ''}{entry.profit} profit
            </div>
          </div>
          <div className='mt-2 text-sm text-gray-300'>
            Value {entry.currentValue} · Spent {entry.totalSpent} · Shares {entry.yesSharesOwned}Y/{entry.noSharesOwned}N
          </div>
          <div className='mt-3 flex flex-wrap gap-2'>
            {entry.answers.map((answer) => (
              <span key={`${entry.username}-${answer.answerMarketId}`} className='rounded-full border border-gray-700 bg-gray-800 px-3 py-1 text-xs text-gray-200'>
                {answer.answerLabel}: {answer.profit >= 0 ? '+' : ''}{answer.profit}
              </span>
            ))}
          </div>
        </article>
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

const GroupedActivityTabs = ({ answers, token, refreshTrigger }) => {
  const tabsData = [
    { label: 'Bets', content: <GroupedBetsActivity answers={answers} refreshTrigger={refreshTrigger} /> },
    { label: 'Positions', content: <GroupedPositionsActivity answers={answers} token={token} refreshTrigger={refreshTrigger} /> },
    { label: 'Leaderboard', content: <GroupedLeaderboardActivity answers={answers} refreshTrigger={refreshTrigger} /> },
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
  refetchData,
}) {
  const [groupData, setGroupData] = useState(null);
  const [answers, setAnswers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showFullDescription, setShowFullDescription] = useState(false);
  const [showTradeModal, setShowTradeModal] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [amendmentBody, setAmendmentBody] = useState('');
  const [amendmentReason, setAmendmentReason] = useState('');
  const [amendmentMessage, setAmendmentMessage] = useState('');
  const [amendmentError, setAmendmentError] = useState('');
  const [submittingAmendment, setSubmittingAmendment] = useState(false);

  useEffect(() => {
    let ignore = false;

    const loadGroup = async () => {
      setLoading(true);
      setError('');
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
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Failed to load grouped market.');
        }
      } finally {
        if (!ignore) {
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

  const group = groupData?.group || marketGroup || {};
  const groupCreator = groupData?.creator || creator || {};
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
  const descriptionAmendments = useMemo(() => uniqueGroupedAmendments(sortedAnswers), [sortedAnswers]);
  const anyTradableAnswer = sortedAnswers.some((answer) => canTradeMarket(answer?.market?.market || {}, isLoggedIn));
  const groupStewardUsername = group.stewardUsername || group.creatorUsername || fallbackMarket?.stewardUsername || fallbackMarket?.creatorUsername || '';
  const canProposeDescriptionAmendment =
    isLoggedIn &&
    token &&
    String(username || '').trim() === String(groupStewardUsername || '').trim() &&
    sortedAnswers.length > 0 &&
    closeDate &&
    new Date(closeDate) > new Date() &&
    !['rejected', 'resolved', 'cancelled'].includes(String(group.lifecycleStatus || '').toLowerCase());
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

  const handleTransactionSuccess = () => {
    setShowTradeModal(false);
    setRefreshKey((current) => current + 1);
    if (refetchData) {
      refetchData();
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

  const answerTradeTabs = sortedAnswers.map((answer) => {
    const childMarket = answer?.market?.market || {};
    const tradable = canTradeMarket(childMarket, isLoggedIn);
    return {
      label: answer.answerLabel || `Answer ${answer.displayOrder + 1}`,
      content: tradable ? (
        <TradeTabs
          marketId={answer.marketId}
          market={childMarket}
          token={token}
          onTransactionSuccess={handleTransactionSuccess}
        />
      ) : (
        <div className='rounded-lg bg-blue-950/70 p-4 text-sm text-blue-50'>
          This answer is not open for trading.
        </div>
      ),
    };
  });

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
          <span>@{groupCreator.username || group.creatorUsername || fallbackMarket.creatorUsername}</span>
          <span>•</span>
          <span>{answers.length} answers</span>
          <span>•</span>
          <span>Closes {formatResolutionDate(closeDate)}</span>
        </div>
        <MarketTagChips tags={tags} className='mt-3' />
      </section>

      <div className='mb-4'>
        <GroupedMarketChart answers={answers} title='Probability Changes' />
      </div>

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
                      <article key={`${amendment.body || amendment.Body}-${index}`} className='rounded-md border border-sky-900/70 bg-sky-950/30 p-3'>
                        <div className='mb-2 flex flex-wrap gap-2 text-xs text-sky-100/80'>
                          <span>Amendment {index + 1}</span>
                          <span>Submitted by @{amendment.createdBy || amendment.CreatedBy}</span>
                          {(amendment.approvedAt || amendment.ApprovedAt) && (
                            <span>Approved {new Date(amendment.approvedAt || amendment.ApprovedAt).toLocaleString()}</span>
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
          { label: 'Closes', value: formatResolutionDate(closeDate), icon: '📅' },
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
        {sortedAnswers.map((answer) => (
          <div key={answer.id || answer.marketId} className='rounded-lg bg-gray-800 p-3 text-center'>
            <div className='truncate text-sm font-semibold text-white'>{answer.answerLabel}</div>
            <div className='mt-1 text-xs text-gray-400'>YES {probabilityDisplay(answer)}</div>
          </div>
        ))}
      </div>

      <div className='flex items-center justify-center mb-4 space-x-4 py-4'>
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
        <GroupedActivityTabs answers={sortedAnswers} token={token} refreshTrigger={refreshKey} />
      </div>

      <div className='h-32 md:hidden' />

      {showTradeModal && (
        <div className='fixed inset-0 z-50 flex items-center justify-center bg-gray-600 bg-opacity-50'>
          <div className='bet-modal relative m-6 mx-auto rounded-lg bg-blue-900 p-6 text-white' style={{ width: '350px' }}>
            <SiteTabs tabs={answerTradeTabs} />
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
    </div>
  );
}
