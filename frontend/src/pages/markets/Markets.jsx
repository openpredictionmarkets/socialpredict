import React, { useState, useEffect } from 'react';
import SiteTabs from '../../components/tabs/SiteTabs';
import MarketsByStatusTable from '../../components/tables/MarketsByStatusTable';
import GlobalSearchBar from '../../components/search/GlobalSearchBar';
import SearchResultsTable from '../../components/tables/SearchResultsTable';
import { TAB_TO_STATUS } from '../../utils/statusMap';
import { apiRequest } from '../../api/httpClient';

const defaultDiscoveryLayout = {
    title: 'Markets',
    description: '',
    featuredTopicsEnabled: false,
    featuredMarketsEnabled: false,
    sectionsEnabled: false,
    recommendationLimit: 20,
};

const hasCuratedBlocks = (layout) => (
    layout?.featuredTopicsEnabled || layout?.featuredMarketsEnabled || layout?.sectionsEnabled
);

const DiscoveryPlaceholderPanel = ({ layout }) => {
    if (!hasCuratedBlocks(layout)) return null;

    return (
        <div className="mb-6 grid gap-4 rounded-xl border border-gray-700 bg-gray-900/70 p-4 text-gray-200">
            <div>
                <p className="font-mono text-xs uppercase tracking-[0.18em] text-sky-300">
                    CMS Market Discovery
                </p>
                <h2 className="mt-2 text-lg font-semibold text-white">Curated discovery blocks enabled</h2>
                <p className="mt-1 text-sm text-gray-400">
                    Pins and sections are persisted next; recommendations below are compacted to {layout.recommendationLimit} markets while curation is enabled.
                </p>
            </div>
            <div className="flex flex-wrap gap-2">
                {layout.featuredTopicsEnabled && <span className="rounded-full border border-sky-500/40 bg-sky-950/40 px-3 py-1 text-xs font-semibold text-sky-100">Featured topics</span>}
                {layout.featuredMarketsEnabled && <span className="rounded-full border border-emerald-500/40 bg-emerald-950/40 px-3 py-1 text-xs font-semibold text-emerald-100">Featured markets</span>}
                {layout.sectionsEnabled && <span className="rounded-full border border-violet-500/40 bg-violet-950/40 px-3 py-1 text-xs font-semibold text-violet-100">Sections</span>}
            </div>
        </div>
    );
};

function Markets() {
    const [searchResults, setSearchResults] = useState(null);
    const [isSearching, setIsSearching] = useState(false);
    const [activeTab, setActiveTab] = useState('Active');
    const [discoveryLayout, setDiscoveryLayout] = useState(defaultDiscoveryLayout);

    useEffect(() => {
        let ignore = false;

        apiRequest('/v0/content/market-discovery/markets', {
            fallbackMessage: 'Failed to load market discovery layout',
        })
            .then((layout) => {
                if (!ignore) {
                    setDiscoveryLayout({ ...defaultDiscoveryLayout, ...layout });
                }
            })
            .catch(() => {
                if (!ignore) {
                    setDiscoveryLayout(defaultDiscoveryLayout);
                }
            });

        return () => {
            ignore = true;
        };
    }, []);

    const handleSearchResults = (results) => {
        setSearchResults(results);
    };

    const handleTabChange = (tabLabel) => {
        setActiveTab(tabLabel);
        // Don't clear search when switching tabs - let GlobalSearchBar re-execute with new status
        // The search will automatically re-execute due to currentStatus change in GlobalSearchBar
    };

    const tabsData = [
        { 
            label: 'Active', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="active" limit={discoveryLayout.recommendationLimit} />,
            onSelect: () => handleTabChange('Active')
        },
        { 
            label: 'Closed', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="closed" limit={discoveryLayout.recommendationLimit} />,
            onSelect: () => handleTabChange('Closed')
        },
        { 
            label: 'Resolved', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="resolved" limit={discoveryLayout.recommendationLimit} />,
            onSelect: () => handleTabChange('Resolved')
        },
        { 
            label: 'All', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="all" limit={discoveryLayout.recommendationLimit} />,
            onSelect: () => handleTabChange('All')
        },
    ];

    return (
        <div className='App'>
            <div className='Center-content'>
                <div className='Center-content-header'>
                    <h1 className='text-2xl font-semibold text-gray-300 mb-2'>{discoveryLayout.title || 'Markets'}</h1>
                    {discoveryLayout.description && (
                        <p className="mb-6 max-w-3xl text-sm text-gray-400">{discoveryLayout.description}</p>
                    )}
                </div>
                <div className='Center-content-table'>
                    {/* Global Search Bar - Always visible at top */}
                    <GlobalSearchBar 
                        onSearchResults={handleSearchResults}
                        currentStatus={TAB_TO_STATUS[activeTab]}
                        isSearching={isSearching}
                        setIsSearching={setIsSearching}
                    />
                    {!isSearching && <DiscoveryPlaceholderPanel layout={discoveryLayout} />}
                    
                    {/* Tabs with Content */}
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
