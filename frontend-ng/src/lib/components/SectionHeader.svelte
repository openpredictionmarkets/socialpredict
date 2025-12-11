<script lang="ts">
	export let eyebrow: string | undefined = undefined;
	export let title: string;
	export let subtitle: string | undefined = undefined;
	export let actionLabel: string | undefined = undefined;
	export let actionHref = '#';
</script>

<div class="header">
	<div>
		{#if eyebrow}
			<div class="eyebrow">{eyebrow}</div>
		{/if}
		<h2>{title}</h2>
		{#if subtitle}
			<p>{subtitle}</p>
		{/if}
	</div>
	{#if actionLabel}
		<a class="action" href={actionHref}>{actionLabel}</a>
	{/if}
</div>

<style>
	.header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.eyebrow {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.35rem 0.7rem;
		border-radius: 999px;
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-primary, #2563eb);
		background: var(--color-primary-soft, rgba(37, 99, 235, 0.16));
		border: 1px solid var(--color-primary, #2563eb);
		box-shadow: 0 6px 18px rgba(37, 99, 235, 0.22);
	}

	h2 {
		margin: 0.2rem 0 0;
		font-size: 1.6rem;
		color: var(--text, #0e0f14);
		letter-spacing: -0.02em;
	}

	p {
		margin: 0.4rem 0 0;
		color: var(--text-subtle, #4b5563);
		max-width: 62ch;
	}

	.action {
		padding: 0.6rem 0.9rem;
		border-radius: 0.85rem;
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		color: var(--text, #0e0f14);
		text-decoration: none;
		font-weight: 700;
		background: var(--panel, rgba(255, 255, 255, 0.96));
		transition: transform 150ms ease, border-color 150ms ease, color 150ms ease;
	}

	.action:hover {
		transform: translateY(-1px);
		color: var(--color-primary, #2563eb);
		border-color: var(--color-primary, #2563eb);
		box-shadow: 0 10px 18px rgba(37, 99, 235, 0.2);
	}

	@media (max-width: 768px) {
		.header {
			flex-direction: column;
		}
	}
</style>
