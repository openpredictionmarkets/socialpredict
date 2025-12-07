<script lang="ts" context="module">
	export type FeatureKey =
		| 'category'
		| 'question'
		| 'prices'
		| 'trend'
		| 'liquidity'
		| 'community'
		| 'resolution'
		| 'cta';
</script>

<script lang="ts">
	import type { LiquidityLevel, Market } from './MarketCard.svelte';

	export let market: Market;
	export let highlightFeature: FeatureKey | null = null;
	export let onFeatureHover: ((key: FeatureKey | null) => void) | undefined = undefined;

	const liquidityCopy: Record<LiquidityLevel, string> = {
		deep: 'Deep',
		moderate: 'Moderate',
		thin: 'Thin',
		'very-thin': 'Very Thin'
	};

	const setHover = (key: FeatureKey | null) => {
		onFeatureHover?.(key);
	};
</script>

<article class={`card ${highlightFeature ? 'highlight-mode' : ''}`} aria-labelledby={`demo-${market.title}`}>
	{#if highlightFeature}
		<div class="blur-layer" aria-hidden="true"></div>
	{/if}
	<div class="content-layer">
		<div class="card__top">
			<div
				class={`pill ${highlightFeature === 'category' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('category')}
				onmouseleave={() => setHover(null)}
			>
				{market.category}
			</div>
			<div
				class={`resolves ${highlightFeature === 'resolution' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('resolution')}
				onmouseleave={() => setHover(null)}
			>
				Resolves {market.resolves}
			</div>
		</div>

		<h3
			id={`demo-${market.title}`}
			class={`title ${highlightFeature === 'question' ? 'highlight' : ''}`}
			onmouseenter={() => setHover('question')}
			onmouseleave={() => setHover(null)}
		>
			{market.title}
		</h3>

		<div class="prices">
			<div
				class={`price yes ${highlightFeature === 'prices' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('prices')}
				onmouseleave={() => setHover(null)}
			>
				<span class="label">YES</span>
				<span class="value">{market.yes}¢</span>
			</div>
			<div
				class={`price no ${highlightFeature === 'prices' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('prices')}
				onmouseleave={() => setHover(null)}
			>
				<span class="label">NO</span>
				<span class="value">{market.no}¢</span>
			</div>
			<div
				class={`trend ${market.trend >= 0 ? 'up' : 'down'} ${highlightFeature === 'trend' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('trend')}
				onmouseleave={() => setHover(null)}
			>
				{market.trend >= 0 ? '▲' : '▼'}
				{Math.abs(market.trend)}%
			</div>
		</div>

		{#if market.sparkline}
			<svg
				class="spark"
				viewBox="0 0 120 30"
				role="presentation"
				aria-hidden="true"
				onmouseenter={() => setHover('trend')}
				onmouseleave={() => setHover(null)}
			>
				<polyline points={market.sparkline} />
			</svg>
		{/if}

		<div class="meta">
			<div
				class={`liquidity ${market.liquidity} ${highlightFeature === 'liquidity' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('liquidity')}
				onmouseleave={() => setHover(null)}
			>
				<span class="dot" aria-hidden="true"></span> Liquidity: {liquidityCopy[market.liquidity]}
			</div>
			<div
				class={`community ${highlightFeature === 'community' ? 'highlight' : ''}`}
				role="presentation"
				onmouseenter={() => setHover('community')}
				onmouseleave={() => setHover(null)}
			>
				Community: {market.community}
			</div>
		</div>

		<button
			class={`cta ${highlightFeature === 'cta' ? 'highlight' : ''}`}
			disabled
			onmouseenter={() => setHover('cta')}
			onmouseleave={() => setHover(null)}
		>
			Predict Now
		</button>
	</div>
</article>

<style>
	.card {
		position: relative;
		display: grid;
		gap: 0.75rem;
		padding: 1.25rem;
		min-width: 360px;
		width: 100%;
		border-radius: 1.2rem;
		background: radial-gradient(140% 140% at 0% 0%, rgba(159, 107, 255, 0.18), transparent 45%),
			radial-gradient(160% 120% at 100% 10%, rgba(53, 226, 209, 0.12), transparent 50%),
			rgba(18, 19, 26, 0.95);
		border: 1px solid rgba(255, 255, 255, 0.06);
		box-shadow: 0 18px 45px rgba(0, 0, 0, 0.35), inset 0 0 0 1px rgba(106, 63, 245, 0.24);
		overflow: hidden;
	}

	.content-layer {
		position: relative;
		z-index: 1;
		display: grid;
		gap: inherit;
	}

	.blur-layer {
		position: absolute;
		inset: 0;
		border-radius: inherit;
		backdrop-filter: blur(3px);
		background: rgba(14, 15, 20, 0.4);
		z-index: 0;
	}

	.card__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		font-size: 0.9rem;
		color: #cfcde5;
	}

	.pill {
		padding: 0.35rem 0.5rem;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.06);
		border: 1px solid rgba(255, 255, 255, 0.08);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-weight: 700;
		color: #f4f1ff;
	}

	.resolves {
		color: #9ea3c1;
	}

	.title {
		font-size: 1.2rem;
		line-height: 1.45;
		color: #f8f7ff;
		margin: 0;
		font-weight: 700;
	}

	.prices {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		align-items: center;
		gap: 0.6rem;
	}

	.price {
		padding: 0.8rem 0.25rem;
		border-radius: 0.9rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.25rem;
		font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
		border: 1px solid rgba(255, 255, 255, 0.08);
		font-weight: 700;
		min-width: 0;
	}

	.price.yes {
		background: rgba(53, 226, 209, 0.1);
		color: #c8fff9;
		box-shadow: 0 0 0 1px rgba(53, 226, 209, 0.35);
	}

	.price.no {
		background: rgba(255, 111, 97, 0.08);
		color: #ffd1ca;
		box-shadow: 0 0 0 1px rgba(255, 111, 97, 0.26);
	}

	.label {
		font-size: 0.85rem;
		text-transform: uppercase;
		color: rgba(255, 255, 255, 0.85);
		--padding-right: 0.35rem;
		white-space: nowrap;
	}

	.value {
		font-size: clamp(0.98rem, 2.4vw, 1.08rem);
		white-space: nowrap;
		text-align: right;
	}

	.trend {
		text-align: right;
		font-weight: 800;
		font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
		padding: 0.7rem 0.9rem;
		border-radius: 0.9rem;
		border: 1px solid rgba(255, 255, 255, 0.08);
	}

	.trend.up {
		color: #35e2d1;
		background: rgba(53, 226, 209, 0.08);
	}

	.trend.down {
		color: #ff8d7f;
		background: rgba(255, 111, 97, 0.08);
	}

	.spark {
		width: 100%;
		height: 42px;
	}

	.spark polyline {
		fill: none;
		stroke: #9f6bff;
		stroke-width: 2.6;
		filter: drop-shadow(0 0 10px rgba(159, 107, 255, 0.45));
		stroke-linecap: round;
	}

	.meta {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		color: #c4c8df;
		font-size: 0.95rem;
	}

	.liquidity {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.45rem 0.65rem;
		border-radius: 0.7rem;
		border: 1px solid rgba(255, 255, 255, 0.08);
		font-weight: 700;
	}

	.liquidity .dot {
		width: 0.65rem;
		height: 0.65rem;
		border-radius: 50%;
	}

	.liquidity.deep .dot {
		background: #35e2d1;
		box-shadow: 0 0 10px rgba(53, 226, 209, 0.7);
	}

	.liquidity.moderate .dot {
		background: #f6c947;
		box-shadow: 0 0 10px rgba(246, 201, 71, 0.6);
	}

	.liquidity.thin .dot {
		background: #ff6f61;
		box-shadow: 0 0 10px rgba(255, 111, 97, 0.6);
	}

	.liquidity.very-thin .dot {
		background: #ff4444;
		box-shadow: 0 0 10px rgba(255, 68, 68, 0.7);
	}

	.community {
		padding: 0.45rem 0.65rem;
		border-radius: 0.7rem;
		background: rgba(255, 255, 255, 0.04);
		border: 1px dashed rgba(255, 255, 255, 0.12);
	}

	.cta {
		justify-self: start;
		padding: 0.8rem 1.1rem;
		border-radius: 0.9rem;
		background: linear-gradient(135deg, #6a3ff5, #9f6bff);
		color: #0e0f14;
		border: none;
		font-weight: 800;
		font-size: 0.98rem;
		letter-spacing: 0.02em;
		cursor: not-allowed;
		box-shadow: 0 12px 30px rgba(106, 63, 245, 0.45);
		transition: transform 160ms ease, box-shadow 160ms ease;
		opacity: 0.7;
	}

	.highlight {
		border: 2px solid rgba(255, 255, 255, 0.9);
		box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.6), 0 14px 32px rgba(0, 0, 0, 0.45);
		background: transparent;
	}

	/* Keep CTA fill consistent; only outline/shine for highlight */
	.cta.highlight {
		background: linear-gradient(135deg, #6a3ff5, #9f6bff);
		color: #0e0f14;
	}

	.muted {
		filter: blur(1px) brightness(0.7);
		transition: filter 120ms ease;
	}
</style>
