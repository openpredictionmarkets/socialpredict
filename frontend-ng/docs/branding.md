# Branding configuration guide (`frontend-ng`)

This document explains how the branding system works for the new frontend (`frontend-ng`): where brand files live, what they control, and how they drive the page layout and appearance.

## Files and loading

- Brand definitions live under:
  - `src/config/branding/*.json` (e.g. `default.json`, `cloud.json`, `fire.json`, `earth.json`, `ocean.json`)
- They are loaded dynamically by:
  - `src/lib/config/branding.ts`
    - Uses `import.meta.glob('../../config/branding/*.json', { eager: true })`
    - Builds a `brandings` map keyed by the filename (without `.json`)
    - Exposes:
      - `DEFAULT_BRAND_KEY`
      - `resolveBranding(key)` → `Branding`
- Brand selection at runtime is managed via the store:
  - `src/lib/stores/brandingStore.ts`
    - `currentBrandKey` – writable store of the active brand key
    - `currentBranding` – derived store `resolveBranding($currentBrandKey)`
  - `BrandSelector.svelte` updates `currentBrandKey`
  - `+layout.svelte` and `+page.svelte` use `currentBranding` to drive CSS variables, layout, and content.

## Branding JSON shape

Every branding JSON file must conform to the `Branding` interface defined in `src/lib/config/branding.ts`:

```ts
export interface Branding {
  theme: 'light' | 'dark';
  notation?: string;
  brand: BrandMeta;
  colors: BrandColors;
  features: BrandFeatures;
  categories: BrandCategory[];
  examples: BrandExamples;
  layout: BrandLayout;
}
```

### `brand`: textual identity

```jsonc
"brand": {
  "name": "SocialPredict [Default]",
  "shortName": "SocialPredict",
  "taglinePrimary": "Predict the future.",
  "taglineSecondary": "Influence the conversation.",
  "description": "Trade YES/NO shares on real events...",
  "heroCtaPrimary": "Start Predicting",
  "heroCtaSecondary": "Learn How It Works"
}
```

Used by `HeroPanel`, the hero text in `+page.svelte`, the logo pill, and CTA button labels.

### `notation`: toolbar hint

```jsonc
"notation": "Default brand: dark, card-first layout"
```

- Optional short string that describes the brand.
- Displayed in the center of the `AppToolbar` via `Notation.svelte`, bound from `currentBranding.notation`.

### `theme` and `colors`: visual palette

- `theme`: `"light"` or `"dark"`
  - Used in `+layout.svelte` to set `data-theme` and choose light vs dark variants.
- `colors`: main palette and semantic roles:

```jsonc
"colors": {
  "background": "#0e0f14",
  "textDark": "#0e0f14",
  "textLight": "#e6e4ff",
  "textSubtleDark": "#4b5563",
  "textSubtleLight": "#c4c8df",
  "textMutedDark": "#6b7280",
  "textMutedLight": "#9ea3c1",
  "panelDark": "rgba(255, 255, 255, 0.03)",
  "panelLight": "rgba(0, 0, 0, 0.02)",
  "borderDark": "rgba(255, 255, 255, 0.08)",
  "borderLight": "rgba(0, 0, 0, 0.08)",
  "primary": "#9f6bff",
  "primarySoft": "rgba(159, 107, 255, 0.14)",
  "accent": "#35e2d1",
  "accentSoft": "rgba(53, 226, 209, 0.1)",
  "danger": "#ff6f61",
  "dangerSoft": "rgba(255, 111, 97, 0.1)"
}
```

`+layout.svelte` maps these onto CSS variables (`--color-background`, `--color-primary`, `--color-accent`, `--color-danger`, etc.), which are then used throughout the components (Hero, MarketCard, TutorialModal, MarketsList, etc.).

### `features`: on/off switches

```jsonc
"features": {
  "showLogo": true,
  "showHero": true,
  "showTicker": true,
  "showTrending": true,
  "showMarketsList": false,
  "showCategories": true,
  "showEndingSoon": true,
  "showTutorial": true
}
```

Controls which sections are rendered in `+page.svelte`:

- `showLogo` – render logo section(s)
- `showHero` – render hero section(s)
- `showTicker` – render `TickerBar`
- `showTrending` – render trending markets section
- `showMarketsList` – render tabular `MarketsList`
- `showCategories` – render categories section
- `showEndingSoon` – render ending‑soon section
- `showTutorial` – show the tutorial CTA and modal

### `categories`: icon + color per category

```jsonc
"categories": [
  {
    "id": "politics",
    "name": "Politics",
    "color": "#94a3b8",
    "icon": "Flag"
  },
  ...
]
```

- `name` is used in UIs (CategoryGrid, MarketCard category pill).
- `color` drives category pill color and category icon color.
- `icon` must match a key in `categoryIcons` in `+page.svelte` (e.g. `"Flag"`, `"Coins"`, `"MonitorSmartphone"`).

### `examples`: demo content

```jsonc
"examples": {
  "tickerItems": [ ... ],
  "heroStats": [ ... ],
  "trendingMarkets": [ ... ],
  "endingSoon": [ ... ]
}
```

All of these are used on the landing page:

- `tickerItems` → `TickerBar`
- `heroStats` → stats row in the hero
- `trendingMarkets` → cards in the Trending section and rows in `MarketsList`
- `endingSoon` → cards in the Ending Soon section and rows in `MarketsList`

