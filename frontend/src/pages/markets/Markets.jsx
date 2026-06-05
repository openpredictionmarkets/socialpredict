import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import SiteTabs from '../../components/tabs/SiteTabs';
import MarketsByStatusTable from '../../components/tables/MarketsByStatusTable';
import GlobalSearchBar from '../../components/search/GlobalSearchBar';
import SearchResultsTable from '../../components/tables/SearchResultsTable';
import MarketChart from '../../components/charts/MarketChart';
import { TAB_TO_STATUS } from '../../utils/statusMap';
import { apiRequest } from '../../api/httpClient';

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

const DiscoveryBadge = ({ children, tone = 'sky' }) => {
    const tones = {
        sky: 'border-sky-500/40 bg-sky-950/40 text-sky-100',
        emerald: 'border-emerald-500/40 bg-emerald-950/40 text-emerald-100',
        amber: 'border-amber-500/40 bg-amber-950/40 text-amber-100',
    };

    return (
        <span className={`rounded-full border px-3 py-1 text-xs font-semibold ${tones[tone] || tones.sky}`}>
            {children}
        </span>
    );
};

const TopicNav = ({ topicPins = [] }) => {
    if (!topicPins.length) return null;

    return (
        <nav className="mb-5 overflow-x-auto border-b border-gray-800 pb-3" aria-label="Market topics">
            <div className="flex min-w-max items-center gap-2">
                {topicPins.map((pin) => (
                    <Link
                        key={`topic-${pin.id || pin.targetPageSlug || pin.sortOrder}`}
                        to={pin.targetPageSlug ? `/markets/topic/${pin.targetPageSlug}` : '#'}
                        className="rounded-full border border-gray-700 bg-gray-900/80 px-4 py-2 text-sm font-semibold text-gray-200 transition hover:border-sky-400/70 hover:bg-sky-950/40 hover:text-white"
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
        <section className="grid gap-3" aria-label="Pinned markets">
            {pinnedMarkets.map(({ pin, details }) => {
                const market = details.market || {};
                const marketId = market.id || pin.marketId;
                const probabilityChanges = normalizeProbabilityChanges(details.probabilityChanges);
                const currentProbability = currentProbabilityFromDetails(details);
                return (
                    <Link
                        key={`market-${pin.id || marketId || pin.sortOrder}`}
                        to={marketId ? `/markets/${marketId}` : '#'}
                        className="block w-full rounded-xl border border-gray-700 bg-gray-950/80 p-3 no-underline shadow-lg transition hover:border-sky-400/60 hover:bg-gray-900"
                    >
                        <h3 className="mb-2 line-clamp-2 min-h-[2.5rem] text-sm font-semibold leading-5 text-white">
                            {market.questionTitle || pin.label || `Market #${marketId}`}
                        </h3>
                        <div className="h-36 overflow-hidden rounded-lg border border-gray-800 bg-gray-900/60 sm:h-44">
                            <MarketChart
                                data={probabilityChanges}
                                currentProbability={currentProbability}
                                title=""
                                className="h-full w-full"
                                closeDateTime={market.resolutionDateTime}
                                yesLabel={market.yesLabel || 'YES'}
                                noLabel={market.noLabel || 'NO'}
                                showHeader={false}
                                compact
                                height={176}
                            />
                        </div>
                    </Link>
                );
            })}
        </section>
    );
};

const DiscoveryPanel = ({ layout, isTopicPage = false }) => {
    if (!hasCuratedBlocks(layout)) return null;

    const pins = sortBySortOrder(layout.pins || []);
    const sections = sortBySortOrder(layout.sections || []).filter((section) => section.isActive !== false);
    const topicPins = pins.filter((pin) => pin.pinType === 'discovery_page');
    const marketPins = pins.filter((pin) => pin.pinType === 'market');

    return (
        <div className="mb-6 space-y-5 text-gray-200">
            {!isTopicPage && layout.featuredTopicsEnabled && <TopicNav topicPins={topicPins} />}

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

        apiRequest(`/v0/content/market-discovery/${discoverySlug}`, {
            fallbackMessage: 'Failed to load market discovery layout',
        })
            .then((layout) => {
                if (!ignore) {
                    setDiscoveryLayout(normalizeDiscoveryLayout(layout, fallback));
                }
            })
            .catch(() => {
                if (!ignore) {
                    setDiscoveryLayout(normalizeDiscoveryLayout({}, fallback));
                }
            });

        return () => {
            ignore = true;
        };
    }, [isTopicPage, topicSlug]);

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
                    {isTopicPage && (
                        <Link to="/markets" className="mb-3 inline-flex text-sm font-semibold text-sky-300 hover:text-sky-200">
                            Back to all markets
                        </Link>
                    )}
                    <h1 className='text-2xl font-semibold text-gray-300 mb-2'>{discoveryLayout.title || 'Markets'}</h1>
                    {discoveryLayout.description && (
                        <p className="mb-3 max-w-3xl text-sm text-gray-400">{discoveryLayout.description}</p>
                    )}
                    {tagScope && (
                        <div className="mb-6">
                            <DiscoveryBadge>{`Filtered by tag: ${tagScope}`}</DiscoveryBadge>
                        </div>
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
                    {!isSearching && <DiscoveryPanel layout={discoveryLayout} isTopicPage={isTopicPage} />}
                    
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
