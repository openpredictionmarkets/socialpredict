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
		border: 1px solid rgba(159, 107, 255, 0.25);
		background: linear-gradient(90deg, rgba(159, 107, 255, 0.15), rgba(53, 226, 209, 0.08));
		border-radius: 999px;
		backdrop-filter: blur(8px);
		box-shadow: 0 0 0 1px rgba(106, 63, 245, 0.2), 0 15px 40px rgba(0, 0, 0, 0.35);
	}

	.ticker__label {
		background: linear-gradient(135deg, #6a3ff5, #9f6bff);
		color: #0e0f14;
		font-weight: 700;
		padding: 0.35rem 0.9rem;
		border-radius: 999px;
		text-transform: uppercase;
		font-size: 0.75rem;
		letter-spacing: 0.05em;
		box-shadow: 0 10px 25px rgba(106, 63, 245, 0.35);
	}

	.ticker__rail {
		overflow: hidden;
		position: relative;
	}

	.ticker__track {
		display: flex;
		gap: 1.5rem;
		min-width: max-content;
		animation: scroll 22s linear infinite;
		animation-play-state: running;
	}

	.ticker.paused .ticker__track {
		animation-play-state: paused;
	}

	.ticker__item {
		display: flex;
		align-items: center;
		gap: 0.65rem;
		color: #dcd9f4;
		font-size: 0.95rem;
		padding: 0.4rem 0.8rem;
		border-radius: 999px;
		background: rgba(14, 15, 20, 0.4);
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: inset 0 0 0 1px rgba(159, 107, 255, 0.25);
	}

	.ticker__question {
		font-weight: 600;
		color: #f4f1ff;
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
		background: rgba(53, 226, 209, 0.1);
		color: #35e2d1;
		border: 1px solid rgba(53, 226, 209, 0.35);
	}

	.price.no {
		background: rgba(255, 111, 97, 0.12);
		color: #ff8d7f;
		border: 1px solid rgba(255, 111, 97, 0.28);
	}

	.delta {
		font-weight: 700;
		padding: 0.2rem 0.5rem;
		border-radius: 0.6rem;
		font-size: 0.85rem;
	}

	.delta.up {
		color: #35e2d1;
		background: rgba(53, 226, 209, 0.1);
	}

	.delta.down {
		color: #ff8d7f;
		background: rgba(255, 111, 97, 0.1);
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
			animation-duration: 28s;
		}
	}
</style>
