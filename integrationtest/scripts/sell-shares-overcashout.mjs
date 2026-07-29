#!/usr/bin/env node
import { mkdir, writeFile } from 'node:fs/promises';
import { dirname } from 'node:path';

const args = parseArgs(process.argv.slice(2));
const baseUrl = args['base-url'] || 'http://localhost:8080';
const apiPrefix = args['api-prefix'] || '/v0';
const password = args.password || 'password';
const moderator = args.moderator || 'testuser01';
const bettor = args.bettor || 'testuser02';
const counterparty = args.counterparty || 'testuser03';
const admin = args.admin || 'admin';
const artifact = args.artifact || 'integrationtest/artifacts/sell-shares-overcashout-latest.json';
const keepGoing = Boolean(args['keep-going']);
const delayMs = Number(args.delay || 150);
const results = [];
const marketIds = [];
let cappedSeq = 0;

const attachmentSetupThroughSeq18 = [
  { seq: 1, type: 'buy', outcome: 'NO', amount: 50 },
  { seq: 2, type: 'buy', outcome: 'YES', amount: 100 },
  { seq: 3, type: 'buy', outcome: 'YES', amount: 50 },
  { seq: 4, type: 'sell', outcome: 'YES', amount: 200 },
  { seq: 5, type: 'buy', outcome: 'YES', amount: 100 },
  { seq: 6, type: 'buy', outcome: 'NO', amount: 10 },
  { seq: 7, type: 'buy', outcome: 'NO', amount: 40 },
  { seq: 8, type: 'buy', outcome: 'YES', amount: 100 },
  { seq: 9, type: 'buy', outcome: 'NO', amount: 50 },
  { seq: 10, type: 'sell', outcome: 'YES', amount: 400 },
  { seq: 11, type: 'buy', outcome: 'NO', amount: 200 },
  { seq: 12, type: 'buy', outcome: 'YES', amount: 100 },
  { seq: 13, type: 'sell', outcome: 'NO', amount: 600 },
  { seq: 14, type: 'buy', outcome: 'NO', amount: 10 },
  { seq: 15, type: 'buy', outcome: 'NO', amount: 10 },
  { seq: 16, type: 'sell', outcome: 'NO', amount: 520 },
  { seq: 17, type: 'buy', outcome: 'NO', amount: 10 },
  { seq: 18, type: 'buy', outcome: 'NO', amount: 10 },
];

const attachmentOvercashoutAttempt = { seq: 19, type: 'sell', outcome: 'NO', amount: 507 };

const sequenceBasedUnlockedSequence = [
  { seq: 1, user: 'bettor', type: 'buy', outcome: 'NO', amount: 50 },
  { seq: 2, user: 'counterparty', type: 'buy', outcome: 'NO', amount: 25 },
  { seq: 3, user: 'bettor', type: 'sell', outcome: 'NO', amount: 75 },
  { seq: 4, user: 'bettor', type: 'buy', outcome: 'NO', amount: 75 },
  { seq: 5, user: 'counterparty', type: 'buy', outcome: 'NO', amount: 10 },
  { seq: 6, user: 'bettor', type: 'buy', outcome: 'NO', amount: 20 },
];

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
  if (!ok && !keepGoing) throw new Error(name);
}

