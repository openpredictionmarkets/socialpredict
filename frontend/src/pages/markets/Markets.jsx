import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import SiteTabs from '../../components/tabs/SiteTabs';
import MarketsByStatusTable from '../../components/tables/MarketsByStatusTable';
import GlobalSearchBar from '../../components/search/GlobalSearchBar';
import SearchResultsTable from '../../components/tables/SearchResultsTable';
import { marketTagChipClassFor } from '../../components/markets/MarketTagChips';
import { TAB_TO_STATUS } from '../../utils/statusMap';
import { apiRequest } from '../../api/httpClient';
import { listMarketTags } from '../../api/marketTagsApi';

const defaultDiscoveryLayout = {
    title: 'Markets',
    description: '',
    featuredTopicsEnabled: false,
    featuredMarketsEnabled: false,
    sectionsEnabled: false,
    recommendationLimit: 20,
    primaryTagSlug: '',
    searchScope: 'all',
    sections: [],
    pins: [],
};

const hasCuratedBlocks = (layout) => (
    layout?.featuredTopicsEnabled || layout?.featuredMarketsEnabled || layout?.sectionsEnabled
);

const sortBySortOrder = (items = []) => [...items].sort((a, b) => Number(a.sortOrder || 0) - Number(b.sortOrder || 0));

const toNumber = (value, fallback = 0) => {
    if (typeof value === 'number') {
        return Number.isFinite(value) ? value : fallback;
    }
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
};

const slugTitle = (slug = '') => slug
    .split('-')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');

const normalizeDiscoveryLayout = (layout, fallback = {}) => ({
    ...defaultDiscoveryLayout,
    ...fallback,
    ...layout,
    sections: sortBySortOrder(layout?.sections || []),
    pins: sortBySortOrder(layout?.pins || []),
});

const DiscoveryBadge = ({ children, tag }) => {
    const tones = {
        sky: 'border-sky-500/40 bg-sky-950/40 text-sky-100',
    };
    const className = tag ? marketTagChipClassFor(tag) : tones.sky;

    return (
        <span className={`rounded-full border px-3 py-1 text-xs font-semibold ${className}`}>
            {children}
        </span>
    );
};

const TopicNav = ({ topicPins = [] }) => {
    return (
        <nav className="mb-5 overflow-x-auto border-b border-gray-800 pb-3" aria-label="Market topics">
            <div className="flex min-w-max items-center gap-2">
                <Link
                    to="/markets"
                    className="rounded-full border border-sky-500/50 bg-sky-950/40 px-4 py-2 text-sm font-semibold text-sky-100 transition hover:border-sky-300/70 hover:bg-sky-900/50 hover:text-white"
                >
                    Markets
                </Link>
                {topicPins.map((pin) => (
                    <Link
                        key={`topic-${pin.id || pin.targetPageSlug || pin.sortOrder}`}
                        to={pin.targetPageSlug ? `/markets/topic/${pin.targetPageSlug}` : '#'}
                        className={`rounded-full border px-4 py-2 text-sm font-semibold transition hover:brightness-125 ${marketTagChipClassFor(pin.tag)}`}
                        title={pin.tag?.description || pin.label || pin.targetPageSlug}
                    >
                        {pin.label || slugTitle(pin.targetPageSlug) || 'Topic'}
                    </Link>
                ))}
            </div>
        </nav>
    );
};

const normalizeProbabilityChanges = (raw = []) => (
    Array.isArray(raw)
        ? raw
            .map((change) => ({
                probability: toNumber(change?.probability ?? change?.Probability),
                timestamp: change?.timestamp ?? change?.Timestamp ?? change?.createdAt ?? change?.CreatedAt,
            }))
            .filter((change) => change.timestamp)
        : []
);

const currentProbabilityFromDetails = (details) => {
    const changes = normalizeProbabilityChanges(details?.probabilityChanges ?? details?.ProbabilityChanges);
    if (changes.length > 0) {
        return toNumber(changes[changes.length - 1].probability);
    }
    return toNumber(details?.lastProbability ?? details?.LastProbability ?? details?.market?.initialProbability);
};

const chartSamplesFromDetails = (details) => {
    const changes = normalizeProbabilityChanges(details?.probabilityChanges ?? details?.ProbabilityChanges);
    const currentProbability = currentProbabilityFromDetails(details);
    const market = details?.market || {};
    const fallbackTimestamp = market.createdAt || market.CreatedAt || new Date().toISOString();
    const samples = changes.length > 0
        ? changes.map((change) => ({
            probability: toNumber(change.probability, currentProbability),
            timestamp: new Date(change.timestamp).getTime(),
        }))
        : [{
            probability: currentProbability,
            timestamp: new Date(fallbackTimestamp).getTime(),
        }];

    const validSamples = samples
        .filter((sample) => Number.isFinite(sample.timestamp))
        .sort((left, right) => left.timestamp - right.timestamp);

    if (validSamples.length === 0) {
        return [{
            probability: currentProbability,
            timestamp: Date.now(),
        }];
    }

    return validSamples;
};

