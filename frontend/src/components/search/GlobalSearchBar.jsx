import React, { useState, useEffect } from 'react';
import { searchMarkets } from '../../api/marketsApi';
import { RegularInput } from '../inputs/InputBar';

const GlobalSearchBar = ({ onSearchResults, currentStatus, isSearching, setIsSearching }) => {
    const [query, setQuery] = useState('');
    const [loading, setLoading] = useState(false);

    // Debounce search queries - trigger on query OR status change
    useEffect(() => {
        const timeoutId = setTimeout(() => {
            if (query.trim().length > 0) {
                performSearch(query.trim());
            } else {
                // Clear search when query is empty
                onSearchResults(null);
                setIsSearching(false);
            }
        }, 300); // 300ms debounce

        return () => clearTimeout(timeoutId);
    }, [query, currentStatus]);

    const performSearch = async (searchQuery) => {
        setLoading(true);
        setIsSearching(true);
        
        try {
            const results = await searchMarkets(searchQuery, currentStatus, 20);
            onSearchResults(results);
        } catch (error) {
            console.error('Search error:', error);
            onSearchResults({ error: 'Search failed. Please try again.' });
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e) => {
        setQuery(e.target.value);
    };

    const handleClear = () => {
        setQuery('');
        onSearchResults(null);
        setIsSearching(false);
    };

    return (
        <div className="mb-4 w-full">
            <div className="relative">
                <RegularInput
                    type="text"
                    value={query}
                    onChange={handleInputChange}
                    placeholder={`Search ${currentStatus} markets...`}
                    className="w-full pr-10"
                />
                {query && (
                    <button
                        onClick={handleClear}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white"
                        title="Clear search"
                    >
                        ✕
                    </button>
                )}
                {loading && (
                    <div className="absolute right-10 top-1/2 transform -translate-y-1/2">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-pink"></div>
                    </div>
                )}
            </div>
            {isSearching && query && (
                <div className="mt-2 text-sm text-gray-400">
                    Searching in <span className="text-primary-pink font-medium">{currentStatus}</span> markets for: 
                    <span className="text-white font-medium"> "{query}"</span>
                </div>
            )}
        </div>
    );
};

export default GlobalSearchBar;
