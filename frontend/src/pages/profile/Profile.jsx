import React, { useEffect, useMemo, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import PrivateUserInfoLayout from '../../components/layouts/profile/private/PrivateUserInfoLayout';
import PortfolioTabContent from '../../components/layouts/profile/public/PortfolioTabContent';
import UserFinancialStatementsLayout from '../../components/layouts/profile/public/UserFinancialStatementsLayout';
import MarketLifecycleTable, { groupLifecycleMarketRows } from '../../components/layouts/profile/MarketLifecycleTable';
import SiteTabs from '../../components/tabs/SiteTabs';
import useUserData from '../../hooks/useUserData';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { listMyLifecycleMarkets } from '../../api/lifecycleMarketsApi';
import { listMyMarketDescriptionAmendments } from '../../api/marketDescriptionAmendmentsApi';
import MarkdownLite from '../../components/markdown/MarkdownLite';
import {
  listMarketGroupAnswerAdditionsForReview,
  reviewMarketGroupAnswerAdditionForSteward,
} from '../../api/marketsApi';

const ErrorMessage = ({ message }) => (
  <div
    className='bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative'
    role='alert'
  >
    <strong className='font-bold'>Error:</strong>
    <span className='block sm:inline'> {message}</span>
  </div>
);

const lifecycleTabByStatus = {
  proposed: 'Proposed Markets',
  published: 'Published Markets',
  rejected: 'Rejected Markets',
};

const amendmentTabByStatus = {
  pending: 'Pending Amendments',
  approved: 'Approved Amendments',
  rejected: 'Rejected Amendments',
};

const answerOptionTabByStatus = {
  pending: 'Pending Answer Options',
  approved: 'Approved Answer Options',
  rejected: 'Rejected Answer Options',
};

const lifecycleShortLabelByStatus = {
  proposed: 'Proposed',
  published: 'Published',
  rejected: 'Rejected',
};

const reviewShortLabelByStatus = {
  pending: 'Pending',
  approved: 'Approved',
  rejected: 'Rejected',
};

const accountTabLabels = ['User Info', 'Portfolio', 'Financials'];
const PROFILE_PAGE_SIZE = 20;
const PROFILE_FETCH_BATCH_SIZE = 100;
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

const pagedRows = (rows = [], page = 0) => {
  const total = rows.length;
  const maxPage = Math.max(0, Math.ceil(total / PROFILE_PAGE_SIZE) - 1);
  const currentPage = Math.min(Math.max(0, page), maxPage);
  const start = currentPage * PROFILE_PAGE_SIZE;
  return {
    currentPage,
    start,
    total,
    rows: rows.slice(start, start + PROFILE_PAGE_SIZE),
    hasPrevious: currentPage > 0,
    hasNext: start + PROFILE_PAGE_SIZE < total,
  };
};

const ProfilePaginationControls = ({ label, pageInfo, onPageChange }) => {
  if (!pageInfo.total) {
    return null;
  }

  return (
    <div className='flex flex-col gap-2 rounded-lg border border-gray-700 bg-gray-900/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between'>
      <div className='text-xs uppercase tracking-[0.16em] text-gray-400'>
        {label} page {pageInfo.currentPage + 1} ({pageInfo.start + 1}-{pageInfo.start + pageInfo.rows.length} of {pageInfo.total})
      </div>
      <div className='flex gap-2'>
        <button
          type='button'
          onClick={() => onPageChange((current) => Math.max(0, current - 1))}
          disabled={!pageInfo.hasPrevious}
          className={paginationButtonClass}
        >
          Previous
        </button>
        <button
          type='button'
          onClick={() => onPageChange((current) => current + 1)}
          disabled={!pageInfo.hasNext}
          className={paginationButtonClass}
        >
          Next
        </button>
      </div>
    </div>
  );
};

const statusFromLegacyTab = (label, lookup) => {
  const entry = Object.entries(lookup).find(([, tabLabel]) => tabLabel === label);
  return entry?.[0] || '';
};

const ProfileAccountTabs = ({ username, userData, defaultTab }) => {
  const tabsData = [
    {
      label: 'User Info',
      content: <PrivateUserInfoLayout userData={userData} />,
    },
    {
      label: 'Portfolio',
      content: <PortfolioTabContent username={username} />,
    },
    {
      label: 'Financials',
      content: <UserFinancialStatementsLayout username={username} />,
    },
  ];

  return <SiteTabs tabs={tabsData} defaultTab={defaultTab} />;
};

const ProfileMarketsTabs = ({ defaultStatus }) => {
  const tabsData = Object.entries(lifecycleShortLabelByStatus).map(([status, label]) => ({
    label,
    content: <ProfileMarketLifecycleTab status={status} />,
  }));

  return <SiteTabs tabs={tabsData} defaultTab={lifecycleShortLabelByStatus[defaultStatus] || 'Proposed'} />;
};

const amendmentNumber = (version) => Math.max(1, Number(version || 2) - 1);

const isGroupedAmendment = (amendment) => Boolean(amendment?.marketGroup?.id);

const amendmentRowKey = (amendment) => {
  if (!isGroupedAmendment(amendment)) {
    return `amendment:${amendment?.id}`;
  }
  return [
    'group-amendment',
    amendment.marketGroup.id,
    amendment.status,
    amendment.body,
    amendment.createdBy,
    amendment.submitReason || '',
  ].join('|');
};

const compactDescriptionAmendments = (amendments = []) => {
  const rows = [];
  const groups = new Map();

  amendments.forEach((amendment) => {
    if (!isGroupedAmendment(amendment)) {
      rows.push({ ...amendment, childAmendments: [amendment] });
      return;
    }

    const key = amendmentRowKey(amendment);
    const existing = groups.get(key);
    if (!existing) {
      const row = {
        ...amendment,
        marketTitle: amendment.marketGroup.questionTitle || amendment.marketTitle,
        marketDescription: amendment.marketGroup.description || amendment.marketDescription,
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

const marketChangeKey = ({ marketType, amendment, answerOption }) => {
  if (marketType === 'grouped') {
    return `group:${amendment?.marketGroup?.id || answerOption?.groupId || 0}`;
  }
  return `market:${amendment?.marketId || 0}`;
};

const marketChangeTitle = ({ marketType, amendment, answerOption }) => {
  if (marketType === 'grouped') {
    return amendment?.marketGroup?.questionTitle ||
      amendment?.marketTitle ||
      answerOption?.marketGroup?.questionTitle ||
      answerOption?.groupTitle ||
      `Grouped market #${answerOption?.groupId || amendment?.marketGroup?.id || ''}`;
  }
  return amendment?.marketTitle || `Market #${amendment?.marketId || ''}`;
};

const marketChangeHref = ({ marketType, amendment, answerOption }) => {
  if (marketType === 'grouped') {
    const childMarketId = amendment?.marketId || answerOption?.marketId;
    if (childMarketId) {
      return `/markets/${childMarketId}`;
    }
    const groupId = amendment?.marketGroup?.id || answerOption?.groupId;
    return groupId ? `/markets/group/${groupId}` : '#';
  }
  return amendment?.marketId ? `/markets/${amendment.marketId}` : '#';
};

const groupProfileMarketChanges = ({ marketType, amendments = [], answerOptions = [] }) => {
  const rows = new Map();

  const ensureRow = ({ amendment = null, answerOption = null }) => {
    const key = marketChangeKey({ marketType, amendment, answerOption });
    if (!rows.has(key)) {
      rows.set(key, {
        key,
        marketType,
        title: marketChangeTitle({ marketType, amendment, answerOption }),
        href: marketChangeHref({ marketType, amendment, answerOption }),
        amendments: [],
        answerOptions: [],
        groupId: amendment?.marketGroup?.id || answerOption?.groupId || 0,
        marketId: amendment?.marketId || answerOption?.marketId || 0,
      });
    }
    const row = rows.get(key);
    if (row.href === '#' || row.href === '/markets/group/0') {
      row.href = marketChangeHref({ marketType, amendment, answerOption });
    }
    if (!row.title || row.title.endsWith('#')) {
      row.title = marketChangeTitle({ marketType, amendment, answerOption });
    }
    return row;
  };

  compactDescriptionAmendments(amendments).forEach((amendment) => {
    ensureRow({ amendment }).amendments.push(amendment);
  });

  answerOptions.forEach((answerOption) => {
    ensureRow({ answerOption }).answerOptions.push(answerOption);
  });

  return Array.from(rows.values()).sort((left, right) => left.title.localeCompare(right.title));
};

const normalizedSearch = (value) => String(value || '').trim().toLowerCase();

const marketChangeSearchText = (change) => [
  change.title,
  change.key,
  change.marketType,
  change.groupId ? `group ${change.groupId}` : '',
  change.marketId ? `market ${change.marketId}` : '',
  ...(change.amendments || []).flatMap((amendment) => [
    amendment.marketTitle,
    amendment.marketDescription,
    amendment.body,
    amendment.submitReason,
    amendment.createdBy,
    amendment.marketId ? `market ${amendment.marketId}` : '',
    amendment.marketGroup?.questionTitle,
    amendment.marketGroup?.answerLabel,
    amendment.marketGroup?.id ? `group ${amendment.marketGroup.id}` : '',
    ...(amendment.childAmendments || []).map((child) => child.marketGroup?.answerLabel || `market ${child.marketId}`),
  ]),
  ...(change.answerOptions || []).flatMap((addition) => [
    addition.groupTitle,
    addition.answerLabel,
    addition.proposedBy,
    addition.reviewedBy,
    addition.rejectionReason,
    addition.groupId ? `group ${addition.groupId}` : '',
    addition.marketId ? `market ${addition.marketId}` : '',
    addition.marketGroup?.questionTitle,
  ]),
].filter(Boolean).join(' ').toLowerCase();

const filterMarketChanges = (changes, query) => {
  const needle = normalizedSearch(query);
  if (!needle) {
    return changes;
  }
  return changes.filter((change) => marketChangeSearchText(change).includes(needle));
};

const ProfileMarketChangesTabs = ({ defaultMarketType, defaultStatus }) => {
  const tabsData = [
    {
      label: 'Grouped Markets',
      content: <ProfileMarketChangesByType marketType='grouped' defaultStatus={defaultStatus} />,
    },
    {
      label: 'Binary Markets',
      content: <ProfileMarketChangesByType marketType='binary' defaultStatus={defaultStatus} />,
    },
  ];

  return <SiteTabs tabs={tabsData} defaultTab={defaultMarketType === 'binary' ? 'Binary Markets' : 'Grouped Markets'} />;
};

const ProfileMarketChangesByType = ({ marketType, defaultStatus }) => {
  const tabsData = Object.entries(reviewShortLabelByStatus).map(([status, label]) => ({
    label,
    content: <ProfileMarketChangeStatusTab marketType={marketType} status={status} />,
  }));

  return <SiteTabs tabs={tabsData} defaultTab={reviewShortLabelByStatus[defaultStatus] || 'Pending'} />;
};

const ProfileMarketChangeStatusTab = ({ marketType, status }) => {
  const { token } = useAuth();
  const [amendments, setAmendments] = useState([]);
  const [answerOptions, setAnswerOptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyAdditionId, setBusyAdditionId] = useState(null);
  const [reasonById, setReasonById] = useState({});
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);

  useEffect(() => {
    setPage(0);
  }, [marketType, status, searchQuery]);

  useEffect(() => {
    let ignore = false;

    const loadChanges = async () => {
      setLoading(true);
      setError('');
      setSuccessMessage('');
      try {
        const amendmentData = await listMyMarketDescriptionAmendments({ token, status, limit: 200 });
        const rawAmendments = amendmentData.amendments || [];
        const filteredAmendments = rawAmendments.filter((amendment) => (
          marketType === 'grouped' ? isGroupedAmendment(amendment) : !isGroupedAmendment(amendment)
        ));

        let groupedAnswerOptions = [];
        if (marketType === 'grouped') {
          const answerOptionData = await listMarketGroupAnswerAdditionsForReview({
            token,
            status,
            limit: 200,
          });
          groupedAnswerOptions = answerOptionData.additions || [];
        }

        if (!ignore) {
          setAmendments(filteredAmendments);
          setAnswerOptions(groupedAnswerOptions);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load market changes.');
          setAmendments([]);
          setAnswerOptions([]);
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadChanges();
    return () => {
      ignore = true;
    };
  }, [marketType, status, token]);

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
      await reviewMarketGroupAnswerAdditionForSteward({
        token,
        additionId: addition.id,
        status: nextStatus,
        reason,
        confirm: nextStatus === 'approved',
      });
      updateReason(addition.id, '');
      setAnswerOptions((current) => current.filter((item) => item.id !== addition.id));
      setSuccessMessage(`Answer option "${addition.answerLabel}" ${nextStatus}.`);
    } catch (err) {
      setError(err.message || 'Unable to review grouped answer option.');
    } finally {
      setBusyAdditionId(null);
    }
  };

  const groupedChanges = useMemo(() => groupProfileMarketChanges({
    marketType,
    amendments,
    answerOptions,
  }), [marketType, amendments, answerOptions]);
  const visibleChanges = useMemo(
    () => filterMarketChanges(groupedChanges, searchQuery),
    [groupedChanges, searchQuery],
  );
  const pageInfo = pagedRows(visibleChanges, page);

  if (loading) {
    return <LoadingSpinner />;
  }

  const emptyLabel = marketType === 'grouped' ? 'grouped market' : 'binary market';

  return (
    <div className='grid gap-4'>
      {error && <ErrorMessage message={error} />}
      {successMessage && (
        <div className='rounded-md bg-emerald-700 p-3 text-sm text-white'>
          {successMessage}
        </div>
      )}
      {marketType === 'binary' && (
        <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-4 text-sm text-gray-300'>
          Binary markets only support description amendments. Answer options are fixed when the market is created.
        </div>
      )}
      <div className='grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
        <label htmlFor={`profile-market-change-search-${marketType}-${status}`} className='text-xs font-mono uppercase tracking-[0.16em] text-gray-400'>
          Search {reviewShortLabelByStatus[status] || status} {marketType === 'grouped' ? 'grouped market' : 'binary market'} changes
        </label>
        <input
          id={`profile-market-change-search-${marketType}-${status}`}
          type='search'
          value={searchQuery}
          onChange={(event) => setSearchQuery(event.target.value)}
          placeholder='Search title, ID, amendment text, answer option, or user'
          className='w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
        />
      </div>
      {groupedChanges.length === 0 ? (
        <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300'>
          No {status} {emptyLabel} changes found.
        </div>
      ) : visibleChanges.length === 0 ? (
        <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300'>
          No {status} {emptyLabel} changes match "{searchQuery}".
        </div>
      ) : (
        <>
          <ProfilePaginationControls
            label={`${reviewShortLabelByStatus[status] || status} ${emptyLabel} changes`}
            pageInfo={pageInfo}
            onPageChange={setPage}
          />
          {pageInfo.rows.map((change, index) => (
            <React.Fragment key={change.key}>
              {index > 0 && (
                <div className='my-2 h-px bg-gradient-to-r from-transparent via-primary-pink/50 to-transparent' />
              )}
              <ProfileMarketChangeCard
                change={change}
                status={status}
                reasonById={reasonById}
                busyAdditionId={busyAdditionId}
                onReasonChange={updateReason}
                onReviewAddition={reviewAddition}
              />
            </React.Fragment>
          ))}
        </>
      )}
    </div>
  );
};

const ProfileMarketChangeCard = ({
  change,
  status,
  reasonById,
  busyAdditionId,
  onReasonChange,
  onReviewAddition,
}) => {
  const isGrouped = change.marketType === 'grouped';
  const canReviewAnswerOptions = isGrouped && status === 'pending';

  return (
    <article className='grid gap-4 rounded-xl border border-gray-600 border-l-4 border-l-primary-pink/70 bg-gray-900/80 p-5 shadow-lg shadow-black/20'>
      <div className='grid gap-2'>
        <div className='flex flex-wrap items-center gap-2 text-sm text-gray-300'>
          <span className='rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100'>
            {isGrouped ? `Group #${change.groupId}` : `Market #${change.marketId}`}
          </span>
          <span className='rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200'>
            {isGrouped ? 'Grouped Market' : 'Binary Market'}
          </span>
        </div>
        <Link
          to={change.href}
          className='text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200'
        >
          {change.title}
        </Link>
      </div>

      <section className='grid gap-3'>
        <div className='flex items-center justify-between gap-3'>
          <h3 className='text-sm font-semibold uppercase tracking-[0.14em] text-sky-200'>
            Description Amendments
          </h3>
          <span className='text-xs text-gray-400'>{change.amendments.length}</span>
        </div>
        {change.amendments.length === 0 ? (
          <div className='rounded-md border border-gray-700 bg-gray-950 p-3 text-sm text-gray-400'>
            No {status} description amendments for this market.
          </div>
        ) : (
          change.amendments.map((amendment) => (
            <ProfileDescriptionAmendmentCard
              key={amendmentRowKey(amendment)}
              amendment={amendment}
              status={status}
            />
          ))
        )}
      </section>

      {isGrouped && (
        <section className='grid gap-3'>
          <div className='flex items-center justify-between gap-3'>
            <h3 className='text-sm font-semibold uppercase tracking-[0.14em] text-emerald-200'>
              Answer Options
            </h3>
            <span className='text-xs text-gray-400'>{change.answerOptions.length}</span>
          </div>
          {change.answerOptions.length === 0 ? (
            <div className='rounded-md border border-gray-700 bg-gray-950 p-3 text-sm text-gray-400'>
              No {status} answer options for this grouped market.
            </div>
          ) : (
            change.answerOptions.map((addition) => {
              const reason = reasonById[addition.id] || '';
              return (
                <article key={addition.id} className='grid gap-3 rounded-md border border-emerald-900/70 bg-emerald-950/20 p-4'>
                  <div className='flex flex-wrap items-center gap-2 text-sm text-emerald-50/90'>
                    <span className='rounded-full border border-emerald-700/70 bg-emerald-900/50 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-emerald-100'>
                      {addition.status}
                    </span>
                    <span>
                      Proposed by{' '}
                      <Link to={`/user/${addition.proposedBy}`} className='text-emerald-200 hover:text-emerald-100'>
                        @{addition.proposedBy}
                      </Link>
                    </span>
                    <span>{addition.createdAt ? new Date(addition.createdAt).toLocaleString() : ''}</span>
                  </div>
                  <div>
                    <p className='text-xs font-semibold uppercase tracking-[0.14em] text-emerald-200'>Answer Option</p>
                    <p className='mt-1 text-xl font-semibold text-white'>{addition.answerLabel}</p>
                    <p className='mt-2 text-sm text-emerald-100/80'>Add-answer cost: {addition.additionCost} credits</p>
                  </div>
                  {addition.status === 'rejected' && addition.rejectionReason && (
                    <div className='rounded-md border border-rose-800/70 bg-rose-950/30 p-3 text-sm text-rose-100'>
                      Rejection reason: {addition.rejectionReason}
                    </div>
                  )}
                  {addition.status === 'approved' && (
                    <div className='rounded-md border border-emerald-800/70 bg-emerald-950/30 p-3 text-sm text-emerald-100'>
                      Approved by @{addition.reviewedBy || 'reviewer'}{addition.reviewedAt ? ` at ${new Date(addition.reviewedAt).toLocaleString()}` : ''}.
                    </div>
                  )}
                  {canReviewAnswerOptions && (
                    <div className='grid gap-3 md:grid-cols-[minmax(0,1fr),auto,auto] md:items-start'>
                      <textarea
                        value={reason}
                        onChange={(event) => onReasonChange(addition.id, event.target.value)}
                        rows={3}
                        placeholder='Decision reason required for rejection'
                        className='rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
                      />
                      <button
                        type='button'
                        disabled={busyAdditionId === addition.id}
                        onClick={() => onReviewAddition(addition, 'approved')}
                        className='rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50'
                      >
                        Approve Answer
                      </button>
                      <button
                        type='button'
                        disabled={busyAdditionId === addition.id || !reason.trim()}
                        onClick={() => onReviewAddition(addition, 'rejected')}
                        className='rounded-md bg-rose-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50'
                      >
                        Reject
                      </button>
                    </div>
                  )}
                </article>
              );
            })
          )}
        </section>
      )}
    </article>
  );
};

const ProfileDescriptionAmendmentCard = ({ amendment, status }) => {
  const previousAmendments = Array.isArray(amendment.previousApprovedAmendments)
    ? amendment.previousApprovedAmendments
    : [];
  const childAmendments = Array.isArray(amendment.childAmendments) ? amendment.childAmendments : [amendment];

  return (
    <article className='grid gap-4 rounded-md border border-sky-900/70 bg-sky-950/20 p-4'>
      <div className='flex flex-wrap items-center gap-2 text-sm text-gray-300'>
        <span className='rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200'>
          Amendment {amendmentNumber(amendment.version)}
        </span>
        <span>Submitted by @{amendment.createdBy}</span>
        <span>{amendment.createdAt ? new Date(amendment.createdAt).toLocaleString() : ''}</span>
      </div>
      {isGroupedAmendment(amendment) && (
        <div className='flex flex-wrap gap-2'>
          {childAmendments.map((child) => (
            <span key={child.id} className='rounded-full border border-sky-800/70 bg-sky-900/40 px-2.5 py-1 text-xs text-sky-100'>
              {child.marketGroup?.answerLabel || `Market #${child.marketId}`} · Amendment {amendmentNumber(child.version)}
            </span>
          ))}
        </div>
      )}
      <div className='grid gap-3 rounded-md border border-gray-700 bg-gray-950 p-4'>
        <div>
          <p className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400'>Description</p>
          <p className='whitespace-pre-wrap text-sm leading-6 text-gray-200'>
            {amendment.marketDescription || 'No market description was returned.'}
          </p>
        </div>
        {previousAmendments.length > 0 && (
          <div className='grid gap-2'>
            {previousAmendments.map((previous) => (
              <article key={previous.id || previous.version} className='rounded-md border border-gray-700 bg-gray-900 p-3'>
                <div className='mb-2 flex flex-wrap gap-2 text-xs text-gray-400'>
                  <span>Amendment {amendmentNumber(previous.version)}</span>
                  <span>Approved by @{previous.approvedBy || 'admin'}</span>
                  {previous.approvedAt && <span>{new Date(previous.approvedAt).toLocaleString()}</span>}
                </div>
                <MarkdownLite>{previous.body}</MarkdownLite>
              </article>
            ))}
          </div>
        )}
      </div>
      {amendment.submitReason && (
        <div className='rounded-md border border-gray-700 bg-gray-950 p-3 text-sm text-gray-300'>
          <span className='font-semibold text-gray-100'>Submit reason:</span> {amendment.submitReason}
        </div>
      )}
      <div className='rounded-md border border-sky-900/70 bg-sky-950/30 p-4'>
        <p className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-sky-200'>
          Proposed Amendment {amendmentNumber(amendment.version)}
        </p>
        <MarkdownLite>{amendment.body}</MarkdownLite>
      </div>
      {status === 'rejected' && amendment.rejectionReason && (
        <div className='rounded-md border border-rose-800/70 bg-rose-950/30 p-3 text-sm text-rose-100'>
          Rejection reason: {amendment.rejectionReason}
        </div>
      )}
    </article>
  );
};

const ProfileMarketLifecycleTab = ({ status }) => {
  const { token } = useAuth();
  const [markets, setMarkets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);

  useEffect(() => {
    setPage(0);
  }, [status, searchQuery]);

  useEffect(() => {
    let ignore = false;

    const loadMarkets = async () => {
      setLoading(true);
      setError('');
      try {
        const rows = [];
        let offset = 0;
        let keepFetching = true;
        while (keepFetching) {
          const data = await listMyLifecycleMarkets({
            token,
            status,
            query: searchQuery,
            limit: PROFILE_FETCH_BATCH_SIZE,
            offset,
          });
          const batch = data.markets || [];
          rows.push(...batch);
          keepFetching = batch.length === PROFILE_FETCH_BATCH_SIZE;
          offset += PROFILE_FETCH_BATCH_SIZE;
        }
        if (!ignore) {
          setMarkets(groupLifecycleMarketRows(rows));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load market queue.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    const timeoutId = window.setTimeout(loadMarkets, 300);
    return () => {
      ignore = true;
      window.clearTimeout(timeoutId);
    };
  }, [status, token, searchQuery]);

  const pageInfo = pagedRows(markets, page);

  return (
    <div className='grid gap-4'>
      {error && <ErrorMessage message={error} />}
      <div className='grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
        <label htmlFor={`profile-market-search-${status}`} className='text-xs font-mono uppercase tracking-[0.16em] text-gray-400'>
          Search markets
        </label>
        <div className='relative'>
          <input
            id={`profile-market-search-${status}`}
            type='search'
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            placeholder={`Search ${status} markets by title or description`}
            className='w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 pr-10 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
          />
          {loading && (
            <div className='absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin rounded-full border-b-2 border-primary-pink' />
          )}
        </div>
      </div>
      {loading ? (
        <LoadingSpinner />
      ) : (
        <>
          <ProfilePaginationControls
            label={`${lifecycleShortLabelByStatus[status] || status} markets`}
            pageInfo={pageInfo}
            onPageChange={setPage}
          />
          <MarketLifecycleTable
            markets={pageInfo.rows}
            emptyMessage={`No ${status} markets found.`}
          />
        </>
      )}
    </div>
  );
};

const Profile = () => {
  const { username } = useAuth();
  const location = useLocation();
  const { userData, userLoading, userError } = useUserData(username, true);

  const defaultTab = useMemo(() => {
    const params = new URLSearchParams(location.search);
    return params.get('tab') || 'Proposed Markets';
  }, [location.search]);

  if (userLoading) {
    return <LoadingSpinner />;
  }

  if (userError) {
    return <ErrorMessage message={`Error loading user data: ${userError}`} />;
  }

  const proposedMarket = location.state?.proposedMarket;
  const marketCreationCost = location.state?.marketCreationCost;
  const isActiveModerator =
    String(userData?.usertype || '').toUpperCase() === 'MODERATOR' &&
    String(userData?.moderatorStatus || '').toLowerCase() === 'active';
  const defaultMarketStatus = statusFromLegacyTab(defaultTab, lifecycleTabByStatus) || 'proposed';
  const accountDefaultTab = accountTabLabels.includes(defaultTab) ? defaultTab : 'User Info';
  const defaultChangeStatus = statusFromLegacyTab(defaultTab, answerOptionTabByStatus) ||
    statusFromLegacyTab(defaultTab, amendmentTabByStatus) ||
    'pending';
  const resolvedDefaultTab = (() => {
    if (!isActiveModerator) {
      return 'Account';
    }
    if (statusFromLegacyTab(defaultTab, lifecycleTabByStatus)) {
      return 'Markets';
    }
    if (
      statusFromLegacyTab(defaultTab, amendmentTabByStatus) ||
      statusFromLegacyTab(defaultTab, answerOptionTabByStatus)
    ) {
      return 'Market Changes';
    }
    if (accountTabLabels.includes(defaultTab)) {
      return 'Account';
    }
    return 'Markets';
  })();

  const profileTabs = [
    {
      label: 'Account',
      content: (
        <ProfileAccountTabs
          username={username}
          userData={userData}
          defaultTab={accountDefaultTab}
        />
      ),
    },
    ...(isActiveModerator ? [
      {
        label: 'Markets',
        content: <ProfileMarketsTabs defaultStatus={defaultMarketStatus} />,
      },
      {
        label: 'Market Changes',
        content: (
          <ProfileMarketChangesTabs
            defaultMarketType='grouped'
            defaultStatus={defaultChangeStatus}
          />
        ),
      },
    ] : []),
  ];

  return (
    <div className='min-h-screen bg-primary-background text-white'>
      <div className='container mx-auto px-4 py-6'>
        <div className='mb-6'>
          <p className='text-xs uppercase tracking-[0.22em] text-primary-pink'>Private Profile</p>
          <h1 className='mt-2 text-3xl font-bold'>Profile Details</h1>
          <p className='mt-2 text-gray-400'>Manage your account, markets, change queues, portfolio, and financial history.</p>
        </div>

        {isActiveModerator && proposedMarket && (
          <div className='mb-6 rounded-lg border border-amber-500 bg-amber-950/40 p-4 text-amber-50'>
            <p className='text-sm uppercase tracking-[0.18em] text-amber-300'>Proposed market created</p>
            <h2 className='mt-2 text-xl font-semibold'>{proposedMarket.questionTitle}</h2>
            <p className='mt-2 text-sm'>
              Market ID <span className='font-mono'>{proposedMarket.id}</span> is awaiting admin review.
              {marketCreationCost !== undefined && marketCreationCost !== null && (
                <> The proposal cost was <span className='font-semibold'>{marketCreationCost}</span> credits.</>
              )}
            </p>
          </div>
        )}

        <SiteTabs tabs={profileTabs} defaultTab={resolvedDefaultTab} />
      </div>
    </div>
  );
};

export default Profile;