const sampleToPoint = (sample, minTime, maxTime, width, height, inset) => {
    const usableWidth = width - inset * 2;
    const usableHeight = height - inset * 2;
    const timeSpan = Math.max(maxTime - minTime, 1);
    const probability = Math.max(0, Math.min(1, sample.probability));

    return {
        x: inset + ((sample.timestamp - minTime) / timeSpan) * usableWidth,
        y: inset + (1 - probability) * usableHeight,
    };
};

const buildStepPath = (points) => {
    if (points.length === 0) {
        return '';
    }
    const [first, ...rest] = points;
    const commands = [`M ${first.x.toFixed(2)} ${first.y.toFixed(2)}`];

    rest.forEach((point) => {
        commands.push(`H ${point.x.toFixed(2)}`);
        commands.push(`V ${point.y.toFixed(2)}`);
    });

    if (points.length === 1) {
        commands.push(`H ${first.x.toFixed(2)}`);
    }

    return commands.join(' ');
};

const PinnedMarketSparkline = ({ details }) => {
    const width = 960;
    const height = 260;
    const inset = 18;
    const samples = chartSamplesFromDetails(details);
    const currentTime = Date.now();
    const minTime = Math.min(...samples.map((sample) => sample.timestamp));
    const maxSampleTime = Math.max(...samples.map((sample) => sample.timestamp));
    const maxTime = Math.max(maxSampleTime, currentTime, minTime + 1);
    const points = samples.map((sample) => sampleToPoint(sample, minTime, maxTime, width, height, inset));
    const firstX = inset;
    const lastX = width - inset;
    const floorY = height - inset;
    const lastPoint = points[points.length - 1] || { x: lastX, y: height / 2 };
    const stepPath = `${buildStepPath(points)} H ${lastX.toFixed(2)}`;
    const areaPath = `${stepPath} L ${lastX} ${floorY} L ${firstX} ${floorY} Z`;
    const currentProbability = currentProbabilityFromDetails(details);

    return (
        <svg
            viewBox={`0 0 ${width} ${height}`}
            className="h-full w-full"
            preserveAspectRatio="none"
            role="img"
            aria-label={`Pinned market probability ${(currentProbability * 100).toFixed(0)} percent`}
        >
            <defs>
                <linearGradient id="pinned-market-area" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#22d3ee" stopOpacity="0.58" />
                    <stop offset="100%" stopColor="#0e7490" stopOpacity="0.16" />
                </linearGradient>
            </defs>
            <rect width={width} height={height} rx="18" fill="#101827" />
            <line x1="0" x2={width} y1={height / 2} y2={height / 2} stroke="#334155" strokeWidth="2" strokeDasharray="10 12" opacity="0.45" />
            <path d={areaPath} fill="url(#pinned-market-area)" />
            <path d={stepPath} fill="none" stroke="#22d3ee" strokeWidth="8" strokeLinecap="round" strokeLinejoin="round" />
            <circle cx={lastX} cy={lastPoint.y || height / 2} r="10" fill="#22d3ee" />
            <text x={width - 28} y="58" textAnchor="end" fill="#e2e8f0" fontSize="24" fontWeight="700">
                {(currentProbability * 100).toFixed(0)}%
            </text>
        </svg>
    );
};

