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

const inflightGetRequests = new Map();

const requestDedupeKey = ({ url, token, headers }) => {
  const headerEntries = Object.entries(headers || {})
    .filter(([key]) => key.toLowerCase() !== 'authorization')
    .sort(([left], [right]) => left.localeCompare(right));
  return JSON.stringify({
    url,
    token: token || '',
    headers: headerEntries,
  });
};

export const apiRequest = async (
  path,
  {
    headers,
    reasonMessages,
    fallbackMessage,
    unwrap = true,
    authenticated = false,
    authToken,
    dedupe = true,
    ...options
  } = {},
) => {
  const token = authenticated ? (authToken || authStorage.getToken()) : null;
  const method = (options.method || 'GET').toUpperCase();
  const url = buildUrl(path);
  const mergedHeaders = mergeHeaders(headers, token);

  const runRequest = async () => {
    const response = await fetch(url, {
      ...options,
      headers: mergedHeaders,
    });
    const text = await response.text();
    const data = parseApiResponseText(text);

    if (!response.ok) {
      throw new Error(getApiErrorMessage(response, data, fallbackMessage, reasonMessages));
    }

    return unwrap ? unwrapApiResponse(data) : data;
  };

  if (method !== 'GET' || !dedupe) {
    return runRequest();
  }

  const key = requestDedupeKey({ url, token, headers: mergedHeaders });
  if (inflightGetRequests.has(key)) {
    return inflightGetRequests.get(key);
  }

  const promise = runRequest().finally(() => {
    inflightGetRequests.delete(key);
  });
  inflightGetRequests.set(key, promise);
  return promise;
};

export const authenticatedApiRequest = (path, options = {}) => apiRequest(path, {
  ...options,
  authenticated: true,
});
