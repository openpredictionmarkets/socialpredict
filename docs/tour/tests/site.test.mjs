import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile, stat } from 'node:fs/promises';
import { sources } from '../sources.mjs';
const root = new URL('../', import.meta.url);

test('all 41 appendix captures ship locally at original dimensions with provenance', async () => {
  assert.equal(sources.length, 41);
  assert.equal(new Set(sources.map(s => s.id)).size, 41);
  let bytes = 0;
  for (const s of sources) {
    const source = new URL(s.source);
    assert.equal(source.origin, 'https://docs.google.com');
    assert.equal(source.pathname, '/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit');
    assert.ok(['t.3sqetzgfjt4c', 't.stbpowk1kp38', 't.z1m4er8t1vu2'].includes(source.searchParams.get('tab')));
    assert.match(s.verification, /^(Verified|Captured):/);
    assert.ok(s.caption.length > 100);
    const image = await readFile(new URL(s.image, root));
    assert.equal(image.subarray(1, 4).toString(), 'PNG');
    assert.equal(image.readUInt32BE(16), 1440);
    assert.equal(image.readUInt32BE(20), 1050);
    bytes += image.length;
  }
  assert.ok(bytes < 6 * 1024 * 1024, `Screenshot payload: ${bytes}`);
});

test('entry point has unique IDs and only local runtime assets', async () => {
  const html = await readFile(new URL('index.html', root), 'utf8');
  const ids = [...html.matchAll(/\bid="([^"]+)"/g)].map(m => m[1]);
  assert.equal(new Set(ids).size, ids.length, 'Duplicate HTML IDs');
  for (const match of html.matchAll(/<(?:script|link|img)\b[^>]*(?:src|href)="([^"]+)"/g)) {
    assert.doesNotMatch(match[1], /^(https?:|\/)/);
    assert.ok((await stat(new URL(match[1], root))).isFile());
  }
  assert.equal([...html.matchAll(/src="app.mjs"/g)].length, 1);
});