const FeaturedMarketPins = ({ marketPins = [] }) => {
    const [pinnedMarkets, setPinnedMarkets] = useState([]);
    const pinKey = marketPins
        .map((pin) => `${pin.id || ''}:${pin.marketId || ''}:${pin.sortOrder || ''}`)
        .join('|');

    useEffect(() => {
        let ignore = false;
        const pinsWithIds = marketPins.filter((pin) => Number(pin.marketId) > 0);

        if (pinsWithIds.length === 0) {
            setPinnedMarkets([]);
            return () => {
                ignore = true;
            };
        }

        Promise.all(
            pinsWithIds.map((pin) => (
                apiRequest(`/v0/markets/${pin.marketId}`, {
                    fallbackMessage: `Failed to load pinned market ${pin.marketId}`,
                })
                    .then((details) => ({ pin, details }))
                    .catch(() => null)
            )),
        ).then((items) => {
            if (!ignore) {
                setPinnedMarkets(items.filter((item) => item?.details?.market));
            }
        });

        return () => {
            ignore = true;
        };
    }, [pinKey]);

    if (!pinnedMarkets.length) return null;

    return (
        <section className="grid gap-3 lg:grid-cols-2" aria-label="Pinned markets">
            {pinnedMarkets.map(({ pin, details }) => {
                const market = details.market || {};
                const marketId = market.id || pin.marketId;
                return (
                    <Link
                        key={`market-${pin.id || marketId || pin.sortOrder}`}
                        to={marketId ? `/markets/${marketId}` : '#'}
                        className="block w-full rounded-xl border border-gray-700 bg-gray-950/80 p-3 no-underline shadow-lg transition hover:border-sky-400/60 hover:bg-gray-900"
                    >
                        <h3 className="mb-2 line-clamp-2 min-h-[2.5rem] text-sm font-semibold leading-5 text-white">
                            {market.questionTitle || pin.label || `Market #${marketId}`}
                        </h3>
                        <div className="h-36 overflow-hidden rounded-lg border border-gray-800 bg-gray-900/60 sm:h-52">
                            <PinnedMarketSparkline details={details} />
                        </div>
                    </Link>
                );
            })}
        </section>
    );
};

const DiscoveryPanel = ({ layout, persistentTopicPins = [], useLayoutTopicPins = true }) => {
    const pins = sortBySortOrder(layout.pins || []);
    const sections = sortBySortOrder(layout.sections || []).filter((section) => section.isActive !== false);
    const topicPins = persistentTopicPins.length > 0
        ? persistentTopicPins
        : (useLayoutTopicPins ? pins.filter((pin) => pin.pinType === 'discovery_page') : []);
    const marketPins = pins.filter((pin) => pin.pinType === 'market');
    const hasTopicNav = topicPins.length > 0;

    if (!hasCuratedBlocks(layout) && !hasTopicNav) return null;

    return (
        <div className="mb-6 space-y-5 text-gray-200">
            {hasTopicNav && <TopicNav topicPins={topicPins} />}

            {layout.featuredMarketsEnabled && <FeaturedMarketPins marketPins={marketPins} />}

            {layout.sectionsEnabled && sections.length > 0 && (
                <section>
                    <div className="flex flex-wrap gap-2" aria-label="Market discovery sections">
                        {sections.map((section) => (
                            <Link
                                key={`section-${section.id || section.slug || section.sortOrder}`}
                                to={section.tagFilterSlug ? `/markets/topic/${section.tagFilterSlug}` : '#'}
                                className="rounded-full border border-amber-500/30 bg-amber-950/20 px-3 py-1.5 text-xs font-semibold text-amber-100 transition hover:border-amber-300/60 hover:bg-amber-900/30"
                                title={section.description || section.title}
                            >
                                {section.title}
                            </Link>
                        ))}
                    </div>
                </section>
            )}
        </div>
    );
};

