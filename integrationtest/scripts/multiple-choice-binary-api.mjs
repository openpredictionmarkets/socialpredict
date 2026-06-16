#!/usr/bin/env node
import { mkdir, writeFile } from 'node:fs/promises';
import { dirname } from 'node:path';

const args = parseArgs(process.argv.slice(2));
const baseUrl = args['base-url'] || 'http://localhost:8080';
const apiPrefix = args['api-prefix'] || '/v0';
const password = args.password || 'password';
const moderator = args.moderator || 'testuser01';
const bettor = args.bettor || 'testuser02';
const admin = args.admin || 'admin';
const artifact = args.artifact || 'integrationtest/artifacts/multiple-choice-binary-latest.json';
const keepGoing = Boolean(args['keep-going']);
const results = [];
let groupId = 0;

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

async function api(method, path, { body, token, expect = [200], retries = 2 } = {}) {
  const url = `${baseUrl.replace(/\/$/, '')}${apiPrefix.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
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
    if (expect.includes(response.status)) return unwrap(data);
    if (response.status === 429 && attempt < retries) {
      await new Promise((resolve) => setTimeout(resolve, 1200 * (attempt + 1)));
      continue;
    }
    throw new Error(`${method} ${url} -> ${response.status}: ${text.slice(0, 300)}`);
  }
}

function check(name, ok, detail = '') {
  results.push({ name, ok: Boolean(ok), detail });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` - ${detail}` : ''}`);
  if (!ok && !keepGoing) throw new Error(name);
}

async function login(username) {
  return (await api('POST', '/login', { body: { username, password } })).token;
}

async function main() {
  const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14);
  const closeAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
  const modToken = await login(moderator);
  const bettorToken = await login(bettor);
  const adminToken = await login(admin);
  check('login seeded moderator, bettor, admin', true);

  const duplicate = await api('POST', '/market-groups', {
    token: modToken,
    expect: [400, 409],
    body: {
      questionTitle: `Ponytail duplicate ${stamp}`,
      description: 'duplicate-label rejection check',
      resolutionDateTime: closeAt,
      answerLabels: ['Red', 'Blue', 'red'],
    },
  });
  check('duplicate answer labels are rejected', Boolean(duplicate));

  let details = await api('POST', '/market-groups', {
    token: modToken,
    expect: [201],
    body: {
      questionTitle: `Ponytail MCB ${stamp}`,
      description: 'stdlib integration run',
      resolutionDateTime: closeAt,
      answerLabels: ['Spain', 'France', 'Germany', 'Brazil'],
      autoApproveAnswerAdditions: true,
    },
  });
  groupId = details.group.id;
  check('group created', groupId > 0, `groupId=${groupId}`);
  check('four child answers created', details.answers.length === 4);
  check('independent binary policy', details.group.probabilityPolicy === 'INDEPENDENT_BINARY');

  if (details.group.lifecycleStatus === 'proposed') {
    await api('PATCH', `/admin/market-groups/${groupId}/approve`, {
      token: adminToken,
      body: { confirm: true },
    });
    check('admin approved proposed group', true);
  }

  details = await api('GET', `/market-groups/${groupId}`);
  const answerIds = details.answers.map((answer) => answer.marketId);
  check('child titles include answer labels', details.answers.every((answer) => answer.market.market.questionTitle.includes(answer.answerLabel)));
  check('child proposal costs are zero', details.answers.every((answer) => (answer.market.market.proposalCost || 0) === 0));

  await api('POST', '/bet', { token: bettorToken, expect: [201], body: { marketId: answerIds[0], amount: 3, outcome: 'YES' } });
  await api('POST', '/bet', { token: bettorToken, expect: [201], body: { marketId: answerIds[1], amount: 2, outcome: 'YES' } });
  details = await api('GET', `/market-groups/${groupId}`);
  const latest = details.answers.map((answer) => answer.probabilityChanges.at(-1).probability);
  check('children trade independently', latest[0] !== latest[2] && latest[1] !== latest[2], JSON.stringify(latest));
  check('probabilities are not normalized', latest.reduce((sum, value) => sum + value, 0) > 1.1, `sum=${latest.reduce((sum, value) => sum + value, 0).toFixed(3)}`);

  const bets = await api('GET', `/market-groups/${groupId}/bets?limit=20&offset=0`);
  const positions = await api('GET', `/market-groups/${groupId}/positions?limit=20&offset=0`, { token: bettorToken });
  const leaderboard = await api('GET', `/market-groups/${groupId}/leaderboard?limit=20&offset=0`);
  check('grouped bets aggregate child markets', bets.total >= 2);
  check('grouped positions include bettor', positions.positions.some((row) => row.username === bettor));
  check('grouped leaderboard includes bettor', leaderboard.leaderboard.some((row) => row.username === bettor));

  const addition = await api('POST', `/market-groups/${groupId}/answers`, {
    token: modToken,
    expect: [200, 201],
    body: { answerLabel: `Italy ${stamp.slice(-4)}` },
  });
  check('steward answer addition auto-approved', addition.status === 'approved', addition.status);
  details = await api('GET', `/market-groups/${groupId}`);
  check('answer addition creates child market', details.answers.length === 5);

  const winner = details.answers[0].marketId;
  await api('POST', `/market-groups/${groupId}/resolve`, {
    token: modToken,
    body: { mode: 'exclusive_yes', winningMarketId: winner },
  });
  details = await api('GET', `/market-groups/${groupId}`);
  const outcomes = details.answers.map((answer) => answer.market.market.resolutionResult);
  check('exclusive resolution resolves one YES', outcomes.filter((x) => x === 'YES').length === 1 && outcomes.filter((x) => x === 'NO').length === outcomes.length - 1, JSON.stringify(outcomes));
  check('parent group resolved', details.group.lifecycleStatus === 'resolved');
}

async function finish() {
  await mkdir(dirname(artifact), { recursive: true });
  await writeFile(artifact, `${JSON.stringify({ groupId, results }, null, 2)}\n`);
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
