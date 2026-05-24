import { API_URL } from '../config';
import {
  getApiErrorMessage,
  parseApiResponseText,
  unwrapApiResponse,
} from '../utils/apiResponse';

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

  const response = await fetch(`${API_URL}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await response.text();
  const payload = parseApiResponseText(text);

  if (!response.ok) {
    throw new Error(getApiErrorMessage(
      response,
      payload,
      `Admin user request failed with status ${response.status}.`,
      adminUserReasonMessages,
    ));
  }

  return unwrapApiResponse(payload);
};

export const listAdminUsers = ({ token, limit = 100, offset = 0 }) => {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
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
