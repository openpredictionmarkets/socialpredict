<script lang="ts">
	import type { BrandKey } from '$lib/config/branding';
	import { brandOptions, currentBrandKey } from '$lib/stores/brandingStore';

	let value: BrandKey;

	// Keep local value in sync with global currentBrandKey
	$: value = $currentBrandKey;

	function handleChange(event: Event) {
		const next = (event.currentTarget as HTMLSelectElement).value as BrandKey;
		currentBrandKey.set(next);
	}
</script>

<label class="brand-selector">
	<span class="brand-selector__label">Brand</span>
	<select class="brand-selector__select" bind:value={value} on:change={handleChange}>
		{#each brandOptions as brand}
			<option value={brand}>{brand}</option>
		{/each}
	</select>
</label>

<style>
	.brand-selector {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.9rem;
		color: var(--text-subtle, #4b5563);
	}

	.brand-selector__label {
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-weight: 600;
	}

	.brand-selector__select {
		padding: 0.35rem 0.6rem;
		border-radius: 0.45rem;
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		background: var(--panel, #ffffff);
		color: var(--text, #0f172a);
		font-size: 0.9rem;
	}
</style>
