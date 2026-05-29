import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter } from 'k6/metrics';

export const betsAttempted = new Counter('sp_bets_attempted');
export const betsSucceeded = new Counter('sp_bets_succeeded');
export const betsFailed = new Counter('sp_bets_failed');
export const loginFailures = new Counter('sp_login_failures');

export const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
export const USERS_FILE = __ENV.USERS_FILE || 'loadtest/fixtures/users.csv';
export const MARKETS_FILE = __ENV.MARKETS_FILE || 'loadtest/fixtures/markets.csv';
export const DEFAULT_PASSWORD = __ENV.LOADTEST_PASSWORD || '';
export const BET_AMOUNT = Number(__ENV.BET_AMOUNT || '1');
export const HOT_MARKET_WEIGHT = Number(__ENV.HOT_MARKET_WEIGHT || '8');

function parseCsv(raw) {
  const lines = raw.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  if (lines.length < 2) return [];

  const headers = lines[0].split(',').map((header) => header.trim());
  return lines.slice(1).map((line) => {
    const values = line.split(',').map((value) => value.trim());
    return headers.reduce((record, header, index) => {
      record[header] = values[index] || '';
      return record;
    }, {});
  });
}

export const users = new SharedArray('loadtest users', () => {
  const parsed = parseCsv(open(USERS_FILE));
  return parsed
    .filter((row) => row.username)
    .map((row) => ({
      username: row.username,
      password: row.password || DEFAULT_PASSWORD,
    }))
    .filter((row) => row.password);
});

export const markets = new SharedArray('loadtest markets', () => {
  const parsed = parseCsv(open(MARKETS_FILE));
  return parsed
    .map((row) => ({
      id: Number(row.market_id || row.id || row.marketId),
      kind: row.kind || 'normal',
    }))
    .filter((row) => Number.isFinite(row.id) && row.id > 0);
});

const tokenCache = {};

export function requireFixtures() {
  if (users.length === 0) {
    throw new Error(`No load-test users found in ${USERS_FILE}`);
  }
  if (markets.length === 0) {
    throw new Error(`No load-test markets found in ${MARKETS_FILE}`);
  }
}

export function pick(list) {
  return list[Math.floor(Math.random() * list.length)];
}

export function pickUser() {
  return pick(users);
}

export function pickMarket({ hotOnly = false } = {}) {
  const hotMarkets = markets.filter((market) => market.kind === 'hot');
  if (hotOnly && hotMarkets.length > 0) return pick(hotMarkets);
  if (hotMarkets.length > 0 && Math.random() < HOT_MARKET_WEIGHT / (HOT_MARKET_WEIGHT + 1)) {
    return pick(hotMarkets);
  }
  return pick(markets);
}

export function authHeaders(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
}

export function login(user) {
  if (tokenCache[user.username]) return tokenCache[user.username];

  const response = http.post(
    `${BASE_URL}/v0/login`,
    JSON.stringify({ username: user.username, password: user.password }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const ok = check(response, {
    'login returned 200': (r) => r.status === 200,
    'login returned token': (r) => Boolean(r.json('result.token')),
  });
  if (!ok) {
    loginFailures.add(1);
    return '';
  }

  const token = response.json('result.token');
  tokenCache[user.username] = token;
  return token;
}

export function placeBet({ hotOnly = false } = {}) {
  const user = pickUser();
  const market = pickMarket({ hotOnly });
  const token = login(user);
  if (!token) {
    betsFailed.add(1);
    return;
  }

  betsAttempted.add(1);
  const outcome = Math.random() < 0.5 ? 'YES' : 'NO';
  const response = http.post(
    `${BASE_URL}/v0/bet`,
    JSON.stringify({ marketId: market.id, amount: BET_AMOUNT, outcome }),
    authHeaders(token),
  );

  const ok = check(response, {
    'bet returned 201': (r) => r.status === 201,
  });
  if (ok) {
    betsSucceeded.add(1);
  } else {
    betsFailed.add(1);
  }
}

export function readMarket() {
	const market = pickMarket();
	const response = http.get(`${BASE_URL}/v0/markets/${market.id}`);
	check(response, {
		'market detail returned 200': (r) => r.status === 200,
	});
}

export function readMarketList() {
	const response = http.get(`${BASE_URL}/v0/markets`);
	check(response, {
		'market list returned 200': (r) => r.status === 200,
	});
}

export function checkProbes() {
  for (const path of ['/health', '/readyz', '/ops/status']) {
    const response = http.get(`${BASE_URL}${path}`);
    check(response, {
      [`${path} returned 200`]: (r) => r.status === 200,
    });
  }
}
