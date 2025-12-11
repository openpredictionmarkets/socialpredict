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
		background:
			radial-gradient(
				140% 140% at 0% 0%,
				var(--color-primary-soft, rgba(159, 107, 255, 0.16)),
				transparent 45%
			),
			radial-gradient(
				160% 120% at 100% 10%,
				var(--color-accent-soft, rgba(53, 226, 209, 0.12)),
				transparent 50%
			),
			var(--panel, rgba(255, 255, 255, 0.96));
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		box-shadow: 0 18px 45px rgba(15, 23, 42, 0.18);
		overflow: hidden;
	}

	.content-layer {
		position: relative;
		z-index: 1;
		display: grid;
		gap: inherit;
	}

	.card__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		font-size: 0.9rem;
		color: var(--text-subtle, #4b5563);
	}

	.pill {
		padding: 0.35rem 0.5rem;
		border-radius: 999px;
		background: color-mix(in srgb, var(--color-primary, #2563eb) 10%, transparent);
		border: 1px solid color-mix(in srgb, var(--color-primary, #2563eb) 45%, transparent);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-weight: 700;
		color: var(--color-primary, #2563eb);
	}

	.resolves {
		color: var(--text-subtle, #6b7280);
	}

	.title {
		font-size: 1.2rem;
		line-height: 1.45;
		color: var(--text, #0e0f14);
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
		border: 1px solid rgba(148, 163, 184, 0.35);
		font-weight: 700;
		min-width: 0;
	}

	.price.yes {
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
		color: var(--color-accent, #16a34a);
		box-shadow: 0 0 0 1px rgba(22, 163, 74, 0.35);
	}

	.price.no {
		background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
		color: var(--color-danger, #dc2626);
		box-shadow: 0 0 0 1px rgba(220, 38, 38, 0.26);
	}

	.label {
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--text-subtle, #4b5563);
		--padding-right: 0.35rem;
		white-space: nowrap;
	}

	.value {
		font-size: 0.95rem;
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
		border: 1px solid rgba(148, 163, 184, 0.35);
	}

	.trend.up {
		color: var(--color-accent, #16a34a);
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
	}

	.trend.down {
		color: var(--color-danger, #dc2626);
		background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
	}

	.spark {
		width: 100%;
		height: 42px;
	}

	.spark polyline {
		fill: none;
		stroke: var(--color-primary, #2563eb);
		stroke-width: 2.6;
		filter: drop-shadow(0 0 10px rgba(37, 99, 235, 0.35));
		stroke-linecap: round;
	}

	.meta {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		color: var(--text-subtle, #4b5563);
		font-size: 0.95rem;
	}

	.liquidity {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.45rem 0.65rem;
		border-radius: 0.7rem;
		border: 1px solid rgba(148, 163, 184, 0.35);
		font-weight: 700;
	}

	.liquidity .dot {
		width: 0.65rem;
		height: 0.65rem;
		border-radius: 50%;
	}

	.liquidity.deep .dot {
		background: var(--color-accent, #16a34a);
		box-shadow: 0 0 10px rgba(22, 163, 74, 0.5);
	}

	.liquidity.moderate .dot {
		background: #f6c947;
		box-shadow: 0 0 10px rgba(246, 201, 71, 0.6);
	}

	.liquidity.thin .dot {
		background: var(--color-danger, #dc2626);
		box-shadow: 0 0 10px rgba(220, 38, 38, 0.5);
	}

	.liquidity.very-thin .dot {
		background: #f97316;
		box-shadow: 0 0 10px rgba(249, 115, 22, 0.6);
	}

	.community {
		padding: 0.45rem 0.65rem;
		border-radius: 0.7rem;
		background: rgba(148, 163, 184, 0.06);
		border: 1px dashed rgba(148, 163, 184, 0.4);
	}

	.cta {
		justify-self: start;
		padding: 0.8rem 1.1rem;
		border-radius: 0.9rem;
		background: linear-gradient(
			135deg,
			var(--color-primary, #2563eb),
			var(--color-accent, #16a34a)
		);
		color: var(--color-text-light, #e9edf6);
		border: none;
		font-weight: 800;
		font-size: 0.98rem;
		letter-spacing: 0.02em;
		cursor: not-allowed;
		box-shadow: 0 12px 30px rgba(37, 99, 235, 0.35);
		transition: transform 160ms ease, box-shadow 160ms ease;
		opacity: 0.7;
	}

	.highlight {
		border: 2px solid var(--color-primary, #2563eb);
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary, #2563eb) 45%, transparent),
			0 14px 32px rgba(15, 23, 42, 0.45);
		background: transparent;
	}

	/* Keep CTA fill consistent; only outline/shine for highlight */
	.cta.highlight {
		background: linear-gradient(
			135deg,
			var(--color-primary, #2563eb),
			var(--color-accent, #16a34a)
		);
		color: var(--color-text-light, #e9edf6);
	}

	.muted {
		filter: blur(1px) brightness(0.7);
		transition: filter 120ms ease;
	}
</style>
