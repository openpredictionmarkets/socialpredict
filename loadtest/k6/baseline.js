import { sleep } from 'k6';
import { checkProbes, placeBet, readMarket, readMarketList, requireFixtures, secureRandomBoolean } from './lib/common.js';

const duration = __ENV.DURATION || '5m';
const browseRate = Number(__ENV.BROWSE_RATE || '20');
const betRate = Number(__ENV.BET_RATE || '5');

export const options = {
  scenarios: {
    browse: {
      executor: 'constant-arrival-rate',
      rate: browseRate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: Number(__ENV.BROWSE_VUS || '20'),
      maxVUs: Number(__ENV.BROWSE_MAX_VUS || '200'),
      exec: 'browse',
    },
    bet: {
      executor: 'constant-arrival-rate',
      rate: betRate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: Number(__ENV.BET_VUS || '20'),
      maxVUs: Number(__ENV.BET_MAX_VUS || '200'),
      exec: 'bet',
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

export function browse() {
  if (secureRandomBoolean()) {
    readMarketList();
  } else {
    readMarket();
  }
  sleep(0.1);
}

export function bet() {
  placeBet();
  sleep(0.1);
}
