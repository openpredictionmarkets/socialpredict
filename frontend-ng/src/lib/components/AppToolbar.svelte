<script lang="ts">
	import BrandSelector from '$lib/components/BrandSelector.svelte';
	import Notation from '$lib/components/Notation.svelte';
	import { currentBranding } from '$lib/stores/brandingStore';
</script>

<header class="app-toolbar">
	<button class="icon-button" type="button" aria-label="Open menu">
		<span class="bar"></span>
		<span class="bar"></span>
		<span class="bar"></span>
	</button>

	<div class="toolbar-center">
		<Notation text={$currentBranding.notation ?? 'Branding driven by config/branding/*.json'} />
	</div>

	<div class="toolbar-right">
		<BrandSelector />
	</div>
</header>

<style>
	:global(:root) {
		--toolbar-height: 64px;
	}

	.app-toolbar {
		position: sticky;
		top: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.9rem 1.5rem;
		margin: 0 auto;
		max-width: 1200px;
		backdrop-filter: blur(10px);
		border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.08));
		min-height: var(--toolbar-height);
	}

	.toolbar-center {
		flex: 1;
		display: flex;
		justify-content: center;
	}

	.toolbar-right {
		margin-left: auto;
	}

	.icon-button {
		display: grid;
		gap: 0.2rem;
		background: var(--panel, rgba(255, 255, 255, 0.05));
		border: 1px solid var(--border, rgba(255, 255, 255, 0.12));
		border-radius: 0.75rem;
		padding: 0.45rem 0.6rem;
		cursor: pointer;
	}

	.icon-button .bar {
		width: 18px;
		height: 2px;
		background: currentColor;
		border-radius: 4px;
	}
</style>
