import { authenticatedApiRequest } from './httpClient';

const adminUserReasonMessages = {
  AUTHORIZATION_DENIED: 'Only admins can manage users.',
  INVALID_REQUEST: 'Check the request and try again.',
  INVALID_STATE: 'That moderator transition is not valid for the selected user.',
  USER_NOT_FOUND: 'No user was found for that username.',
  VALIDATION_FAILED: 'Check the user fields and try again.',
  RATE_LIMITED: 'Too many admin requests. Wait and try again.',
};

const adminRequest = async ({ path, token, method = 'GET', body }) => {
  if (!token) {
    throw new Error('Admin authentication token is missing. Please log in again.');
  }

  return authenticatedApiRequest(path, {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
    body: body ? JSON.stringify(body) : undefined,
    authToken: token,
    reasonMessages: adminUserReasonMessages,
    fallbackMessage: 'Admin user request failed. Please try again.',
  });
};

export const listAdminUsers = ({
  token,
  limit = 100,
  offset = 0,
  usertype = '',
  query = '',
}) => {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  if (usertype) {
    params.set('usertype', usertype);
  }
  if (query) {
    params.set('query', query);
  }
  return adminRequest({
    path: `/v0/admin/users?${params.toString()}`,
    token,
  });
};

export const promoteUserToModerator = ({ token, username, reason }) => adminRequest({
  path: `/v0/admin/users/${encodeURIComponent(username)}/role`,
  token,
  method: 'PATCH',
  body: {
    usertype: 'MODERATOR',
    reason: reason || 'approved for moderator mode',
  },
});

export const updateModeratorSuspension = ({ token, username, suspended, reason }) => adminRequest({
  path: `/v0/admin/moderators/${encodeURIComponent(username)}/suspension`,
  token,
  method: 'PATCH',
  body: {
    suspended,
    reason,
  },
});
