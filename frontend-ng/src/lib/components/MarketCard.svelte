<script lang="ts" context="module">
	export type LiquidityLevel = 'deep' | 'moderate' | 'thin' | 'very-thin';

	export type Market = {
		title: string;
		yes: number;
		no: number;
		community: string;
		resolves: string;
		trend: number;
		liquidity: LiquidityLevel;
		category: string;
		categoryColor?: string;
		sparkline?: string;
	};

</script>

<script lang="ts">
	export let market: Market;

	const liquidityCopy: Record<LiquidityLevel, string> = {
		deep: 'Deep',
		moderate: 'Moderate',
		thin: 'Thin',
		'very-thin': 'Very Thin'
	};
</script>

<article class="card" aria-labelledby={`market-${market.title}`}>
	<div class="card__top">
		<div
			class="pill"
			style={`--category-color: ${market.categoryColor ?? 'var(--color-primary, #2563eb)'}`}
		>
			{market.category}
		</div>
		<div class="resolves">Resolves {market.resolves}</div>
	</div>

	<h3 id={`market-${market.title}`} class="title">{market.title}</h3>

	<div class="prices">
		<div class="price yes">
			<span class="label">YES</span>
			<span class="value">{market.yes}¢</span>
		</div>
		<div class="price no">
			<span class="label">NO</span>
			<span class="value">{market.no}¢</span>
		</div>
		<div class="trend {market.trend >= 0 ? 'up' : 'down'}">
			{market.trend >= 0 ? '▲' : '▼'}
			{Math.abs(market.trend)}%
		</div>
	</div>

	{#if market.sparkline}
		<svg class="spark" viewBox="0 0 120 30" role="presentation" aria-hidden="true">
			<polyline points={market.sparkline} />
		</svg>
	{/if}

	<div class="meta">
		<div class={`liquidity ${market.liquidity}`}>
			<span class="dot" aria-hidden="true"></span> Liquidity: {liquidityCopy[market.liquidity]}
		</div>
		<div class="community">Community: {market.community}</div>
	</div>

	<button class="cta">Predict Now</button>
</article>

<style>
  /* * {
    border: 1px dashed yellow;
  } */

	.card {
		display: grid;
		gap: 0.75rem;
		padding: 1.25rem;
		border-radius: 1.2rem;
		background: var(--panel, rgba(255, 255, 255, 0.96));
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		box-shadow: 0 12px 30px rgba(15, 23, 42, 0.12);
		transition: transform 200ms ease, box-shadow 200ms ease, border-color 200ms ease;
	}

	.card:hover {
		transform: translateY(-6px);
		border-color: var(--color-primary, #2563eb);
		box-shadow: 0 18px 40px rgba(15, 23, 42, 0.16), 0 0 0 1px rgba(37, 99, 235, 0.16);
	}

	.card__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		font-size: 0.9rem;
		color: var(--text-muted, #6b7280);
	}

	.pill {
		padding: 0.35rem 0.5rem;
		border-radius: 999px;
		background: color-mix(in srgb, var(--category-color) 12%, transparent);
		border: 1px solid color-mix(in srgb, var(--category-color) 55%, transparent);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-weight: 700;
		color: var(--category-color, var(--color-primary, #2563eb));
	}

	.resolves {
		color: var(--text-subtle, #4b5563);
	}

	.title {
		font-size: 1.2rem;
		line-height: 1.45;
		color: var(--text, #0e0f14);
		margin: 0;
		font-weight: 700;
		min-height: calc(1.2rem * 1.45 * 4);
		display: -webkit-box;
		-webkit-line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
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

  .price .label {
    min-width: 3ch;
    text-align: left;
  }

  .price .value {
    min-width: 4ch;
    text-align: right;
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
		white-space: nowrap;
	}

	.value {
    font-size: 0.85rem;
		white-space: nowrap;
		text-align: right;
	}

	.trend {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    justify-content: flex-end;
    white-space: nowrap;
		text-align: right;
		font-weight: 800;
		----font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
    font-size: 0.85rem;
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

	.spark::before {
		content: '';
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
		cursor: pointer;
		box-shadow: 0 12px 30px rgba(37, 99, 235, 0.35);
		transition: transform 160ms ease, box-shadow 160ms ease;
	}

	.cta:hover {
		transform: translateY(-2px);
		box-shadow: 0 16px 36px rgba(37, 99, 235, 0.45);
	}

	@media (max-width: 720px) {
		.card {
			padding: 1.1rem;
		}

		.prices {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.trend {
			grid-column: span 2;
			text-align: left;
		}
	}
</style>
