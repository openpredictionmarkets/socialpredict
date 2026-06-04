import React, { useEffect, useMemo, useState } from 'react';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';

const layoutModes = {
  top: {
    slug: 'markets',
    label: 'Market page layout (TOP)',
    route: '/markets',
    purpose: 'Search-first discovery page for all public markets.',
    pageType: 'top',
    fixedBlocks: [
      'Search bar',
      'Status tabs: Active, Closed, Resolved, All',
      'Recommendation fallback',
    ],
  },
  secondary: {
    slug: 'topic-template',
    label: 'Topic page layout (SECONDARY)',
    route: '/markets/topic/:slug',
    purpose: 'Tag-scoped market discovery page for a CMS-managed topic.',
    pageType: 'secondary',
    fixedBlocks: [
      'Topic title and description',
      'Search bar filtered by topic tag',
      'Status tabs: Active, Closed, Resolved, All',
    ],
  },
};

const persistedTables = [
  'market_discovery_pages',
  'market_discovery_sections',
  'market_discovery_pins',
];

const defaultPageState = {
  slug: 'markets',
  title: 'Markets',
  description: 'Browse and search prediction markets.',
  pageType: 'top',
  primaryTagSlug: '',
  searchScope: 'all',
  featuredTopicsEnabled: false,
  featuredMarketsEnabled: false,
  sectionsEnabled: false,
  defaultRecommendationLimit: 20,
  curatedRecommendationLimit: 5,
  recommendationLimit: 20,
  isPublished: true,
  version: 0,
};

const normalizePage = (page = {}, modeKey = 'top') => {
  const mode = layoutModes[modeKey];
  const hasCuratedBlocks = !!(page.featuredTopicsEnabled || page.featuredMarketsEnabled || page.sectionsEnabled);
  const defaultRecommendationLimit = Number(page.defaultRecommendationLimit || 20);
  const curatedRecommendationLimit = Number(page.curatedRecommendationLimit || 5);

  return {
    ...defaultPageState,
    slug: page.slug || mode.slug,
    title: page.title || (modeKey === 'secondary' ? 'Topic Markets' : 'Markets'),
    description: page.description || (modeKey === 'secondary' ? 'Browse and search markets in this topic.' : 'Browse and search prediction markets.'),
    pageType: page.pageType || mode.pageType,
    primaryTagSlug: page.primaryTagSlug || '',
    searchScope: page.searchScope || (modeKey === 'secondary' ? 'tag' : 'all'),
    featuredTopicsEnabled: !!page.featuredTopicsEnabled,
    featuredMarketsEnabled: !!page.featuredMarketsEnabled,
    sectionsEnabled: !!page.sectionsEnabled,
    defaultRecommendationLimit,
    curatedRecommendationLimit,
    recommendationLimit: hasCuratedBlocks ? curatedRecommendationLimit : defaultRecommendationLimit,
    isPublished: page.isPublished !== false,
    version: Number(page.version || 0),
  };
};

const ToggleCard = ({ title, description, checked, onChange }) => (
  <label className={`flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition ${
    checked
      ? 'border-primary-pink bg-primary-pink/10'
      : 'border-gray-700 bg-gray-900/70 hover:border-gray-500'
  }`}>
    <input
      type="checkbox"
      checked={checked}
      onChange={(event) => onChange(event.target.checked)}
      className="mt-1 h-4 w-4 accent-primary-pink"
    />
    <span>
      <span className="block font-semibold text-white">{title}</span>
      <span className="mt-1 block text-sm text-gray-400">{description}</span>
    </span>
  </label>
);

const Pill = ({ children }) => (
  <span className="rounded-full border border-sky-500/40 bg-sky-950/40 px-3 py-1 text-xs font-semibold text-sky-100">
    {children}
  </span>
);

