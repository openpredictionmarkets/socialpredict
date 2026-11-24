import test from 'node:test';
import assert from 'node:assert/strict';
import { freshTrade, buy, project, probability, discover, review, replaySale, catalogue, selectMarket } from '../model.mjs';

test('WPAM matches the appendix first purchase and later projection', () => {
  assert.equal(probability(0, 0), 0.5);
  assert.equal(project(freshTrade(), 20, 'YES'), 25 / 30);
  assert.equal(probability(40, 30), 0.5625);
  assert.equal(project({ yes: 35, no: 30 }, 5, 'YES'), 0.5625);
});
test('buy charges the initial fee once and keeps input state immutable', () => {
  const initial = freshTrade();
  const first = buy(initial, 20, 'YES');
  assert.equal(initial.credit, 500);
  assert.equal(first.credit, 479);
  assert.equal(first.history.length, 1);
  assert.equal(buy(first, 5, 'NO').credit, 474);
});
test('invalid or unaffordable trades cannot change credit', () => {
  for (const amount of [0, -1, NaN, Infinity, 1.5, 500]) {
    assert.throws(() => buy(freshTrade(), amount, 'YES'));
  }
  assert.throws(() => buy(freshTrade(), 10, 'MAYBE'));
});
test('switching questions preserves separate positions, initial fees and shared account credit', () => {
  const first = buy(freshTrade(), 20, 'YES');
  const secondMarket = selectMarket({}, 'decisions', first, 'custom');
  assert.equal(secondMarket.trade.credit, 479);
  assert.equal(secondMarket.trade.history.length, 0);
  assert.equal(probability(secondMarket.trade.yes, secondMarket.trade.no), .5);
  const second = buy(secondMarket.trade, 10, 'NO');
  assert.equal(second.credit, 468);
  const back = selectMarket(secondMarket.book, 'custom', second, 'decisions');
  assert.equal(back.trade.history.length, 1);
  assert.equal(back.trade.credit, 468);
  assert.equal(buy(back.trade, 5, 'YES').credit, 463);
});
test('discovery intersects text, topic and lifecycle; grouped question appears once', () => {
  assert.equal(discover(catalogue, 'independent', 'research-replication', 'active').length, 1);
  assert.equal(discover(catalogue, 'meetup', 'research-replication', 'active').length, 0);
  assert.equal(discover(catalogue, 'meetup', 'all', 'active').length, 2);
  assert.equal(discover(catalogue, '', 'all', 'closed').length, 0);
});
test('review requires a rejection reason and refunds the proposal once', () => {
  const proposed = { status: 'proposed', credit: 490, log: [] };
  assert.throws(() => review(proposed, 'reject', ''));
  const rejected = review(proposed, 'reject', 'Add an evidence URL.');
  assert.equal(rejected.credit, 500);
  assert.equal(rejected.status, 'rejected');
  assert.throws(() => review(rejected, 'reject', 'Again'));
  assert.equal(review(proposed, 'approve', '').status, 'published');
});
test('recorded replay requires later activity, quote and confirmation in order', () => {
  assert.throws(() => replaySale('first', 'sell'));
  assert.equal(replaySale('first', 'activity'), 'activity');
  assert.equal(replaySale('activity', 'quote'), 'quoted');
  assert.equal(replaySale('quoted', 'sell'), 'sold');
  assert.throws(() => replaySale('sold', 'sell'));
});
