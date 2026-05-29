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
  return summary.metrics?.[name]?.values?.[value] ?? fallback;
}

function count(summary, name) {
  return metric(summary, name, 'count', 0);
}

function rate(summary, name) {
  return metric(summary, name, 'rate', 0);
}

function percentile(summary, name, p) {
  return metric(summary, name, `p(${p})`, 0);
}

const args = parseArgs(process.argv.slice(2));
if (!args.summary || !args.out) {
  console.error('Usage: node loadtest/dossier/summarize.mjs --summary FILE --out FILE [--metadata FILE] [--decision VALUE]');
  process.exit(1);
}

const summary = readJSON(args.summary);
const metadata = args.metadata ? readJSON(args.metadata) : {};
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
  },
  infrastructureObservations: metadata.infrastructureObservations || {},
  decision: args.decision || metadata.decision || 'inconclusive',
  knownRisks: metadata.knownRisks || [],
  source: {
    k6Summary: args.summary,
    metadata: args.metadata || null,
  },
};

fs.mkdirSync(path.dirname(args.out), { recursive: true });
fs.writeFileSync(args.out, `${JSON.stringify(dossier, null, 2)}\n`);
console.log(`Wrote ${args.out}`);
