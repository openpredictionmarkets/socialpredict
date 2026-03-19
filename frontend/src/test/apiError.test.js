import { describe, it, expect } from 'vitest';
import { parseApiError } from '../utils/apiError';

function makeResponse(status, body) {
  return {
    status,
    text: async () => body,
  };
}

describe('parseApiError', () => {
  it('extracts the error field from a JSON {"error":"..."} backend response', async () => {
    const response = makeResponse(400, JSON.stringify({ error: 'bet amount too low' }));
    const msg = await parseApiError(response);
    expect(msg).toBe('bet amount too low');
  });

  it('falls back to message field when error field is absent', async () => {
    const response = makeResponse(400, JSON.stringify({ message: 'something went wrong' }));
    const msg = await parseApiError(response);
    expect(msg).toBe('something went wrong');
  });

  it('returns plain text body when body is not valid JSON', async () => {
    const response = makeResponse(500, 'Internal Server Error');
    const msg = await parseApiError(response);
    expect(msg).toBe('Internal Server Error');
  });

  it('returns HTTP status string when body is empty', async () => {
    const response = makeResponse(503, '');
    const msg = await parseApiError(response);
    expect(msg).toBe('HTTP 503');
  });

  it('returns HTTP status string when text() throws', async () => {
    const response = {
      status: 502,
      text: async () => { throw new Error('network error'); },
    };
    const msg = await parseApiError(response);
    expect(msg).toBe('HTTP 502');
  });

  it('uses error field over message field when both present', async () => {
    const response = makeResponse(422, JSON.stringify({ error: 'specific error', message: 'generic message' }));
    const msg = await parseApiError(response);
    expect(msg).toBe('specific error');
  });
});
