import React, { useEffect, useMemo, useState } from 'react';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';
import { getMarketDetails, searchMarkets } from '../../api/marketsApi';
import { listAdminMarketTags } from '../../api/marketTagsApi';

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
    label: 'Topic page pins (SECONDARY)',
    route: '/markets/topic/:slug',
    purpose: 'Tag-scoped market discovery page with independent per-topic market pins.',
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
  sections: [],
  pins: [],
};

const sortBySortOrder = (items = []) => [...items].sort((a, b) => Number(a.sortOrder || 0) - Number(b.sortOrder || 0));

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
    sections: sortBySortOrder(page.sections || []),
    pins: sortBySortOrder(page.pins || []),
  };
};

const nextSortOrder = (items = []) => items.reduce((max, item) => Math.max(max, Number(item.sortOrder || 0)), 0) + 1;

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

const Field = ({ label, children, className = '' }) => (
  <label className={`grid gap-2 text-sm text-gray-300 ${className}`}>
    {label}
    {children}
  </label>
);

const textInputClass = 'w-full min-w-0 rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40';

const tagOptionLabel = (tag) => `${tag.displayName || tag.slug} (${tag.slug})`;

const tagOptionsWithCurrent = (tags, currentSlug) => {
  const normalizedCurrent = String(currentSlug || '').trim();
  if (!normalizedCurrent || tags.some((tag) => tag.slug === normalizedCurrent)) {
    return tags;
  }

  return [
    ...tags,
    {
      slug: normalizedCurrent,
      displayName: `Saved slug: ${normalizedCurrent}`,
      isActive: false,
    },
  ];
};

const marketOverviewTitle = (overview) => (
  overview?.market?.questionTitle
  || overview?.questionTitle
  || `Market #${overview?.market?.id || overview?.id || 'unknown'}`
);

const marketOverviewId = (overview) => overview?.market?.id || overview?.id || overview?.marketId || 0;

