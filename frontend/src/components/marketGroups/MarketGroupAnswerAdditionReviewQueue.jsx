import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import LoadingSpinner from '../loaders/LoadingSpinner';
import {
  listMarketGroupAnswerAdditionsForReview,
  reviewMarketGroupAnswerAdditionForSteward,
} from '../../api/marketsApi';

const statusLabel = (status) => String(status || 'pending').toLowerCase();
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

export default function MarketGroupAnswerAdditionReviewQueue({
  token,
  groupId = '',
  status = 'pending',
  emptyMessage = '',
  onReviewed,
}) {
  const [additions, setAdditions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [busyAdditionId, setBusyAdditionId] = useState(null);
  const [reasonById, setReasonById] = useState({});
  const normalizedStatus = statusLabel(status);
  const canReview = normalizedStatus === 'pending';

  const loadAdditions = async () => {
    if (!token) {
      setAdditions([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError('');
    try {
      const data = await withRateLimitRetry(() => listMarketGroupAnswerAdditionsForReview({
        groupId,
        token,
        status: normalizedStatus,
        limit: 100,
      }));
      setAdditions(data.additions || []);
    } catch (err) {
      setError(err.message || 'Unable to load grouped answer options.');
      setAdditions([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAdditions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId, normalizedStatus, token]);

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
      setSuccessMessage(`Answer option "${addition.answerLabel}" ${nextStatus}.`);
      await loadAdditions();
      if (onReviewed) {
        onReviewed();
      }
    } catch (err) {
      setError(err.message || 'Unable to review grouped answer option.');
    } finally {
      setBusyAdditionId(null);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className='grid gap-4'>
      {error && (
        <div className='rounded-md bg-red-700 p-3 text-sm text-white'>
          {error}
        </div>
      )}
      {successMessage && (
        <div className='rounded-md bg-emerald-700 p-3 text-sm text-white'>
          {successMessage}
        </div>
      )}
      {additions.length === 0 && (
        <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300'>
          {emptyMessage || `No ${normalizedStatus} answer options found.`}
        </div>
      )}
      {additions.map((addition) => {
        const group = addition.marketGroup || {};
        const marketHref = addition.marketId
          ? `/markets/${addition.marketId}`
          : (group.id ? `/markets/group/${group.id}` : '#');
        const reason = reasonById[addition.id] || '';
        return (
          <article key={addition.id} className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
            <div className='flex flex-wrap items-center gap-2 text-sm text-gray-300'>
              <span className='rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100'>
                Group #{addition.groupId}
              </span>
              <span className='rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200'>
                {addition.status}
              </span>
              <span>
                Proposed by{' '}
                <Link to={`/user/${addition.proposedBy}`} className='text-sky-300 hover:text-sky-200'>
                  @{addition.proposedBy}
                </Link>
              </span>
              <span>{addition.createdAt ? new Date(addition.createdAt).toLocaleString() : ''}</span>
            </div>
            <div className='grid gap-2'>
              <Link
                to={marketHref}
                className='text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200'
              >
                {group.questionTitle || addition.groupTitle || `Grouped market #${addition.groupId}`}
              </Link>
              <div className='rounded-md border border-sky-900/70 bg-sky-950/20 p-4'>
                <p className='text-xs font-semibold uppercase tracking-[0.14em] text-sky-200'>Answer Option</p>
                <p className='mt-1 text-xl font-semibold text-white'>{addition.answerLabel}</p>
                <p className='mt-2 text-sm text-sky-100/80'>Add-answer cost: {addition.additionCost} credits</p>
              </div>
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
            {canReview && (
              <div className='grid gap-3 md:grid-cols-[minmax(0,1fr),auto,auto] md:items-start'>
                <textarea
                  value={reason}
                  onChange={(event) => updateReason(addition.id, event.target.value)}
                  rows={3}
                  placeholder='Decision reason required for rejection'
                  className='rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40'
                />
                <button
                  type='button'
                  disabled={busyAdditionId === addition.id}
                  onClick={() => reviewAddition(addition, 'approved')}
                  className='rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50'
                >
                  Approve Answer
                </button>
                <button
                  type='button'
                  disabled={busyAdditionId === addition.id || !reason.trim()}
                  onClick={() => reviewAddition(addition, 'rejected')}
                  className='rounded-md bg-rose-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50'
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
}
