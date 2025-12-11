const brandingModules = import.meta.glob('../../config/branding/*.json', {
	eager: true
}) as Record<string, { default: unknown }>;

export interface BrandMeta {
  name: string;
  shortName: string;
  taglinePrimary: string;
  taglineSecondary: string;
  description: string;
  heroCtaPrimary: string;
  heroCtaSecondary: string;
}

export interface BrandColors {
  background: string;
  textDark: string;
  textLight: string;
  textSubtleDark: string;
  textSubtleLight: string;
  textMutedDark: string;
  textMutedLight: string;
  panelDark: string;
  panelLight: string;
  borderDark: string;
  borderLight: string;
  primary: string;
  primarySoft: string;
  accent: string;
  accentSoft: string;
  danger: string;
  dangerSoft: string;
}

export interface BrandFeatures {
  showHero: boolean;
  showTicker: boolean;
  showTrending: boolean;
  showMarketsList: boolean;
  showCategories: boolean;
  showEndingSoon: boolean;
  showTutorial: boolean;
  showLogo: boolean;
}

export interface BrandCategory {
  id: string;
  name: string;
  color: string;
  icon: string;
}

export interface BrandTickerItem {
  label: string;
  yes: number;
  no: number;
  trend: number;
}

export interface BrandHeroStat {
  label: string;
  value: string;
}

export interface BrandExampleMarket {
  title: string;
  yes: number;
  no: number;
  community: string;
  resolves: string;
  trend: number;
  liquidity: 'deep' | 'moderate' | 'thin' | 'very-thin';
  category: string;
  sparkline?: string;
}

export interface BrandExamples {
  tickerItems: BrandTickerItem[];
  heroStats: BrandHeroStat[];
  trendingMarkets: BrandExampleMarket[];
  endingSoon: BrandExampleMarket[];
}

export type LayoutSectionType =
  | 'logo'
  | 'hero'
  | 'ticker'
  | 'trending'
  | 'marketsList'
  | 'categories'
  | 'endingSoon';

export interface BrandLayoutSection {
  type: LayoutSectionType;
  variant?: string;
}

export interface BrandLayout {
  sections: BrandLayoutSection[];
}

export interface Branding {
  theme: 'light' | 'dark';
  brand: BrandMeta;
  colors: BrandColors;
  features: BrandFeatures;
  categories: BrandCategory[];
  examples: BrandExamples;
  layout: BrandLayout;
}

// Build brandings map dynamically from files under config/branding
export const brandings = Object.entries(brandingModules).reduce(
  (acc, [path, module]) => {
    const match = path.match(/\/([^/]+)\.json$/);
    if (!match) return acc;
    const key = match[1];
    acc[key] = module.default as Branding;
    return acc;
  },
  {} as Record<string, Branding>
);

export type BrandKey = keyof typeof brandings;

export const brandKeys = Object.keys(brandings) as BrandKey[];

export const DEFAULT_BRAND_KEY: BrandKey =
  (brandKeys.includes('default' as BrandKey) ? 'default' : brandKeys[0]) as BrandKey;

export function resolveBranding(key: BrandKey | string | null | undefined): Branding {
  if (key && key in brandings) {
    return brandings[key as BrandKey];
  }
  return brandings[DEFAULT_BRAND_KEY];
}

export const branding: Branding = resolveBranding(DEFAULT_BRAND_KEY);