const MarketPinSearch = ({ pin, onSelect, tagSlug = '' }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState('');
  const [selectedTitle, setSelectedTitle] = useState('');
  const [loadingSelectedTitle, setLoadingSelectedTitle] = useState(false);

  useEffect(() => {
    const marketId = Number(pin.marketId);
    if (!marketId) {
      setSelectedTitle('');
      setLoadingSelectedTitle(false);
      return undefined;
    }

    let ignore = false;
    setLoadingSelectedTitle(true);
    getMarketDetails(marketId)
      .then((details) => {
        if (!ignore) {
          setSelectedTitle(details?.market?.questionTitle || `Market #${marketId}`);
        }
      })
      .catch(() => {
        if (!ignore) {
          setSelectedTitle(`Market #${marketId}`);
        }
      })
      .finally(() => {
        if (!ignore) {
          setLoadingSelectedTitle(false);
        }
      });

    return () => {
      ignore = true;
    };
  }, [pin.marketId]);

  useEffect(() => {
    const trimmed = query.trim();
    if (trimmed.length < 2) {
      setResults([]);
      setSearchError('');
      return undefined;
    }

    let ignore = false;
    const timeoutId = setTimeout(async () => {
      setSearching(true);
      setSearchError('');
      try {
        const response = await searchMarkets(trimmed, 'active', 8, tagSlug ? { tagSlug } : {});
        if (!ignore) {
          setResults(response.primaryResults || []);
        }
      } catch (err) {
        if (!ignore) {
          setSearchError(err.message || 'Unable to search active markets.');
          setResults([]);
        }
      } finally {
        if (!ignore) {
          setSearching(false);
        }
      }
    }, 250);

    return () => {
      ignore = true;
      clearTimeout(timeoutId);
    };
  }, [query, tagSlug]);

  return (
    <div className="grid gap-3">
      <div className="rounded-lg border border-gray-700 bg-gray-900/70 p-3">
        <div className="font-mono text-xs uppercase tracking-[0.14em] text-gray-400">Selected Market</div>
        {Number(pin.marketId) > 0 ? (
          <>
            <div className="mt-1 text-sm font-semibold text-white">
              {loadingSelectedTitle ? 'Loading selected market...' : selectedTitle || `Market #${pin.marketId}`}
            </div>
            <div className="mt-1 text-xs text-gray-400">Market #{pin.marketId}</div>
          </>
        ) : (
          <div className="mt-1 text-sm text-white">No market selected yet.</div>
        )}
      </div>
      <Field label="Search Active Markets">
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={tagSlug ? `Search active ${tagSlug} markets...` : 'Type at least 2 characters...'}
          className={textInputClass}
        />
      </Field>
      {searching && <p className="text-xs text-gray-400">Searching active markets...</p>}
      {searchError && <p className="text-xs text-red-300">{searchError}</p>}
      {results.length > 0 && (
        <div className="grid gap-2">
          {results.map((overview) => {
            const id = marketOverviewId(overview);
            return (
              <button
                key={id}
                type="button"
                onClick={() => {
                  setSelectedTitle(marketOverviewTitle(overview));
                  onSelect(id);
                  setQuery('');
                  setResults([]);
                }}
                className={`rounded-lg border px-3 py-2 text-left text-sm transition ${
                  Number(pin.marketId) === Number(id)
                    ? 'border-primary-pink bg-primary-pink/15 text-white'
                    : 'border-gray-700 bg-gray-800 text-gray-200 hover:border-sky-400/60'
                }`}
              >
                <span className="block font-semibold">{marketOverviewTitle(overview)}</span>
                <span className="mt-1 block text-xs text-gray-400">Active market #{id}</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
};

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
            <div className="text-sm font-semibold text-white">Topic Pins</div>
            <div className="mt-1 text-xs text-gray-400">{state.pins.filter((pin) => pin.pinType === 'discovery_page').length} page pins configured.</div>
          </div>
        )}
        {state.featuredMarketsEnabled && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Market Pins</div>
            <div className="mt-1 text-xs text-gray-400">{state.pins.filter((pin) => pin.pinType === 'market').length} market pins configured.</div>
          </div>
        )}
        {state.sectionsEnabled && (
          <div className="rounded-lg border border-gray-700 bg-gray-900 p-4">
            <div className="text-sm font-semibold text-white">Sections</div>
            <div className="mt-1 text-xs text-gray-400">{state.sections.length} named sections configured; otherwise the page has an implicit All section.</div>
          </div>
        )}
      </div>
    </div>
  );
};

function MarketDiscoveryLayoutEditor() {
  const [selectedMode, setSelectedMode] = useState('top');
  const [selectedSecondarySlug, setSelectedSecondarySlug] = useState('');
  const [layoutState, setLayoutState] = useState({
    top: normalizePage({}, 'top'),
    secondary: normalizePage({}, 'secondary'),
  });
  const [loading, setLoading] = useState(true);
  const [loadingSecondary, setLoadingSecondary] = useState(false);
  const [saving, setSaving] = useState(false);
  const [savingSections, setSavingSections] = useState(false);
  const [savingPins, setSavingPins] = useState(false);
  const [activeTags, setActiveTags] = useState([]);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadPages = async () => {
      setLoading(true);
      setError('');
      try {
        const topPage = await apiRequest('/v0/content/market-discovery/markets', { fallbackMessage: 'Failed to load TOP layout.' });
        if (!ignore) {
          setLayoutState({
            top: normalizePage(topPage, 'top'),
            secondary: normalizePage({}, 'secondary'),
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

  useEffect(() => {
    let ignore = false;

    const loadActiveTags = async () => {
      try {
        const result = await listAdminMarketTags({ includeInactive: false });
        if (!ignore) {
          const tags = result.tags || [];
          setActiveTags(tags);
          setSelectedSecondarySlug((current) => current || tags[0]?.slug || '');
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load active market tags.');
        }
      }
    };

    loadActiveTags();

    return () => {
      ignore = true;
    };
  }, []);

  useEffect(() => {
    if (!selectedSecondarySlug) {
      return undefined;
    }

    let ignore = false;

    const loadSecondaryPage = async () => {
      setLoadingSecondary(true);
      setError('');
      try {
        const page = await apiRequest(`/v0/content/market-discovery/${selectedSecondarySlug}`, {
          fallbackMessage: 'Failed to load topic page layout.',
        });
        if (!ignore) {
          setLayoutState((current) => ({
            ...current,
            secondary: {
              ...normalizePage(page, 'secondary'),
              slug: selectedSecondarySlug,
              primaryTagSlug: selectedSecondarySlug,
              searchScope: 'tag',
              featuredTopicsEnabled: false,
              featuredMarketsEnabled: true,
              sectionsEnabled: false,
            },
          }));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load topic page layout.');
        }
      } finally {
        if (!ignore) {
          setLoadingSecondary(false);
        }
      }
    };

    loadSecondaryPage();

    return () => {
      ignore = true;
    };
  }, [selectedSecondarySlug]);

  const mode = layoutModes[selectedMode];
  const state = layoutState[selectedMode];
  const selectedSecondaryTag = activeTags.find((tag) => tag.slug === selectedSecondarySlug);
  const selectedPageSlug = selectedMode === 'secondary' ? selectedSecondarySlug : mode.slug;
  const selectedPageLabel = selectedMode === 'secondary' && selectedSecondarySlug
    ? `${mode.label}: ${tagOptionLabel(selectedSecondaryTag || { slug: selectedSecondarySlug })}`
    : mode.label;
  const selectedRoute = selectedMode === 'secondary' && selectedSecondarySlug
    ? `/markets/topic/${selectedSecondarySlug}`
    : mode.route;
  const canEditSelectedPage = selectedMode !== 'secondary' || !!selectedSecondarySlug;
  const topicPins = state.pins
    .map((pin, index) => ({ pin, index }))
    .filter(({ pin }) => pin.pinType === 'discovery_page');
  const marketPins = state.pins
    .map((pin, index) => ({ pin, index }))
    .filter(({ pin }) => pin.pinType === 'market');
  const activeTagOptions = useMemo(() => (
    [...activeTags].sort((left, right) => (
      (left.displayName || left.slug).localeCompare(right.displayName || right.slug)
    ))
  ), [activeTags]);
  const firstActiveTagSlug = activeTagOptions[0]?.slug || '';

  const replaceSelectedState = (saved) => {
    setLayoutState((current) => ({
      ...current,
      [selectedMode]: selectedMode === 'secondary'
        ? {
          ...normalizePage(saved, selectedMode),
          slug: selectedSecondarySlug,
          primaryTagSlug: selectedSecondarySlug,
          searchScope: 'tag',
          featuredTopicsEnabled: false,
          featuredMarketsEnabled: true,
          sectionsEnabled: false,
        }
        : normalizePage(saved, selectedMode),
    }));
  };

  const updateState = (updates) => {
    setLayoutState((current) => ({
      ...current,
      [selectedMode]: {
        ...current[selectedMode],
        ...updates,
      },
    }));
  };

  const updateSection = (index, updates) => {
    updateState({
      sections: state.sections.map((section, currentIndex) => (
        currentIndex === index ? { ...section, ...updates } : section
      )),
    });
  };

  const updatePin = (index, updates) => {
    updateState({
      pins: state.pins.map((pin, currentIndex) => (
        currentIndex === index ? { ...pin, ...updates } : pin
      )),
    });
  };

  const selectedBlocks = useMemo(() => [
    ...mode.fixedBlocks,
    ...(state.featuredTopicsEnabled ? ['Topic Pins'] : []),
    ...(state.featuredMarketsEnabled ? ['Market Pins'] : []),
    ...(state.sectionsEnabled ? ['CMS sections'] : []),
  ], [mode, state]);

  const saveLayout = async () => {
    setSaving(true);
    setMessage('');
    setError('');
    try {
      const saved = await authenticatedApiRequest(`/v0/admin/content/market-discovery/${selectedPageSlug}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...state,
          pageType: mode.pageType,
          primaryTagSlug: selectedMode === 'secondary' ? selectedSecondarySlug : state.primaryTagSlug,
          searchScope: selectedMode === 'secondary' ? 'tag' : state.searchScope,
          featuredTopicsEnabled: selectedMode === 'secondary' ? false : state.featuredTopicsEnabled,
          featuredMarketsEnabled: selectedMode === 'secondary' ? true : state.featuredMarketsEnabled,
          sectionsEnabled: selectedMode === 'secondary' ? false : state.sectionsEnabled,
        }),
        fallbackMessage: 'Failed to save market discovery layout.',
      });
      replaceSelectedState(saved);
      setMessage(`${selectedPageLabel} saved.`);
    } catch (err) {
      setError(err.message || 'Unable to save market discovery layout.');
    } finally {
      setSaving(false);
    }
  };

  const saveSections = async () => {
    setSavingSections(true);
    setMessage('');
    setError('');
    try {
      const saved = await authenticatedApiRequest(`/v0/admin/content/market-discovery/${selectedPageSlug}/sections`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sections: state.sections }),
        fallbackMessage: 'Failed to save discovery sections.',
      });
      replaceSelectedState(saved);
      setMessage(`${selectedPageLabel} sections saved.`);
    } catch (err) {
      setError(err.message || 'Unable to save discovery sections.');
    } finally {
      setSavingSections(false);
    }
  };

  const savePins = async () => {
    setSavingPins(true);
    setMessage('');
    setError('');
    try {
      const pins = selectedMode === 'secondary'
        ? state.pins.filter((pin) => pin.pinType === 'market')
        : state.pins;
      const saved = await authenticatedApiRequest(`/v0/admin/content/market-discovery/${selectedPageSlug}/pins`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pins }),
        fallbackMessage: 'Failed to save discovery pins.',
      });
      replaceSelectedState(saved);
      setMessage(`${selectedPageLabel} pins saved.`);
    } catch (err) {
      setError(err.message || 'Unable to save discovery pins.');
    } finally {
      setSavingPins(false);
    }
  };

  const addSection = () => {
    updateState({
      sections: [
        ...state.sections,
        {
          slug: '',
          title: 'New Section',
          description: '',
          tagFilterSlug: '',
          sortOrder: nextSortOrder(state.sections),
          isActive: true,
        },
      ],
    });
  };

  const addPin = (pinType) => {
    updateState({
      pins: [
        ...state.pins,
        {
          pinType,
          marketId: pinType === 'market' ? '' : 0,
          targetPageSlug: pinType === 'discovery_page' ? firstActiveTagSlug : '',
          label: pinType === 'market' ? '' : 'Featured Topic',
          sortOrder: nextSortOrder(state.pins),
        },
      ],
    });
  };

  const removePin = (index) => {
    updateState({ pins: state.pins.filter((_, currentIndex) => currentIndex !== index) });
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
            The TOP record controls the public /markets page title, description, fallback list size, section cards, and featured pins.
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
                  <h2 className="text-2xl font-bold text-white">{selectedPageLabel}</h2>
                  <p className="mt-2 max-w-2xl text-sm text-gray-300">{mode.purpose}</p>
                  <p className="mt-1 text-xs font-semibold text-sky-200">{selectedRoute}</p>
                </div>
                <button
                  type="button"
                  disabled={saving || loadingSecondary || !canEditSelectedPage}
                  onClick={saveLayout}
                  className="rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-pink/80 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Layout'}
                </button>
              </div>

              {selectedMode === 'secondary' && (
                <div className="mt-5 rounded-lg border border-sky-500/30 bg-sky-950/20 p-4">
                  <Field label="Secondary Topic Page">
                    <select
                      value={selectedSecondarySlug}
                      onChange={(event) => setSelectedSecondarySlug(event.target.value)}
                      className={textInputClass}
                    >
                      {activeTagOptions.length === 0 && <option value="">Create an active tag first</option>}
                      {activeTagOptions.map((tag) => (
                        <option key={tag.slug} value={tag.slug}>{tagOptionLabel(tag)}</option>
                      ))}
                    </select>
                  </Field>
                  <p className="mt-3 text-xs text-sky-100/80">
                    Secondary pages are stored per active tag. The topic nav is inherited from the TOP /markets page; market pins here are independent and searched within this topic.
                  </p>
                  {loadingSecondary && <p className="mt-2 text-xs text-gray-300">Loading selected topic page...</p>}
                </div>
              )}

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <Field label="Page Title">
                  <input
                    value={state.title}
                    maxLength={160}
                    onChange={(event) => updateState({ title: event.target.value })}
                    className={textInputClass}
                  />
                </Field>
                <Field label="Search Scope">
                  <select
                    value={state.searchScope}
                    onChange={(event) => updateState({ searchScope: event.target.value })}
                    disabled={selectedMode === 'secondary'}
                    className={textInputClass}
                  >
                    <option value="all">All public markets</option>
                    <option value="tag">Current topic/tag by default</option>
                  </select>
                </Field>
                <Field label="Page Description" className="md:col-span-2">
                  <textarea
                    value={state.description}
                    maxLength={500}
                    rows={3}
                    onChange={(event) => updateState({ description: event.target.value })}
                    className={textInputClass}
                  />
                </Field>
                <Field label="Empty CMS Recommendation Limit">
                  <input
                    type="number"
                    min="1"
                    max="50"
                    value={state.defaultRecommendationLimit}
                    onChange={(event) => updateState({ defaultRecommendationLimit: Number(event.target.value || 1) })}
                    className={textInputClass}
                  />
                </Field>
                <Field label="Curated CMS Recommendation Limit">
                  <input
                    type="number"
                    min="1"
                    max="50"
                    value={state.curatedRecommendationLimit}
                    onChange={(event) => updateState({ curatedRecommendationLimit: Number(event.target.value || 1) })}
                    className={textInputClass}
                  />
                </Field>
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                {selectedMode === 'top' && (
                  <ToggleCard
                    title="Topic Pins"
                    description="Create the persistent topic navigation bar. TOP Topic Pins appear on /markets and every secondary topic page."
                    checked={state.featuredTopicsEnabled}
                    onChange={(checked) => updateState({ featuredTopicsEnabled: checked })}
                  />
                )}
                <ToggleCard
                  title="Market Pins"
                  description={selectedMode === 'secondary' ? 'Pin active markets for this topic page only.' : 'Pin active markets as large chart cards before long fallback lists.'}
                  checked={selectedMode === 'secondary' ? true : state.featuredMarketsEnabled}
                  onChange={(checked) => updateState({ featuredMarketsEnabled: checked })}
                />
                {selectedMode === 'top' && (
                  <ToggleCard
                    title="CMS sections"
                    description="Allow named sections beyond the implicit All section."
                    checked={state.sectionsEnabled}
                    onChange={(checked) => updateState({ sectionsEnabled: checked })}
                  />
                )}
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

            {selectedMode === 'top' && (
              <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-xl font-bold text-white">Sections</h3>
                  <p className="mt-1 text-sm text-gray-400">Named section cards can optionally point at an active tag. Empty pages still use the implicit All section.</p>
                </div>
                <div className="flex gap-2">
                  <button type="button" onClick={addSection} className="rounded-md border border-sky-500/50 px-3 py-2 text-sm font-semibold text-sky-100 hover:bg-sky-950/50">Add Section</button>
                  <button type="button" disabled={savingSections} onClick={saveSections} className="rounded-md bg-primary-pink px-3 py-2 text-sm font-semibold text-white hover:bg-primary-pink/80 disabled:opacity-50">
                    {savingSections ? 'Saving...' : 'Save Sections'}
                  </button>
                </div>
              </div>
              <div className="mt-4 space-y-3">
                {state.sections.length === 0 && <p className="rounded-lg border border-dashed border-gray-700 p-4 text-sm text-gray-400">No explicit sections yet.</p>}
                {state.sections.map((section, index) => (
                  <div key={`${section.slug}-${index}`} className="grid gap-3 rounded-lg border border-gray-700 bg-gray-950 p-4 md:grid-cols-[1fr_1fr_110px_auto]">
                    <Field label="Title">
                      <input value={section.title || ''} onChange={(event) => updateSection(index, { title: event.target.value })} className={textInputClass} />
                    </Field>
                    <Field label="Slug">
                      <input value={section.slug || ''} onChange={(event) => updateSection(index, { slug: event.target.value })} placeholder="auto-from-title" className={textInputClass} />
                    </Field>
                    <Field label="Order">
                      <input type="number" value={section.sortOrder || index + 1} onChange={(event) => updateSection(index, { sortOrder: Number(event.target.value || 0) })} className={textInputClass} />
                    </Field>
                    <label className="flex items-end gap-2 pb-2 text-sm text-gray-300">
                      <input type="checkbox" checked={section.isActive !== false} onChange={(event) => updateSection(index, { isActive: event.target.checked })} className="h-4 w-4 accent-primary-pink" />
                      Active
                    </label>
                    <Field label="Tag Filter" className="md:col-span-2">
                      <select
                        value={section.tagFilterSlug || ''}
                        onChange={(event) => updateSection(index, { tagFilterSlug: event.target.value })}
                        className={textInputClass}
                      >
                        <option value="">No tag filter</option>
                        {tagOptionsWithCurrent(activeTagOptions, section.tagFilterSlug).map((tag) => (
                          <option key={tag.slug} value={tag.slug}>
                            {tagOptionLabel(tag)}{tag.isActive === false ? ' - inactive or missing' : ''}
                          </option>
                        ))}
                      </select>
                    </Field>
                    <Field label="Description" className="md:col-span-2">
                      <input value={section.description || ''} onChange={(event) => updateSection(index, { description: event.target.value })} className={textInputClass} />
                    </Field>
                    <div className="md:col-span-4">
                      <button type="button" onClick={() => updateState({ sections: state.sections.filter((_, currentIndex) => currentIndex !== index) })} className="text-sm font-semibold text-red-300 hover:text-red-200">
                        Remove section
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            )}

            {selectedMode === 'top' && (
              <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-xl font-bold text-white">Topic Pins</h3>
                  <p className="mt-1 text-sm text-gray-400">
                    Topic pins build the persistent market topic nav bar. Each entry points to an active tag and appears on /markets and every /markets/topic/:slug page.
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => addPin('discovery_page')}
                    disabled={activeTagOptions.length === 0}
                    className="rounded-md border border-sky-500/50 px-3 py-2 text-sm font-semibold text-sky-100 hover:bg-sky-950/50 disabled:cursor-not-allowed disabled:opacity-50"
                    title={activeTagOptions.length === 0 ? 'Create an active tag before adding topic pins.' : 'Add a pinned topic from active tags.'}
                  >
                    Add Topic Pin
                  </button>
                  <button type="button" disabled={savingPins} onClick={savePins} className="rounded-md bg-primary-pink px-3 py-2 text-sm font-semibold text-white hover:bg-primary-pink/80 disabled:opacity-50">
                    {savingPins ? 'Saving...' : 'Save Pins'}
                  </button>
                </div>
              </div>
              <div className="mt-3 flex flex-wrap gap-2 text-xs text-gray-400">
                <Pill>{topicPins.length} topic pins</Pill>
              </div>
              <div className="mt-4 space-y-3">
                {topicPins.length === 0 && <p className="rounded-lg border border-dashed border-gray-700 p-4 text-sm text-gray-400">No topic pins yet.</p>}
                {topicPins.map(({ pin, index }) => (
                  <div key={`topic-${pin.id || pin.targetPageSlug || index}`} className="grid gap-3 rounded-lg border border-gray-700 bg-gray-950 p-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,90px)]">
                    <Field label="Nav Label">
                      <input value={pin.label || ''} onChange={(event) => updatePin(index, { label: event.target.value })} className={textInputClass} />
                    </Field>
                    <Field label="Target Topic Tag">
                      <select
                        value={pin.targetPageSlug || ''}
                        onChange={(event) => updatePin(index, { targetPageSlug: event.target.value })}
                        className={textInputClass}
                      >
                        {activeTagOptions.length === 0 && !pin.targetPageSlug && (
                          <option value="">Create an active tag first</option>
                        )}
                        {tagOptionsWithCurrent(activeTagOptions, pin.targetPageSlug).map((tag) => (
                          <option key={tag.slug} value={tag.slug}>
                            {tagOptionLabel(tag)}{tag.isActive === false ? ' - inactive or missing' : ''}
                          </option>
                        ))}
                      </select>
                    </Field>
                    <Field label="Order">
                      <input type="number" value={pin.sortOrder || index + 1} onChange={(event) => updatePin(index, { sortOrder: Number(event.target.value || 0) })} className={textInputClass} />
                    </Field>
                    <div className="md:col-span-3">
                      <button type="button" onClick={() => removePin(index)} className="text-sm font-semibold text-red-300 hover:text-red-200">
                        Remove topic pin
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            )}

            <div className="rounded-xl border border-gray-700 bg-gray-900/80 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-xl font-bold text-white">Market Pins</h3>
                  <p className="mt-1 text-sm text-gray-400">
                    {selectedMode === 'secondary'
                      ? `Market pins render selected active ${selectedSecondarySlug || 'topic'} markets on this topic page only.`
                      : 'Market pins render selected active markets as featured chart cards. Search active markets, select one, then save.'}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button type="button" onClick={() => addPin('market')} disabled={!canEditSelectedPage || loadingSecondary} className="rounded-md border border-emerald-500/50 px-3 py-2 text-sm font-semibold text-emerald-100 hover:bg-emerald-950/50 disabled:cursor-not-allowed disabled:opacity-50">Add Market Pin</button>
                  <button type="button" disabled={savingPins || !canEditSelectedPage || loadingSecondary} onClick={savePins} className="rounded-md bg-primary-pink px-3 py-2 text-sm font-semibold text-white hover:bg-primary-pink/80 disabled:opacity-50">
                    {savingPins ? 'Saving...' : 'Save Pins'}
                  </button>
                </div>
              </div>
              <div className="mt-3 flex flex-wrap gap-2 text-xs text-gray-400">
                <Pill>{marketPins.length} market pins</Pill>
              </div>
              <div className="mt-4 space-y-3">
                {marketPins.length === 0 && <p className="rounded-lg border border-dashed border-gray-700 p-4 text-sm text-gray-400">No market pins yet.</p>}
                {marketPins.map(({ pin, index }) => (
                  <div key={`market-${pin.id || pin.marketId || index}`} className="grid gap-3 rounded-lg border border-gray-700 bg-gray-950 p-4 md:grid-cols-[minmax(0,1fr)_minmax(0,90px)]">
                    <MarketPinSearch pin={pin} tagSlug={selectedMode === 'secondary' ? selectedSecondarySlug : ''} onSelect={(marketId) => updatePin(index, { marketId: Number(marketId), label: '' })} />
                    <Field label="Order">
                      <input type="number" value={pin.sortOrder || index + 1} onChange={(event) => updatePin(index, { sortOrder: Number(event.target.value || 0) })} className={textInputClass} />
                    </Field>
                    <div className="md:col-span-2">
                      <button type="button" onClick={() => removePin(index)} className="text-sm font-semibold text-red-300 hover:text-red-200">
                        Remove market pin
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <LayoutPreview mode={{ ...mode, route: selectedRoute }} state={state} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default MarketDiscoveryLayoutEditor;
