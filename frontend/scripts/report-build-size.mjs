import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import { gzipSync } from 'node:zlib';

const buildDir = join(process.cwd(), 'build');

const formatKb = (bytes) => `${(bytes / 1024).toFixed(2)} kB`;

const walkFiles = (dir) => {
  const entries = readdirSync(dir);

  return entries.flatMap((entry) => {
    const fullPath = join(dir, entry);
    const stats = statSync(fullPath);

    if (stats.isDirectory()) {
      return walkFiles(fullPath);
    }

    return [fullPath];
  });
};

if (!existsSync(buildDir)) {
  console.error('Build directory not found. Run npm run build before reporting build size.');
  process.exit(1);
}

const files = walkFiles(buildDir)
  .map((filePath) => {
    const contents = readFileSync(filePath);

    return {
      path: relative(buildDir, filePath),
      bytes: contents.byteLength,
      gzipBytes: gzipSync(contents).byteLength,
    };
  })
  .sort((left, right) => right.bytes - left.bytes);

const totals = files.reduce(
  (memo, file) => ({
    bytes: memo.bytes + file.bytes,
    gzipBytes: memo.gzipBytes + file.gzipBytes,
  }),
  { bytes: 0, gzipBytes: 0 },
);

console.log('Frontend build size report');
console.log('Informational only: this report does not enforce a bundle budget.');
console.log('');
console.log('| File | Size | Gzip |');
console.log('| --- | ---: | ---: |');

files.forEach((file) => {
  console.log(`| ${file.path} | ${formatKb(file.bytes)} | ${formatKb(file.gzipBytes)} |`);
});

console.log('| --- | ---: | ---: |');
console.log(`| Total | ${formatKb(totals.bytes)} | ${formatKb(totals.gzipBytes)} |`);
