#!/usr/bin/env node
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { spawnSync } from 'node:child_process';

const args = parseArgs(process.argv.slice(2));
const root = resolve(dirname(new URL(import.meta.url).pathname), '../..');
const baseUrl = args['base-url'] || process.env.BASE_URL || 'http://localhost:8080';
const apiPrefix = args['api-prefix'] || process.env.API_PREFIX || '/v0';
const schema = args.schema || process.env.SCHEMA || `${root}/backend/docs/openapi.yaml`;
const password = args.password || process.env.TEST_PASSWORD || 'password';
const moderator = args.moderator || process.env.MODERATOR_USER || 'testuser01';
const bettor = args.bettor || process.env.BETTOR_USER || 'testuser02';
const admin = args.admin || process.env.ADMIN_USER || 'admin';
const maxExamples = String(args['max-examples'] || process.env.MAX_EXAMPLES || '5');
const phases = args.phases || process.env.PHASES || 'fuzzing';
const seed = String(args.seed || process.env.SCHEMATHESIS_SEED || '289592472906227242590307531853135667664');
const rateLimit = args['rate-limit'] || process.env.SCHEMATHESIS_RATE_LIMIT || '1/s';
const schemathesisDelayMs = Number(args['schemathesis-delay-ms'] || process.env.SCHEMATHESIS_DELAY_MS || '5000');
const report = args.report || process.env.REPORT || '';
const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14);
const out = args.out || process.env.OUT || `${root}/integrationtest/artifacts/schemathesis-grouped-${stamp}`;
const results = [];
let activeGroupId = 0;
let naGroupId = 0;

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

function apiUrl(path) {
  return `${baseUrl.replace(/\/$/, '')}${apiPrefix.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
}

function unwrap(data) {
  return data?.ok === true && Object.hasOwn(data, 'result') ? data.result : data;
}

async function api(method, path, { body, token, expect = [200], retries = 2 } = {}) {
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
    if (expect.includes(response.status)) return unwrap(data);
    if (response.status === 429 && attempt < retries) {
      await new Promise((resolveRetry) => setTimeout(resolveRetry, 1200 * (attempt + 1)));
      continue;
    }
    throw new Error(`${method} ${url} -> ${response.status}: ${text.slice(0, 300)}`);
  }
  throw new Error(`${method} ${url} exhausted retries`);
}

function check(name, ok, detail = '') {
  results.push({ name, ok: Boolean(ok), detail });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` - ${detail}` : ''}`);
  if (!ok) throw new Error(name);
}

function assertLiveFreshness(label, payload) {
  const freshness = payload?.freshness;
  check(`${label} includes freshness`, Boolean(freshness), JSON.stringify(payload));
  check(`${label} freshness source is live`, freshness.source === 'live', JSON.stringify(freshness));
  check(`${label} freshness is not transaction-safe`, freshness.transactionSafeRead === false, JSON.stringify(freshness));
  check(`${label} freshness target is zero`, freshness.targetFreshnessSeconds === 0, JSON.stringify(freshness));
}

async function login(username) {
  return (await api('POST', '/login', { body: { username, password } })).token;
}

async function createApprovedGroup(token, adminToken, title, answerLabels) {
  const closeAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
  let details = await api('POST', '/market-groups', {
    token,
    expect: [201],
    body: {
      questionTitle: title,
      description: 'Schemathesis grouped-market integration scenario',
      resolutionDateTime: closeAt,
      answerLabels,
      autoApproveAnswerAdditions: true,
    },
  });
  if (details.group.lifecycleStatus === 'proposed') {
    await api('PATCH', `/admin/market-groups/${details.group.id}/approve`, {
      token: adminToken,
      body: { confirm: true },
    });
    details = await api('GET', `/market-groups/${details.group.id}`);
  }
  return details;
}

function forceGroupIDForPath(openapi, path, groupId) {
  const marker = `  ${path}:`;
  const start = openapi.indexOf(marker);
  if (start < 0) throw new Error(`OpenAPI path not found: ${path}`);
  let end = openapi.indexOf('\n  /', start + marker.length);
  if (end < 0) end = openapi.indexOf('\ncomponents:', start + marker.length);
  if (end < 0) end = openapi.length;

  const block = openapi.slice(start, end);
  const patched = block.replace(
    /schema:\n(\s*)type: integer\n\1format: int64\n\1minimum: 1/,
    `schema:\n$1type: integer\n$1format: int64\n$1enum: [${groupId}]`,
  );
  if (patched === block) throw new Error(`Could not patch id parameter for ${path}`);
  return `${openapi.slice(0, start)}${patched}${openapi.slice(end)}`;
}

async function writeConcreteSchema(groupId) {
  const groupedPaths = [
    '/v0/market-groups/{id}',
    '/v0/market-groups/{id}/bets',
    '/v0/market-groups/{id}/positions',
    '/v0/market-groups/{id}/leaderboard',
  ];
  let openapi = await readFile(schema, 'utf8');
  for (const path of groupedPaths) {
    openapi = forceGroupIDForPath(openapi, path, groupId);
  }
  const concreteSchema = `${out}/openapi-grouped-${groupId}.yaml`;
  await writeFile(concreteSchema, openapi);
  return { concreteSchema, groupedPaths };
}

