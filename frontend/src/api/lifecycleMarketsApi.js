import { authenticatedApiRequest } from './httpClient';

const lifecycleReasonMessages = {
  AUTHORIZATION_DENIED: 'You are not allowed to view this market queue.',
  INVALID_REQUEST: 'Check the market queue filter and try again.',
  RATE_LIMITED: 'Too many requests. Wait and try again.',
};

const lifecycleRequest = async ({ path, token }) => {
  if (!token) {
    throw new Error('Authentication token is missing. Please log in again.');
  }

  return authenticatedApiRequest(path, {
    headers: {
      'Content-Type': 'application/json',
    },
    authToken: token,
    reasonMessages: lifecycleReasonMessages,
    fallbackMessage: 'Market queue request failed. Please try again.',
  });
};

const buildLifecycleQuery = ({ status, query, limit = 50, offset = 0 }) => {
  const params = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  if (String(query || '').trim()) {
    params.set('query', String(query).trim());
  }
  return params.toString();
};

export const listMyLifecycleMarkets = ({ token, status, query, limit, offset }) => lifecycleRequest({
  path: `/v0/profile/markets?${buildLifecycleQuery({ status, query, limit, offset })}`,
  token,
});

export const listAdminLifecycleMarkets = ({ token, status, query, limit, offset }) => lifecycleRequest({
  path: `/v0/admin/markets?${buildLifecycleQuery({ status, query, limit, offset })}`,
  token,
});
