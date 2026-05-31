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

function parseCSV(file) {
  const raw = fs.readFileSync(file, 'utf8').trim();
  if (!raw) return { header: [], rows: [] };
  const lines = raw.split(/\r?\n/).filter(Boolean);
  const header = lines.shift().split(',');
  const rows = lines.map((line) => {
    const values = line.split(',');
    return Object.fromEntries(header.map((key, index) => [key, values[index] ?? '']));
  });
  return { header, rows };
}

function numeric(row, key) {
  const value = Number.parseFloat(row[key]);
  return Number.isFinite(value) ? value : null;
}

function stats(rows, key) {
  const values = rows.map((row) => numeric(row, key)).filter((value) => value !== null);
  if (values.length === 0) return null;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const avg = values.reduce((sum, value) => sum + value, 0) / values.length;
  return {
    min: round(min),
    avg: round(avg),
    max: round(max),
  };
}

function minValue(rows, key) {
  const s = stats(rows, key);
  return s ? s.min : null;
}

function maxValue(rows, key) {
  const s = stats(rows, key);
  return s ? s.max : null;
}

function round(value) {
  if (value === null || value === undefined) return null;
  return Math.round(value * 100) / 100;
}

function printLine(label, value, unit = '') {
  if (value === null || value === undefined) return;
  console.log(`${label}: ${value}${unit}`);
}

const args = parseArgs(process.argv.slice(2));
if (!args.csv) {
  console.error('Usage: node loadtest/dossier/summarize-host.mjs --csv FILE [--profile FILE] [--out FILE]');
  process.exit(1);
}

const { rows } = parseCSV(args.csv);
if (rows.length === 0) {
  console.error(`No telemetry rows found: ${args.csv}`);
  process.exit(1);
}

const first = rows[0];
const last = rows[rows.length - 1];
const profile = args.profile ? JSON.parse(fs.readFileSync(args.profile, 'utf8')) : null;
const summary = {
  schemaVersion: '0.1.0',
  generatedAt: new Date().toISOString(),
  source: {
    csv: args.csv,
    profile: args.profile || null,
  },
  sampleCount: rows.length,
  startedAt: first.timestamp_utc || null,
  endedAt: last.timestamp_utc || null,
  host: {
    load1: stats(rows, 'load1'),
    load5: stats(rows, 'load5'),
    load15: stats(rows, 'load15'),
    cpuUserPct: stats(rows, 'cpu_user_pct'),
    cpuSystemPct: stats(rows, 'cpu_system_pct'),
    cpuIdlePct: stats(rows, 'cpu_idle_pct'),
    memTotalMiB: maxValue(rows, 'mem_total_mib'),
    memUsedMiB: stats(rows, 'mem_used_mib'),
    memAvailableMiB: stats(rows, 'mem_available_mib'),
    minMemAvailableMiB: minValue(rows, 'mem_available_mib'),
    swapUsedMiB: stats(rows, 'swap_used_mib'),
    diskUsedPct: stats(rows, 'disk_used_pct'),
    diskUsedMiB: stats(rows, 'disk_used_mib'),
    diskAvailableMiB: stats(rows, 'disk_available_mib'),
    diskReadKiBPerSec: stats(rows, 'disk_read_kib_per_sec'),
    diskWriteKiBPerSec: stats(rows, 'disk_write_kib_per_sec'),
    netRxKiBPerSec: stats(rows, 'net_rx_kib_per_sec'),
    netTxKiBPerSec: stats(rows, 'net_tx_kib_per_sec'),
  },
  docker: {
    cpuPctSum: stats(rows, 'docker_cpu_pct_sum'),
    memMiBSum: stats(rows, 'docker_mem_mib_sum'),
    containerCount: maxValue(rows, 'container_count'),
    backendCpuPct: stats(rows, 'backend_cpu_pct'),
    postgresCpuPct: stats(rows, 'postgres_cpu_pct'),
    traefikCpuPct: stats(rows, 'traefik_cpu_pct'),
  },
  profile,
};

if (args.out) {
  fs.mkdirSync(path.dirname(args.out), { recursive: true });
  fs.writeFileSync(args.out, `${JSON.stringify(summary, null, 2)}\n`);
}

console.log('Host telemetry summary');
console.log(`Samples: ${summary.sampleCount}`);
console.log(`Window: ${summary.startedAt} -> ${summary.endedAt}`);
printLine('Max CPU user', summary.host.cpuUserPct?.max, '%');
printLine('Max CPU system', summary.host.cpuSystemPct?.max, '%');
printLine('Min CPU idle', summary.host.cpuIdlePct?.min, '%');
printLine('Min RAM available', summary.host.minMemAvailableMiB, ' MiB');
printLine('Max RAM used', summary.host.memUsedMiB?.max, ' MiB');
printLine('Max disk used', summary.host.diskUsedPct?.max, '%');
printLine('Max disk read', summary.host.diskReadKiBPerSec?.max, ' KiB/s');
printLine('Max disk write', summary.host.diskWriteKiBPerSec?.max, ' KiB/s');
printLine('Max network RX', summary.host.netRxKiBPerSec?.max, ' KiB/s');
printLine('Max network TX', summary.host.netTxKiBPerSec?.max, ' KiB/s');
printLine('Max Docker CPU sum', summary.docker.cpuPctSum?.max, '%');
printLine('Max Docker RAM sum', summary.docker.memMiBSum?.max, ' MiB');
printLine('Max backend CPU', summary.docker.backendCpuPct?.max, '%');
printLine('Max Postgres CPU', summary.docker.postgresCpuPct?.max, '%');
printLine('Max Traefik CPU', summary.docker.traefikCpuPct?.max, '%');
if (summary.profile) {
  printLine('Host CPU count', summary.profile.host?.cpu_count);
  printLine('Host RAM total', summary.profile.host?.mem_total_mib, ' MiB');
  printLine('Docker CPU count', summary.profile.docker?.ncpu);
  printLine('Docker RAM total', summary.profile.docker?.mem_total_mib, ' MiB');
  console.log(`Explicit container CPU limits: ${summary.profile.docker?.containersWithCpuLimits ?? 0}/${summary.profile.containers?.length ?? 0}`);
  console.log(`Explicit container memory limits: ${summary.profile.docker?.containersWithMemoryLimits ?? 0}/${summary.profile.containers?.length ?? 0}`);
}
if (args.out) console.log(`Host summary JSON: ${args.out}`);
