<script lang="ts">
	'use runes';

	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { currentBranding } from '$lib/stores/brandingStore';

	const { children } = $props();

	const branding = $derived($currentBranding);

	$effect(() => {
		if (typeof document === 'undefined') return;

		const root = document.documentElement;
		const colors = branding.colors;
		const theme = branding.theme ?? 'light';

		root.dataset.theme = theme;

		root.style.setProperty('--color-background', colors.background);

		root.style.setProperty('--color-text-dark', colors.textDark);
		root.style.setProperty('--color-text-light', colors.textLight);
		root.style.setProperty('--color-text-subtle-dark', colors.textSubtleDark);
		root.style.setProperty('--color-text-subtle-light', colors.textSubtleLight);
		root.style.setProperty('--color-text-muted-dark', colors.textMutedDark);
		root.style.setProperty('--color-text-muted-light', colors.textMutedLight);

		root.style.setProperty('--color-panel-dark', colors.panelDark);
		root.style.setProperty('--color-panel-light', colors.panelLight);
		root.style.setProperty('--color-border-dark', colors.borderDark);
		root.style.setProperty('--color-border-light', colors.borderLight);

		root.style.setProperty('--color-primary', colors.primary);
		root.style.setProperty('--color-primary-soft', colors.primarySoft);
		root.style.setProperty('--color-accent', colors.accent);
		root.style.setProperty('--color-accent-soft', colors.accentSoft);
		root.style.setProperty('--color-danger', colors.danger);
		root.style.setProperty('--color-danger-soft', colors.dangerSoft);
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
