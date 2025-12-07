<script lang="ts">
	import TickerBar, { type TickerItem } from '$lib/components/TickerBar.svelte';
	import SectionHeader from '$lib/components/SectionHeader.svelte';
	import MarketCard, { type Market } from '$lib/components/MarketCard.svelte';
	import CategoryGrid, { type Category } from '$lib/components/CategoryGrid.svelte';
	import TutorialModal, { type TutorialStep } from '$lib/components/TutorialModal.svelte';
	import {
		Flag,
		ChartNoAxesCombined,
		Coins,
		MonitorSmartphone,
		Trophy,
		Sparkles,
		FlaskConical
	} from 'lucide-svelte';

	const tickerItems: TickerItem[] = [
		{ label: 'BTC above $100k?', yes: 41, no: 59, trend: 3 },
		{ label: 'New iPhone folds?', yes: 62, no: 38, trend: -1 },
		{ label: 'Taylor Swift wins AOTY?', yes: 55, no: 45, trend: 7 },
		{ label: 'Fed cuts rates this year?', yes: 48, no: 52, trend: 2 }
	];

	const trendingMarkets: Market[] = [
		{
			title: 'Will Apple release a foldable iPhone in 2025?',
			yes: 62,
			no: 38,
			community: '1,842 predictions',
			resolves: 'Oct 9 · Apple Event',
			trend: 12,
			liquidity: 'deep',
			category: 'Tech',
			sparkline: '0,20 10,18 20,16 30,12 40,15 50,10 60,12 70,9 80,11 90,7 100,10 110,6 120,9'
		},
		{
			title: 'Will BTC close above $90k on Dec 31?',
			yes: 41,
			no: 59,
			community: '9,231 predictions',
			resolves: 'Dec 31',
			trend: 4,
			liquidity: 'deep',
			category: 'Crypto',
			sparkline: '0,10 10,12 20,9 30,13 40,12 50,15 60,16 70,18 80,14 90,16 100,15 110,17 120,14'
		},
		{
			title: 'Will the US enter a technical recession in 2025?',
			yes: 33,
			no: 67,
			community: '4,105 predictions',
			resolves: 'Mar 31',
			trend: -2,
			liquidity: 'moderate',
			category: 'Economy',
			sparkline: '0,18 10,15 20,17 30,14 40,12 50,13 60,10 70,11 80,8 90,7 100,9 110,6 120,5'
		},
		{
			title: 'Will Team USA win the 2026 World Cup?',
			yes: 22,
			no: 78,
			community: '3,210 predictions',
			resolves: 'Jul 19',
			trend: 1,
			liquidity: 'thin',
			category: 'Sports',
			sparkline: '0,9 10,10 20,11 30,13 40,12 50,11 60,13 70,12 80,14 90,13 100,12 110,11 120,10'
		}
	];

	const endingSoon: Market[] = [
		{
			title: 'Will a new US stimulus bill pass by Friday?',
			yes: 58,
			no: 42,
			community: '2,018 predictions',
			resolves: 'Fri · 11:59 PM',
			trend: 6,
			liquidity: 'moderate',
			category: 'Politics'
		},
		{
			title: 'Will ETH flip BTC in market cap by year end?',
			yes: 17,
			no: 83,
			community: '6,523 predictions',
			resolves: 'Dec 31',
			trend: -3,
			liquidity: 'very-thin',
			category: 'Crypto'
		},
		{
			title: 'Will an AI-first phone ship over 5M units?',
			yes: 46,
			no: 54,
			community: '1,130 predictions',
			resolves: 'Nov 30',
			trend: 4,
			liquidity: 'thin',
			category: 'Tech'
		}
	];

	const categories: Category[] = [
		{ name: 'Politics', icon: Flag, color: '#888888' },
		{ name: 'Economy', icon: ChartNoAxesCombined, color: '#F6C947' },
		{ name: 'Crypto', icon: Coins, color: '#35E2D1' },
		{ name: 'Tech', icon: MonitorSmartphone, color: '#9F6BFF' },
		{ name: 'Sports', icon: Trophy, color: '#F6C947' },
		{ name: 'Culture', icon: Sparkles, color: '#FF6F61' },
		{ name: 'Science', icon: FlaskConical, color: '#35E2D1' }
	];

	const tutorialSteps: TutorialStep[] = [
		{
			title: '<strong>A prediction market works like this…</strong>',
			description: [
				'<p>Buy YES/NO shares on real-world events.</p><br/>',
				'<p>Share prices move with demand, reflecting the crowd’s collective forecast. This is the "<em>Wisdom of the Crowd</em>" in action.</p>'
			].join('\n'),
			note: [
				'<p>Tip: Share prices represent outcome probabilities. For example, if the market price for YES is 62¢, the market expects that there is a ~62% chance the market will resolve YES.</p>'
			].join('\n')
		},
		{
			title: 'Tour a market...',
			description: [
				{ key: 'category', label: 'Market type', detail: 'Which arena this market belongs to (Politics, Tech, Sports, etc).' },
        { key: 'resolution', label: 'Resolution time', detail: 'When the market will settle and pay out.' },
				{ key: 'question', label: 'Market question', detail: 'The exact phrasing of what resolves YES vs NO.' },
				{ key: 'prices', label: 'Prices', detail: 'YES/NO quotes show implied probability; they change with trades.' },
				{ key: 'trend', label: 'Trend', detail: 'Recent momentum signal to see which outcome is heating up.' },
				{ key: 'liquidity', label: 'Liquidity', detail: 'Indicates how well funded the market is and how many participants are active within it.' },
				{ key: 'community', label: 'Community', detail: 'How many predictions/comments—shows activity and interest.' },
				{ key: 'cta', label: 'Action', detail: 'How users place trades.' }
			],
			note: 'Hover over each bullet to highlight its feature on the market card.'
		},
		{
			title: 'See how prices change...',
			description: [
        '<p style="margin-bottom: 1rem !important;">All markets start at 50¢/50¢ representing equal probability for YES and NO outcomes.</p>',
				'<p style="margin-bottom: 1rem !important;">Drag the sliders to see how market prices react to trades by users.</p>',
				'<p style="margin-bottom: 1rem !important;">Notice how buying more shares of one outcome pushes its price up, while the opposing outcome\'s price drops accordingly.</p>',
				'<p>Prices are determined by supply and demand. The more users buy YES shares, the higher the price for YES becomes, indicating increased confidence in that outcome.</p>'
			].join('\n'),
			note: 'Tip: If you were actually trading, your actual cost would be set by the price at the time of your trade.'
		},
		{
			title: 'Watch the market resolve — rewards arrive instantly',
			description: [
				'<p style="margin-bottom: 1rem !important;">When an event settles (when its market "resolves"), shareholders are paid out.</p>',
				'<p style="margin-bottom: 1rem !important;">If you hold shares of the winning outcome, you receive 100% value for each share you own.</p>',
        '<p style="margin-bottom: 1rem !important;>If the market resolves to "YES", holders of YES shares are paid $1 (100% value) for each of their shares.</p>',
				'<p style="margin-bottom: 1rem !important;">Losing shares become worthless.</p>',
        '<p style="margin-bottom: 1rem !important;">Funds are credited to your account instantly, ready for withdrawal or reinvestment.</p>',
        '<p>Let\'s assume you bought your shares for the current price of your chosen outcome.</p>'
			].join('\n'),
			note: [
        'Press the "Resolve Market to YES" button to see how your payout is calculated based on your purchase price versus the resolution outcome.'
      ].join('\n')
		},
		{
			title: 'Join the community layer',
			description:
				'Follow top forecasters, join prediction threads, and see sentiment move in real time.'
		}
	];

	let showTutorial = false;
