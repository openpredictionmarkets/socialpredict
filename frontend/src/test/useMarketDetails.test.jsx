import { describe, expect, it } from 'vitest';
import { mergeMarketDetailsWithSummary } from '../hooks/useMarketDetails';

const liveDetails = {
  probabilityChanges: [{ probability: 0.58 }],
  numUsers: 2,
  totalVolume: 25,
  marketDust: 1,
  freshness: null,
};

describe('mergeMarketDetailsWithSummary', () => {
  it('does not let stale summaries overwrite live market accounting', () => {
    const staleSummary = {
      probabilityChanges: [{ probability: 0.5 }],
      numUsers: 1,
      totalVolume: 10,
      marketDust: 0,
      freshness: {
        isStale: true,
        staleReason: 'bet_accepted',
      },
    };

    expect(mergeMarketDetailsWithSummary(liveDetails, staleSummary)).toEqual(liveDetails);
  });

  it('uses fresh summary accounting when available', () => {
    const freshSummary = {
      probabilityChanges: [{ probability: 0.6 }],
      numUsers: 3,
      totalVolume: 40,
      marketDust: 2,
      freshness: {
        isStale: false,
      },
    };

    expect(mergeMarketDetailsWithSummary(liveDetails, freshSummary)).toEqual({
      ...liveDetails,
      probabilityChanges: freshSummary.probabilityChanges,
      numUsers: 3,
      totalVolume: 40,
      marketDust: 2,
      freshness: freshSummary.freshness,
    });
  });
});
