import React, { useMemo, useState } from 'react';

const layoutModes = {
  top: {
    label: 'Market page layout (TOP)',
    route: '/markets',
    purpose: 'Search-first discovery page for all public markets.',
    defaultRecommendationLimit: 20,
    curatedRecommendationLimit: 5,
    fixedBlocks: [
      'Search bar',
      'Status tabs: Active, Closed, Resolved, All',
      'Recommendation fallback',
    ],
    optionalBlocks: [
      'Featured topic/category cards',
      'Featured market cards',
      'CMS sections',
    ],
  },
  secondary: {
    label: 'Topic page layout (SECONDARY)',
    route: '/markets/topic/:slug',
    purpose: 'Tag-scoped market discovery page for a CMS-managed topic.',
    defaultRecommendationLimit: 20,
    curatedRecommendationLimit: 5,
    fixedBlocks: [
      'Topic title and description',
      'Search bar filtered by topic tag',
      'Status tabs: Active, Closed, Resolved, All',
    ],
    optionalBlocks: [
      'Pinned markets for this topic',
      'Topic sections',
      'Section-specific featured markets',
    ],
  },
};

const plannedPersistence = [
  'market_discovery_pages',
  'market_discovery_sections',
  'market_discovery_pins',
];

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
  const hasCuratedBlocks = state.featuredTopics || state.featuredMarkets || state.sections;
  const recommendationLimit = hasCuratedBlocks
    ? mode.curatedRecommendationLimit
    : mode.defaultRecommendationLimit;

  return (
    <div className="rounded-xl border border-gray-700 bg-gray-950 p-5 shadow-lg">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-sky-300">Preview Contract</p>
          <h3 className="mt-2 text-xl font-bold text-white">{mode.label}</h3>
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
        {state.featuredTopics && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Featured topic/category cards</div>
            <div className="mt-1 text-xs text-gray-400">Pinned secondary pages appear after compact recommendations.</div>
          </div>
        )}
        {state.featuredMarkets && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Featured market cards</div>
            <div className="mt-1 text-xs text-gray-400">Admin-pinned markets appear in configured order.</div>
          </div>
        )}
        {state.sections && (
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
    top: {
      searchScope: 'all',
      featuredTopics: true,
      featuredMarkets: true,
      sections: false,
    },
    secondary: {
      searchScope: 'tag',
      featuredTopics: false,
      featuredMarkets: true,
      sections: true,
    },
  });

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
    ...(state.featuredTopics ? ['Featured topic/category cards'] : []),
    ...(state.featuredMarkets ? ['Featured market cards'] : []),
    ...(state.sections ? ['CMS sections'] : []),
  ], [mode, state]);

  return (
    <div className="min-h-screen bg-primary-background p-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-sky-300">CMS</p>
          <h1 className="mt-2 text-3xl font-bold text-white">Market Discovery Layout</h1>
          <p className="mt-2 max-w-3xl text-gray-300">
            Scaffold the TOP market page and SECONDARY topic page layout options from FEATURE/09.
            This panel defines the admin-facing CMS shape before page/section/pin persistence is added.
          </p>
        </div>

        <div className="rounded-lg border border-amber-500/40 bg-amber-950/40 p-4 text-sm text-amber-100">
          These controls are a planning scaffold only. Saving will be enabled after backend CMS tables
          and APIs are added for discovery pages, sections, and pins.
        </div>

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
              <h2 className="font-semibold text-white">Planned Backend Tables</h2>
              <div className="mt-3 flex flex-wrap gap-2">
                {plannedPersistence.map((table) => (
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
                  disabled
                  className="rounded-md bg-gray-700 px-4 py-2 text-sm font-semibold text-gray-400 disabled:cursor-not-allowed"
                  title="Backend persistence is planned in FEATURE/09 sections 06 and 07."
                >
                  Save Layout Later
                </button>
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <ToggleCard
                  title="Featured topic/category cards"
                  description="Show pinned SECONDARY pages on the TOP page. Usually disabled inside SECONDARY topic pages."
                  checked={state.featuredTopics}
                  onChange={(checked) => updateState({ featuredTopics: checked })}
                />
                <ToggleCard
                  title="Featured market cards"
                  description="Show manually pinned markets before long fallback lists."
                  checked={state.featuredMarkets}
                  onChange={(checked) => updateState({ featuredMarkets: checked })}
                />
                <ToggleCard
                  title="CMS sections"
                  description="Allow named sections beyond the implicit All section."
                  checked={state.sections}
                  onChange={(checked) => updateState({ sections: checked })}
                />
                <div className="rounded-lg border border-gray-700 bg-gray-950 p-4">
                  <label className="block font-semibold text-white" htmlFor="search-scope">
                    Search Scope
                  </label>
                  <select
                    id="search-scope"
                    value={state.searchScope}
                    onChange={(event) => updateState({ searchScope: event.target.value })}
                    className="mt-3 w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                  >
                    <option value="all">All public markets</option>
                    <option value="tag">Current topic/tag by default</option>
                  </select>
                  <p className="mt-2 text-sm text-gray-400">
                    TOP should usually use all markets; SECONDARY should usually use topic/tag scope.
                  </p>
                </div>
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
