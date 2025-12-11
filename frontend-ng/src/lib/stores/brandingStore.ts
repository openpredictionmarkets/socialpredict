import { derived, writable } from 'svelte/store';
import {
  brandings,
  brandKeys,
  DEFAULT_BRAND_KEY,
  resolveBranding,
  type BrandKey,
  type Branding
} from '$lib/config/branding';

export const currentBrandKey = writable<BrandKey>(DEFAULT_BRAND_KEY);

export const currentBranding = derived<typeof currentBrandKey, Branding>(
  currentBrandKey,
  (key) => resolveBranding(key)
);

export const brandOptions = brandKeys;

