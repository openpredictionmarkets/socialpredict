import { authenticatedApiRequest } from './httpClient';

const marketReviewReasonMessages = {
  AUTHORIZATION_DENIED: 'Only admins can review proposed markets.',
  INVALID_REQUEST: 'Check the market ID and request fields.',
  INVALID_STATE: 'This market is not in a state that can be reviewed.',
  MARKET_NOT_FOUND: 'No market was found for that ID.',
  USER_NOT_FOUND: 'No active moderator was found for that steward username.',
  VALIDATION_FAILED: 'Check the review fields and try again.',
};

const reviewMarket = async ({ marketId, token, action, body }) => {
  const normalizedMarketId = String(marketId || '').trim();
  if (!normalizedMarketId) {
    throw new Error('Market ID is required.');
  }
  if (!token) {
    throw new Error('Admin authentication token is missing. Please log in again.');
  }

  return authenticatedApiRequest(`/v0/admin/markets/${normalizedMarketId}/${action}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
    authToken: token,
    reasonMessages: marketReviewReasonMessages,
    fallbackMessage: 'Market review failed. Please try again.',
  });
};

export const approveProposedMarket = ({ marketId, token }) => reviewMarket({
  marketId,
  token,
  action: 'approve',
  body: { confirm: true },
});

export const rejectProposedMarket = ({ marketId, token, reason }) => {
  const normalizedReason = String(reason || '').trim();
  if (!normalizedReason) {
    throw new Error('A rejection reason is required.');
  }

  return reviewMarket({
    marketId,
    token,
    action: 'reject',
    body: { reason: normalizedReason },
  });
};

export const reassignMarketSteward = ({ marketId, token, stewardUsername, reason }) => {
  const normalizedStewardUsername = String(stewardUsername || '').trim();
  const normalizedReason = String(reason || '').trim();
  if (!normalizedStewardUsername) {
    throw new Error('A steward username is required.');
  }
  if (!normalizedReason) {
    throw new Error('A reassignment reason is required.');
  }

  return reviewMarket({
    marketId,
    token,
    action: 'steward',
    body: {
      stewardUsername: normalizedStewardUsername,
      reason: normalizedReason,
    },
  });
};
