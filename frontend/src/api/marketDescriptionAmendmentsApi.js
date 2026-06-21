import { authenticatedApiRequest } from './httpClient';

const amendmentReasonMessages = {
  AUTHORIZATION_DENIED: 'Only the active market steward can propose description amendments.',
  INVALID_STATE: 'This market is not in a state that accepts description amendments.',
  MARKET_NOT_FOUND: 'No market was found for that ID.',
  USER_NOT_FOUND: 'The steward user could not be verified.',
  VALIDATION_FAILED: 'Check the amendment text and markdown-lite formatting.',
  PASSWORD_CHANGE_REQUIRED: 'Change your password before proposing or reviewing amendments.',
  RATE_LIMITED: 'Too many amendment requests. Wait and try again.',
};

const adminAmendmentReasonMessages = {
  ...amendmentReasonMessages,
  AUTHORIZATION_DENIED: 'Only admins can review description amendments.',
};

export const proposeMarketDescriptionAmendment = ({ token, marketId, body, submitReason = '' }) => {
  const normalizedBody = String(body || '').trim();
  if (!normalizedBody) {
    throw new Error('Amendment text is required.');
  }
  return authenticatedApiRequest(`/v0/markets/${marketId}/description-amendments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    authToken: token,
    body: JSON.stringify({
      body: normalizedBody,
      bodyFormat: 'markdown_lite',
      submitReason: String(submitReason || '').trim(),
    }),
    reasonMessages: amendmentReasonMessages,
    fallbackMessage: 'Description amendment request failed. Please try again.',
  });
};

export const listAdminMarketDescriptionAmendments = ({ token, status = 'pending', marketId = '', query = '', limit = 100, offset = 0 }) => {
  const params = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  if (marketId) {
    params.set('marketId', String(marketId));
  }
  if (String(query || '').trim()) {
    params.set('query', String(query).trim());
  }
  return authenticatedApiRequest(`/v0/admin/market-description-amendments?${params.toString()}`, {
    authToken: token,
    reasonMessages: adminAmendmentReasonMessages,
    fallbackMessage: 'Unable to load description amendments.',
  });
};

export const listMyMarketDescriptionAmendments = ({ token, status = 'pending', marketId = '', limit = 100, offset = 0 }) => {
  const params = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  if (marketId) {
    params.set('marketId', String(marketId));
  }
  return authenticatedApiRequest(`/v0/profile/market-description-amendments?${params.toString()}`, {
    authToken: token,
    reasonMessages: amendmentReasonMessages,
    fallbackMessage: 'Unable to load your description amendments.',
  });
};

export const getMarketGovernanceSettings = ({ token }) => authenticatedApiRequest('/v0/admin/market-description-amendments/settings', {
  authToken: token,
  reasonMessages: adminAmendmentReasonMessages,
  fallbackMessage: 'Unable to load market governance settings.',
});

export const updateMarketGovernanceSettings = ({
  token,
  autoApproveDescriptionAmendments,
  autoApproveMarketProposals,
  autoApproveMarketGroupAnswers,
  marketGroupAnswerAdditionApprovalPolicy,
  version = 0,
}) => authenticatedApiRequest('/v0/admin/market-description-amendments/settings', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
  },
  authToken: token,
  body: JSON.stringify({
    autoApproveDescriptionAmendments: Boolean(autoApproveDescriptionAmendments),
    autoApproveMarketProposals: Boolean(autoApproveMarketProposals),
    autoApproveMarketGroupAnswers: Boolean(autoApproveMarketGroupAnswers),
    marketGroupAnswerAdditionApprovalPolicy,
    version,
  }),
  reasonMessages: adminAmendmentReasonMessages,
  fallbackMessage: 'Unable to save market governance settings.',
});

export const reviewMarketDescriptionAmendment = ({ token, amendmentId, status, reason }) => {
  const normalizedStatus = String(status || '').trim();
  const normalizedReason = String(reason || '').trim();
  if (!normalizedStatus || !normalizedReason) {
    throw new Error('Review status and reason are required.');
  }
  return authenticatedApiRequest(`/v0/admin/market-description-amendments/${amendmentId}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    authToken: token,
    body: JSON.stringify({ status: normalizedStatus, reason: normalizedReason }),
    reasonMessages: adminAmendmentReasonMessages,
    fallbackMessage: 'Unable to review description amendment.',
  });
};

export const reviewGroupedMarketDescriptionAmendments = ({ token, amendmentIds, status, reason }) => {
  const normalizedStatus = String(status || '').trim();
  const normalizedReason = String(reason || '').trim();
  const normalizedIds = Array.from(new Set((amendmentIds || [])
    .map((id) => Number(id))
    .filter((id) => Number.isInteger(id) && id > 0)));
  if (!normalizedStatus || !normalizedReason || normalizedIds.length === 0) {
    throw new Error('Grouped review status, reason, and amendment IDs are required.');
  }
  return authenticatedApiRequest('/v0/admin/market-description-amendments/grouped-review', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    authToken: token,
    body: JSON.stringify({ amendmentIds: normalizedIds, status: normalizedStatus, reason: normalizedReason }),
    reasonMessages: adminAmendmentReasonMessages,
    fallbackMessage: 'Unable to review grouped description amendment.',
  });
};
