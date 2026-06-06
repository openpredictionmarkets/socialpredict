import { apiRequest, authenticatedApiRequest } from './httpClient';

const tagReasonMessages = {
  AUTHORIZATION_DENIED: 'Only admins can manage market tags.',
  INVALID_REQUEST: 'Check the tag fields and try again.',
  VALIDATION_FAILED: 'Check the tag fields and try again.',
};

export const listMarketTags = () => apiRequest('/v0/market-tags', {
  fallbackMessage: 'Unable to load market tags.',
});

export const listAdminMarketTags = ({ token, includeInactive = true }) => authenticatedApiRequest(
  `/v0/admin/market-tags?includeInactive=${includeInactive ? 'true' : 'false'}`,
  {
    authToken: token,
    fallbackMessage: 'Unable to load market tags.',
    reasonMessages: tagReasonMessages,
  },
);

export const createAdminMarketTag = ({ token, tag }) => authenticatedApiRequest('/v0/admin/market-tags', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(tag),
  authToken: token,
  fallbackMessage: 'Unable to create market tag.',
  reasonMessages: tagReasonMessages,
});

export const updateAdminMarketTag = ({ token, slug, tag }) => authenticatedApiRequest(`/v0/admin/market-tags/${slug}`, {
  method: 'PATCH',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(tag),
  authToken: token,
  fallbackMessage: 'Unable to update market tag.',
  reasonMessages: tagReasonMessages,
});
