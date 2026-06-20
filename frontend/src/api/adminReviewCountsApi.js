import { listAdminLifecycleMarkets } from './lifecycleMarketsApi';
import { listAdminMarketDescriptionAmendments } from './marketDescriptionAmendmentsApi';
import { listAdminMarketGroupAnswerAdditions } from './moderationApi';

export const emptyPendingAdminReviewCounts = {
  pendingMarkets: 0,
  pendingAmendments: 0,
  pendingAnswers: 0,
};

export const totalPendingReviewCount = (counts) => (
  Number(counts?.pendingMarkets || 0) +
  Number(counts?.pendingAmendments || 0) +
  Number(counts?.pendingAnswers || 0)
);

const countCacheTTL = 10000;
let cachedToken = '';
let cachedAt = 0;
let cachedCounts = emptyPendingAdminReviewCounts;
let inflightCountsRequest = null;

export const clearPendingAdminReviewCountsCache = () => {
  cachedToken = '';
  cachedAt = 0;
  cachedCounts = emptyPendingAdminReviewCounts;
  inflightCountsRequest = null;
};

export const getPendingAdminReviewCounts = async ({ token, force = false }) => {
  if (!token) {
    return emptyPendingAdminReviewCounts;
  }

  const now = Date.now();
  if (!force && cachedToken === token && now - cachedAt < countCacheTTL) {
    return cachedCounts;
  }

  if (!force && inflightCountsRequest) {
    return inflightCountsRequest;
  }

  inflightCountsRequest = Promise.all([
    listAdminLifecycleMarkets({ token, status: 'proposed', limit: 1, offset: 0 }),
    listAdminMarketDescriptionAmendments({ token, status: 'pending', limit: 1, offset: 0 }),
    listAdminMarketGroupAnswerAdditions({ token, status: 'pending', limit: 1, offset: 0 }),
  ]).then(([marketsResult, amendmentsResult, answersResult]) => {
    const nextCounts = {
      pendingMarkets: Number(marketsResult.total ?? marketsResult.markets?.length ?? 0),
      pendingAmendments: Number(amendmentsResult.total ?? amendmentsResult.amendments?.length ?? 0),
      pendingAnswers: Number(answersResult.total ?? answersResult.additions?.length ?? 0),
    };
    cachedToken = token;
    cachedAt = Date.now();
    cachedCounts = nextCounts;
    return nextCounts;
  }).finally(() => {
    inflightCountsRequest = null;
  });

  return inflightCountsRequest;
};
