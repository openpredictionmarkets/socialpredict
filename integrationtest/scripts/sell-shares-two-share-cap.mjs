#!/usr/bin/env node
import { mkdir, writeFile } from 'node:fs/promises';
import { dirname } from 'node:path';

const args = parseArgs(process.argv.slice(2));
const baseUrl = args['base-url'] || 'http://localhost:8080';
const apiPrefix = args['api-prefix'] || '/v0';
const password = args.password || 'password';
const moderator = args.moderator || 'testuser01';
const bettor = args.bettor || 'testuser03';
const follower = args.follower || 'testuser04';
const admin = args.admin || 'admin';
const artifact = args.artifact || 'integrationtest/artifacts/sell-shares-two-share-cap-latest.json';
const delayMs = Number(args.delay || 150);
const results = [];
const marketIds = [];

function parseArgs(items) {
  const out = {};
  for (let i = 0; i < items.length; i += 1) {
    const item = items[i];
    if (!item.startsWith('--')) continue;
    const key = item.slice(2);
    if (items[i + 1] && !items[i + 1].startsWith('--')) out[key] = items[++i];
    else out[key] = true;
  }
  return out;
}

function unwrap(data) {
  return data?.ok === true && Object.hasOwn(data, 'result') ? data.result : data;
}

function apiUrl(path) {
  return `${baseUrl.replace(/\/$/, '')}${apiPrefix.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
}

async function wait(ms) {
  if (ms > 0) await new Promise((resolve) => setTimeout(resolve, ms));
}

async function apiRaw(method, path, { body, token, expect = [200], retries = 4 } = {}) {
  const url = apiUrl(path);
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    const response = await fetch(url, {
      method,
      headers: {
        Accept: 'application/json',
        ...(body ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    const text = await response.text();
    const data = text ? JSON.parse(text) : null;
    if (expect.includes(response.status)) {
      return { status: response.status, data, result: unwrap(data), text };
    }
    if (response.status === 429 && attempt < retries) {
      await wait(1200 * (attempt + 1));
      continue;
    }
    throw new Error(`${method} ${url} -> ${response.status}: ${text.slice(0, 500)}`);
  }
  throw new Error(`${method} ${url} did not complete`);
}

async function api(method, path, options) {
  return (await apiRaw(method, path, options)).result;
}

function check(name, ok, detail = '') {
  results.push({ name, ok: Boolean(ok), detail });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` - ${detail}` : ''}`);
  if (!ok) throw new Error(name);
}

function sameInt(name, got, want) {
  check(name, Number(got) === Number(want), `got=${got}, want=${want}`);
}

function reason(raw) {
  return raw?.data?.reason || raw?.result?.reason || raw?.data?.code || '';
}

function message(raw) {
  return raw?.data?.message || raw?.result?.message || '';
}

function numberField(row, camelName, exportedName) {
  return Number(row?.[camelName] ?? row?.[exportedName] ?? 0);
}

function shares(position, outcome) {
  return outcome === 'YES'
    ? numberField(position, 'yesSharesOwned', 'YesSharesOwned')
    : numberField(position, 'noSharesOwned', 'NoSharesOwned');
}

function positionValue(position) {
  return numberField(position, 'value', 'Value');
}

async function login(username) {
  return (await api('POST', '/login', { body: { username, password } })).token;
}

async function createMarket(moderatorToken, adminToken) {
  const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14);
  const closeAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
  const created = await api('POST', '/markets', {
    token: moderatorToken,
    expect: [201],
    body: {
      questionTitle: `Sell two-share backend cap ${stamp}`,
      description: 'backend regression for capped sell settlement',
      outcomeType: 'BINARY',
      resolutionDateTime: closeAt,
      yesLabel: 'YES',
      noLabel: 'NO',
    },
  });
  const marketId = Number(created.id);
  check('market created', marketId > 0, `marketId=${marketId}`);
  marketIds.push(marketId);

  if (created.lifecycleStatus === 'proposed' || created.status === 'proposed') {
    await api('PATCH', `/admin/markets/${marketId}/approve`, {
      token: adminToken,
      body: { confirm: true },
    });
    check('market approved', true);
  }
  return marketId;
}

