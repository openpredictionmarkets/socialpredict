import React, { useState, useEffect } from 'react';
import SiteTabs from '../../components/tabs/SiteTabs';
import MarketsByStatusTable from '../../components/tables/MarketsByStatusTable';
import GlobalSearchBar from '../../components/search/GlobalSearchBar';
import SearchResultsTable from '../../components/tables/SearchResultsTable';
import { TAB_TO_STATUS } from '../../utils/statusMap';

function Markets() {
    const [searchResults, setSearchResults] = useState(null);
    const [isSearching, setIsSearching] = useState(false);
    const [activeTab, setActiveTab] = useState('Active');

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
                <MarketsByStatusTable status="active" />,
            onSelect: () => handleTabChange('Active')
        },
        { 
            label: 'Closed', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="closed" />,
            onSelect: () => handleTabChange('Closed')
        },
        { 
            label: 'Resolved', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="resolved" />,
            onSelect: () => handleTabChange('Resolved')
        },
        { 
            label: 'All', 
            content: isSearching ? 
                <SearchResultsTable searchResults={searchResults} /> : 
                <MarketsByStatusTable status="all" />,
            onSelect: () => handleTabChange('All')
        },
    ];

    return (
        <div className='App'>
            <div className='Center-content'>
                <div className='Center-content-header'>
                    <h1 className='text-2xl font-semibold text-gray-300 mb-6'>Markets</h1>
                </div>
                <div className='Center-content-table'>
                    {/* Global Search Bar - Always visible at top */}
                    <GlobalSearchBar 
                        onSearchResults={handleSearchResults}
                        currentStatus={TAB_TO_STATUS[activeTab]}
                        isSearching={isSearching}
                        setIsSearching={setIsSearching}
                    />
                    
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
