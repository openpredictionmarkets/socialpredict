import { freshTrade, buy, project, probability, topics, catalogue, discover, review, replaySale, selectMarket } from './model.mjs';
import { sources } from './sources.mjs';

const $ = s => document.querySelector(s);
const esc = value => String(value).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const percent = n => `${Math.round(n * 100)}%`;
const chapters = [
  ['overview', 'The whole round', 'Overview'], ['community', 'Make it yours', 'Administrator'],
  ['create', 'Write a question', 'Moderator'], ['review', 'Review & publish', 'Administrator'],
  ['discover', 'Find your question', 'Participant'], ['trade', 'Place a forecast', 'Participant'],
  ['sell', 'Preview a sale', 'Participant'], ['groups', 'Related markets', 'Moderator & participant'],
  ['start', 'Run your own round', 'Self-hosting'],
];
const defaultQuestion = 'Will at least three meetup forecasts inform a recorded decision?';
const defaultCriteria = 'Resolve YES if three distinct entries in the decision log cite a forecast made before the decision. Each entry must name a decision owner and link the timestamped forecast. Evidence cutoff: 31 December 2026, 23:59 UTC. Otherwise resolve NO.';
function initialState() {
  return { configTab: 'home', home: { title: 'Workshop Forecast Hub', intro: 'Forecast milestones, compare evidence, and learn together.' },
    people: [{ name: 'hostalice', role: 'Moderator' }, { name: 'learnercara', role: 'Participant' }], metricsPublic: false, leaderboardPublic: true,
    share: { name: 'Workshop Forecast Hub', description: 'Forecast with our community.' },
    proposal: { title: defaultQuestion, criteria: defaultCriteria, type: 'binary', tags: ['useful-forecasts', 'community-experiments'], close: '2026-12-31T23:59', status: 'proposed', credit: 490, log: [] },
    tradeId: 'decisions', tradeBook: {}, viewingProposal: false, submitted: false, auto: false, reviewTab: 'publishing', amendment: false, amendmentReason: '', steward: 'hostalice', stewardshipReason: '',
    topic: 'all', query: '', lifecycle: 'active', trade: freshTrade(), outcome: 'YES', amount: 20, replay: 'first',
    selectedGroup: 0, groupSourceCount: 3, groupGovernanceOpen: false, answerPolicy: 'review', groupAnswers: ['Decision-linked forecasts', 'Weekly meetup questions', 'Replication forecast round'], pendingAnswer: '', notice: '' };
}
let state = initialState();
let index = Math.max(0, chapters.findIndex(c => c[0] === location.hash.slice(1)));
const main = $('#main');
function announce(message) { $('#announcement').textContent = message; }
function heading(title, subtitle) { return `<div class="slide-heading"><h1>${title}</h1><p>${subtitle}</p></div>`; }
function panelHead(title, meta = '') { return `<div class="panel-head"><span>${title}</span><span class="meta">${meta}</span></div>`; }
function screenshot(id, note = '') {
  const s = sources.find(s => s.id === id);
  return `<aside class="proof"><button class="capture-button" data-evidence="${id}" aria-label="Open original screenshot: ${esc(s.title)}"><img src="${s.image}" alt="SocialPredict running app: ${esc(s.title)}" width="1440" height="1050" loading="lazy"><span>Inspect the original screenshot <span aria-hidden="true">↗</span></span></button><p>${note || esc(s.title)}</p><button class="text-button" data-evidence="${id}">Browse appendix captures</button></aside>`;
}
function tagChips(tags) { return tags.map(t => `<span class="tag">${esc(topics.find(x => x.slug === t)?.name || t)}</span>`).join(''); }
function tabs(items, active, attr) { return `<div class="tabs" role="group">${items.map(([key, label]) => `<button ${attr}="${key}" aria-pressed="${key === active}">${label}</button>`).join('')}</div>`; }
function overview() {
  return `<section class="hero"><div class="hero-copy"><p class="eyebrow">Open source community forecasting</p><h1>Forecast together.<br><span>Learn together.</span></h1><p class="hero-description">Give your community a place to ask, predict, and compare evidence.</p><button class="primary large" data-go="community">Try the tour</button></div><div class="hero-visual"><div class="window-caption"><span>Workshop Forecast Hub</span><span>Original app capture</span></div><button class="hero-image" data-evidence="C1" aria-label="Inspect SocialPredict market catalogue screenshot"><img src="${sources.find(s => s.id === 'C1').image}" alt="SocialPredict catalogue with topics and published questions" width="1440" height="1050"></button></div></section><div class="round-path" aria-label="Tour workflow">${[['community', 'Organize', 'Set up your community'], ['create', 'Ask', 'Write measurable questions'], ['discover', 'Discover', 'Follow topics and search'], ['trade', 'Forecast', 'Trade with play credits']].map(([id, title, sub]) => `<button data-go="${id}"><strong>${title}</strong><span>${sub}</span></button>`).join('')}</div>`;
}
function community() {
  const selected = state.configTab;
  let content;
  if (selected === 'home') content = `<form id="home-form" class="stack"><label>Homepage title<input name="title" value="${esc(state.home.title)}" maxlength="80" required></label><label>Welcome text<textarea name="intro" maxlength="180" required>${esc(state.home.intro)}</textarea></label><button class="primary" type="submit">Save homepage</button></form><div class="live-preview"><span class="meta">Saved homepage preview</span><h3 id="home-preview-title">${esc(state.home.title)}</h3><p id="home-preview-intro">${esc(state.home.intro)}</p><div class="preview-rule"><strong>How to join</strong><p>Ask the host for an account. Forecast with play credits.</p></div></div>`;
  if (selected === 'people') content = `<form id="people-form" class="inline-form"><label>New participant<input name="username" placeholder="e.g. hostben" pattern="[A-Za-z0-9_]{3,24}" required maxlength="24"></label><button class="primary" type="submit">Add account</button></form><div class="people-list">${state.people.map((p, i) => `<div class="person"><span class="avatar">${esc(p.name.slice(0, 1).toUpperCase())}</span><div><strong>${esc(p.name)}</strong><span>${p.role}</span></div>${p.role === 'Participant' ? `<button class="quiet" data-promote="${i}">Promote to moderator</button>` : '<span class="meta">Can propose questions</span>'}</div>`).join('')}</div><p class="small">The app generates initial credentials. This tour creates a sample roster entry.</p>`;
  if (selected === 'reporting') content = `<div class="stack"><label class="switch-row"><span><strong>System metrics</strong><small>Allow logged-out visitors</small></span><input id="metrics-public" type="checkbox" ${state.metricsPublic ? 'checked' : ''}></label><label class="switch-row"><span><strong>Global leaderboard</strong><small>Allow logged-out visitors</small></span><input id="leaderboard-public" type="checkbox" ${state.leaderboardPublic ? 'checked' : ''}></label><div class="live-preview"><span class="meta">Logged-out visitor preview</span><div class="visibility-row"><span>System metrics</span><strong>${state.metricsPublic ? 'Public' : 'Sign in required'}</strong></div><div class="visibility-row"><span>Leaderboard</span><strong>${state.leaderboardPublic ? 'Public' : 'Sign in required'}</strong></div></div></div>`;
  if (selected === 'sharing') content = `<form id="sharing-form" class="stack"><label>Site name<input name="name" value="${esc(state.share.name)}" maxlength="80" required></label><label>Link description<textarea name="description" maxlength="160" required>${esc(state.share.description)}</textarea></label><button class="primary" type="submit">Save sharing defaults</button></form><div class="share-preview"><span class="meta">Market link preview</span><h3>${esc(state.share.name)}</h3><p>${esc(state.share.description)}</p><small>Image metadata is optional in the app.</small></div>`;
  const proof = { home: 'A2', people: 'A3', reporting: 'A16', sharing: 'A17' }[selected];
  return heading('Make it your community.', 'Edit the homepage. Invite hosts. Choose who sees your reports.') + `<div class="split"><div class="panel">${panelHead('Platform settings', 'Administrator')}${tabs([['home', 'Homepage'], ['people', 'People'], ['reporting', 'Reporting'], ['sharing', 'Sharing']], selected, 'data-config')}<div class="panel-body config-content">${content}</div><p class="feedback" role="status">${esc(state.notice)}</p></div>${screenshot(proof, 'Try the controls, then compare with the running application.')}</div>`;
}
function create() {
  const p = state.proposal;
  return heading('Ask a question worth settling.', 'Define the evidence before anyone trades.') + `<div class="split"><div class="panel">${panelHead('Create a market', 'Moderator · 10-credit proposal cost')}<form id="question-form" class="panel-body stack"><label>Question type<select name="type"><option value="binary" ${p.type === 'binary' ? 'selected' : ''}>Binary: YES / NO</option><option value="group" ${p.type === 'group' ? 'selected' : ''}>Group: related binary markets</option></select></label><label>Question<input name="title" value="${esc(p.title)}" maxlength="180" required></label><label>Resolution criteria<textarea name="criteria" rows="3" required maxlength="1500">${esc(p.criteria)}</textarea></label><div class="form-two"><fieldset><legend>Topics</legend>${topics.map(t => `<label class="check-label"><input type="checkbox" name="tags" value="${t.slug}" ${p.tags.includes(t.slug) ? 'checked' : ''}>${esc(t.name)}</label>`).join('')}</fieldset><label>Trading close (UTC)<input name="close" type="datetime-local" value="${p.close}" required><small>Align trading close with the evidence cutoff.</small></label></div><div id="group-fields" ${p.type === 'group' ? '' : 'hidden'}><label>Answer options, one per line<textarea name="answers" rows="3">Decision-linked forecasts\nWeekly meetup questions\nReplication forecast round</textarea></label></div><div class="action-row"><button type="submit" class="primary">Submit proposal</button><span class="small">10 play credits; refund on rejection.</span></div>${state.submitted ? `<div class="result"><strong>${p.status === 'published' ? 'Published' : p.status === 'rejected' ? 'Rejected; proposal credits refunded' : 'Awaiting admin review'}</strong><span>${esc(p.title)}</span><button class="text-button" type="button" data-go="review">Open review queue</button></div>` : ''}</form></div>${screenshot(p.type === 'group' ? 'B5' : 'B2', 'Moderators select shared tags. Administrators maintain the vocabulary.')}</div>`;
}
function reviewSlide() {
  const p = state.proposal;
  let content;
  if (state.reviewTab === 'publishing') content = `<label class="switch-row"><span><strong>Auto-approve new proposals</strong><small>Applies to the next moderator submission</small></span><input id="auto-approve" type="checkbox" ${state.auto ? 'checked' : ''}></label><div class="proposal"><div class="action-row"><span class="status">${p.status}</span><span class="meta">Creator: hostalice · Steward: ${state.steward}</span></div><h3>${esc(p.title)}</h3><div class="tags">${tagChips(p.tags)}</div><details><summary>Read resolution criteria</summary><p>${esc(p.criteria)}</p></details></div><form id="review-form" class="stack"><label>Decision reason<input name="reason" placeholder="e.g. Add a publication URL and a clear deadline." maxlength="240" ${p.status !== 'proposed' ? 'disabled' : ''}></label><div class="action-row"><button class="yes" type="submit" name="action" value="approve" ${p.status !== 'proposed' ? 'disabled' : ''}>Approve & publish</button><button class="quiet" type="submit" name="action" value="reject" ${p.status !== 'proposed' ? 'disabled' : ''}>Reject & refund 10</button><button class="text-button" type="button" data-retry-review>Replay review</button></div></form><div class="audit"><span class="meta">Decision trail · proposer has ${p.credit} available credits</span>${p.log.map(l => `<p><strong>${l.action === 'approve' ? 'Administrator approved' : 'Administrator rejected'}</strong> ${esc(l.reason)}${l.refund ? ` <span class="positive">+${l.refund} credits refunded</span>` : ''}</p>`).join('') || '<p>Submit a decision to add its reason here.</p>'}</div>`;
  if (state.reviewTab === 'amendments') content = `<div class="contract-block"><span class="meta">Original contract · preserved</span><p>${esc(defaultCriteria)}</p></div><div class="contract-block amendment"><span class="meta">Proposed addition · hostalice</span><p>Count each distinct decision once, even if it cites several forecasts.</p></div><form id="amendment-form" class="stack"><label>Review reason<input name="reason" value="${esc(state.amendmentReason)}" required placeholder="Explain the clarification" maxlength="240"></label><button class="primary" ${state.amendment ? 'disabled' : ''}>Approve clarification</button></form>${state.amendment ? `<div class="audit"><strong>Approved amendment history</strong><p>Administrator: ${esc(state.amendmentReason)}</p><p>Original criteria retained. Clarification appended.</p></div>` : ''}`;
  if (state.reviewTab === 'stewardship') content = `<div class="handoff"><div><span class="avatar">A</span><strong>hostalice</strong><small>Original creator</small></div><span class="handoff-line" aria-hidden="true"></span><div><span class="avatar">${state.steward === 'hostalice' ? 'A' : 'B'}</span><strong>${state.steward}</strong><small>Current steward</small></div></div><form id="steward-form" class="stack"><label>New steward<select name="steward"><option>hostben</option><option>hostalice</option></select></label><label>Reason<input name="reason" required maxlength="240" placeholder="Holiday coverage"></label><button class="primary">Transfer stewardship</button></form>${state.stewardshipReason ? `<div class="audit">Administrator reassigned to ${state.steward}: ${esc(state.stewardshipReason)}</div>` : ''}`;
  const proof = { publishing: 'A6', amendments: state.amendment ? 'A10' : 'A9', stewardship: 'A13' }[state.reviewTab];
  return heading('Publish with a decision trail.', 'Approve a proposal, return it with a reason, or hand over stewardship.') + `<div class="split"><div class="panel">${panelHead('Market governance', 'Administrator')}${tabs([['publishing', 'Publishing'], ['amendments', 'Clarifications'], ['stewardship', 'Stewardship']], state.reviewTab, 'data-review-tab')}<div class="panel-body stack">${content}<p class="feedback" role="status">${esc(state.notice)}</p></div></div>${screenshot(proof)}</div>`;
}
function visibleCatalogue() {
  const markets = state.submitted && state.proposal.status === 'published'
    ? [{ id: 'custom', title: state.proposal.title, tags: state.proposal.tags, status: 'active', probability: state.proposal.type === 'group' ? null : .5, type: state.proposal.type === 'group' ? `${state.groupAnswers.length} binary answers` : 'Binary', close: state.proposal.close.slice(0, 10) }, ...catalogue]
    : catalogue;
  return markets.map(m => {
    const trades = m.id === state.tradeId ? state.trade : state.tradeBook[m.id];
    return m.probability !== null && trades ? { ...m, probability: probability(trades.yes, trades.no) } : m;
  });
}
function marketRows() {
  const rows = discover(visibleCatalogue(), state.query, state.topic, state.lifecycle);
  return rows.length ? rows.map(m => `<tr><td><button class="market-link" data-go="${m.id === 'experiments' || (m.id === 'custom' && state.proposal.type === 'group') ? 'groups' : 'trade'}" data-market="${m.id}">${esc(m.title)}</button><div class="tags">${tagChips(m.tags)}</div></td><td><strong class="probability-cell">${m.probability === null ? `${state.groupAnswers.length} answers` : percent(m.probability)}</strong><span class="meta">${m.probability === null ? `${state.groupAnswers.length} binary answers` : m.type}</span></td><td class="close-cell">${m.close}</td></tr>`).join('') : '<tr><td colspan="3" class="empty">No matching questions. Clear the search or choose another scope.</td></tr>';
}
function discovery() {
  return heading('Find your question.', 'Choose a topic. Search inside it. Keep the same question in several topics.') + `<div class="panel discovery-panel">${panelHead('Markets', 'Participant')}${tabs([['all', 'All topics'], ...topics.map(t => [t.slug, esc(t.name)])], state.topic, 'data-topic')}<div class="panel-body"><div class="search-row"><label>Search questions<input type="search" id="market-search" value="${esc(state.query)}" placeholder="Try “independent” or “meetup”"></label><label>Market status<select id="lifecycle">${['active', 'closed', 'resolved', 'all'].map(s => `<option value="${s}" ${s === state.lifecycle ? 'selected' : ''}>${s[0].toUpperCase() + s.slice(1)}</option>`).join('')}</select></label></div><p id="search-scope" class="small">${scopeText()}</p><div class="table-wrap"><table><thead><tr><th>Question</th><th>YES probability</th><th>Closes</th></tr></thead><tbody id="market-results">${marketRows()}</tbody></table></div><p id="search-count" class="small" role="status">${discover(visibleCatalogue(), state.query, state.topic, state.lifecycle).length} questions</p></div></div><div class="evidence-strip"><span>Shared tags, topic navigation, and scoped search in the app.</span><button class="quiet" data-evidence="C4">Inspect scoped search</button><button class="quiet" data-evidence="B9">Inspect topic setup</button></div>`;
}
function scopeText() { return `Searching ${state.lifecycle === 'all' ? 'all' : state.lifecycle} questions ${state.topic === 'all' ? 'across all topics' : `tagged ${topics.find(t => t.slug === state.topic).name}`}.`; }
function chart(values, projected) {
  const all = [...values, ...(projected == null ? [] : [projected])];
  const pts = all.map((v, i) => `${45 + i * (500 / Math.max(1, all.length - 1))},${205 - v * 170}`);
  return `<svg class="probability-chart" viewBox="0 0 590 245" role="img" aria-label="YES probability history: ${values.map(percent).join(', ')}${projected == null ? '' : `. Projected next value ${percent(projected)}`}"><g class="chart-grid">${[0, .5, 1].map(v => `<line x1="45" y1="${205 - v * 170}" x2="550" y2="${205 - v * 170}"/><text x="0" y="${210 - v * 170}">${percent(v)}</text>`).join('')}</g><polyline class="chart-line" points="${pts.slice(0, values.length).join(' ')}"/>${projected == null ? '' : `<line class="chart-projection" x1="${pts[pts.length - 2]?.split(',')[0] || 45}" y1="${pts[pts.length - 2]?.split(',')[1] || 120}" x2="${pts.at(-1).split(',')[0]}" y2="${pts.at(-1).split(',')[1]}"/>`}${all.map((v, i) => `<circle class="${i === all.length - 1 && projected != null ? 'projected-dot' : 'chart-dot'}" cx="${pts[i].split(',')[0]}" cy="${pts[i].split(',')[1]}" r="5"/>`).join('')}<text x="45" y="237">Market opened</text><text x="460" y="237">${projected == null ? 'Latest trade' : 'Your preview'}</text></svg>`;
}
function trading() {
  const question = state.viewingProposal ? state.proposal.title : defaultQuestion;
  const criteria = state.viewingProposal ? state.proposal.criteria : defaultCriteria;
  const t = state.trade; const current = probability(t.yes, t.no);
  let projected = null; try { projected = project(t, state.amount, state.outcome); } catch { /* Validation appears on submission. */ }
  return heading('See your forecast change the price.', 'Read the contract, choose an outcome, and check the fee before you confirm.') + `<div class="trade-grid"><div class="panel market-detail">${panelHead('Market detail', 'Participant')}<div class="panel-body"><div class="tags">${tagChips(state.viewingProposal ? state.proposal.tags : ['useful-forecasts', 'community-experiments'])}</div><h2>${esc(question)}</h2><div class="price-line"><div><strong>${percent(current)}</strong><span>Current YES probability</span></div><div><strong class="pink">${projected == null ? '…' : percent(projected)}</strong><span>After your ${state.outcome} order</span></div></div>${chart([.5, ...t.history.map(h => h.probability)], projected)}<details class="contract-details"><summary>Read the full contract</summary><p>${esc(criteria)}</p><p class="small">Synthetic evidence log. Trading close: ${state.viewingProposal ? esc(state.proposal.close.replace('T', ' ')) + ' UTC' : '31 Dec 2026, 23:59 UTC'}.</p></details><div class="activity"><strong>Recent activity</strong>${t.history.length ? t.history.slice(-3).reverse().map(h => `<div><span>learnercara bought ${h.amount} ${h.outcome}</span><span>${percent(h.probability)}</span></div>`).join('') : '<p class="small">Your first order will appear here.</p>'}</div></div></div><div class="panel order-panel">${panelHead('Purchase shares', 'Play credits')}<form id="trade-form" class="panel-body stack"><div class="balance"><span>Available credit</span><strong id="trade-credit">${t.credit}</strong></div><div class="outcomes" role="group" aria-label="Outcome">${['YES', 'NO'].map(o => `<button type="button" data-outcome="${o}" aria-pressed="${o === state.outcome}" class="${o === 'YES' ? 'yes' : 'no'}">${o}</button>`).join('')}</div><label>Order amount<input id="order-amount" name="amount" type="number" min="1" step="1" max="${t.credit}" required value="${state.amount}"></label><div class="order-terms"><div><span>Initial trade fee</span><strong>${t.history.length ? 0 : 1}</strong></div><div><span>Total credits</span><strong id="order-total">${Number.isSafeInteger(state.amount) ? state.amount + (t.history.length ? 0 : 1) : '…'}</strong></div><small>You pay the initial fee once per market.</small></div><button class="primary" type="submit">Confirm ${state.outcome} purchase</button><p class="feedback" id="trade-feedback" role="status">${esc(state.notice)}</p><button class="text-button" type="button" data-evidence="C6">Inspect the app’s trade preview</button></form></div></div>`;
}
function selling() {
  const stage = state.replay;
  const step = ['first', 'activity', 'quoted', 'sold'].indexOf(stage);
  const titles = ['You bought 20 YES', 'Another participant trades', 'You preview a five-credit sale', 'You confirm and check your portfolio'];
  const texts = ['You have 479 available credits after the one-credit initial fee. Your first position cannot yet fund a sale.', 'The recorded example adds 30 NO and then 20 YES from another synthetic account.', 'The recorded quote allows five shares for five credits, with zero dust. The app revalidates at confirmation.', 'You have 484 available credits and 33 YES shares in the recorded portfolio.'];
  const proof = ['C8', 'C9', 'C9', 'C11'][step];
  return heading('Check what you can sell.', 'Replay the recorded appendix trade. Owning shares does not guarantee an executable sale.') + `<div class="split"><div class="panel">${panelHead('Appendix C replay', 'Separate from your sandbox trades')}<div class="replay-steps">${['Buy', 'Later activity', 'Quote', 'Portfolio'].map((label, n) => `<span class="${n <= step ? 'complete' : ''}" ${n === step ? 'aria-current="step"' : ''}>${label}</span>`).join('')}</div><div class="panel-body stack"><h2>${titles[step]}</h2><p>${texts[step]}</p><div class="sale-numbers"><div><span>Available credits</span><strong>${stage === 'sold' ? 484 : 479}</strong></div><div><span>${stage === 'sold' ? 'YES shares' : 'Executable sale'}</span><strong>${stage === 'first' ? '0' : stage === 'sold' ? '33' : '5 credits'}</strong></div></div>${stage === 'first' ? '<div class="guard"><strong>Initial value cannot yet be sold</strong><p>Wait for later activity from another participant.</p></div>' : stage === 'sold' ? '<div class="result"><strong>Sale complete</strong><span>+5 credits · 5 shares sold · 0 dust</span><span>Remaining shares reflect the full market history.</span></div>' : `<dl class="quote"><div><dt>Sale order</dt><dd>5 credits</dd></div><div><dt>Shares sold</dt><dd>5</dd></div><div><dt>Credits per share</dt><dd>1</dd></div><div><dt>Dust</dt><dd>0</dd></div><div><dt>Quote status</dt><dd>${stage === 'quoted' ? 'Allowed' : 'Request a preview'}</dd></div></dl>`}<div class="action-row">${stage === 'first' ? '<button class="primary" data-replay="activity">Add recorded later trades</button><button class="quiet" disabled>Confirm sale</button>' : stage === 'activity' ? '<button class="primary" data-replay="quote">Preview sale</button>' : stage === 'quoted' ? '<button class="primary" data-replay="sell">Confirm five-credit sale</button>' : '<button class="primary" data-go="groups">Compare grouped questions</button>'}<button class="text-button" data-replay-reset>Restart replay</button></div><p class="small">Recorded fixtures, not a general sale calculator. No question was resolved in the walkthrough.</p></div></div>${screenshot(proof)}</div>`;
}
function groups() {
  const selected = state.groupAnswers[state.selectedGroup];
  const openingPrice = percent(probability(0, 0));
  return heading('Compare related markets.', 'Read price-implied probabilities together. Consider how the questions might connect.') + `<div class="split"><div class="panel">${panelHead('Community forecasting experiments', 'Separate YES / NO markets')}<div class="panel-body stack"><div class="group-context"><span class="meta">Shared question</span><h2>${state.groupSourceCount === 0 ? esc(state.proposal.title) : 'Which experiments will meet their success criteria?'}</h2></div><div class="group-markets" role="group" aria-label="Related markets">${state.groupAnswers.map((a, i) => `<button class="group-market" data-group-market="${i}" aria-pressed="${i === state.selectedGroup}"><span><strong>${esc(a)}</strong><small>${i < state.groupSourceCount ? 'Recorded YES probability' : 'Opening YES probability'}</small></span><span class="group-market-price">${openingPrice}</span></button>`).join('')}</div><div class="group-selected"><span class="meta">Selected market</span><h3>${esc(selected)}</h3><div class="group-quotes"><div><span>YES</span><strong>${openingPrice}</strong></div><div><span>NO</span><strong>${openingPrice}</strong></div></div><p class="small">Participants judge possible links. Each question keeps its own price, trades, and resolution.</p><button class="text-button" data-evidence="C12">Inspect the grouped price history</button></div><p class="small">Original capture: three questions at 50%, before any trades. Trading determines their later prices.</p><details class="group-governance" ${state.groupGovernanceOpen ? 'open' : ''}><summary>Moderator: add an answer</summary><form id="answer-form" class="stack"><label>New-answer policy<select id="answer-policy" name="policy"><option value="review" ${state.answerPolicy === 'review' ? 'selected' : ''}>Administrator review</option><option value="steward" ${state.answerPolicy === 'steward' ? 'selected' : ''}>Steward chooses per market</option><option value="auto" ${state.answerPolicy === 'auto' ? 'selected' : ''}>Auto-approve active moderators</option></select></label>${state.answerPolicy === 'steward' ? '<label class="check-label"><input type="checkbox" id="steward-auto" checked>Steward allows auto-approval for this group</label>' : ''}<label>Propose another answer<input name="answer" required maxlength="80" placeholder="e.g. Public health forecasting" ${state.groupAnswers.length >= 6 || state.pendingAnswer ? 'disabled' : ''}></label><div class="action-row"><button class="primary" ${state.groupAnswers.length >= 6 || state.pendingAnswer ? 'disabled' : ''}>Propose answer · 2 credits</button>${state.pendingAnswer ? '<button class="quiet" type="button" data-approve-answer>Approve queued answer</button>' : ''}</div><p class="feedback" role="status">${esc(state.notice)}</p>${state.pendingAnswer ? `<div class="audit">Awaiting administrator review: ${esc(state.pendingAnswer)}</div>` : ''}</form></details></div></div>${screenshot('C12', 'Participants compare related markets and their price histories in one view.')}</div>`;
}
function start() {
  return heading('Run a round of your own.', 'Self-host SocialPredict. Choose your economics, invite hosts, and publish your first questions.') + `<div class="start-grid"><div class="panel setup-panel">${panelHead('Your infrastructure', 'Docker · domain · database')}<div class="panel-body"><h2>Start from the repository.</h2><p>Follow the setup guide for prerequisites, secrets, and domain configuration.</p><pre><code>git clone https://github.com/openpredictionmarkets/socialpredict.git\ncd socialpredict\n./SocialPredict install\n./SocialPredict up</code></pre><div class="action-row"><a class="primary" href="https://github.com/openpredictionmarkets/socialpredict/blob/main/README/LOCAL_SETUP.md">Read the setup guide</a><a class="quiet" href="https://github.com/openpredictionmarkets/socialpredict">Explore the code</a></div><details><summary>Choose setup-time economics</summary><p>Set starting credit, debt allowance, proposal cost, subsidy, bonuses, and fees in <code>backend/setup/setup.yaml</code>.</p><p>Provision the instance and configure secrets before launch. The appendix demonstrates a self-hosted instance.</p></details></div></div><div class="evidence-index"><h2>Inspect the appendix evidence.</h2><p class="small">41 original captures from the running app. Synthetic accounts and questions.</p>${[['A1', 'Run a platform', 'Homepage, accounts, roles, publishing, reporting, sharing'], ['B1', 'Create & organize', 'Contracts, grouped questions, tags, topic navigation'], ['C1', 'Browse & trade', 'Search, purchase, sale guard, quotes, portfolio']].map(([id, title, sub]) => `<button data-evidence="${id}"><strong>${title}</strong><span>${sub}</span><span aria-hidden="true">↗</span></button>`).join('')}<details><summary>Scope and source notes</summary><p>Appendices A, B and C, 23 November 2026. Captured source: <a href="https://github.com/openpredictionmarkets/socialpredict/tree/d133d22751b21f38fa54acbfccb6db98e719fa9a">d133d22</a>.</p><p>Suspension and tag deactivation controls were captured without exercising them. Image uploads and social-crawler rendering were not tested. Topic navigation was exercised in Appendix B; optional chart pins were disabled after a display mismatch.</p><p>The browser examples run in memory. Reload or Reset tour clears them. Credit is a configured play-credit allowance; it is not cash. The appendices report no resolved outcomes or real-world adoption results.</p></details><button class="quiet" data-go="overview">Return to the beginning</button></div></div>`;
}
const views = [overview, community, create, reviewSlide, discovery, trading, selling, groups, start];
function render(focus = false) {
  const active = document.activeElement;
  const wasInside = main.contains(active);
  const focusAttribute = ['id', 'data-config', 'data-review-tab', 'data-topic', 'data-outcome', 'data-group-market'].find(a => active?.hasAttribute(a));
  const focusSelector = wasInside && focusAttribute ? `[${focusAttribute}="${CSS.escape(active.getAttribute(focusAttribute))}"]` : null;
  main.innerHTML = `<div class="slide" data-slide="${chapters[index][0]}">${views[index]()}</div>`;
  $('#rail').innerHTML = chapters.map((c, i) => `<a href="#${c[0]}" class="rail-item ${i === index ? 'active' : ''}" ${i === index ? 'aria-current="step"' : ''}><span>${String(i + 1).padStart(2, '0')}</span><span>${c[1]}</span></a>`).join('');
  $('#counter').textContent = `${String(index + 1).padStart(2, '0')} / 09`;
  $('#chapter-name').textContent = chapters[index][2];
  $('#previous').disabled = index === 0;
  $('#next').textContent = index === 8 ? 'Start again' : 'Next';
  document.title = `${chapters[index][1]} | SocialPredict`;
  if (!focus && wasInside) (focusSelector && main.querySelector(focusSelector) || main).focus({ preventScroll: true });
  if (focus) { main.focus({ preventScroll: true }); window.scrollTo({ top: 0, behavior: 'instant' }); announce(`Slide ${index + 1} of 9. ${chapters[index][1]}.`); }
}
function go(id) { const next = chapters.findIndex(c => c[0] === id); if (next < 0) return; state.notice = ''; if (location.hash === `#${id}`) { index = next; render(true); } else location.hash = id; }
addEventListener('hashchange', () => { index = Math.max(0, chapters.findIndex(c => c[0] === location.hash.slice(1))); state.notice = ''; render(true); });
$('#previous').addEventListener('click', () => go(chapters[Math.max(0, index - 1)][0]));
$('#next').addEventListener('click', () => go(chapters[(index + 1) % 9][0]));
$('#reset').addEventListener('click', () => { state = initialState(); go('overview'); announce('Tour reset. Sample changes cleared.'); });
$('#chapters').addEventListener('click', () => $('#chapter-dialog').showModal());
$('#close-chapters').addEventListener('click', () => $('#chapter-dialog').close());
$('#chapter-list').innerHTML = chapters.map((c, i) => `<a href="#${c[0]}" data-chapter="${c[0]}"><span>${i + 1}</span><strong>${c[1]}</strong><small>${c[2]}</small></a>`).join('');
$('#chapter-list').addEventListener('click', e => { const a = e.target.closest('a'); if (a) { $('#chapter-dialog').close(); go(a.dataset.chapter); } });
addEventListener('keydown', e => {
  if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey || document.querySelector('dialog[open]') || e.target.closest('input, textarea, select, button, summary, [contenteditable]')) return;
  if (['ArrowRight', 'ArrowLeft', 'Home', 'End'].includes(e.key)) {
    e.preventDefault(); const next = e.key === 'Home' ? 0 : e.key === 'End' ? 8 : Math.min(8, Math.max(0, index + (e.key === 'ArrowRight' ? 1 : -1)));
    go(chapters[next][0]);
  }
});
function showEvidence(id) {
  const s = sources.find(s => s.id === id) || sources[0];
  $('#capture-select').value = s.id; $('#capture-image').src = s.image; $('#capture-image').alt = `${s.id}: ${s.title}`;
  $('#full-image').href = s.image; $('#capture-title').textContent = `${s.id}. ${s.title}`;
  $('#capture-verification').textContent = s.verification; $('#capture-caption').textContent = s.caption; $('#capture-source').href = s.source;
  $('#capture-count').textContent = `${sources.indexOf(s) + 1} / ${sources.length}`;
  $('#previous-capture').disabled = sources.indexOf(s) === 0; $('#next-capture').disabled = sources.indexOf(s) === sources.length - 1;
  if (!$('#evidence').open) $('#evidence').showModal();
}
$('#capture-select').innerHTML = sources.map(s => `<option value="${s.id}">${s.id}. ${esc(s.title)}</option>`).join('');
$('#capture-select').addEventListener('change', e => showEvidence(e.target.value));
$('#close-evidence').addEventListener('click', () => $('#evidence').close());
for (const [selector, delta] of [['#previous-capture', -1], ['#next-capture', 1]]) $(selector).addEventListener('click', () => showEvidence(sources[sources.findIndex(s => s.id === $('#capture-select').value) + delta].id));

