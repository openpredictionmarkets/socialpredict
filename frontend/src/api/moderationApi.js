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

const reviewMarketGroup = async ({ groupId, token, action, body }) => {
  const normalizedGroupId = String(groupId || '').trim();
  if (!normalizedGroupId) {
    throw new Error('Market group ID is required.');
  }
  if (!token) {
    throw new Error('Admin authentication token is missing. Please log in again.');
  }

  return authenticatedApiRequest(`/v0/admin/market-groups/${normalizedGroupId}/${action}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
    authToken: token,
    reasonMessages: marketReviewReasonMessages,
    fallbackMessage: 'Market group review failed. Please try again.',
  });
};

export const approveProposedMarket = ({ marketId, token }) => reviewMarket({
  marketId,
  token,
  action: 'approve',
  body: { confirm: true },
});

export const approveProposedMarketGroup = ({ groupId, token }) => reviewMarketGroup({
  groupId,
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

export const rejectProposedMarketGroup = ({ groupId, token, reason }) => {
  const normalizedReason = String(reason || '').trim();
  if (!normalizedReason) {
    throw new Error('A rejection reason is required.');
  }

  return reviewMarketGroup({
    groupId,
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

export const reassignMarketGroupSteward = ({ groupId, token, stewardUsername, reason }) => {
  const normalizedStewardUsername = String(stewardUsername || '').trim();
  const normalizedReason = String(reason || '').trim();
  if (!normalizedStewardUsername) {
    throw new Error('A steward username is required.');
  }
  if (!normalizedReason) {
    throw new Error('A reassignment reason is required.');
  }

  return reviewMarketGroup({
    groupId,
    token,
    action: 'steward',
    body: {
      stewardUsername: normalizedStewardUsername,
      reason: normalizedReason,
    },
  });
};

export const updateMarketTags = ({ marketId, token, tagSlugs }) => {
  if (!Array.isArray(tagSlugs)) {
    throw new Error('Market tags must be provided as a list.');
  }

  return reviewMarket({
    marketId,
    token,
    action: 'tags',
    body: { tagSlugs },
    });
};

export const listAdminMarketGroupAnswerAdditions = ({ token, status = 'pending', groupId = '', limit = 100, offset = 0 }) => {
  const params = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  if (groupId) {
    params.set('groupId', String(groupId));
  }
  return authenticatedApiRequest(`/v0/admin/market-group-answer-additions?${params.toString()}`, {
    authToken: token,
    reasonMessages: marketReviewReasonMessages,
    fallbackMessage: 'Unable to load grouped answer additions.',
  });
};

export const reviewMarketGroupAnswerAddition = ({ additionId, token, status, reason = '', confirm = false }) => {
  const normalizedAdditionId = String(additionId || '').trim();
  const normalizedStatus = String(status || '').trim();
  if (!normalizedAdditionId || !normalizedStatus) {
    throw new Error('Answer addition ID and review status are required.');
  }
  return authenticatedApiRequest(`/v0/admin/market-group-answer-additions/${normalizedAdditionId}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    authToken: token,
    body: JSON.stringify({
      status: normalizedStatus,
      reason: String(reason || '').trim(),
      confirm: Boolean(confirm),
    }),
    reasonMessages: {
      ...marketReviewReasonMessages,
      INSUFFICIENT_BALANCE: 'The proposing moderator no longer has enough credit to add this answer.',
    },
    fallbackMessage: 'Unable to review grouped answer addition.',
  });
};
