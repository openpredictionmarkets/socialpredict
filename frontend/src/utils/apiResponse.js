export const unwrapApiResponse = (payload) => {
  if (payload && typeof payload === 'object' && 'ok' in payload) {
    if (payload.ok === false) {
      throw new Error(payload.message || payload.reason || 'Request failed');
    }

    if (payload.ok === true && 'result' in payload) {
      return payload.result;
    }
  }

  return payload;
};

export const parseApiResponseText = (text) => {
  if (!text) {
    return {};
  }

  try {
    return JSON.parse(text);
  } catch {
    return { message: text };
  }
};

export const getApiErrorMessage = (
  response,
  payload,
  fallbackMessage,
  reasonMessages = {},
) => {
  if (payload?.message) {
    return payload.message;
  }

  if (payload?.error) {
    return payload.error;
  }

  if (payload?.reason && reasonMessages[payload.reason]) {
    return reasonMessages[payload.reason];
  }

  if (typeof payload === 'string' && payload.trim()) {
    return payload;
  }

  return fallbackMessage || `Request failed with status ${response.status}`;
};
