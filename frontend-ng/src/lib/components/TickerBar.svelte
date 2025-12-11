<script lang="ts" context="module">
	export type TickerItem = {
		label: string;
		yes: number;
		no: number;
		trend: number;
	};
</script>

<script lang="ts">
	export let items: TickerItem[] = [];

	let paused = false;
</script>

<div
	class={`ticker ${paused ? 'paused' : ''}`}
	role="group"
	aria-label="Live market ticker (hover to pause)"
	on:mouseenter={() => (paused = true)}
	on:mouseleave={() => (paused = false)}
>
	<div class="ticker__label">Live</div>
	<div class="ticker__rail">
		<div class="ticker__track" aria-label="Live market ticker">
			{#each items.concat(items) as item, index}
				<div class="ticker__item" aria-label={`${item.label} at YES ${item.yes} NO ${item.no}`}>
					<span class="ticker__question">{item.label}</span>
					<span class="ticker__prices">
						<span class="price yes">YES {item.yes}¢</span>
						<span class="price no">NO {item.no}¢</span>
					</span>
					<span class={`delta ${item.trend >= 0 ? 'up' : 'down'}`}>
						{item.trend >= 0 ? '▲' : '▼'} {Math.abs(item.trend)}%
					</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.ticker {
		display: grid;
		grid-template-columns: auto 1fr;
		align-items: center;
		gap: 0.75rem;
		padding: 0.65rem 1rem;
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		background: linear-gradient(
			90deg,
			var(--color-primary-soft, rgba(37, 99, 235, 0.12)),
			var(--color-accent-soft, rgba(22, 163, 74, 0.08))
		);
		border-radius: 999px;
		backdrop-filter: blur(8px);
		box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.2), 0 10px 25px rgba(15, 23, 42, 0.18);
	}

	.ticker__label {
		background: linear-gradient(
			135deg,
			var(--color-primary, #2563eb),
			var(--color-accent, #16a34a)
		);
		color: var(--color-text-light, #e9edf6);
		font-weight: 700;
		padding: 0.35rem 0.9rem;
		border-radius: 999px;
		text-transform: uppercase;
		font-size: 0.75rem;
		letter-spacing: 0.05em;
		box-shadow: 0 10px 25px rgba(37, 99, 235, 0.3);
	}

	.ticker__rail {
		overflow: hidden;
		position: relative;
	}

	.ticker__track {
		display: flex;
		gap: 1.5rem;
		min-width: max-content;
		animation: scroll 40s linear infinite;
		animation-play-state: running;
	}

	.ticker.paused .ticker__track {
		animation-play-state: paused;
	}

	.ticker__item {
		display: flex;
		align-items: center;
		gap: 0.65rem;
		color: var(--text, #0e0f14);
		font-size: 0.95rem;
		padding: 0.4rem 0.8rem;
		border-radius: 999px;
		background: var(--panel, rgba(255, 255, 255, 0.78));
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.25);
	}

	.ticker__question {
		font-weight: 600;
		color: var(--text, #0e0f14);
	}

	.ticker__prices {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
	}

	.price {
		padding: 0.25rem 0.6rem;
		border-radius: 0.7rem;
		font-weight: 700;
		font-size: 0.9rem;
	}

	.price.yes {
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
		color: var(--color-accent, #16a34a);
		border: 1px solid rgba(22, 163, 74, 0.4);
	}

	.price.no {
		background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
		color: var(--color-danger, #dc2626);
		border: 1px solid rgba(220, 38, 38, 0.32);
	}

	.delta {
		font-weight: 700;
		padding: 0.2rem 0.5rem;
		border-radius: 0.6rem;
		font-size: 0.85rem;
	}

	.delta.up {
		color: var(--color-accent, #16a34a);
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
	}

	.delta.down {
		color: var(--color-danger, #dc2626);
		background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
	}

	@keyframes scroll {
		from {
			transform: translateX(0);
		}
		to {
			transform: translateX(-50%);
		}
	}

	@media (max-width: 768px) {
		.ticker {
			grid-template-columns: 1fr;
			border-radius: 1rem;
		}

		.ticker__label {
			justify-self: start;
		}

		.ticker__track {
			animation-duration: 52s;
		}
	}
</style>