main.addEventListener('click', e => {
  const b = e.target.closest('button'); if (!b) return;
  if (b.dataset.evidence) { showEvidence(b.dataset.evidence); return; }
  if (b.dataset.go) {
    // Search examples link to their own contracts, never the unrelated binary demo.
    if (b.dataset.market && !['decisions', 'experiments', 'custom'].includes(b.dataset.market)) { showEvidence(b.dataset.market === 'replication' ? 'C2' : 'B4'); return; }
    if (b.dataset.go === 'trade') {
      const nextId = b.dataset.market === 'custom' ? 'custom' : 'decisions';
      const selected = selectMarket(state.tradeBook, state.tradeId, state.trade, nextId);
      state.tradeBook = selected.book; state.trade = selected.trade; state.tradeId = nextId;
      state.viewingProposal = nextId === 'custom';
    }
    go(b.dataset.go); return;
  }
  if (b.dataset.groupMarket !== undefined) { state.selectedGroup = Number(b.dataset.groupMarket); render(); }
  if (b.dataset.config) { state.configTab = b.dataset.config; state.notice = ''; render(); }
  if (b.dataset.promote !== undefined) { state.people[Number(b.dataset.promote)].role = 'Moderator'; state.notice = 'Sample role updated. This host can now propose questions.'; render(); }
  if (b.dataset.reviewTab) { state.reviewTab = b.dataset.reviewTab; state.notice = ''; render(); }
  if (b.hasAttribute('data-retry-review')) { state.proposal = { ...state.proposal, status: 'proposed', credit: 490, log: [] }; state.notice = ''; render(); }
  if (b.dataset.topic) { state.topic = b.dataset.topic; render(); }
  if (b.dataset.outcome) { const amount = $('#order-amount').valueAsNumber; state.amount = amount; state.outcome = b.dataset.outcome; state.notice = ''; render(); }
  if (b.dataset.replay) { state.replay = replaySale(state.replay, b.dataset.replay); render(); }
  if (b.hasAttribute('data-replay-reset')) { state.replay = 'first'; render(); }
  if (b.hasAttribute('data-approve-answer')) { addAnswer(state.pendingAnswer); state.pendingAnswer = ''; state.notice = 'Administrator approved. New binary child market added.'; render(); }
});
function addAnswer(label) { state.groupAnswers.push(label);  }
main.addEventListener('toggle', e => { if (e.target.matches('.group-governance')) state.groupGovernanceOpen = e.target.open; }, true);
main.addEventListener('change', e => {
  const t = e.target;
  if (t.id === 'metrics-public') { state.metricsPublic = t.checked; render(); }
  if (t.id === 'leaderboard-public') { state.leaderboardPublic = t.checked; render(); }
  if (t.id === 'auto-approve') { state.auto = t.checked; announce('Policy changed for the next proposal.'); }
  if (t.name === 'type') { $('#group-fields').hidden = t.value !== 'group'; }
  if (t.id === 'lifecycle') { state.lifecycle = t.value; updateSearch(); }
  if (t.id === 'answer-policy') { state.answerPolicy = t.value; state.notice = ''; render(); }
});
function updateSearch() { $('#market-results').innerHTML = marketRows(); $('#search-scope').textContent = scopeText(); $('#search-count').textContent = `${discover(visibleCatalogue(), state.query, state.topic, state.lifecycle).length} questions`; }
main.addEventListener('input', e => {
  const t = e.target;
  if (t.id === 'market-search') { state.query = t.value; updateSearch(); }
  if (t.id === 'order-amount') {
    state.amount = t.valueAsNumber;
    $('#order-total').textContent = Number.isSafeInteger(state.amount) ? state.amount + (state.trade.history.length ? 0 : 1) : '…';
    $('#trade-feedback').textContent = '';
    let p = null; try { p = project(state.trade, state.amount, state.outcome); } catch { /* Keep input editable. */ }
    $('.price-line .pink').textContent = p == null ? '…' : percent(p);
    $('.probability-chart').outerHTML = chart([.5, ...state.trade.history.map(h => h.probability)], p);
  }

});
main.addEventListener('submit', e => {
  e.preventDefault(); const form = e.target; const d = new FormData(form);
  try {
    if (form.id === 'home-form') { state.home = { title: d.get('title').trim(), intro: d.get('intro').trim() }; state.notice = 'Homepage saved in this browser preview.'; }
    if (form.id === 'sharing-form') { state.share = { name: d.get('name').trim(), description: d.get('description').trim() }; state.notice = 'Sharing defaults saved in this browser preview.'; }
    if (form.id === 'people-form') {
      const name = d.get('username').trim();
      if (state.people.some(p => p.name.toLowerCase() === name.toLowerCase())) throw new Error('That username is already in the sample roster.');
      state.people.push({ name, role: 'Participant' }); state.notice = `Added ${name} to the sample roster.`;
    }
    if (form.id === 'question-form') {
      if (!d.get('title').trim() || !d.get('criteria').trim()) throw new Error('Add a question and measurable criteria.');
      const answers = d.get('answers').split('\n').map(a => a.trim()).filter(Boolean);
      if (d.get('type') === 'group' && (answers.length < 2 || answers.length > 6)) throw new Error('This tour supports two to six answer options.');
      state.proposal = { title: d.get('title').trim(), criteria: d.get('criteria').trim(), close: d.get('close'), type: d.get('type'), tags: d.getAll('tags'), credit: 490, status: 'proposed', log: [] };
      if (state.auto) state.proposal = review(state.proposal, 'approve', 'Auto-approved under the platform policy.');
      if (d.get('type') === 'group') { state.groupAnswers = answers.slice(0, 6); state.selectedGroup = 0; state.groupSourceCount = 0; }
      delete state.tradeBook.custom;
      if (state.tradeId === 'custom') state.trade = { ...freshTrade(), credit: state.trade.credit };
      state.submitted = true;
    }
    if (form.id === 'review-form') { state.proposal = review(state.proposal, e.submitter.value, d.get('reason')); state.notice = state.proposal.status === 'published' ? 'Published in this simulation.' : 'Rejected. Proposal credits refunded in this simulation.'; }
    if (form.id === 'amendment-form') { if (!d.get('reason').trim()) throw new Error('Add a review reason.'); state.amendment = true; state.amendmentReason = d.get('reason').trim(); }
    if (form.id === 'steward-form') { if (!d.get('reason').trim()) throw new Error('Add a transfer reason.'); state.steward = d.get('steward'); state.stewardshipReason = d.get('reason').trim(); }
    if (form.id === 'trade-form') { state.trade = buy(state.trade, Number(d.get('amount')), state.outcome); state.notice = `Bought ${state.amount} ${state.outcome}. Available credit: ${state.trade.credit}.`; }
    if (form.id === 'answer-form') {
      const label = d.get('answer').trim(); if (!label) throw new Error('Enter an answer option.');
      if (state.groupAnswers.length >= 6 || state.pendingAnswer) throw new Error('Complete the pending answer review first.');
      if (state.answerPolicy === 'auto' || (state.answerPolicy === 'steward' && $('#steward-auto').checked)) { addAnswer(label); state.notice = 'Auto-approved. New binary child market added.'; }
      else { state.pendingAnswer = label; state.notice = 'Answer queued for administrator review.'; }
    }
    render(); announce(state.notice || 'Demo updated.');
  } catch (error) {
    state.notice = error.message;
    let feedback = form.querySelector('.feedback');
    if (!feedback) { feedback = document.createElement('p'); feedback.className = 'feedback error'; feedback.setAttribute('role', 'alert'); form.append(feedback); }
    feedback.textContent = error.message; announce(error.message);
  }
});
render();

document.querySelector('.skip-link').addEventListener('click', e => { e.preventDefault(); main.focus(); });