function Markets() {
    const { slug: topicSlugParam } = useParams();
    const topicSlug = topicSlugParam || '';
    const isTopicPage = !!topicSlug;
    const [searchResults, setSearchResults] = useState(null);
    const [isSearching, setIsSearching] = useState(false);
    const [activeTab, setActiveTab] = useState('Active');
    const [discoveryLayout, setDiscoveryLayout] = useState(defaultDiscoveryLayout);
    const [persistentTopicPins, setPersistentTopicPins] = useState([]);
    const [marketTagsBySlug, setMarketTagsBySlug] = useState({});

    useEffect(() => {
        let ignore = false;

        listMarketTags()
            .then((result) => {
                if (ignore) return;
                const bySlug = {};
                (result.tags || []).forEach((tag) => {
                    if (tag.slug) {
                        bySlug[tag.slug] = tag;
                    }
                });
                setMarketTagsBySlug(bySlug);
            })
            .catch(() => {
                if (!ignore) {
                    setMarketTagsBySlug({});
                }
            });

        return () => {
            ignore = true;
        };
    }, []);

    useEffect(() => {
        let ignore = false;
        const discoverySlug = isTopicPage ? topicSlug : 'markets';
        const fallback = isTopicPage
            ? {
                title: slugTitle(topicSlug) || 'Topic Markets',
                description: 'Browse and search markets in this topic.',
                primaryTagSlug: topicSlug,
                searchScope: 'tag',
                featuredMarketsEnabled: true,
                sectionsEnabled: true,
                recommendationLimit: 5,
            }
            : {};

        setSearchResults(null);
        setIsSearching(false);

        const loadDiscovery = async () => {
            try {
                const [layout, topLayout] = await Promise.all([
                    apiRequest(`/v0/content/market-discovery/${discoverySlug}`, {
                        fallbackMessage: 'Failed to load market discovery layout',
                    }),
                    isTopicPage
                        ? apiRequest('/v0/content/market-discovery/markets', {
                            fallbackMessage: 'Failed to load market topic navigation',
                        })
                        : Promise.resolve(null),
                ]);
                if (!ignore) {
                    const normalizedLayout = normalizeDiscoveryLayout(layout, fallback);
                    const topNavLayout = isTopicPage
                        ? normalizeDiscoveryLayout(topLayout || {}, {})
                        : normalizedLayout;
                    const topTopicPins = topNavLayout.featuredTopicsEnabled
                        ? sortBySortOrder(topNavLayout.pins || [])
                            .filter((pin) => pin.pinType === 'discovery_page')
                            .map((pin) => ({ ...pin, tag: marketTagsBySlug[pin.targetPageSlug] }))
                        : [];

                    setDiscoveryLayout(normalizedLayout);
                    setPersistentTopicPins(topTopicPins);
                }
            } catch {
                if (!ignore) {
                    setDiscoveryLayout(normalizeDiscoveryLayout({}, fallback));
                    setPersistentTopicPins([]);
                }
            }
        };

        loadDiscovery();

        return () => {
            ignore = true;
        };
    }, [isTopicPage, topicSlug, marketTagsBySlug]);

    const tagScope = isTopicPage
        ? topicSlug
        : (discoveryLayout.searchScope === 'tag' ? discoveryLayout.primaryTagSlug : '');

    const handleSearchResults = (results) => {
        setSearchResults(results);
    };

    const handleTabChange = (tabLabel) => {
        setActiveTab(tabLabel);
        // Don't clear search when switching tabs; GlobalSearchBar re-executes with currentStatus.
    };

    const tabsData = [
        { 
            label: 'Active', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="active" limit={discoveryLayout.recommendationLimit} tagSlug={tagScope} />,
            onSelect: () => handleTabChange('Active')
        },
        { 
            label: 'Closed', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="closed" limit={discoveryLayout.recommendationLimit} tagSlug={tagScope} />,
            onSelect: () => handleTabChange('Closed')
        },
        { 
            label: 'Resolved', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="resolved" limit={discoveryLayout.recommendationLimit} tagSlug={tagScope} />,
            onSelect: () => handleTabChange('Resolved')
        },
        { 
            label: 'All', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="all" limit={discoveryLayout.recommendationLimit} tagSlug={tagScope} />,
            onSelect: () => handleTabChange('All')
        },
    ];

    return (
        <div className='App'>
            <div className='Center-content'>
                <div className='Center-content-header'>
                    {isTopicPage ? (
                        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(260px,420px)] md:items-start">
                            <div>
                                <h1 className='text-2xl font-semibold text-gray-300 mb-2'>{discoveryLayout.title || 'Markets'}</h1>
                                {discoveryLayout.description && (
                                    <p className="mb-3 max-w-3xl text-sm text-gray-400">{discoveryLayout.description}</p>
                                )}
                            </div>
                            <div className="space-y-2 md:text-right">
                                {tagScope && (
                                    <div className="md:flex md:justify-end">
                                        <DiscoveryBadge tag={marketTagsBySlug[tagScope]}>{`Filtered by tag: ${tagScope}`}</DiscoveryBadge>
                                    </div>
                                )}
                            </div>
                        </div>
                    ) : (
                        <>
                            <h1 className='text-2xl font-semibold text-gray-300 mb-2'>{discoveryLayout.title || 'Markets'}</h1>
                            {discoveryLayout.description && (
                                <p className="mb-3 max-w-3xl text-sm text-gray-400">{discoveryLayout.description}</p>
                            )}
                            {tagScope && (
                                <div className="mb-6">
                                    <DiscoveryBadge tag={marketTagsBySlug[tagScope]}>{`Filtered by tag: ${tagScope}`}</DiscoveryBadge>
                                </div>
                            )}
                        </>
                    )}
                </div>
                <div className='Center-content-table'>
                    <GlobalSearchBar 
                        onSearchResults={handleSearchResults}
                        currentStatus={TAB_TO_STATUS[activeTab]}
                        isSearching={isSearching}
                        setIsSearching={setIsSearching}
                        tagSlug={tagScope}
                    />
                    {!isSearching && (
                        <DiscoveryPanel
                            layout={discoveryLayout}
                            persistentTopicPins={persistentTopicPins}
                            useLayoutTopicPins={!isTopicPage}
                        />
                    )}
                    
                    <SiteTabs 
                        tabs={tabsData} 
                        onTabChange={handleTabChange}
                        activeTab={activeTab}
                    />
                </div>
            </div>
        </div>
    );
}

export default Markets;
