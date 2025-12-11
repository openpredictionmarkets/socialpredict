<div class="hero__panel">
	<div class="panel__glass">
		<div class="panel__header">
			<div>
				<div class="eyebrow">Order Ticket</div>
				<h3>Will Apple release a foldable iPhone in 2025?</h3>
			</div>
			<div class="score">
				<div>SP Score</div>
				<strong>72</strong>
			</div>
		</div>
		<div class="panel__row">
			<div class="pill yes highlight">YES 62¢</div>
			<div class="pill no">NO 38¢</div>
			<div class="pill trend up">▲ +12%</div>
		</div>
		<div class="slider">
			<div class="slider__label">Choose stake</div>
			<div class="slider__track">
				<div class="slider__fill" style="width: 64%"></div>
				<div class="slider__thumb" style="left: 64%"></div>
			</div>
			<div class="slider__values">
				<span>$25</span>
				<span>Potential payout $40.50</span>
			</div>
		</div>
		<button class="primary block">Confirm YES</button>
		<div class="panel__footer">
			<div>Maker fee 0.15%</div>
			<div>Depth preview · Live</div>
		</div>
	</div>
</div>

<style>
	.hero__panel {
		position: relative;
		color: var(--text, #0e0f14);
	}

	.panel__glass {
		position: relative;
		padding: 1.4rem;
		border-radius: 1.2rem;
		background: var(--panel, rgba(255, 255, 255, 0.96));
		border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
		box-shadow: 0 18px 40px rgba(15, 23, 42, 0.18);
		backdrop-filter: blur(8px);
		display: grid;
		gap: 1rem;
	}

	.panel__header {
		display: flex;
		justify-content: space-between;
		gap: 0.75rem;
		align-items: center;
	}

	.panel__header h3 {
		margin: 0.2rem 0 0;
		font-size: 1.05rem;
	}

	.score {
		text-align: right;
		color: var(--text-subtle, #4b5563);
	}

	.score strong {
		display: block;
		font-size: 1.8rem;
		color: var(--color-accent, #16a34a);
		text-shadow: 0 0 12px rgba(22, 163, 74, 0.45);
	}

	.eyebrow {
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.8rem;
		color: var(--color-primary, #2563eb);
	}

	.panel__row {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.6rem;
	}

	.pill {
		padding: 0.75rem 0.9rem;
		border-radius: 0.85rem;
		text-align: center;
		font-weight: 800;
		border: 1px solid rgba(148, 163, 184, 0.4);
	}

	.pill.highlight {
		border: 3px solid var(--color-primary, #2563eb);
	}

	.pill.yes {
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
		color: var(--color-accent, #16a34a);
	}

	.pill.no {
		background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
		color: var(--color-danger, #dc2626);
	}

	.pill.trend.up {
		color: var(--color-accent, #16a34a);
		background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
	}

	.slider {
		display: grid;
		gap: 0.35rem;
	}

	.slider__label {
		color: var(--text-subtle);
		font-size: 0.95rem;
	}

	.slider__track {
		position: relative;
		height: 12px;
		border-radius: 999px;
		background: rgba(148, 163, 184, 0.3);
		overflow: hidden;
	}

	.slider__fill {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			90deg,
			var(--color-primary, #2563eb),
			var(--color-accent, #16a34a)
		);
	}

	.slider__thumb {
		position: absolute;
		top: 50%;
		width: 18px;
		height: 18px;
		transform: translate(-50%, -50%);
		border-radius: 50%;
		background: #fff;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.35);
		border: 2px solid var(--color-primary, #2563eb);
	}

	.slider__values {
		display: flex;
		justify-content: space-between;
		color: var(--text);
		font-weight: 700;
		font-family: 'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo,
			Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
	}

	.panel__footer {
		display: flex;
		justify-content: space-between;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	@media (max-width: 640px) {
		.panel__row {
			grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		}
	}
</style>
