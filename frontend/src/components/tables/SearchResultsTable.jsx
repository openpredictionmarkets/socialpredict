import React from 'react';
import MarketTables from './MarketTables';

const SearchResultsTable = ({ searchResults }) => {
    if (!searchResults) {
        return null;
    }

    if (searchResults.error) {
        return (
            <div className="text-center py-8">
                <p className="text-red-400">{searchResults.error}</p>
            </div>
        );
    }

    if (searchResults.totalCount === 0) {
        return (
            <div className="text-center py-8">
                <p className="text-gray-400">No markets found for "{searchResults.query}"</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Primary Results */}
            {searchResults.primaryCount > 0 && (
                <div>
                    <div className="mb-3">
                        <h3 className="text-lg font-medium text-white">
                            {searchResults.primaryStatus === 'all' 
                                ? `All Markets` 
                                : `${searchResults.primaryStatus.charAt(0).toUpperCase() + searchResults.primaryStatus.slice(1)} Markets`
                            } ({searchResults.primaryCount})
                        </h3>
                        <p className="text-sm text-gray-400">
                            Results from {searchResults.primaryStatus} markets matching "{searchResults.query}"
                        </p>
                    </div>
                    <MarketTables markets={searchResults.primaryResults} />
                </div>
            )}

            {/* Fallback Results */}
            {searchResults.fallbackUsed && searchResults.fallbackCount > 0 && (
                <div>
                    {/* Horizontal Rule */}
                    <div className="flex items-center py-4">
                        <div className="flex-grow border-t border-gray-600"></div>
                        <div className="mx-4 text-sm text-gray-400 font-medium">
                            More Results from Other Categories
                        </div>
                        <div className="flex-grow border-t border-gray-600"></div>
                    </div>

                    <div className="mb-3">
                        <h3 className="text-lg font-medium text-gray-300">
                            Additional Markets ({searchResults.fallbackCount})
                        </h3>
                        <p className="text-sm text-gray-400">
                            Other markets matching "{searchResults.query}"
                        </p>
                    </div>
                    <MarketTables markets={searchResults.fallbackResults} />
                </div>
            )}

            {/* Summary */}
            <div className="text-center py-4 border-t border-gray-700">
                <p className="text-sm text-gray-400">
                    Found {searchResults.totalCount} markets matching "{searchResults.query}"
                    {searchResults.fallbackUsed && 
                        ` (${searchResults.primaryCount} in ${searchResults.primaryStatus}, ${searchResults.fallbackCount} in other categories)`
                    }
                </p>
            </div>
        </div>
    );
};

export default SearchResultsTable;
