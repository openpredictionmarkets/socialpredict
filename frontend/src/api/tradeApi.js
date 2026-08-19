import { apiRequest, authenticatedApiRequest } from './httpClient';

export const NO_SELLABLE_SHARES_MESSAGE = [
  'No sellable shares yet.',
  'Initial value cannot be sold until a follow-up order from another user has been placed.',
  'Wait for another order from another user, then try selling again.',
].join(' ');

const jsonHeaders = {
  'Content-Type': 'application/json',
};

const validateOrder = ({ marketId, amount, outcome }, actionLabel) => {
  if (!marketId || !amount || !outcome) {
    throw new Error(`Missing required ${actionLabel} data (marketId, amount, outcome)`);
  }
  if (Number(amount) < 1) {
    throw new Error(`${actionLabel} amount must be at least 1`);
  }
  if (outcome !== 'YES' && outcome !== 'NO') {
    throw new Error(`${actionLabel} outcome must be YES or NO`);
  }
};

export const placeBet = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to place a bet.');
  }
  validateOrder({ marketId, amount, outcome }, 'Bet');

  return authenticatedApiRequest('/v0/bet', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Bet failed. Please try again.',
  });
};

export const fetchUserPosition = ({ token, marketId }) => {
  if (!token) {
    throw new Error('Please log in again to view your Position.');
  }
  if (!marketId) {
    throw new Error('Market ID is required.');
  }

  return authenticatedApiRequest(`/v0/userposition/${marketId}`, {
    authToken: token,
    reasonMessages: {
      INVALID_TOKEN: 'Please log in again to view your Position.',
      AUTHORIZATION_DENIED: 'Please log in again to view your Position.',
      PASSWORD_CHANGE_REQUIRED: 'Please update your password before viewing your Position.',
      NO_POSITION: NO_SELLABLE_SHARES_MESSAGE,
    },
    fallbackMessage: NO_SELLABLE_SHARES_MESSAGE,
  });
};

export const quoteSale = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to sell shares.');
  }
  validateOrder({ marketId, amount, outcome }, 'Sale');

  return authenticatedApiRequest('/v0/sell/quote', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Sale quote failed. Please try again.',
  });
};

export const sellShares = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to sell shares.');
  }
  validateOrder({ marketId, amount, outcome }, 'Sale');

  return authenticatedApiRequest('/v0/sell', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Sale failed. Please try again.',
  });
};

export const fetchTradingFees = async ({ token } = {}) => {
  const setup = await apiRequest('/v0/setup', {
    authenticated: Boolean(token),
    authToken: token,
    fallbackMessage: 'Failed to load trading fees.',
  });
  return setup?.betting?.betFees || null;
};
