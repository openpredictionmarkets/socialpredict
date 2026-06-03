import { sleep } from 'k6';
import { checkProbes, placeBet, preAuthenticateUsers, requireFixtures } from './lib/common.js';

const duration = __ENV.DURATION || '60s';
const targetRate = Number(__ENV.TARGET_RATE || '50');
const targetTimeUnit = __ENV.TARGET_TIME_UNIT || '1s';
const preauthUsers = Number(__ENV.PREAUTH_USERS || '100');
const setupTimeout = __ENV.SETUP_TIMEOUT || '5m';

export const options = {
  setupTimeout,
  scenarios: {
    hot_market_burst: {
      executor: 'constant-arrival-rate',
      rate: targetRate,
      timeUnit: targetTimeUnit,
      duration,
      preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || '100'),
      maxVUs: Number(__ENV.MAX_VUS || '1000'),
    },
  },
  thresholds: {
    checks: ['rate>0.90'],
    http_req_failed: ['rate<0.10'],
  },
};

export function setup() {
  requireFixtures();
  checkProbes();
  return {
    authenticatedUsers: preAuthenticateUsers(preauthUsers),
  };
}

export default function (data) {
  placeBet({ hotOnly: true, authenticatedUsers: data.authenticatedUsers });
  sleep(0.01);
}