Two helper components consume these:

- `MarketsCards.svelte` – renders a responsive grid of `MarketCard` items from a list of markets.
- `MarketsList.svelte` – renders a table view of markets, driven by an optional `filter` prop that selects which fields (columns) to show.

The Svelte `Market` type is defined in `MarketCard.svelte` and must match the shape of the markets in `trendingMarkets` / `endingSoon`.

## Layout configuration

The most important part of branding is `layout.sections`, which controls **which** sections appear on the main page and in what order.

```jsonc
"layout": {
  "sections": [
    { "type": "logo", "variant": "default", "span": "full" },
    { "type": "hero", "variant": "default", "span": "full" },
    { "type": "ticker", "span": "full" },
    { "type": "trending", "span": "half" },
    { "type": "marketsList", "span": "half" },
    { "type": "categories", "span": "full" },
    { "type": "endingSoon", "span": "half" }
  ]
}
```

### Section types

Valid `type` values are:

- `"logo"` – renders the logo pill at the top
- `"hero"` – hero marketing block
- `"ticker"` – sticky ticker bar
- `"trending"` – Trending Right Now card grid
- `"marketsList"` – tabular list of markets (`MarketsList`)
- `"categories"` – CategoryGrid
- `"endingSoon"` – Ending Soon card stack

These section types map directly to the `{:else if section.type === '...'}` branches in `src/routes/+page.svelte`.

### `variant` on hero and logo sections

Some sections support `variant` hints to change layout:

- `logo`:
  - `"default"` – left aligned
  - `"logoRight"` – right aligned
  - `"logoCenter"` – centered
- `hero`:
  - `"default"` – panel on the right
  - `"leftPanel"` – panel on the left
  - `"logoRight"` – hero with panel on the right and logo aside
  - `"logoCenter"` – centered hero text
  - `"bottomPanel"` – panel below hero copy
  - `"noPanel"` – hero copy and stats only (no `HeroPanel`)

These are interpreted in the hero/logo branches in `+page.svelte` and only affect layout, not content.

### `span`: column span hint

The page uses a **two‑column grid** at the `main.page` level:

```css
.page {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1.2fr);
  gap: 2rem;
}

.section,
.hero,
.ticker-wrap,
.logo-section {
  grid-column: 1 / -1; /* full width by default */
}

.section--span-half,
.hero.section--span-half,
.ticker-wrap.section--span-half,
.logo-section.section--span-half {
  grid-column: auto; /* one column */
}
```

Each section in `layout.sections` may set:

- `"span": "full"` – optional, the default; section spans both columns
- `"span": "half"` – section occupies one column of the grid

Example:

```jsonc
"sections": [
  { "type": "logo", "variant": "default" },                 // full width (default)
  { "type": "hero", "variant": "default", "span": "full" }, // full width
  { "type": "ticker", "span": "half" },                     // column 1
  { "type": "marketsList", "span": "half" },                // column 2
  { "type": "categories", "span": "full" },                 // full width below
  { "type": "endingSoon", "span": "half" }                  // column 1 or 2 depending on position
]
```

In `+page.svelte`, this is interpreted by adding a `section--span-half` class when `section.span === 'half'` for all relevant section wrappers.

## Adding a new brand

1. Create a new JSON file under `src/config/branding`, e.g. `cloud.json`.
2. Copy the shape from `default.json` and adjust:
   - `brand` text
   - `theme` and `colors`
   - `features` toggles
   - `layout.sections` (order, type, `variant`, `span`)
   - `categories` and `examples`
3. The new brand is picked up automatically by `branding.ts` and appears in:
   - `brandKeys`
   - `BrandSelector` options
4. Select it via the dropdown or set `DEFAULT_BRAND_KEY` (in `branding.ts`) during development.

## Common patterns

- **Minimal hero‑only layout**:

  ```jsonc
  "features": {
    "showLogo": false,
    "showHero": true,
    "showTicker": false,
    "showTrending": false,
    "showMarketsList": false,
    "showCategories": false,
    "showEndingSoon": false,
    "showTutorial": true
  },
  "layout": {
    "sections": [
      { "type": "hero", "variant": "default", "span": "full" }
    ]
  }
  ```

- **Data‑heavy layout with table and cards side‑by‑side**:

  ```jsonc
  "layout": {
    "sections": [
      { "type": "logo", "variant": "default" },
      { "type": "hero", "variant": "default", "span": "full" },
      { "type": "ticker", "span": "full" },
      { "type": "trending", "span": "half" },
      { "type": "marketsList", "span": "half" },
      { "type": "categories", "span": "full" },
      { "type": "endingSoon", "span": "half" }
    ]
  }
  ```

## Where to change what

- Change text, palette, examples, features, and layout:
  - `src/config/branding/*.json`
- Change how brands are loaded and typed:
  - `src/lib/config/branding.ts`
- Change how brands are applied to CSS variables:
  - `src/routes/+layout.svelte`
- Change how layout sections map to UI:
  - `src/routes/+page.svelte`
- Change which properties are exposed in the Brand selector:
  - `src/lib/components/BrandSelector.svelte`

This setup lets you build very different landing page experiences by editing only the JSON branding files, without touching the Svelte components for most changes.
