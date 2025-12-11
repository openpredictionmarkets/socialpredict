<script lang="ts">
	'use runes';

	import TickerBar, {
		type TickerItem
	} from '$lib/components/TickerBar.svelte';

	import SectionHeader from '$lib/components/SectionHeader.svelte';

	import MarketCard, {
		type Market
	} from '$lib/components/MarketCard.svelte';

	import CategoryGrid, { type Category } from '$lib/components/CategoryGrid.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import TutorialModal from '$lib/components/TutorialModal.svelte';
	import AppToolbar from '$lib/components/AppToolbar.svelte';
	import HeroPanel from '$lib/components/HeroPanel.svelte';
  import BrandDebugPanel from '$lib/components/BrandDebugPanel.svelte';
	import { currentBranding } from '$lib/stores/brandingStore';

	import {
		Flag,
		ChartNoAxesCombined,
		Coins,
		MonitorSmartphone,
		Trophy,
		Sparkles,
		FlaskConical
	} from 'lucide-svelte';

	const { data } = $props();

	const branding = $derived($currentBranding);

	const tickerItems: TickerItem[] = $derived(branding.examples.tickerItems);

	const trendingMarkets: Market[] = $derived(
		branding.examples.trendingMarkets.map((market) => ({
			...market,
			categoryColor: branding.categories.find((c) => c.name === market.category)?.color
		}))
	);

	const endingSoon: Market[] = $derived(
		branding.examples.endingSoon.map((market) => ({
			...market,
			categoryColor: branding.categories.find((c) => c.name === market.category)?.color
		}))
	);

	const categoryIcons: Record<string, Category['icon']> = {
		Flag,
		ChartNoAxesCombined,
		Coins,
		MonitorSmartphone,
		Trophy,
		Sparkles,
		FlaskConical
	};

	const categories: Category[] = $derived(
		branding.categories.map((category) => ({
			name: category.name,
			icon: categoryIcons[category.icon] ?? Flag,
			color: category.color
		}))
	);

	let showTutorial = $state(false);
</script>

<AppToolbar/>


