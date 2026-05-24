import React, { useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import {
  approveProposedMarket,
  rejectProposedMarket,
} from '../../../api/moderationApi';

const DetailRow = ({ label, value }) => (
  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 border-b border-gray-700 py-2">
    <dt className="text-sm text-gray-400">{label}</dt>
    <dd className="font-mono text-sm text-white break-all">{value || 'n/a'}</dd>
  </div>
);

function ModeratorMarketReview() {
  const { token } = useAuth();
  const [marketId, setMarketId] = useState('');
  const [rejectionReason, setRejectionReason] = useState('');
  const [confirmApproval, setConfirmApproval] = useState(false);
  const [reviewResult, setReviewResult] = useState(null);
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const runReview = async (reviewAction) => {
    setError('');
    setReviewResult(null);
    setIsSubmitting(true);

    try {
      const result = reviewAction === 'approve'
        ? await approveProposedMarket({ marketId, token })
        : await rejectProposedMarket({ marketId, token, reason: rejectionReason });

      setReviewResult(result);
      if (reviewAction === 'reject') {
        setRejectionReason('');
      }
      setConfirmApproval(false);
    } catch (err) {
      setError(err.message || 'Market review failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section className="p-6 bg-primary-background shadow-md rounded-lg text-white">
      <div className="mb-6">
        <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">
          Moderator mode smoke test
        </p>
        <h1 className="text-2xl font-bold mt-2">Review Proposed Market</h1>
        <p className="text-sm text-gray-300 mt-2 max-w-3xl">
          Use this temporary admin surface to approve or reject a known proposed
          market ID. It intentionally does not replace the future proposal queue.
        </p>
      </div>

      <div className="grid gap-5">
        <label className="grid gap-2">
          <span className="text-sm font-medium text-gray-300">Market ID</span>
          <input
            type="number"
            min="1"
            value={marketId}
            onChange={(event) => setMarketId(event.target.value)}
            placeholder="Enter proposed market ID"
            className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
        </label>

        <label className="flex items-start gap-3 rounded-md border border-gray-700 bg-gray-800/70 p-3">
          <input
            type="checkbox"
            checked={confirmApproval}
            onChange={(event) => setConfirmApproval(event.target.checked)}
            className="mt-1"
          />
          <span className="text-sm text-gray-300">
            Confirm that this proposal should be published and become tradable.
          </span>
        </label>

        <div className="flex flex-col sm:flex-row gap-3">
          <button
            type="button"
            disabled={isSubmitting || !confirmApproval}
            onClick={() => runReview('approve')}
            className="rounded-md bg-emerald-600 px-4 py-2 font-semibold text-white transition hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Approve Market
          </button>
        </div>

        <label className="grid gap-2">
          <span className="text-sm font-medium text-gray-300">Rejection Reason</span>
          <textarea
            value={rejectionReason}
            onChange={(event) => setRejectionReason(event.target.value)}
            placeholder="Briefly explain why the proposal is being rejected"
            rows={4}
            className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
        </label>

        <div className="flex flex-col sm:flex-row gap-3">
          <button
            type="button"
            disabled={isSubmitting || !rejectionReason.trim()}
            onClick={() => runReview('reject')}
            className="rounded-md bg-rose-700 px-4 py-2 font-semibold text-white transition hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Reject Market
          </button>
        </div>
      </div>

      {error && (
        <div className="mt-5 rounded-md bg-red-700 p-3 text-sm text-white">
          {error}
        </div>
      )}

      {reviewResult && (
        <div className="mt-6 rounded-md border border-gray-700 bg-gray-900 p-4">
          <h2 className="text-lg font-semibold text-white">Review Result</h2>
          <dl className="mt-3">
            <DetailRow label="Market ID" value={reviewResult.id} />
            <DetailRow label="Public status" value={reviewResult.status} />
            <DetailRow label="Lifecycle status" value={reviewResult.lifecycleStatus} />
            <DetailRow label="Approved by" value={reviewResult.approvedBy} />
            <DetailRow label="Rejected by" value={reviewResult.rejectedBy} />
            <DetailRow label="Rejection reason" value={reviewResult.rejectionReason} />
          </dl>
        </div>
      )}
    </section>
  );
}

export default ModeratorMarketReview;
