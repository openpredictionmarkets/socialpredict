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
		color: #9f6bff;
		background: rgba(159, 107, 255, 0.14);
		border: 1px solid rgba(159, 107, 255, 0.35);
		box-shadow: 0 10px 22px rgba(159, 107, 255, 0.25);
	}

	h2 {
		margin: 0.2rem 0 0;
		font-size: 1.6rem;
		color: #f7f6ff;
		letter-spacing: -0.02em;
	}

	p {
		margin: 0.4rem 0 0;
		color: #c4c8df;
		max-width: 62ch;
	}

	.action {
		padding: 0.6rem 0.9rem;
		border-radius: 0.85rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: #e6e4ff;
		text-decoration: none;
		font-weight: 700;
		background: rgba(255, 255, 255, 0.04);
		transition: transform 150ms ease, border-color 150ms ease, color 150ms ease;
	}

	.action:hover {
		transform: translateY(-1px);
		color: #ffffff;
		border-color: rgba(159, 107, 255, 0.45);
		box-shadow: 0 10px 18px rgba(159, 107, 255, 0.25);
	}

	@media (max-width: 768px) {
		.header {
			flex-direction: column;
		}
	}
</style>