async function place(token, marketId, outcome, amount) {
  const result = await api('POST', '/bet', {
    token,
    expect: [201],
    body: { marketId, outcome, amount },
  });
  await wait(delayMs);
  return result;
}

async function quoteRaw(token, marketId, outcome, amount, expect = [200]) {
  return apiRaw('POST', '/sell/quote', {
    token,
    expect,
    body: { marketId, outcome, amount },
  });
}

async function sellRaw(token, marketId, outcome, amount, expect = [201]) {
  const result = await apiRaw('POST', '/sell', {
    token,
    expect,
    body: { marketId, outcome, amount },
  });
  await wait(delayMs);
  return result;
}

async function position(token, marketId) {
  return api('GET', `/userposition/${marketId}`, { token });
}

async function details(marketId) {
  return api('GET', `/markets/${marketId}`);
}

async function financial(username) {
  const response = await api('GET', `/users/${username}/financial`);
  return response.financial || {};
}

function assertSale(prefix, sale) {
  sameInt(`${prefix} sharesSold`, sale.sharesSold, 2);
  sameInt(`${prefix} saleValue`, sale.saleValue, 2);
  sameInt(`${prefix} dust`, sale.dust, 1);
  sameInt(`${prefix} netProceeds`, sale.netProceeds, 1);
  sameInt(`${prefix} net formula`, sale.netProceeds, sale.saleValue - sale.dust);
}