async function pauseBeforeSchemathesis() {
  if (!Number.isFinite(schemathesisDelayMs) || schemathesisDelayMs <= 0) return;
  console.log(`Waiting ${schemathesisDelayMs}ms before Schemathesis to let local rate-limit buckets settle.`);
  await new Promise((resolveDelay) => setTimeout(resolveDelay, schemathesisDelayMs));
}

function runSchemathesis(concreteSchema, groupedPaths) {
  const cmd = 'schemathesis';
  const schemathesisArgs = [
    'run',
    concreteSchema,
    '--url',
    baseUrl,
    '--checks',
    'not_a_server_error,status_code_conformance,content_type_conformance,response_schema_conformance',
    '--phases',
    phases,
    '--max-examples',
    maxExamples,
    '--mode',
    'positive',
    '--generation-database',
    'none',
    '--request-timeout',
    '5',
    '--request-retries',
    '2',
    '--rate-limit',
    rateLimit,
    '--seed',
    seed,
    '--no-color',
  ];
  if (report) {
    schemathesisArgs.push('--report', report, '--report-dir', out);
    if (report.split(',').includes('junit')) {
      schemathesisArgs.push('--report-junit-path', `${out}/junit.xml`);
    }
  }
  for (const path of groupedPaths) {
    schemathesisArgs.push('--include-path', path);
  }

  console.log('Schemathesis grouped market contract run');
  console.log(`Base URL: ${baseUrl}`);
  console.log(`Schema: ${concreteSchema}`);
  console.log(`Group ID: ${activeGroupId}`);
  console.log(`Paths: ${groupedPaths.join(' ')}`);
  const result = spawnSync(cmd, schemathesisArgs, { stdio: 'inherit' });
  check('schemathesis grouped read contract passed', result.status === 0, `exit=${result.status}`);
}

async function main() {
  await mkdir(out, { recursive: true });

  const modToken = await login(moderator);
  const bettorToken = await login(bettor);
  const adminToken = await login(admin);
  check('login seeded moderator, bettor, admin', true);

  const spec = await readFile(schema, 'utf8');
  check('OpenAPI grouped answer cap is 50', spec.includes('maxItems: 50'));
  check('OpenAPI grouped resolve supports na mode', spec.includes('enum: [exclusive_yes, manual, na]'));

  const activeDetails = await createApprovedGroup(
    modToken,
    adminToken,
    `Schemathesis grouped active ${stamp}`,
    ['Alpha', 'Beta', 'Gamma'],
  );
  activeGroupId = activeDetails.group.id;
  check('active grouped market created', activeGroupId > 0, `groupId=${activeGroupId}`);

  const activeAnswerIDs = activeDetails.answers.map((answer) => answer.marketId);
  await api('POST', '/bet', { token: bettorToken, expect: [201], body: { marketId: activeAnswerIDs[0], amount: 3, outcome: 'YES' } });
  await api('POST', '/bet', { token: bettorToken, expect: [201], body: { marketId: activeAnswerIDs[1], amount: 2, outcome: 'NO' } });

  assertLiveFreshness('grouped bets', await api('GET', `/market-groups/${activeGroupId}/bets?limit=20&offset=0`));
  assertLiveFreshness('grouped positions', await api('GET', `/market-groups/${activeGroupId}/positions?limit=20&offset=0`, { token: bettorToken }));
  assertLiveFreshness('grouped leaderboard', await api('GET', `/market-groups/${activeGroupId}/leaderboard?limit=20&offset=0`));

  const naDetails = await createApprovedGroup(
    modToken,
    adminToken,
    `Schemathesis grouped NA ${stamp}`,
    ['Home', 'Away'],
  );
  naGroupId = naDetails.group.id;
  await api('POST', '/bet', { token: bettorToken, expect: [201], body: { marketId: naDetails.answers[0].marketId, amount: 4, outcome: 'YES' } });
  await api('POST', `/market-groups/${naGroupId}/resolve`, {
    token: modToken,
    body: { mode: 'na' },
  });
  const resolvedNA = await api('GET', `/market-groups/${naGroupId}`);
  const outcomes = resolvedNA.answers.map((answer) => answer.market.market.resolutionResult);
  check('group N/A resolves every child N/A', outcomes.every((outcome) => outcome === 'N/A'), JSON.stringify(outcomes));
  check('group N/A marks parent resolved', resolvedNA.group.lifecycleStatus === 'resolved', resolvedNA.group.lifecycleStatus);

  const { concreteSchema, groupedPaths } = await writeConcreteSchema(activeGroupId);
  await pauseBeforeSchemathesis();
  runSchemathesis(concreteSchema, groupedPaths);
}

async function finish() {
  await mkdir(out, { recursive: true });
  const failed = results.filter((result) => !result.ok);
  await writeFile(`${out}/summary.json`, `${JSON.stringify({
    activeGroupId,
    naGroupId,
    results,
  }, null, 2)}\n`);
  console.log(`Artifacts: ${out}`);
  console.log(`Summary: ${results.length - failed.length} passed, ${failed.length} failed`);
  process.exitCode = failed.length ? 1 : process.exitCode || 0;
}

try {
  await main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  results.push({ name: 'runner completed', ok: false, detail: message });
  console.error(`FAIL runner completed - ${message}`);
  process.exitCode = 1;
} finally {
  await finish();
}