const LayoutPreview = ({ mode, state }) => {
  const hasCuratedBlocks = state.featuredTopicsEnabled || state.featuredMarketsEnabled || state.sectionsEnabled;
  const recommendationLimit = hasCuratedBlocks
    ? state.curatedRecommendationLimit
    : state.defaultRecommendationLimit;

  return (
    <div className="rounded-xl border border-gray-700 bg-gray-950 p-5 shadow-lg">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-sky-300">Preview Contract</p>
          <h3 className="mt-2 text-xl font-bold text-white">{state.title}</h3>
          {state.description && <p className="mt-1 text-sm text-gray-400">{state.description}</p>}
        </div>
        <Pill>{mode.route}</Pill>
      </div>

      <div className="mt-5 space-y-3">
        <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
          <div className="text-sm font-semibold text-white">1. Search first</div>
          <div className="mt-1 text-xs text-gray-400">
            {state.searchScope === 'tag' ? 'Search defaults to the page tag.' : 'Search defaults to all public markets.'}
          </div>
        </div>
        <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
          <div className="text-sm font-semibold text-white">2. Recommendations</div>
          <div className="mt-1 text-xs text-gray-400">
            Show {recommendationLimit} fallback markets when CMS content is {hasCuratedBlocks ? 'present' : 'empty'}.
          </div>
        </div>
        {state.featuredTopicsEnabled && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Featured topic/category cards</div>
            <div className="mt-1 text-xs text-gray-400">Pinned secondary pages appear after compact recommendations.</div>
          </div>
        )}
        {state.featuredMarketsEnabled && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Featured market cards</div>
            <div className="mt-1 text-xs text-gray-400">Admin-pinned markets appear in configured order.</div>
          </div>
        )}
        {state.sectionsEnabled && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Sections</div>
            <div className="mt-1 text-xs text-gray-400">Explicit sections render after pins; otherwise the page has an implicit All section.</div>
          </div>
        )}
      </div>
    </div>
  );
};