async function main() {
  const modToken = await login(moderator);
  const bettorToken = await login(bettor);
  const followerToken = await login(follower);
  const adminToken = await login(admin);
  check('login seeded moderator, bettor, follower, admin', true);

  const marketId = await createMarket(modToken, adminToken);
  await place(bettorToken, marketId, 'NO', 1);

  const lockedFinancial = await financial(bettor);
  const lockedPosition = await position(bettorToken, marketId);
  const lockedDetails = await details(marketId);
  sameInt('single buy shows aggregate NO share', shares(lockedPosition, 'NO'), 1);
  sameInt('single buy aggregate value', positionValue(lockedPosition), 1);

  const lockedQuote = await quoteRaw(bettorToken, marketId, 'NO', 1, [422]);
  check('quote rejected before another user follow-up', lockedQuote.status === 422, `status=${lockedQuote.status}`);
  check('locked quote uses no-position contract', reason(lockedQuote) === 'NO_POSITION', `reason=${reason(lockedQuote)}`);
  check('locked quote reports follow-up requirement', message(lockedQuote).includes('follow-up order from another user'), message(lockedQuote));

  const lockedSell = await sellRaw(bettorToken, marketId, 'NO', 1, [422]);
  check('sell rejected before another user follow-up', lockedSell.status === 422, `status=${lockedSell.status}`);
  check('locked sell uses no-position contract', reason(lockedSell) === 'NO_POSITION', `reason=${reason(lockedSell)}`);
  check('locked sell reports follow-up requirement', message(lockedSell).includes('follow-up order from another user'), message(lockedSell));
  sameInt('locked sell balance unchanged', Number((await financial(bettor)).accountBalance || 0), Number(lockedFinancial.accountBalance || 0));
  sameInt('locked sell position unchanged', positionValue(await position(bettorToken, marketId)), positionValue(lockedPosition));
  sameInt('locked sell dust unchanged', Number((await details(marketId)).marketDust || 0), Number(lockedDetails.marketDust || 0));

  await place(followerToken, marketId, 'NO', 1);

  const followerQuote = await quoteRaw(followerToken, marketId, 'NO', 1, [422]);
  check('latest follower quote remains locked', followerQuote.status === 422, `status=${followerQuote.status}`);
  check('latest follower quote uses no-position contract', reason(followerQuote) === 'NO_POSITION', `reason=${reason(followerQuote)}`);

  const beforeFinancial = await financial(bettor);
  const beforePosition = await position(bettorToken, marketId);
  const beforeDetails = await details(marketId);
  const ownedNoShares = shares(beforePosition, 'NO');
  const currentValue = positionValue(beforePosition);
  const oversizedAmount = 3;

  sameInt('different-user follow-up unlocks bettor NO shares', ownedNoShares, 2);
  sameInt('setup YES shares', shares(beforePosition, 'YES'), 0);
  sameInt('setup position value', currentValue, 2);

  const quote = (await quoteRaw(bettorToken, marketId, 'NO', oversizedAmount)).result;
  check('quote allowed', quote.allowed === true);
  assertSale('quote', quote);
  sameInt('quote valuePerShare', quote.valuePerShare, 1);
  check('quote never exceeds owned shares', quote.sharesSold <= ownedNoShares, `quoted=${quote.sharesSold}, owned=${ownedNoShares}`);
  check('quote never exceeds current position value', quote.saleValue <= currentValue, `saleValue=${quote.saleValue}, value=${currentValue}`);

  const sell = (await sellRaw(bettorToken, marketId, 'NO', oversizedAmount)).result;
  assertSale('sell', sell);
  check('sell never exceeds owned shares', sell.sharesSold <= ownedNoShares, `sold=${sell.sharesSold}, owned=${ownedNoShares}`);
  check('sell gross value never exceeds current position value', sell.saleValue <= currentValue, `saleValue=${sell.saleValue}, value=${currentValue}`);

  const afterFinancial = await financial(bettor);
  const afterPosition = await position(bettorToken, marketId);
  const afterDetails = await details(marketId);
  sameInt('position NO shares zero after capped sell', shares(afterPosition, 'NO'), 0);
  sameInt('position YES shares unchanged', shares(afterPosition, 'YES'), 0);
  sameInt('position value zero after capped sell', positionValue(afterPosition), 0);
  sameInt('balance increases by net proceeds', Number(afterFinancial.accountBalance || 0) - Number(beforeFinancial.accountBalance || 0), sell.netProceeds);
  sameInt('market dust retained', Number(afterDetails.marketDust || 0) - Number(beforeDetails.marketDust || 0), sell.dust);
  sameInt('market volume accounts for sell row and dust', Number(afterDetails.totalVolume || 0) - Number(beforeDetails.totalVolume || 0), -sell.sharesSold + sell.dust);

  const emptyQuote = await quoteRaw(bettorToken, marketId, 'NO', 1, [422]);
  check('quote rejected after position is exhausted', emptyQuote.status === 422, `status=${emptyQuote.status}`);
  check('exhausted quote uses existing contract', reason(emptyQuote) === 'NO_POSITION' || reason(emptyQuote) === 'INSUFFICIENT_SHARES', `reason=${reason(emptyQuote)}`);

  const emptySell = await sellRaw(bettorToken, marketId, 'NO', 1, [422]);
  check('sell rejected after position is exhausted', emptySell.status === 422, `status=${emptySell.status}`);
  check('exhausted sell uses existing contract', reason(emptySell) === 'NO_POSITION' || reason(emptySell) === 'INSUFFICIENT_SHARES', `reason=${reason(emptySell)}`);
}

async function finish() {
  await mkdir(dirname(artifact), { recursive: true });
  await writeFile(artifact, `${JSON.stringify({ marketIds, results }, null, 2)}\n`);
  const failed = results.filter((result) => !result.ok);
  console.log(`artifact: ${artifact}`);
  console.log(`summary: ${results.length - failed.length} passed, ${failed.length} failed`);
  process.exitCode = failed.length ? 1 : 0;
}

try {
  await main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  results.push({ name: 'runner completed', ok: false, detail: message });
  console.error(`FAIL runner completed - ${message}`);
} finally {
  await finish();
}
