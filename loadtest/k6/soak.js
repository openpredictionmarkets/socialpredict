import { sleep } from 'k6';
import { checkProbes, placeBet, readMarket, readMarketList, requireFixtures } from './lib/common.js';

const duration = __ENV.DURATION || '30m';
const rate = Number(__ENV.RATE || '10');

export const options = {
  scenarios: {
    soak: {
      executor: 'constant-arrival-rate',
      rate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || '50'),
      maxVUs: Number(__ENV.MAX_VUS || '500'),
    },
  },
  thresholds: {
    checks: ['rate>0.95'],
    http_req_failed: ['rate<0.05'],
  },
};

export function setup() {
  requireFixtures();
  checkProbes();
}

export default function () {
  const roll = Math.random();
  if (roll < 0.35) {
    readMarketList();
  } else if (roll < 0.7) {
    readMarket();
  } else {
    placeBet();
  }
  sleep(0.25);
}