function MarketDiscoveryLayoutEditor() {
  const [selectedMode, setSelectedMode] = useState('top');
  const [layoutState, setLayoutState] = useState({
    top: normalizePage({}, 'top'),
    secondary: normalizePage({}, 'secondary'),
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadPages = async () => {
      setLoading(true);
      setError('');
      try {
        const [topPage, secondaryPage] = await Promise.all([
          apiRequest('/v0/content/market-discovery/markets', { fallbackMessage: 'Failed to load TOP layout.' }),
          apiRequest('/v0/content/market-discovery/topic-template', { fallbackMessage: 'Failed to load SECONDARY layout.' }),
        ]);
        if (!ignore) {
          setLayoutState({
            top: normalizePage(topPage, 'top'),
            secondary: normalizePage(secondaryPage, 'secondary'),
          });
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load market discovery layouts.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadPages();

    return () => {
      ignore = true;
    };
  }, []);

  const mode = layoutModes[selectedMode];
  const state = layoutState[selectedMode];
  const updateState = (updates) => {
    setLayoutState((current) => ({
      ...current,
      [selectedMode]: {
        ...current[selectedMode],
        ...updates,
      },
    }));
  };

  const selectedBlocks = useMemo(() => [
    ...mode.fixedBlocks,
    ...(state.featuredTopicsEnabled ? ['Featured topic/category cards'] : []),
    ...(state.featuredMarketsEnabled ? ['Featured market cards'] : []),
    ...(state.sectionsEnabled ? ['CMS sections'] : []),
  ], [mode, state]);

  const saveLayout = async () => {
    setSaving(true);
    setMessage('');
    setError('');
    try {
      const saved = await authenticatedApiRequest(`/v0/admin/content/market-discovery/${mode.slug}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...state,
          pageType: mode.pageType,
        }),
        fallbackMessage: 'Failed to save market discovery layout.',
      });
      setLayoutState((current) => ({
        ...current,
        [selectedMode]: normalizePage(saved, selectedMode),
      }));
      setMessage(`${mode.label} saved. /markets will use the TOP layout immediately after refresh.`);
    } catch (err) {
      setError(err.message || 'Unable to save market discovery layout.');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-primary-background p-8 text-white">
        Loading market discovery layouts...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-primary-background p-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-sky-300">CMS</p>
          <h1 className="mt-2 text-3xl font-bold text-white">Market Discovery Layout</h1>
          <p className="mt-2 max-w-3xl text-gray-300">
            Configure persisted TOP market page and SECONDARY topic page layout options from FEATURE/09.
            The TOP record now controls the public /markets page title, description, fallback list size, and CMS block visibility.
          </p>
        </div>

        {message && <div className="rounded-lg bg-emerald-700 p-4 text-sm text-white">{message}</div>}
        {error && <div className="rounded-lg bg-red-700 p-4 text-sm text-white">{error}</div>}

        <div className="grid gap-6 lg:grid-cols-[320px,minmax(0,1fr)]">
          <div className="space-y-4">
            <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-4">
              <h2 className="font-semibold text-white">Layout Type</h2>
              <div className="mt-4 grid gap-3">
                {Object.entries(layoutModes).map(([key, item]) => (
                  <button
                    key={key}
                    type="button"
                    onClick={() => setSelectedMode(key)}
                    className={`rounded-lg border p-4 text-left transition ${
                      selectedMode === key
                        ? 'border-primary-pink bg-primary-pink/15 text-white'
                        : 'border-gray-700 bg-gray-950 text-gray-300 hover:border-gray-500'
                    }`}
                  >
                    <span className="block font-semibold">{item.label}</span>
                    <span className="mt-1 block text-xs text-gray-400">{item.route}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-4">
              <h2 className="font-semibold text-white">Persisted Backend Tables</h2>
              <div className="mt-3 flex flex-wrap gap-2">
                {persistedTables.map((table) => (
                  <Pill key={table}>{table}</Pill>
                ))}
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-5">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <h2 className="text-2xl font-bold text-white">{mode.label}</h2>
                  <p className="mt-2 max-w-2xl text-sm text-gray-300">{mode.purpose}</p>
                </div>
                <button
                  type="button"
                  disabled={saving}
                  onClick={saveLayout}
                  className="rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-pink/80 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Layout'}
                </button>
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <label className="grid gap-2 text-sm text-gray-300">
                  Page Title
                  <input
                    value={state.title}
                    maxLength={160}
                    onChange={(event) => updateState({ title: event.target.value })}
                    className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  />
                </label>
                <label className="grid gap-2 text-sm text-gray-300">
                  Search Scope
                  <select
                    value={state.searchScope}
                    onChange={(event) => updateState({ searchScope: event.target.value })}
                    className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  >
                    <option value="all">All public markets</option>
                    <option value="tag">Current topic/tag by default</option>
                  </select>
                </label>
                <label className="grid gap-2 text-sm text-gray-300 md:col-span-2">
                  Page Description
                  <textarea
                    value={state.description}
                    maxLength={500}
                    rows={3}
                    onChange={(event) => updateState({ description: event.target.value })}
                    className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  />
                </label>
                <label className="grid gap-2 text-sm text-gray-300">
                  Empty CMS Recommendation Limit
                  <input
                    type="number"
                    min="1"
                    max="50"
                    value={state.defaultRecommendationLimit}
                    onChange={(event) => updateState({ defaultRecommendationLimit: Number(event.target.value || 1) })}
                    className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  />
                </label>
                <label className="grid gap-2 text-sm text-gray-300">
                  Curated CMS Recommendation Limit
                  <input
                    type="number"
                    min="1"
                    max="50"
                    value={state.curatedRecommendationLimit}
                    onChange={(event) => updateState({ curatedRecommendationLimit: Number(event.target.value || 1) })}
                    className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  />
                </label>
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <ToggleCard
                  title="Featured topic/category cards"
                  description="Show pinned SECONDARY pages on the TOP page. Usually disabled inside SECONDARY topic pages."
                  checked={state.featuredTopicsEnabled}
                  onChange={(checked) => updateState({ featuredTopicsEnabled: checked })}
                />
                <ToggleCard
                  title="Featured market cards"
                  description="Show manually pinned markets before long fallback lists."
                  checked={state.featuredMarketsEnabled}
                  onChange={(checked) => updateState({ featuredMarketsEnabled: checked })}
                />
                <ToggleCard
                  title="CMS sections"
                  description="Allow named sections beyond the implicit All section."
                  checked={state.sectionsEnabled}
                  onChange={(checked) => updateState({ sectionsEnabled: checked })}
                />
                <ToggleCard
                  title="Published"
                  description="Unpublished secondary layouts stay editable but should not appear publicly once topic routes are added."
                  checked={state.isPublished}
                  onChange={(checked) => updateState({ isPublished: checked })}
                />
              </div>

              <div className="mt-6 rounded-lg border border-gray-700 bg-gray-950 p-4">
                <h3 className="font-semibold text-white">Selected Render Blocks</h3>
                <ol className="mt-3 list-decimal space-y-2 pl-5 text-sm text-gray-300">
                  {selectedBlocks.map((block) => (
                    <li key={block}>{block}</li>
                  ))}
                </ol>
              </div>
            </div>

            <LayoutPreview mode={mode} state={state} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default MarketDiscoveryLayoutEditor;
