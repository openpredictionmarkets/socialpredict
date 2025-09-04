import React, { useState, useEffect } from 'react';
import { searchMarkets } from '../../api/marketsApi';
import MarketTables from '../tables/MarketTables';

const MarketSearch = () => {
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [selectedStatus, setSelectedStatus] = useState('all');

    const handleSearch = async (query = searchQuery, status = selectedStatus) => {
        if (!query.trim()) {
            setSearchResults(null);
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const results = await searchMarkets(query.trim(), status);
            setSearchResults(results);
        } catch (err) {
            console.error('Search error:', err);
            setError('Failed to search markets. Please try again.');
            setSearchResults(null);
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e) => {
        const value = e.target.value;
        setSearchQuery(value);
        
        // Debounce search - only search after user stops typing for 500ms
        if (value.trim()) {
            const timeoutId = setTimeout(() => {
                handleSearch(value, selectedStatus);
            }, 500);
            
            // Clear previous timeout
            return () => clearTimeout(timeoutId);
        } else {
            setSearchResults(null);
        }
    };

    const handleStatusChange = (e) => {
        const status = e.target.value;
        setSelectedStatus(status);
        if (searchQuery.trim()) {
            handleSearch(searchQuery, status);
        }
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        handleSearch();
    };

    return (
        <div className="space-y-4">
            {/* Search Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
                <div className="flex flex-col sm:flex-row gap-2">
                    <div className="flex-1">
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={handleInputChange}
                            placeholder="Search markets by title..."
                            className="w-full px-4 py-2 bg-custom-gray-light border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-pink focus:border-transparent"
                            maxLength={100}
                        />
                    </div>
                    
                    <div className="flex gap-2">
                        <select
                            value={selectedStatus}
                            onChange={handleStatusChange}
                            className="px-3 py-2 bg-custom-gray-light border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-pink focus:border-transparent"
                        >
                            <option value="all">All Status</option>
                            <option value="active">Active Only</option>
                            <option value="closed">Closed Only</option>
                            <option value="resolved">Resolved Only</option>
                        </select>
                        
                        <button
                            type="submit"
                            disabled={!searchQuery.trim() || loading}
                            className="px-4 py-2 bg-primary-pink text-white rounded-lg hover:bg-pink-600 focus:outline-none focus:ring-2 focus:ring-primary-pink focus:ring-offset-2 focus:ring-offset-custom-gray-dark disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {loading ? 'Searching...' : 'Search'}
                        </button>
                    </div>
                </div>
            </form>

            {/* Error Display */}
            {error && (
                <div className="bg-red-900 border border-red-700 text-red-100 px-4 py-3 rounded-lg">
                    {error}
                </div>
            )}

            {/* Loading State */}
            {loading && (
                <div className="text-center py-8">
                    <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-pink"></div>
                    <p className="text-gray-400 mt-2">Searching markets...</p>
                </div>
            )}

            {/* Search Results */}
            {searchResults && !loading && (
                <div className="space-y-4">
                    <div className="text-sm text-gray-400">
                        Search results for "{searchResults.query}" 
                        {searchResults.primaryStatus !== 'all' && ` in ${searchResults.primaryStatus} markets`}
                        {' '}({searchResults.totalCount} result{searchResults.totalCount !== 1 ? 's' : ''})
                    </div>

                    {/* Primary Results */}
                    {searchResults.primaryResults && searchResults.primaryResults.length > 0 && (
                        <div>
                            {searchResults.primaryStatus !== 'all' && (
                                <h3 className="text-lg font-semibold text-gray-300 mb-3 capitalize">
                                    {searchResults.primaryStatus} Markets ({searchResults.primaryCount})
                                </h3>
                            )}
                            <MarketTables markets={searchResults.primaryResults} />
                        </div>
                    )}

                    {/* Fallback Results */}
                    {searchResults.fallbackUsed && searchResults.fallbackResults && searchResults.fallbackResults.length > 0 && (
                        <div>
                            <h3 className="text-lg font-semibold text-gray-300 mb-3">
                                Other Markets ({searchResults.fallbackCount})
                            </h3>
                            <MarketTables markets={searchResults.fallbackResults} />
                        </div>
                    )}

                    {/* No Results */}
                    {searchResults.totalCount === 0 && (
                        <div className="text-center py-8">
                            <p className="text-gray-400">No markets found for "{searchResults.query}"</p>
                            <p className="text-sm text-gray-500 mt-2">Try different keywords or check spelling</p>
                        </div>
                    )}
                </div>
            )}

            {/* Initial State */}
            {!searchQuery && !searchResults && !loading && (
                <div className="text-center py-8">
                    <p className="text-gray-400">Enter a search term to find markets</p>
                </div>
            )}
        </div>
    );
};

export default MarketSearch;
