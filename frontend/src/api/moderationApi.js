import { API_URL } from '../config';
import {
  getApiErrorMessage,
  parseApiResponseText,
  unwrapApiResponse,
} from '../utils/apiResponse';

const marketReviewReasonMessages = {
  AUTHORIZATION_DENIED: 'Only admins can review proposed markets.',
  INVALID_REQUEST: 'Check the market ID and request fields.',
  INVALID_STATE: 'This market is not in a state that can be reviewed.',
  MARKET_NOT_FOUND: 'No market was found for that ID.',
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

  const response = await fetch(`${API_URL}/v0/admin/markets/${normalizedMarketId}/${action}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });

  const text = await response.text();
  const payload = parseApiResponseText(text);

  if (!response.ok) {
    throw new Error(getApiErrorMessage(
      response,
      payload,
      `Market review failed with status ${response.status}.`,
      marketReviewReasonMessages,
    ));
  }

  return unwrapApiResponse(payload);
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
