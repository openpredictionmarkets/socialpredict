<script module lang="ts">
	import type { FeatureKey } from '$lib/components/DemoMarketCard.svelte';

	export type FeatureDescription = {
		key: FeatureKey;
		label: string;
		detail: string;
	};

	export type TutorialStep = {
		title: string;
		description: string | FeatureDescription[];
		note?: string;
	};

	export type Props = {
		open: boolean;
		steps: TutorialStep[];
		onClose?: () => void;
		onComplete?: () => void;
	};
</script>

<script lang="ts">
	'use runes';

	import DemoMarketCard from '$lib/components/DemoMarketCard.svelte';
	import MarketCard, { type Market } from '$lib/components/MarketCard.svelte';

let { open = false, steps = [], onClose, onComplete } = $props();

	const b = 40; // liquidity parameter for mock LMSR
	const baseYes = 50;
	const baseNo = 50;

	let index = $state(0);
	let completed = $state(false);
	let resolved = $state(false);
	let hoveredFeature = $state<FeatureKey | null>(null);

	let yesBuy = $state<number>(0);
	let noBuy = $state<number>(0);

	const mockMarket: Market = {
		title: 'Will Apple release a foldable iPhone in 2025?',
		yes: 62,
		no: 38,
		community: '1,842 predictions',
		resolves: 'Oct 9 · Apple Event',
		trend: 12,
		liquidity: 'deep',
		category: 'Tech',
		sparkline: '0,20 10,18 20,16 30,12 40,15 50,10 60,12 70,9 80,11 90,7 100,10 110,6 120,9'
	};

	const total = $derived(steps.length);
	const current = $derived(steps[index]);
	const description = $derived(current?.description);
	const priceYes = $derived(toCents(priceLmsr(baseYes + yesBuy, baseNo + noBuy)));
	const priceNo = $derived(100 - priceYes);
	const displayedMarket = $derived(
		resolved
			? { ...mockMarket, yes: 100, no: 0 }
			: { ...mockMarket, yes: priceYes, no: priceNo }
	);
	const costYes = $derived(tradeCost(yesBuy, 0));
	const costNo = $derived(tradeCost(0, noBuy));
	const activeSide = $derived(yesBuy > 0 ? 'yes' : noBuy > 0 ? 'no' : null);
	const shareCost = $derived(
		yesBuy > 0
			? Number(((yesBuy * priceYes) / 100).toFixed(2))
			: noBuy > 0
				? Number(((noBuy * priceNo) / 100).toFixed(2))
				: 0
	);
	const payoutYes = $derived(Number((yesBuy * 1).toFixed(2)));
	const netProfit = $derived(resolved ? Number((payoutYes - shareCost).toFixed(2)) : 0);
	const netProfitPercent = $derived(
		resolved && shareCost > 0 ? Number(((netProfit / shareCost) * 100).toFixed(2)) : 0
	);
	const payoutPositive = $derived(payoutYes > 0);

	$effect(() => {
		if (!open) {
			index = 0;
			completed = false;
			resolved = false;
			yesBuy = 0;
			noBuy = 0;
			hoveredFeature = null;
		}
	});

	function close() {
		index = 0;
		completed = false;
		onClose?.();
	}

	function next() {
		if (index < total - 1) {
			index += 1;
		} else {
			completed = true;
			onComplete?.();
		}
	}

	function back() {
		if (index === 0) return;
		index = Math.max(0, index - 1);
	}

	function resolveMarket() {
		if (resolved) return;
		resolved = true;
	}

	function resetResolution() {
		if (!resolved) return;
		resolved = false;
	}

	function priceLmsr(qYes: number, qNo: number) {
		const expYes = Math.exp(qYes / b);
		const expNo = Math.exp(qNo / b);
		return expYes / (expYes + expNo);
	}

	function toCents(value: number) {
		return Math.round(value * 100);
	}

	function cost(qYes: number, qNo: number) {
		return b * Math.log(Math.exp(qYes / b) + Math.exp(qNo / b));
	}

	function tradeCost(deltaYes: number, deltaNo: number) {
		const before = cost(baseYes, baseNo);
		const after = cost(baseYes + deltaYes, baseNo + deltaNo);
		return Math.max(0, after - before);
	}

	function handleYesInput(event: Event) {
		const value = Number((event.currentTarget as HTMLInputElement).value);
		yesBuy = value;
		noBuy = 0;
	}

	function handleNoInput(event: Event) {
		const value = Number((event.currentTarget as HTMLInputElement).value);
		noBuy = value;
		yesBuy = 0;
	}