function sameInt(name, got, want) {
  check(name, Number(got) === Number(want), `got=${got}, want=${want}`);
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

function opposite(outcome) {
  return outcome === 'YES' ? 'NO' : 'YES';
}

async function login(username) {
  return (await api('POST', '/login', { body: { username, password } })).token;
}

async function createMarket(label, moderatorToken, adminToken) {
  const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14);
  const closeAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
  const created = await api('POST', '/markets', {
    token: moderatorToken,
    expect: [201],
    body: {
      questionTitle: `Sell overcashout ${label} ${stamp}`,
      description: 'sell shares over-cashout integration scenario',
      outcomeType: 'BINARY',
      resolutionDateTime: closeAt,
      yesLabel: 'YES',
      noLabel: 'NO',
    },
  });
  const marketId = Number(created.id);
  check(`${label} market created`, marketId > 0, `marketId=${marketId}`);
  marketIds.push(marketId);

  if (created.lifecycleStatus === 'proposed' || created.status === 'proposed') {
    await api('PATCH', `/admin/markets/${marketId}/approve`, {
      token: adminToken,
      body: { confirm: true },
    });
    check(`${label} market approved`, true);
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

function assertSaleFields(prefix, sale, expected) {
  sameInt(`${prefix} sharesSold`, sale.sharesSold, expected.sharesSold);
  sameInt(`${prefix} saleValue`, sale.saleValue, expected.saleValue);
  sameInt(`${prefix} dust`, sale.dust, expected.dust);
  sameInt(`${prefix} netProceeds`, sale.netProceeds, expected.netProceeds);
}

async function assertQuoteAndSell({ prefix, token, marketId, outcome, amount, expected, requireDust = false }) {
  const beforePosition = await position(token, marketId);
  const beforeFinancial = await financial(bettor);
  const beforeDetails = await details(marketId);

  const quote = (await quoteRaw(token, marketId, outcome, amount)).result;
  check(`${prefix} quote allowed`, quote.allowed === true);
  assertSaleFields(`${prefix} quote`, quote, expected);
  sameInt(`${prefix} quote net formula`, quote.netProceeds, quote.saleValue - quote.dust);
  if (requireDust) check(`${prefix} quote has retained dust`, quote.dust > 0, `dust=${quote.dust}`);

  const sell = (await sellRaw(token, marketId, outcome, amount)).result;
  assertSaleFields(`${prefix} sell`, sell, expected);
  sameInt(`${prefix} sell matches quote shares`, sell.sharesSold, quote.sharesSold);
  sameInt(`${prefix} sell matches quote value`, sell.saleValue, quote.saleValue);
  sameInt(`${prefix} sell credits only net proceeds`, sell.netProceeds, sell.saleValue - sell.dust);

  const afterPosition = await position(token, marketId);
  const afterFinancial = await financial(bettor);
  const afterDetails = await details(marketId);
  const soldDelta = shares(beforePosition, outcome) - shares(afterPosition, outcome);
  const oppositeDelta = shares(afterPosition, opposite(outcome)) - shares(beforePosition, opposite(outcome));
  const maxProjectedValue = Math.max(0, positionValue(beforePosition) - sell.saleValue);
  check(`${prefix} position sold-outcome shares decrease`, soldDelta >= sell.sharesSold, `delta=${soldDelta}, sold=${sell.sharesSold}`);
  check(`${prefix} position opposite shares do not increase`, oppositeDelta <= 0, `delta=${oppositeDelta}`);
  check(`${prefix} position value reduced by gross sale value`, positionValue(afterPosition) <= maxProjectedValue, `after=${positionValue(afterPosition)}, max=${maxProjectedValue}`);
  sameInt(`${prefix} balance increases by net proceeds`, Number(afterFinancial.accountBalance || 0) - Number(beforeFinancial.accountBalance || 0), sell.netProceeds);

  if (requireDust) {
    const dustDelta = Number(afterDetails.marketDust || 0) - Number(beforeDetails.marketDust || 0);
    const volumeDelta = Number(afterDetails.totalVolume || 0) - Number(beforeDetails.totalVolume || 0);
    sameInt(`${prefix} market dust retained`, dustDelta, sell.dust);
    sameInt(`${prefix} market volume accounts for sell row and dust`, volumeDelta, -sell.sharesSold + sell.dust);
  } else {
    check(`${prefix} market detail dust numeric`, Number.isFinite(Number(afterDetails.marketDust || 0)));
    check(`${prefix} market detail volume numeric`, Number.isFinite(Number(afterDetails.totalVolume || 0)));
  }
}

async function replaySetupThroughSeq18(token, marketId) {
  for (const step of attachmentSetupThroughSeq18) {
    if (step.type === 'buy') {
      await place(token, marketId, step.outcome, step.amount);
      continue;
    }
    const quote = (await quoteRaw(token, marketId, step.outcome, step.amount)).result;
    check(`setup seq ${step.seq} quote allowed`, quote.allowed === true);
    const sell = (await sellRaw(token, marketId, step.outcome, step.amount)).result;
    check(`setup seq ${step.seq} sell succeeded`, sell.sharesSold > 0 && sell.netProceeds >= 0, `shares=${sell.sharesSold}, net=${sell.netProceeds}`);
  }
}

async function assertOvercashoutCapped(token, marketId) {
  const beforeFinancial = await financial(bettor);
  const beforePosition = await position(token, marketId);
  const beforeDetails = await details(marketId);
  const attempt = attachmentOvercashoutAttempt;

  const quote = (await quoteRaw(token, marketId, attempt.outcome, attempt.amount, [200])).result;
  check(`seq ${attempt.seq} quote caps oversized request`, quote.allowed === true && quote.sharesSold > 0, JSON.stringify(quote));
  check(`seq ${attempt.seq} quote avoids tiny-tail over-cashout`, !(quote.sharesSold <= 4 && quote.netProceeds >= 400), JSON.stringify(quote));
  sameInt(`seq ${attempt.seq} quote net formula`, quote.netProceeds, quote.saleValue - quote.dust);

  const sell = (await sellRaw(token, marketId, attempt.outcome, attempt.amount, [201])).result;
  sameInt(`seq ${attempt.seq} sell matches quote shares`, sell.sharesSold, quote.sharesSold);
  sameInt(`seq ${attempt.seq} sell matches quote value`, sell.saleValue, quote.saleValue);
  sameInt(`seq ${attempt.seq} sell matches quote dust`, sell.dust, quote.dust);
  sameInt(`seq ${attempt.seq} sell credits only net proceeds`, sell.netProceeds, sell.saleValue - sell.dust);
  check(`seq ${attempt.seq} sell avoids tiny-tail over-cashout`, !(sell.sharesSold <= 4 && sell.netProceeds >= 400), JSON.stringify(sell));

  const afterFinancial = await financial(bettor);
  const afterPosition = await position(token, marketId);
  const afterDetails = await details(marketId);
  sameInt(`seq ${attempt.seq} balance increases by net proceeds`, Number(afterFinancial.accountBalance || 0) - Number(beforeFinancial.accountBalance || 0), sell.netProceeds);
  check(`seq ${attempt.seq} sold shares do not exceed pre-sale NO shares`, sell.sharesSold <= shares(beforePosition, attempt.outcome), `sold=${sell.sharesSold}, before=${shares(beforePosition, attempt.outcome)}`);
  const dustDelta = Number(afterDetails.marketDust || 0) - Number(beforeDetails.marketDust || 0);
  const volumeDelta = Number(afterDetails.totalVolume || 0) - Number(beforeDetails.totalVolume || 0);
  sameInt(`seq ${attempt.seq} market dust retained`, dustDelta, sell.dust);
  sameInt(`seq ${attempt.seq} market volume accounts for capped sell`, volumeDelta, -sell.sharesSold + sell.dust);
  cappedSeq = attempt.seq;
}

async function replaySequenceBasedUnlockedFlow(tokens, marketId) {
  for (const step of sequenceBasedUnlockedSequence) {
    const token = tokens[step.user];
    if (step.type === 'buy') {
      await place(token, marketId, step.outcome, step.amount);
      continue;
    }
    const quote = (await quoteRaw(token, marketId, step.outcome, step.amount)).result;
    check(`sequence setup seq ${step.seq} quote allowed`, quote.allowed === true);
    const sell = (await sellRaw(token, marketId, step.outcome, step.amount)).result;
    check(`sequence setup seq ${step.seq} sell succeeded`, sell.sharesSold > 0 && sell.netProceeds >= 0, `shares=${sell.sharesSold}, net=${sell.netProceeds}`);
  }
}

async function assertSequenceBasedUnlockedSaleAllowed({ token, username, marketId, amount }) {
  const beforeFinancial = await financial(username);
  const beforePosition = await position(token, marketId);
  const beforeDetails = await details(marketId);

  check(`${username} sequence has aggregate value`, positionValue(beforePosition) > 0, `value=${positionValue(beforePosition)}`);
  check(`${username} sequence has NO shares`, shares(beforePosition, 'NO') > 0, `shares=${shares(beforePosition, 'NO')}`);

  const quote = (await quoteRaw(token, marketId, 'NO', amount, [200])).result;
  check(`${username} sequence quote allowed`, quote.allowed === true, JSON.stringify(quote));
  check(`${username} sequence quote executable shares`, quote.sharesSold > 0, JSON.stringify(quote));
  check(`${username} sequence quote executable value`, quote.saleValue > 0, JSON.stringify(quote));
  check(`${username} sequence quote no over-cashout`, quote.saleValue <= positionValue(beforePosition), `sale=${quote.saleValue}, value=${positionValue(beforePosition)}`);
  sameInt(`${username} sequence quote net formula`, quote.netProceeds, quote.saleValue - quote.dust);

  const sell = (await sellRaw(token, marketId, 'NO', amount, [201])).result;
  sameInt(`${username} sequence sell matches quote shares`, sell.sharesSold, quote.sharesSold);
  sameInt(`${username} sequence sell matches quote value`, sell.saleValue, quote.saleValue);
  sameInt(`${username} sequence sell matches quote dust`, sell.dust, quote.dust);
  sameInt(`${username} sequence sell credits only net proceeds`, sell.netProceeds, sell.saleValue - sell.dust);

  const afterFinancial = await financial(username);
  const afterDetails = await details(marketId);
  sameInt(`${username} sequence balance net proceeds`, Number(afterFinancial.accountBalance || 0) - Number(beforeFinancial.accountBalance || 0), sell.netProceeds);
  const dustDelta = Number(afterDetails.marketDust || 0) - Number(beforeDetails.marketDust || 0);
  sameInt(`${username} sequence market dust retained`, dustDelta, sell.dust);
}

async function main() {
  const modToken = await login(moderator);
  const bettorToken = await login(bettor);
  const counterpartyToken = await login(counterparty);
  const adminToken = await login(admin);
  check('login seeded moderator, bettor, counterparty, admin', true);

  const happyMarketId = await createMarket('happy', modToken, adminToken);
  await place(bettorToken, happyMarketId, 'NO', 50);
  await place(bettorToken, happyMarketId, 'YES', 100);
  await place(bettorToken, happyMarketId, 'YES', 50);
  await assertQuoteAndSell({
    prefix: 'happy attachment seq 4 exact sell',
    token: bettorToken,
    marketId: happyMarketId,
    outcome: 'YES',
    amount: 200,
    expected: { sharesSold: 100, saleValue: 200, dust: 0, netProceeds: 200 },
  });

  await place(bettorToken, happyMarketId, 'YES', 100);
  await place(bettorToken, happyMarketId, 'NO', 10);
  await place(bettorToken, happyMarketId, 'NO', 40);
  await place(bettorToken, happyMarketId, 'YES', 100);
  await place(bettorToken, happyMarketId, 'NO', 50);
  await assertQuoteAndSell({
    prefix: 'happy attachment seq 10 dust sell',
    token: bettorToken,
    marketId: happyMarketId,
    outcome: 'YES',
    amount: 400,
    expected: { sharesSold: 98, saleValue: 392, dust: 1, netProceeds: 391 },
    requireDust: true,
  });

  const sadMarketId = await createMarket('sad', modToken, adminToken);
  await replaySetupThroughSeq18(bettorToken, sadMarketId);
  await assertOvercashoutCapped(bettorToken, sadMarketId);
  check('attachment over-cashout sequence was capped safely', cappedSeq === attachmentOvercashoutAttempt.seq, `seq=${cappedSeq}`);

  const sequenceBettorMarketId = await createMarket('sequence-bettor', modToken, adminToken);
  await replaySequenceBasedUnlockedFlow({ bettor: bettorToken, counterparty: counterpartyToken }, sequenceBettorMarketId);
  await assertSequenceBasedUnlockedSaleAllowed({
    token: bettorToken,
    username: bettor,
    marketId: sequenceBettorMarketId,
    amount: 1,
  });

  const sequenceCounterpartyMarketId = await createMarket('sequence-counterparty', modToken, adminToken);
  await replaySequenceBasedUnlockedFlow({ bettor: bettorToken, counterparty: counterpartyToken }, sequenceCounterpartyMarketId);
  await assertSequenceBasedUnlockedSaleAllowed({
    token: counterpartyToken,
    username: counterparty,
    marketId: sequenceCounterpartyMarketId,
    amount: 1,
  });
}

async function finish() {
  await mkdir(dirname(artifact), { recursive: true });
  await writeFile(artifact, `${JSON.stringify({ marketIds, cappedSeq, results }, null, 2)}\n`);
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
