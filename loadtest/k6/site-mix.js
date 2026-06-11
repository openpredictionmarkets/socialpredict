import { sleep } from 'k6';
import {
  checkProbes,
  fetchActiveTagSlugs,
  pickMarket,
  pickTagSlug,
  placeBet,
  preAuthenticateUsers,
  readConfigDisplay,
  readMarket,
  readMarketBetsPage,
  readMarketDiscovery,
  readMarketLeaderboardPage,
  readMarketPositionsPage,
  readMarketSummary,
  readStatsDisplay,
  readUserDisplay,
  requireFixtures,
  secureRandomFraction,
} from './lib/common.js';

const duration = __ENV.DURATION || '5m';
const browseRate = Number(__ENV.BROWSE_RATE || '250');
const browseTimeUnit = __ENV.BROWSE_TIME_UNIT || '1s';
const betRate = Number(__ENV.BET_RATE || '25');
const betTimeUnit = __ENV.BET_TIME_UNIT || '1s';
const preauthUsers = Number(__ENV.PREAUTH_USERS || '500');
const setupTimeout = __ENV.SETUP_TIMEOUT || '10m';

export const options = {
  setupTimeout,
  scenarios: {
    site_reads: {
      executor: 'constant-arrival-rate',
      rate: browseRate,
      timeUnit: browseTimeUnit,
      duration,
      preAllocatedVUs: Number(__ENV.BROWSE_VUS || '100'),
      maxVUs: Number(__ENV.BROWSE_MAX_VUS || '1000'),
      exec: 'siteRead',
    },
    hot_market_bets: {
      executor: 'constant-arrival-rate',
      rate: betRate,
      timeUnit: betTimeUnit,
      duration,
      preAllocatedVUs: Number(__ENV.BET_VUS || '100'),
      maxVUs: Number(__ENV.BET_MAX_VUS || '1000'),
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
  return {
    authenticatedUsers: preAuthenticateUsers(preauthUsers),
    tagSlugs: fetchActiveTagSlugs(),
  };
}

export function siteRead(data) {
  const roll = secureRandomFraction();
  const market = pickMarket();
  const tagSlug = pickTagSlug(data.tagSlugs);

  if (roll < 0.05) {
    readConfigDisplay();
  } else if (roll < 0.25) {
    readMarketDiscovery({ slug: 'markets', status: secureRandomFraction() < 0.8 ? 'active' : 'all' });
  } else if (roll < 0.35) {
    const slug = tagSlug || 'markets';
    readMarketDiscovery({ slug, status: 'active', tagSlug });
  } else if (roll < 0.55) {
    readMarketSummary(market);
  } else if (roll < 0.60) {
    readMarket();
  } else if (roll < 0.67) {
    readMarketBetsPage(market);
  } else if (roll < 0.74) {
    readMarketPositionsPage(market);
  } else if (roll < 0.80) {
    readMarketLeaderboardPage(market);
  } else if (roll < 0.90) {
    readUserDisplay(data.authenticatedUsers);
  } else {
    readStatsDisplay(data.authenticatedUsers);
  }

  sleep(0.01);
}

export function bet(data) {
  placeBet({ hotOnly: true, authenticatedUsers: data.authenticatedUsers });
  sleep(0.01);
}
