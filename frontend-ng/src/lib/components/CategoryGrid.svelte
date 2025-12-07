<script lang="ts" context="module">
	import type { ComponentType } from 'svelte';

	export type Category = {
		name: string;
		icon: ComponentType;
		color?: string;
	};
</script>

<script lang="ts">
	export let categories: Category[] = [];
</script>

<div class="grid">
	{#each categories as category}
		<button
			class="tile"
			type="button"
			style={`--icon-color: ${category.color ?? '#9f6bff'}; --icon-bg: ${
				category.color ? category.color + '1A' : 'rgba(159, 107, 255, 0.12)'
			};`}
		>
			<span class="icon" aria-hidden="true">
				<svelte:component this={category.icon} />
			</span>
			<span class="label">{category.name}</span>
		</button>
	{/each}
</div>

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 0.75rem;
	}

	.tile {
		position: relative;
		padding: 0.95rem 1.1rem;
		border-radius: 1rem;
		border: 1px solid rgba(255, 255, 255, 0.08);
		background: rgba(255, 255, 255, 0.03);
		color: #f8f7ff;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-weight: 700;
		cursor: pointer;
		transition: transform 160ms ease, border-color 160ms ease, box-shadow 160ms ease;
	}

	.tile::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: inherit;
		background: linear-gradient(135deg, rgba(159, 107, 255, 0.12), rgba(53, 226, 209, 0.08));
		opacity: 0;
		transition: opacity 200ms ease;
	}

	.tile:hover {
		transform: translateY(-2px);
		border-color: rgba(159, 107, 255, 0.4);
		box-shadow: 0 16px 28px rgba(0, 0, 0, 0.3);
	}

	.tile:hover::after {
		opacity: 1;
	}

	.icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.8rem;
		background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.16), transparent 50%),
			linear-gradient(135deg, var(--icon-bg), rgba(53, 226, 209, 0.15));
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.icon :global(svg) {
		width: 1.3rem;
		height: 1.3rem;
		stroke-width: 2.2;
		color: var(--icon-color, #9f6bff);
	}

	.label {
		position: relative;
		z-index: 1;
		letter-spacing: -0.01em;
	}
</style>
