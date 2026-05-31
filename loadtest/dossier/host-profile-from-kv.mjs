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

function numberOrString(value) {
  if (value === '') return null;
  if (/^-?\d+$/.test(value)) return Number.parseInt(value, 10);
  if (/^-?\d+\.\d+$/.test(value)) return Number.parseFloat(value);
  return value;
}

const args = parseArgs(process.argv.slice(2));
if (!args.input || !args.out) {
  console.error('Usage: node loadtest/dossier/host-profile-from-kv.mjs --input FILE --out FILE');
  process.exit(1);
}

const lines = fs.readFileSync(args.input, 'utf8').split(/\r?\n/).filter(Boolean);
const profile = {
  schemaVersion: '0.1.0',
  generatedAt: new Date().toISOString(),
  host: {},
  docker: {},
  containers: [],
};

for (const line of lines) {
  if (line.startsWith('container=')) {
    const raw = line.slice('container='.length);
    const [name, service, nanoCpus, cpuQuota, cpuPeriod, cpusetCpus, memoryBytes, memorySwapBytes] = raw.split('|');
    profile.containers.push({
      name: name || null,
      composeService: service || null,
      nanoCpus: numberOrString(nanoCpus || ''),
      cpuQuota: numberOrString(cpuQuota || ''),
      cpuPeriod: numberOrString(cpuPeriod || ''),
      cpusetCpus: cpusetCpus || null,
      memoryBytes: numberOrString(memoryBytes || ''),
      memorySwapBytes: numberOrString(memorySwapBytes || ''),
      cpuLimited: Number(nanoCpus || 0) > 0 || Number(cpuQuota || 0) > 0 || Boolean(cpusetCpus),
      memoryLimited: Number(memoryBytes || 0) > 0,
    });
    continue;
  }

  const index = line.indexOf('=');
  if (index < 0) continue;
  const key = line.slice(0, index);
  const value = line.slice(index + 1);
  const target = key.startsWith('docker_') ? profile.docker : profile.host;
  const cleanKey = key.replace(/^docker_/, '');
  target[cleanKey] = numberOrString(value);
}

profile.docker.containersWithCpuLimits = profile.containers.filter((container) => container.cpuLimited).length;
profile.docker.containersWithMemoryLimits = profile.containers.filter((container) => container.memoryLimited).length;
profile.docker.hasExplicitContainerCpuLimits = profile.docker.containersWithCpuLimits > 0;
profile.docker.hasExplicitContainerMemoryLimits = profile.docker.containersWithMemoryLimits > 0;

fs.mkdirSync(path.dirname(args.out), { recursive: true });
fs.writeFileSync(args.out, `${JSON.stringify(profile, null, 2)}\n`);
console.log(`Host profile JSON: ${args.out}`);
