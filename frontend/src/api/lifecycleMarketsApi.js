import { API_URL } from '../config';
import {
  getApiErrorMessage,
  parseApiResponseText,
  unwrapApiResponse,
} from '../utils/apiResponse';

const lifecycleReasonMessages = {
  AUTHORIZATION_DENIED: 'You are not allowed to view this market queue.',
  INVALID_REQUEST: 'Check the market queue filter and try again.',
  RATE_LIMITED: 'Too many requests. Wait and try again.',
};

const lifecycleRequest = async ({ path, token }) => {
  if (!token) {
    throw new Error('Authentication token is missing. Please log in again.');
  }

  const response = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  });

  const text = await response.text();
  const payload = parseApiResponseText(text);

  if (!response.ok) {
    throw new Error(getApiErrorMessage(
      response,
      payload,
      `Market queue request failed with status ${response.status}.`,
      lifecycleReasonMessages,
    ));
  }

  return unwrapApiResponse(payload);
};

const buildLifecycleQuery = ({ status, limit = 50, offset = 0 }) => {
  const params = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  return params.toString();
};

export const listMyLifecycleMarkets = ({ token, status, limit, offset }) => lifecycleRequest({
  path: `/v0/profile/markets?${buildLifecycleQuery({ status, limit, offset })}`,
  token,
});

export const listAdminLifecycleMarkets = ({ token, status, limit, offset }) => lifecycleRequest({
  path: `/v0/admin/markets?${buildLifecycleQuery({ status, limit, offset })}`,
  token,
});
