import type { LayoutLoad } from './$types';

// Branding is now selected client-side via a store (no query param).
export const load: LayoutLoad = () => {
	return {};
};