</script>

{#if open}
	<div class="overlay" role="dialog" aria-modal="true" aria-label="SocialPredict tutorial">
		<div class="modal">
			<button class="close" aria-label="Close tutorial" type="button" onclick={close}>✕</button>

			{#if !completed}
				<div class="layout">
					<div class="instructions">
						<div class="eyebrow">Step {index + 1} of {total}</div>
						<h3 class="demo-title">{@html current?.title}</h3>
						{#if typeof description === 'string'}
							<p>{@html description}</p>
						{/if}
						{#if current?.note}
							<div class="note">{@html current.note}</div>
						{/if}

						{#if index === 1 && Array.isArray(description)}
							<ul class="feature-list">
								{#each description as feature}
									<li
										class={feature.key === hoveredFeature ? 'highlighted' : ''}
										onmouseenter={() => (hoveredFeature = feature.key)}
										onmouseleave={() => (hoveredFeature = null)}
									>
										<strong>{feature.label}:</strong>
										{feature.detail}
									</li>
								{/each}
							</ul>
						{/if}
					</div>

					<div class="demo">
						<div class="demo__label">Mock market</div>
						<div
							class={`card-shell ${resolved ? 'resolved' : ''}`}
							role="presentation"
							onmouseleave={() => (hoveredFeature = null)}
						>
							<DemoMarketCard
								market={displayedMarket}
								highlightFeature={index === 1 ? hoveredFeature : null}
								onFeatureHover={(key) => (hoveredFeature = key)}
							/>
							{#if resolved}
								<div class="resolved-badge">Resolved</div>
							{/if}
						</div>

						{#if index === 2}
							<div class="sliders">
								<div class="slider-row">
									<div class="slider-meta">
										<div class="tag yes">YES</div>
										<div class="value">{yesBuy} shares</div>
									</div>
									<input
										type="range"
										min="0"
										max="100"
										step="1"
										value={yesBuy}
										oninput={handleYesInput}
									/>
								</div>
								<div class="slider-row">
									<div class="slider-meta">
										<div class="tag no">NO</div>
										<div class="value">{noBuy} shares</div>
									</div>
									<input
										type="range"
										min="0"
										max="100"
										step="1"
										value={noBuy}
										oninput={handleNoInput}
									/>
								</div>
								<div class="pill">
									<span>Cost</span>
									{#if activeSide === 'yes'}
										<span>Cost to buy {yesBuy} shares YES: ${costYes.toFixed(2)}</span>
									{:else if activeSide === 'no'}
										<span>Cost to buy {noBuy} shares NO: ${costNo.toFixed(2)}</span>
									{:else}
										<span>Move a slider to see cost</span>
									{/if}
								</div>
							</div>
						{:else if index === 3}
							<div class="resolve-box">
								<div class="resolve-actions">
									<button class="primary" type="button" onclick={resolveMarket} disabled={resolved}>
										{resolved ? 'Market is Resolved' : 'Resolve Market to YES'}
									</button>
									<button
										class="ghost"
										type="button"
										onclick={resetResolution}
										disabled={!resolved}
									>
										Reset resolution
									</button>
								</div>
								<div class="resolve-meta">
									<div>You own: {yesBuy} YES / {noBuy} NO</div>
									{#if resolved}
										<div class="cost">What you paid for your shares: ${shareCost.toFixed(2)}</div>
									{/if}
									<div class={`payout ${payoutPositive ? 'up' : 'down'}`}>
										{#if resolved}
											Your Payout: ${payoutYes.toFixed(2)}
											<div>
												Net {netProfit >= 0 ? 'profit' : 'loss'}: ${netProfit.toFixed(2)} ({netProfitPercent}
												%)
											</div>
										{:else}
											<!-- Your projected payout if market is resolved to YES: ${payoutYes.toFixed(2)} -->
										{/if}
									</div>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<div class="progress">
					{#each steps as _, i}
						<div class={`dot ${i === index ? 'active' : ''}`}></div>
					{/each}
				</div>

				<div class="actions">
					{#if index > 0}
						<button class="secondary ghost" type="button" onclick={back}>Back</button>
					{/if}
					<button class="primary" type="button" onclick={next}>
						{index === total - 1 ? 'Finish' : 'Next'}
					</button>
				</div>
			{:else}
				<div class="layout">
					<div class="instructions">
						<div class="header">
							<div class="eyebrow">Ready to trade?</div>
							<h3>Join the SocialPredict crowd</h3>
							<p>
								You just walked through how to place a YES/NO order. Create an account to follow
								forecasters, join threads, and start predicting live markets.
							</p>
						</div>
						<div class="cta-row single">
							<button class="primary" type="button">Create Account</button>
						</div>
					</div>
				</div>
				<div class="actions">
					<button class="secondary" type="button" onclick={close}>Close</button>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.overlay {
		position: fixed;
		inset: 0;
		background: rgba(10, 11, 16, 0.65);
		backdrop-filter: blur(6px);
		display: grid;
		place-items: center;
		z-index: 20;
		padding: 1.5rem;
	}

	.modal {
		position: relative;
		width: min(78vw, 1100px);
		min-height: 60vh;
		padding: 1.6rem 1.9rem;
		border-radius: 1.2rem;
		background: radial-gradient(120% 120% at 0% 0%, rgba(159, 107, 255, 0.15), transparent 45%),
			radial-gradient(140% 140% at 100% 0%, rgba(53, 226, 209, 0.12), transparent 40%),
			#12131a;
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.45), 0 0 0 1px rgba(159, 107, 255, 0.18);
		display: grid;
		gap: 1.2rem;
	}

	.layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.2rem;
		align-items: start;
	}

	.close {
		position: absolute;
		top: 0.8rem;
		right: 0.8rem;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.12);
		color: #e6e4ff;
		border-radius: 50%;
		width: 36px;
		height: 36px;
		cursor: pointer;
		font-size: 1rem;
	}

	.instructions {
		display: grid;
		gap: 0.75rem;
		align-content: start;
	}

	.header {
		display: grid;
		gap: 0.35rem;
	}


	.eyebrow {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.35rem 0.7rem;
		border-radius: 999px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.78rem;
		color: #9f6bff;
		background: rgba(159, 107, 255, 0.14);
		border: 1px solid rgba(159, 107, 255, 0.3);
		width: fit-content;
	}

	h3 {
		margin: 0.1rem 0 0;
		color: #f7f6ff;
	}

  .demo-title {
    margin: 0.1rem 0 0;
    color: #f7f6ff;
    font-weight: 900;
  }

	p {
		margin: 0;
		color: #c4c8df;
	}

	.note {
		margin-top: 0.3rem;
		padding: 0.65rem 0.8rem;
		border-radius: 0.75rem;
		background: rgba(255, 255, 255, 0.04);
		border: 1px dashed rgba(255, 255, 255, 0.12);
		color: #e6e4ff;
	}

	.progress {
		display: flex;
		gap: 0.4rem;
	}

	.dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.12);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.dot.active {
		background: linear-gradient(135deg, #6a3ff5, #35e2d1);
		box-shadow: 0 0 12px rgba(159, 107, 255, 0.45);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.7rem;
		align-items: center;
	}

	.demo {
		display: grid;
		gap: 0.75rem;
		align-content: start;
	}

	.demo__label {
		font-size: 0.9rem;
		color: #c4c8df;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.feature-list {
		list-style: disc;
		padding-left: 1.3rem;
		display: grid;
		gap: 0.5rem;
		color: #dcd9f4;
	}

	.feature-list li {
		list-style: disc;
	}

	.feature-list li.highlighted {
		color: #35e2d1;
		text-shadow: 0 0 10px rgba(53, 226, 209, 0.35);
	}

	.feature-list strong {
		color: #f7f6ff;
	}

	.sliders {
		display: grid;
		gap: 0.85rem;
		margin-top: 0.5rem;
	}

	.slider-row {
		display: grid;
		gap: 0.35rem;
	}

	.slider-meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
		color: #e6e4ff;
		font-weight: 700;
	}

	.slider-row input[type='range'] {
		accent-color: #6a3ff5;
	}

	.slider-row input[type='range']::-webkit-slider-thumb {
		box-shadow: 0 0 0 6px rgba(159, 107, 255, 0.22);
	}

	.tag {
		padding: 0.25rem 0.6rem;
		border-radius: 0.7rem;
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.tag.yes {
		background: rgba(53, 226, 209, 0.14);
		color: #35e2d1;
		border: 1px solid rgba(53, 226, 209, 0.4);
	}

	.tag.no {
		background: rgba(255, 111, 97, 0.12);
		color: #ff8d7f;
		border: 1px solid rgba(255, 111, 97, 0.35);
	}

	.pill {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 0.9rem;
		border-radius: 0.85rem;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		font-weight: 700;
		color: #f4f1ff;
	}

	.cta-row {
		display: flex;
		gap: 0.75rem;
	}

	.cta-row.single {
		justify-content: flex-start;
	}

	.card-shell {
		position: relative;
	}

	.resolved .resolved-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.35rem 0.7rem;
		border-radius: 0.9rem;
		background: rgba(53, 226, 209, 0.18);
		color: #0e0f14;
		font-weight: 800;
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.25);
	}

	.resolve-box {
		display: grid;
		gap: 0.55rem;
		align-content: start;
	}

	.resolve-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.resolve-box .primary {
		width: fit-content;
	}

	.resolve-meta {
		display: grid;
		gap: 0.2rem;
		color: #dcd9f4;
		font-weight: 700;
	}

	.cost {
		color: #a9aec9;
	}

	.payout.up {
		color: #35e2d1;
	}

	.payout.down {
		color: #ff8d7f;
	}

	button.primary,
	button.secondary,
	button.ghost {
		padding: 0.85rem 1.2rem;
		border-radius: 0.95rem;
		font-weight: 800;
		font-size: 1rem;
		letter-spacing: 0.01em;
		border: none;
		cursor: pointer;
	}

	button.primary {
		background: linear-gradient(135deg, #6a3ff5, #9f6bff);
		color: #0e0f14;
		box-shadow: 0 14px 36px rgba(159, 107, 255, 0.45);
	}

	button.secondary {
		background: rgba(255, 255, 255, 0.06);
		color: #e6e4ff;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	button.ghost {
		background: transparent;
		color: #e6e4ff;
		border: 1px solid rgba(255, 255, 255, 0.14);
	}

	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		box-shadow: none;
	}

	button.primary:hover {
		transform: translateY(-1px);
	}

	button.secondary:hover,
	button.ghost:hover {
		border-color: rgba(159, 107, 255, 0.45);
	}

	@media (max-width: 900px) {
		.modal {
			width: min(95vw, 760px);
		}

		.layout {
			grid-template-columns: 1fr;
		}

		.cta-row {
			flex-direction: column;
		}

		.actions {
			flex-direction: column;
			align-items: stretch;
		}

		.actions button {
			width: 100%;
		}
	}
</style>