</script>

<main class="page">
	<section class="hero">
		<div class="hero__copy">
			<div class="tag">SocialPredict</div>
			<h1>
				Predict the future. <span>Influence the conversation.</span>
			</h1>
			<p>
				Trade YES/NO shares on real events with a social layer built in. Follow forecasters, watch
				the crowd move, and stay in sync with the signals that matter.
			</p>
			<div class="hero__actions">
				<button class="primary">Start Predicting</button>
				<button class="secondary" type="button" on:click={() => (showTutorial = true)}>
					Learn How It Works
				</button>
			</div>
			<div class="hero__stats">
				<div>
					<strong>218K</strong>
					<span>Predictions this week</span>
				</div>
				<div>
					<strong>1,204</strong>
					<span>Active markets</span>
				</div>
				<div>
					<strong>82%</strong>
					<span>Community accuracy</span>
				</div>
			</div>
		</div>
		<div class="hero__panel">
			<div class="panel__glass">
				<div class="panel__header">
					<div>
						<div class="eyebrow">Order Ticket</div>
						<h3>Will Apple release a foldable iPhone in 2025?</h3>
					</div>
					<div class="score">
						<div>SP Score</div>
						<strong>72</strong>
					</div>
				</div>
				<div class="panel__row">
					<div class="pill yes">YES 62¢</div>
					<div class="pill no">NO 38¢</div>
					<div class="pill trend up">▲ +12%</div>
				</div>
				<div class="slider">
					<div class="slider__label">Choose stake</div>
					<div class="slider__track">
						<div class="slider__fill" style="width: 64%"></div>
						<div class="slider__thumb" style="left: 64%"></div>
					</div>
					<div class="slider__values">
						<span>$25</span>
						<span>Potential payout $40.50</span>
					</div>
				</div>
				<button class="primary block">Confirm YES</button>
				<div class="panel__footer">
					<div>Maker fee 0.15%</div>
					<div>Depth preview · Live</div>
				</div>
			</div>
		</div>
	</section>

	<section class="ticker-wrap">
		<TickerBar items={tickerItems} />
	</section>

	<section class="section">
		<SectionHeader
			eyebrow="Trending Right Now"
			title="Markets moving with volume, velocity, and social buzz"
			subtitle="A pulse on what the SocialPredict crowd is talking about in real time."
			actionLabel="See all markets"
			actionHref="#"
		/>
		<div class="grid grid--markets">
			{#each trendingMarkets as market}
				<MarketCard {market} />
			{/each}
		</div>
	</section>

	<section class="section section--split">
		<div class="split__left">
			<SectionHeader
				eyebrow="Categories"
				title="Jump into any arena"
				subtitle="From macro to memes, pick a lane and see what the crowd thinks."
			/>
			<CategoryGrid {categories} />
		</div>
		<div class="split__right">
			<SectionHeader
				eyebrow="Ending Soon"
				title="Last chance to get in"
				subtitle="Markets about to resolve — catch the final moves."
			/>
			<div class="stack">
				{#each endingSoon as market}
					<MarketCard {market} />
				{/each}
			</div>
		</div>
	</section>

	<TutorialModal
		open={showTutorial}
		steps={tutorialSteps}
		onClose={() => (showTutorial = false)}
		onComplete={() => (showTutorial = true)}
	/>
</main>

<style>
	:global(body) {
		background: radial-gradient(circle at 10% 20%, rgba(159, 107, 255, 0.18), transparent 25%),
			radial-gradient(circle at 80% 0%, rgba(53, 226, 209, 0.18), transparent 22%),
			#0e0f14;
		color: #e6e4ff;
		font-family: 'Inter', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
	}

	.page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2.5rem 1.5rem 3rem;
		display: grid;
		gap: 2rem;
	}

	.hero {
		display: grid;
		grid-template-columns: 1.1fr 0.9fr;
		gap: 1.5rem;
		align-items: center;
	}

	.hero__copy h1 {
		font-size: clamp(2.2rem, 4vw, 3.3rem);
		line-height: 1.1;
		margin: 0.4rem 0;
		color: #f7f6ff;
	}

	.hero__copy h1 span {
		color: #9f6bff;
		text-shadow: 0 0 18px rgba(159, 107, 255, 0.55);
	}

	.hero__copy p {
		color: #c4c8df;
		font-size: 1.05rem;
		margin: 0.5rem 0 1.1rem;
	}

	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.4rem 0.9rem;
		border-radius: 999px;
		background: rgba(159, 107, 255, 0.14);
		color: #dcd9f4;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-weight: 700;
		border: 1px solid rgba(159, 107, 255, 0.35);
		box-shadow: 0 10px 25px rgba(159, 107, 255, 0.3);
	}

	.hero__actions {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		margin-bottom: 1.1rem;
	}

	button.primary,
	button.secondary {
		padding: 0.85rem 1.2rem;
		border-radius: 0.95rem;
		font-weight: 800;
		font-size: 1rem;
		letter-spacing: 0.01em;
		border: none;
		cursor: pointer;
		transition: transform 150ms ease, box-shadow 150ms ease, background 150ms ease, color 150ms ease;
	}

	button.primary {
		background: linear-gradient(135deg, #6a3ff5, #9f6bff);
		color: #0e0f14;
		box-shadow: 0 14px 36px rgba(159, 107, 255, 0.45);
	}

	button.primary.block {
		width: 100%;
		justify-content: center;
	}

	button.secondary {
		background: rgba(255, 255, 255, 0.06);
		color: #e6e4ff;
		border: 1px solid rgba(255, 255, 255, 0.08);
	}


	button.primary:hover {
		transform: translateY(-1px);
		box-shadow: 0 18px 40px rgba(159, 107, 255, 0.5);
	}

	button.secondary:hover {
		border-color: rgba(159, 107, 255, 0.45);
	}

	.hero__stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 0.85rem;
	}

	.hero__stats div {
		padding: 0.75rem 0.9rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 0.9rem;
	}

	.hero__stats strong {
		display: block;
		font-size: 1.4rem;
	}

	.hero__stats span {
		color: #a9aec9;
		font-size: 0.95rem;
	}

	.hero__panel {
		position: relative;
	}

	.panel__glass {
		position: relative;
		padding: 1.4rem;
		border-radius: 1.2rem;
		background: linear-gradient(135deg, rgba(106, 63, 245, 0.28), rgba(14, 15, 20, 0.9)),
			rgba(14, 15, 20, 0.9);
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 18px 45px rgba(0, 0, 0, 0.35), 0 0 0 1px rgba(159, 107, 255, 0.25);
		backdrop-filter: blur(8px);
		display: grid;
		gap: 1rem;
	}

	.panel__header {
		display: flex;
		justify-content: space-between;
		gap: 0.75rem;
		align-items: center;
	}

	.panel__header h3 {
		margin: 0.2rem 0 0;
		font-size: 1.05rem;
	}

	.score {
		text-align: right;
		color: #c4c8df;
	}

	.score strong {
		display: block;
		font-size: 1.8rem;
		color: #35e2d1;
		text-shadow: 0 0 12px rgba(53, 226, 209, 0.45);
	}

	.eyebrow {
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.8rem;
		color: #9f6bff;
	}

	.panel__row {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.6rem;
	}

	.pill {
		padding: 0.75rem 0.9rem;
		border-radius: 0.85rem;
		text-align: center;
		font-weight: 800;
		border: 1px solid rgba(255, 255, 255, 0.08);
	}

	.pill.yes {
		background: rgba(53, 226, 209, 0.1);
		color: #35e2d1;
	}

	.pill.no {
		background: rgba(255, 111, 97, 0.1);
		color: #ff8d7f;
	}

	.pill.trend.up {
		color: #35e2d1;
		background: rgba(53, 226, 209, 0.08);
	}

	.slider {
		display: grid;
		gap: 0.35rem;
	}

	.slider__label {
		color: #c4c8df;
		font-size: 0.95rem;
	}

	.slider__track {
		position: relative;
		height: 12px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.08);
		overflow: hidden;
	}

	.slider__fill {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, #6a3ff5, #35e2d1);
	}

	.slider__thumb {
		position: absolute;
		top: 50%;
		width: 18px;
		height: 18px;
		transform: translate(-50%, -50%);
		border-radius: 50%;
		background: #fff;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.35);
		border: 2px solid #6a3ff5;
	}

	.slider__values {
		display: flex;
		justify-content: space-between;
		color: #e6e4ff;
		font-weight: 700;
		font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
	}

	.panel__footer {
		display: flex;
		justify-content: space-between;
		color: #9ea3c1;
		font-size: 0.9rem;
	}

	.ticker-wrap {
		position: sticky;
		top: 0.75rem;
		z-index: 2;
	}

	.section {
		display: grid;
		gap: 1rem;
	}

	.grid {
		display: grid;
		gap: 1rem;
	}

	.grid--markets {
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
	}

	.section--split {
		grid-template-columns: 0.95fr 1.05fr;
	}

	.split__left,
	.split__right {
		display: grid;
		gap: 1rem;
		align-content: start;
	}

	.stack {
		display: grid;
		gap: 0.85rem;
	}

	@media (max-width: 960px) {
		.hero {
			grid-template-columns: 1fr;
		}

		.section--split {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 640px) {
		.page {
			padding: 1.8rem 1.2rem 2.6rem;
		}

		.panel__row {
			grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		}
	}
</style>
