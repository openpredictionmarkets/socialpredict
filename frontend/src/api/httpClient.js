import { API_URL } from '../config';
import {
  getApiErrorMessage,
  parseApiResponseText,
  unwrapApiResponse,
} from '../utils/apiResponse';
import { authStorage } from './authStorage';

const buildUrl = (path) => {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  return `${API_URL}${path.startsWith('/') ? path : `/${path}`}`;
};

const mergeHeaders = (headers = {}, token) => {
  const merged = { ...headers };
  if (token) {
    merged.Authorization = `Bearer ${token}`;
  }
  return merged;
};

export const apiRequest = async (
  path,
  {
    headers,
    reasonMessages,
    fallbackMessage,
    unwrap = true,
    authenticated = false,
    ...options
  } = {},
) => {
  const token = authenticated ? authStorage.getToken() : null;
  const response = await fetch(buildUrl(path), {
    ...options,
    headers: mergeHeaders(headers, token),
  });
  const text = await response.text();
  const data = parseApiResponseText(text);

  if (!response.ok) {
    throw new Error(getApiErrorMessage(response, data, fallbackMessage, reasonMessages));
  }

  return unwrap ? unwrapApiResponse(data) : data;
};

export const authenticatedApiRequest = (path, options = {}) => apiRequest(path, {
  ...options,
  authenticated: true,
});