<main class="page">
	{#each branding.layout.sections as section}
		{#if section.type === 'logo' && branding.features.showLogo}
			<section
				class={`section logo-section ${
					section.variant === 'logoCenter'
						? 'logo-section--center'
						: section.variant === 'logoRight'
							? 'logo-section--right'
							: 'logo-section--default'
				}`}
			>
				<Logo text={branding.brand.name} />
			</section>
		{:else if section.type === 'hero' && branding.features.showHero}
			{#if section.variant === 'leftPanel'}
				<section class="hero hero--left-panel">
					<HeroPanel />
					<div class="hero__copy">
						<h1>
							{branding.brand.taglinePrimary}
							<span>{branding.brand.taglineSecondary}</span>
						</h1>
						<p>{branding.brand.description}</p>
						<div class="hero__actions">
							<button class="primary" onclick={() => (showTutorial = true)}>
								{branding.brand.heroCtaPrimary}
							</button>
							{#if branding.features.showTutorial}
								<button class="secondary" type="button" onclick={() => (showTutorial = true)}>
									{branding.brand.heroCtaSecondary}
								</button>
							{/if}
						</div>
						<div class="hero__stats">
							{#each branding.examples.heroStats as stat}
								<div>
									<strong>{stat.value}</strong>
									<span>{stat.label}</span>
								</div>
							{/each}
						</div>
					</div>
				</section>
			{:else if section.variant === 'logoRight'}
				<section class="hero hero--logo-right">
					<div class="hero__copy">
						<h1>
							{branding.brand.taglinePrimary}
							<span>{branding.brand.taglineSecondary}</span>
						</h1>
						<p>{branding.brand.description}</p>
						<div class="hero__actions">
							<button class="primary" onclick={() => (showTutorial = true)}>
								{branding.brand.heroCtaPrimary}
							</button>
							{#if branding.features.showTutorial}
								<button class="secondary" type="button" onclick={() => (showTutorial = true)}>
									{branding.brand.heroCtaSecondary}
								</button>
							{/if}
						</div>
						<div class="hero__stats">
							{#each branding.examples.heroStats as stat}
								<div>
									<strong>{stat.value}</strong>
									<span>{stat.label}</span>
								</div>
							{/each}
						</div>
					</div>
					<div class="hero__aside">
						<HeroPanel />
					</div>
				</section>
			{:else if section.variant === 'logoCenter'}
				<section class="hero hero--logo-center">
					<div class="hero__copy">
						<h1>
							{branding.brand.taglinePrimary}
							<span>{branding.brand.taglineSecondary}</span>
						</h1>
						<p>{branding.brand.description}</p>
						<div class="hero__actions">
							<button class="primary" onclick={() => (showTutorial = true)}>
								{branding.brand.heroCtaPrimary}
							</button>
							{#if branding.features.showTutorial}
								<button class="secondary" type="button" onclick={() => (showTutorial = true)}>
									{branding.brand.heroCtaSecondary}
								</button>
							{/if}
						</div>
						<div class="hero__stats">
							{#each branding.examples.heroStats as stat}
								<div>
									<strong>{stat.value}</strong>
									<span>{stat.label}</span>
								</div>
							{/each}
						</div>
					</div>
					<HeroPanel />
				</section>
			{:else if section.variant === 'bottomPanel'}
				<section class="hero hero--bottom-panel">
					<div class="hero__copy">
						<h1>
							{branding.brand.taglinePrimary}
							<span>{branding.brand.taglineSecondary}</span>
						</h1>
						<p>{branding.brand.description}</p>
						<div class="hero__actions">
							<button class="primary" onclick={() => (showTutorial = true)}>
								{branding.brand.heroCtaPrimary}
							</button>
							{#if branding.features.showTutorial}
								<button class="secondary" type="button" onclick={() => (showTutorial = true)}>
									{branding.brand.heroCtaSecondary}
								</button>
							{/if}
						</div>
						<div class="hero__stats">
							{#each branding.examples.heroStats as stat}
								<div>
									<strong>{stat.value}</strong>
									<span>{stat.label}</span>
								</div>
							{/each}
						</div>
					</div>
					<HeroPanel />
				</section>
			{:else}
				<section class="hero">
					<div class="hero__copy">
						<h1>
							{branding.brand.taglinePrimary}
							<span>{branding.brand.taglineSecondary}</span>
						</h1>
						<p>{branding.brand.description}</p>
						<div class="hero__actions">
							<button class="primary" onclick={() => (showTutorial = true)}>
								{branding.brand.heroCtaPrimary}
							</button>
							{#if branding.features.showTutorial}
								<button class="secondary" type="button" onclick={() => (showTutorial = true)}>
									{branding.brand.heroCtaSecondary}
								</button>
							{/if}
						</div>
						<div class="hero__stats">
							{#each branding.examples.heroStats as stat}
								<div>
									<strong>{stat.value}</strong>
									<span>{stat.label}</span>
								</div>
							{/each}
						</div>
					</div>
					<HeroPanel />
				</section>
			{/if}
		{:else if section.type === 'ticker' && branding.features.showTicker}
			<section class="ticker-wrap">
				<TickerBar items={tickerItems} />
			</section>
		{:else if section.type === 'trending' && branding.features.showTrending}
			<section class="section">
				<SectionHeader
					eyebrow="Trending Right Now"
					title="Markets moving with volume, velocity, and social buzz"
					subtitle="A pulse on what the SocialPredict crowd is talking about in real time."
					actionLabel="See all markets"
					actionHref="#"
				/>
				<div class="grid grid--markets">
					{#each trendingMarkets as market}
						<MarketCard {market} />
					{/each}
				</div>
			</section>
		{:else if section.type === 'categories' && branding.features.showCategories}
			<section class="section section--split">
				<div class="split__left">
					<SectionHeader
						eyebrow="Categories"
						title="Jump into any arena"
						subtitle="From macro to memes, pick a lane and see what the crowd thinks."
					/>
					<CategoryGrid {categories} />
				</div>
				{#if branding.features.showEndingSoon}
					<div class="split__right">
						<SectionHeader
							eyebrow="Ending Soon"
							title="Last chance to get in"
							subtitle="Markets about to resolve — catch the final moves."
						/>
						<div class="stack">
							{#each endingSoon as market}
								<MarketCard {market} />
							{/each}
						</div>
					</div>
				{/if}
			</section>
		{:else if section.type === 'endingSoon' && branding.features.showEndingSoon && !branding.features.showCategories}
			<section class="section">
				<SectionHeader
					eyebrow="Ending Soon"
					title="Last chance to get in"
					subtitle="Markets about to resolve — catch the final moves."
				/>
				<div class="stack">
					{#each endingSoon as market}
						<MarketCard {market} />
					{/each}
				</div>
			</section>
		{/if}
	{/each}

	{#if branding.features.showTutorial}
		<TutorialModal
			open={showTutorial}
			onClose={() => (showTutorial = false)}
			onComplete={() => (showTutorial = true)}
		/>
	{/if}
</main>

<BrandDebugPanel {branding} />

<style>
:global(:root) {
  --bg: var(--color-background, #f5f7fb);
  --text: var(--color-text-dark, #0e0f14);
  --text-subtle: var(--color-text-subtle-dark, #4b5563);
  --text-muted: var(--color-text-muted-dark, #6b7280);
  --panel: var(--color-panel-light, rgba(0, 0, 0, 0.02));
  --border: var(--color-border-light, rgba(0, 0, 0, 0.08));
}

:global(:root[data-theme='light']) {
  --bg: var(--color-background, #f5f7fb);
  --text: var(--color-text-dark, #0e0f14);
  --text-subtle: var(--color-text-subtle-dark, #4b5563);
  --text-muted: var(--color-text-muted-dark, #6b7280);
  --panel: var(--color-panel-light, rgba(0, 0, 0, 0.02));
  --border: var(--color-border-light, rgba(0, 0, 0, 0.08));
}

:global(:root[data-theme='dark']) {
  --bg: var(--color-background, #0e0f14);
  --text: var(--color-text-light, #e6e4ff);
  --text-subtle: var(--color-text-subtle-light, #c4c8df);
  --text-muted: var(--color-text-muted-light, #9ea3c1);
  --panel: var(--color-panel-dark, rgba(255, 255, 255, 0.03));
  --border: var(--color-border-dark, rgba(255, 255, 255, 0.08));
}

:global(body) {
  background: var(--bg);
  color: var(--text);
  font-family: 'Inter', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2.5rem 1.5rem 3rem;
  display: grid;
  gap: 2rem;
}

.hero {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 1.5rem;
  align-items: center;
}

.logo-section {
  display: flex;
}

.logo-section--default {
  justify-content: flex-start;
}

.logo-section--center {
  justify-content: center;
}

.logo-section--right {
  justify-content: flex-end;
}

.hero.hero--logo-right {
  grid-template-columns: 1.1fr 0.9fr;
}

.hero.hero--bottom-panel {
  grid-template-columns: 1fr;
}

.hero.hero--logo-center .hero__copy {
  text-align: center;
}

.hero.hero--logo-center .hero__actions {
  justify-content: center;
}

.hero__copy h1 {
  font-size: clamp(2.2rem, 4vw, 3.3rem);
  line-height: 1.1;
  margin: 0.4rem 0;
  color: var(--text, #0e0f14);
}

.hero__copy h1 span {
  color: var(--color-primary, #16a34a);
  text-shadow: 0 0 18px rgba(159, 107, 255, 0.55);
}

.hero__copy p {
  color: var(--text-subtle);
  font-size: 1.05rem;
  margin: 0.5rem 0 1.1rem;
}

.hero__actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 1.1rem;
}

button.primary,
button.secondary {
  padding: 0.85rem 1.2rem;
  border-radius: 0.95rem;
  font-weight: 800;
  font-size: 1rem;
  letter-spacing: 0.01em;
  border: none;
  cursor: pointer;
  transition: transform 150ms ease, box-shadow 150ms ease, background 150ms ease, color 150ms ease;
}

button.primary {
  background: linear-gradient(135deg, var(--color-primary, #2563eb), var(--color-accent, #16a34a));
  color: var(--color-text-light, #e9edf6);
  box-shadow: 0 14px 36px rgba(37, 99, 235, 0.35);
}

button.secondary {
  background: var(--panel, rgba(0, 0, 0, 0.02));
  color: var(--text);
  border: 1px solid var(--border, rgba(0, 0, 0, 0.08));
}

button.primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 18px 40px rgba(159, 107, 255, 0.5);
}

button.secondary:hover {
  border-color: var(--color-primary, #2563eb);
}

.hero__stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 0.85rem;
}

.hero__aside {
  display: grid;
  gap: 0.75rem;
  align-content: start;
  justify-items: end;
}

.hero__stats div {
  padding: 0.75rem 0.9rem;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 0.9rem;
}

.hero__stats strong {
  display: block;
  font-size: 1.4rem;
}

.hero__stats span {
  color: var(--text-muted);
  font-size: 0.95rem;
}

.ticker-wrap {
  position: sticky;
  top: calc(var(--toolbar-height, 64px) + 0.75rem);
  z-index: 2;
}

.section {
  display: grid;
  gap: 1rem;
}

.grid {
  display: grid;
  gap: 1rem;
}

.grid--markets {
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
}

.section--split {
  grid-template-columns: 0.95fr 1.05fr;
}

.split__left,
.split__right {
  display: grid;
  gap: 1rem;
  align-content: start;
}

.stack {
  display: grid;
  gap: 0.85rem;
}

@media (max-width: 960px) {
  .hero {
    grid-template-columns: 1fr;
  }

  .section--split {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .page {
    padding: 1.8rem 1.2rem 2.6rem;
  }
}
</style>
