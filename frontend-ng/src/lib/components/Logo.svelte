<script lang="ts">
	export let text: string;
</script>

<div class="logo-pill">
	{text}
</div>

<style>
	.logo-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.4rem 0.9rem;
		border-radius: 999px;
		background: var(--color-primary-soft, rgba(37, 99, 235, 0.16));
		color: var(--text);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-weight: 700;
		border: 1px solid var(--color-primary, #2563eb);
		box-shadow: 0 10px 25px rgba(37, 99, 235, 0.25);
	}
</style>

