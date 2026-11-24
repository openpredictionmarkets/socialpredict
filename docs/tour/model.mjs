// Port of weightedAverageProbabilityFormula.Calculate in
// backend/internal/domain/math/probabilities/wpam/wpam_marketprobabilities.go.
// Seeds and fees match backend/setup/setup.yaml. This is browser-only demo state.
export const probability = (yes, no) => (5 + yes) / (10 + yes + no);
export const freshTrade = () => ({ yes: 0, no: 0, credit: 500, history: [] });
export function selectMarket(book, currentId, current, nextId) {
  const updatedBook = { ...book, [currentId]: current };
  return { book: updatedBook, trade: { ...(updatedBook[nextId] || freshTrade()), credit: current.credit } };
}
function validateOrder(amount, outcome) {
  if (!Number.isSafeInteger(amount) || amount < 1) throw new Error('Enter a whole number of credits, at least 1.');
  if (!['YES', 'NO'].includes(outcome)) throw new Error('Choose YES or NO.');
}
export function project(state, amount, outcome) {
  validateOrder(amount, outcome);
  return probability(state.yes + (outcome === 'YES' ? amount : 0), state.no + (outcome === 'NO' ? amount : 0));
}
export function buy(state, amount, outcome) {
  const p = project(state, amount, outcome);
  const fee = state.history.length ? 0 : 1;
  if (amount + fee > state.credit) throw new Error('You need enough available credits for the order and fee.');
  return { yes: state.yes + (outcome === 'YES' ? amount : 0), no: state.no + (outcome === 'NO' ? amount : 0),
    credit: state.credit - amount - fee, history: [...state.history, { amount, outcome, probability: p, fee }] };
}
export const topics = [
  { slug: 'useful-forecasts', name: 'Useful Forecasts' },
  { slug: 'research-replication', name: 'Research & Replication' },
  { slug: 'community-experiments', name: 'Community Experiments' },
];
export const catalogue = [
  { id: 'decisions', title: 'Will at least three meetup forecasts inform a recorded decision?', tags: ['useful-forecasts', 'community-experiments'], status: 'active', probability: .5, type: 'Binary', close: '31 Dec 2026' },
  { id: 'replication', title: 'Will an independent team reproduce the pilot result?', tags: ['research-replication', 'useful-forecasts'], status: 'active', probability: .5, type: 'Binary', close: '31 Dec 2026' },
  { id: 'return', title: 'Will hosts return for a second forecasting round?', tags: ['community-experiments'], status: 'active', probability: .5, type: 'Binary', close: '31 Dec 2026' },
  { id: 'experiments', title: 'Which low-effort meetup experiments will we repeat?', tags: ['useful-forecasts', 'community-experiments'], status: 'active', probability: null, type: '3 binary answers', close: '31 Dec 2026' },
];
export function discover(markets, query = '', topic = 'all', lifecycle = 'active') {
  return markets.filter(m => (topic === 'all' || m.tags.includes(topic)) &&
    (lifecycle === 'all' || m.status === lifecycle) && m.title.toLowerCase().includes(query.trim().toLowerCase()));
}
export function review(state, action, reason) {
  if (state.status !== 'proposed') throw new Error('This proposal already has a decision.');
  if (!['approve', 'reject'].includes(action)) throw new Error('Choose an approval decision.');
  if (action === 'reject' && !reason.trim()) throw new Error('Give the proposer a reason.');
  return { ...state, status: action === 'approve' ? 'published' : 'rejected',
    credit: state.credit + (action === 'reject' ? 10 : 0),
    log: [...state.log, { action, reason: reason.trim() || 'Measurable criteria and evidence cutoff checked.', refund: action === 'reject' ? 10 : 0 }] };
}
// Appendix C's recorded sequence, not a general-purpose sale calculation.
export function replaySale(stage, action) {
  const transitions = { first: ['activity', 'activity'], activity: ['quote', 'quoted'], quoted: ['sell', 'sold'] };
  const next = transitions[stage];
  if (!next || next[0] !== action) throw new Error('Complete the preceding replay step first.');
  return next[1];
}
