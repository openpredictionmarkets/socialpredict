import { afterEach, describe, expect, it, vi } from 'vitest';
import { quoteSale } from './tradeApi';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('tradeApi', () => {
  it('preserves backend projection details on rejected sale quotes', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      ok: false,
      reason: 'INSUFFICIENT_SHARES',
      message: 'Position value exists, but this Sale Order is not executable right now.',
      details: {
        outcome: 'NO',
        requestedCredits: 17,
        positionValue: 34,
        nominalUnlockedValue: 17,
        projectedPositionValue: 34,
        executableSaleValue: 0,
      },
    }), { status: 422, headers: { 'Content-Type': 'application/json' } })));

    await expect(quoteSale({
      token: 'token-1',
      marketId: 7,
      outcome: 'NO',
      amount: 17,
    })).rejects.toMatchObject({
      status: 422,
      reason: 'INSUFFICIENT_SHARES',
      message: 'Position value exists, but this Sale Order is not executable right now.',
      details: {
        outcome: 'NO',
        requestedCredits: 17,
        positionValue: 34,
        nominalUnlockedValue: 17,
        projectedPositionValue: 34,
        executableSaleValue: 0,
      },
    });
  });
});
