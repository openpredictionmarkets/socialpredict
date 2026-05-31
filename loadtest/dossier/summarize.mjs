#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

function parseArgs(argv) {
  const args = {};
  for (let i = 0; i < argv.length; i += 1) {
    const value = argv[i];
    if (!value.startsWith('--')) continue;
    const key = value.slice(2);
    const next = argv[i + 1];
    if (!next || next.startsWith('--')) {
      args[key] = true;
    } else {
      args[key] = next;
      i += 1;
    }
  }
  return args;
}

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function metric(summary, name, value, fallback = 0) {
  const metricValue = summary.metrics?.[name];
  if (!metricValue) return fallback;
  return metricValue[value] ?? metricValue.values?.[value] ?? fallback;
}

function count(summary, name) {
  return metric(summary, name, 'count', 0);
}

function rate(summary, name) {
  return metric(summary, name, 'rate', metric(summary, name, 'value', 0));
}

function percentile(summary, name, p) {
  const fallback = p === 50 ? metric(summary, name, 'med', 0) : 0;
  return metric(summary, name, `p(${p})`, fallback);
}

function round(value) {
  return Math.round(value * 100) / 100;
}

function rateLimitEquivalentUsers(summary, metadata) {
  const successfulBetsPerSecond = rate(summary, 'sp_bets_succeeded');
  const normalPolicy = metadata.normalRateLimitPolicy || {
    profile: 'secure-default',
    loginRatePerSecond: 0.1,
    loginBurst: 3,
    generalRatePerSecond: 1,
    generalBurst: 10,
    cleanupInterval: '5m',
  };
  const generalRatePerSecond = Number(normalPolicy.generalRatePerSecond);
  const oneBetPerTenSeconds = 0.1;

  return {
    note: 'Rate-limit equivalent users are client identities/IPs for the current in-process limiter. They are user-equivalent only when each user has a distinct limiter identity.',
    formula: 'ceil(successful_bets_per_second / allowed_bets_per_second_per_client_identity)',
    measuredSuccessfulBetsPerSecond: round(successfulBetsPerSecond),
    normalRateLimitPolicy: normalPolicy,
    normalGeneralLimitClientIdentities: generalRatePerSecond > 0
      ? Math.ceil(successfulBetsPerSecond / generalRatePerSecond)
      : null,
    oneBetPerTenSecondsClientIdentities: Math.ceil(successfulBetsPerSecond / oneBetPerTenSeconds),
    assumptions: {
      normalGeneralLimitAllowedBetsPerSecondPerClientIdentity: generalRatePerSecond || null,
      oneBetPerTenSecondsAllowedBetsPerSecondPerClientIdentity: oneBetPerTenSeconds,
      burstIsIgnoredForSustainedHotWindowEstimate: true,
    },
  };
}

const args = parseArgs(process.argv.slice(2));
if (!args.summary || !args.out) {
  console.error('Usage: node loadtest/dossier/summarize.mjs --summary FILE --out FILE [--metadata FILE] [--host-summary FILE] [--decision VALUE]');
  process.exit(1);
}

const summary = readJSON(args.summary);
const metadata = args.metadata ? readJSON(args.metadata) : {};
const hostSummary = args['host-summary'] ? readJSON(args['host-summary']) : null;
const dossier = {
  schemaVersion: '0.1.0',
  generatedAt: new Date().toISOString(),
  release: metadata.release || process.env.RELEASE || 'unknown',
  environment: metadata.environment || process.env.ENVIRONMENT || 'unknown',
  baseUrl: metadata.baseUrl || process.env.BASE_URL || 'unknown',
  appTopology: metadata.appTopology || 'unknown',
  databaseTopology: metadata.databaseTopology || 'unknown',
  dropletSize: metadata.dropletSize || 'unknown',
  loadGenerator: metadata.loadGenerator || 'unknown',
  scenario: metadata.scenario || summary.root_group?.name || path.basename(args.summary, path.extname(args.summary)),
  seed: metadata.seed || {},
  traffic: metadata.traffic || {},
  results: {
    requestsTotal: count(summary, 'http_reqs'),
    iterationsTotal: count(summary, 'iterations'),
    checksRate: rate(summary, 'checks'),
    errorRate: rate(summary, 'http_req_failed'),
    httpReqDurationP50Ms: percentile(summary, 'http_req_duration', 50),
    httpReqDurationP95Ms: percentile(summary, 'http_req_duration', 95),
    httpReqDurationP99Ms: percentile(summary, 'http_req_duration', 99),
    betsAttempted: count(summary, 'sp_bets_attempted'),
    betsSucceeded: count(summary, 'sp_bets_succeeded'),
    betsFailed: count(summary, 'sp_bets_failed'),
    loginFailures: count(summary, 'sp_login_failures'),
    rateLimited: count(summary, 'sp_rate_limited'),
    loginRateLimited: count(summary, 'sp_login_rate_limited'),
  },
  rateLimitPolicy: metadata.rateLimitPolicy || {},
  rateLimitEquivalents: rateLimitEquivalentUsers(summary, metadata),
  infrastructureObservations: metadata.infrastructureObservations || {},
  hostTelemetry: hostSummary ? {
    ...(metadata.hostTelemetry || {}),
    summary: hostSummary,
  } : metadata.hostTelemetry || {},
  decision: args.decision || metadata.decision || 'inconclusive',
  knownRisks: metadata.knownRisks || [],
  source: {
    k6Summary: args.summary,
    hostSummary: args['host-summary'] || null,
    metadata: args.metadata || null,
  },
};

fs.mkdirSync(path.dirname(args.out), { recursive: true });
fs.writeFileSync(args.out, `${JSON.stringify(dossier, null, 2)}\n`);
console.log(`Wrote ${args.out}`);
