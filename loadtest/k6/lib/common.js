import http from 'k6/http';
import { check } from 'k6';
import crypto from 'k6/crypto';
import { SharedArray } from 'k6/data';
import { Counter } from 'k6/metrics';

export const betsAttempted = new Counter('sp_bets_attempted');
export const betsSucceeded = new Counter('sp_bets_succeeded');
export const betsFailed = new Counter('sp_bets_failed');
export const loginFailures = new Counter('sp_login_failures');
export const marketReadFailures = new Counter('sp_market_read_failures');
export const rateLimited = new Counter('sp_rate_limited');
export const loginRateLimited = new Counter('sp_login_rate_limited');

export const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
export const API_PREFIX = normalizeApiPrefix(__ENV.API_PREFIX || '');
export const USERS_FILE = __ENV.USERS_FILE || 'loadtest/fixtures/users.csv';
export const MARKETS_FILE = __ENV.MARKETS_FILE || 'loadtest/fixtures/markets.csv';
export const DEFAULT_PASSWORD = __ENV.LOADTEST_PASSWORD || '';
export const BET_AMOUNT = Number(__ENV.BET_AMOUNT || '1');
export const HOT_MARKET_WEIGHT = Number(__ENV.HOT_MARKET_WEIGHT || '8');
export const FAILURE_LOG_LIMIT = Number(__ENV.FAILURE_LOG_LIMIT || '10');
export const PREAUTH_USERS = Number(__ENV.PREAUTH_USERS || '100');

let failureLogCount = 0;

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

function normalizeApiPrefix(prefix) {
  const trimmed = String(prefix || '').trim();
  if (!trimmed || trimmed === '/') return '';
  return trimmed.startsWith('/') ? trimmed.replace(/\/$/, '') : `/${trimmed.replace(/\/$/, '')}`;
}

function apiUrl(path) {
  return `${BASE_URL}${API_PREFIX}${path}`;
}

function responseSnippet(response) {
  if (!response || !response.body) return '';
  return String(response.body).replace(/\s+/g, ' ').slice(0, 240);
}

function responseReason(response) {
  if (!response || !response.body) return '';
  try {
    return response.json('reason') || '';
  } catch {
    return '';
  }
}

function recordFailure(kind, response, context = {}) {
  const status = response ? String(response.status) : 'unknown';
  const reason = context.reason || responseReason(response);
  const tags = { kind, status };
  if (reason) tags.reason = reason;
  if (reason === 'RATE_LIMITED') {
    rateLimited.add(1, tags);
  } else if (reason === 'LOGIN_RATE_LIMITED') {
    loginRateLimited.add(1, tags);
  }
  if (kind === 'login') {
    loginFailures.add(1, tags);
  } else if (kind === 'bet') {
    betsFailed.add(1, tags);
  } else if (kind === 'market_read') {
    marketReadFailures.add(1, tags);
  }

  if (failureLogCount >= FAILURE_LOG_LIMIT) return;
  failureLogCount += 1;

  const details = Object.entries({ ...context, reason })
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${key}=${value}`)
    .join(' ');
  console.warn(`[loadtest] ${kind} failed status=${status} ${details} body="${responseSnippet(response)}"`);
}

export function secureRandomFraction() {
  const bytes = new Uint8Array(crypto.randomBytes(4));
  const value = ((bytes[0] << 24) >>> 0) + (bytes[1] << 16) + (bytes[2] << 8) + bytes[3];
  return value / 0x100000000;
}

export function secureRandomBoolean() {
  return secureRandomFraction() < 0.5;
}

export function requireFixtures() {
  if (users.length === 0) {
    throw new Error(`No load-test users found in ${USERS_FILE}`);
  }
  if (markets.length === 0) {
    throw new Error(`No load-test markets found in ${MARKETS_FILE}`);
  }
}

export function pick(list) {
  return list[Math.floor(secureRandomFraction() * list.length)];
}

export function pickUser() {
  return pick(users);
}

export function pickMarket({ hotOnly = false } = {}) {
  const hotMarkets = markets.filter((market) => market.kind === 'hot');
  if (hotOnly && hotMarkets.length > 0) return pick(hotMarkets);
  if (hotMarkets.length > 0 && secureRandomFraction() < HOT_MARKET_WEIGHT / (HOT_MARKET_WEIGHT + 1)) {
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
    apiUrl('/v0/login'),
    JSON.stringify({ username: user.username, password: user.password }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const ok = check(response, {
    'login returned 200': (r) => r.status === 200,
    'login returned token': (r) => Boolean(r.json('result.token')),
  });
  if (!ok) {
    recordFailure('login', response, { user: user.username, url: apiUrl('/v0/login') });
    return '';
  }

  const token = response.json('result.token');
  tokenCache[user.username] = token;
  return token;
}

export function preAuthenticateUsers(limit = PREAUTH_USERS) {
  const count = Math.min(Math.max(Number(limit) || 0, 0), users.length);
  const authenticatedUsers = [];
  for (const user of users.slice(0, count)) {
    const token = login(user);
    if (token) {
      authenticatedUsers.push({ username: user.username, token });
    }
  }
  if (authenticatedUsers.length === 0) {
    throw new Error('No users could be pre-authenticated for load testing');
  }
  return authenticatedUsers;
}

export function pickAuthenticatedUser(authenticatedUsers) {
  if (!authenticatedUsers || authenticatedUsers.length === 0) {
    return null;
  }
  return pick(authenticatedUsers);
}

export function placeBet({ hotOnly = false, authenticatedUsers = null } = {}) {
  const authenticatedUser = pickAuthenticatedUser(authenticatedUsers);
  const user = authenticatedUser || pickUser();
  const market = pickMarket({ hotOnly });
  const token = authenticatedUser ? authenticatedUser.token : login(user);
  if (!token) {
    betsFailed.add(1, { reason: 'login_failed' });
    return;
  }

  betsAttempted.add(1);
  const outcome = secureRandomBoolean() ? 'YES' : 'NO';
  const response = http.post(
    apiUrl('/v0/bet'),
    JSON.stringify({ marketId: market.id, amount: BET_AMOUNT, outcome }),
    authHeaders(token),
  );

  const ok = check(response, {
    'bet returned 201': (r) => r.status === 201,
  });
  if (ok) {
    betsSucceeded.add(1);
  } else {
    recordFailure('bet', response, { marketId: market.id, outcome, url: apiUrl('/v0/bet') });
  }
}

export function readMarket() {
	const market = pickMarket();
	const response = http.get(apiUrl(`/v0/markets/${market.id}`));
	const ok = check(response, {
		'market detail returned 200': (r) => r.status === 200,
	});
	if (!ok) {
		recordFailure('market_read', response, { marketId: market.id, url: apiUrl(`/v0/markets/${market.id}`) });
	}
}

export function readMarketList() {
	const response = http.get(apiUrl('/v0/markets'));
	const ok = check(response, {
		'market list returned 200': (r) => r.status === 200,
	});
	if (!ok) {
		recordFailure('market_read', response, { url: apiUrl('/v0/markets') });
	}
}

export function checkProbes() {
  for (const path of ['/health', '/readyz', '/ops/status']) {
    const response = http.get(`${BASE_URL}${path}`);
    check(response, {
      [`${path} returned 200`]: (r) => r.status === 200,
    });
  }
}
