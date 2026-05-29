import { sleep } from 'k6';
import { checkProbes, placeBet, readMarket, readMarketList, requireFixtures } from './lib/common.js';

export const options = {
  vus: Number(__ENV.VUS || '1'),
  iterations: Number(__ENV.ITERATIONS || '3'),
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
  readMarketList();
  readMarket();
  placeBet();
  sleep(1);
}
