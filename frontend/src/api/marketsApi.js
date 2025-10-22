import { API_URL } from '../config';

/**
 * Search markets by query and status
 * @param {string} query - Search query
 * @param {string} status - Market status filter ('all', 'active', 'closed', 'resolved')
 * @param {number} limit - Maximum number of results (optional, defaults to 20)
 * @returns {Promise<Object>} Search results object
 */
export const searchMarkets = async (query, status = 'all', limit = 20) => {
    if (!query?.trim()) {
        throw new Error('Search query is required');
    }

    const params = new URLSearchParams({
        query: query.trim(),
        status: status,
        limit: limit.toString()
    });

    const response = await fetch(`${API_URL}/v0/markets/search?${params}`);

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Search failed: ${response.status} ${errorText}`);
    }

    return await response.json();
};

/**
 * Get markets by status (existing endpoint)
 * @param {string} status - Market status ('active', 'closed', 'resolved', 'all')
 * @returns {Promise<Object>} Markets data
 */
export const getMarketsByStatus = async (status = 'all') => {
    const endpoint = status === 'all'
        ? `${API_URL}/v0/markets`
        : `${API_URL}/v0/markets/${status}`;

    const response = await fetch(endpoint);

    if (!response.ok) {
        throw new Error(`Failed to fetch ${status} markets: ${response.status}`);
    }

    return await response.json();
};

/**
 * Get market details by ID
 * @param {string|number} marketId - Market ID
 * @returns {Promise<Object>} Market details
 */
export const getMarketDetails = async (marketId) => {
    if (!marketId) {
        throw new Error('Market ID is required');
    }

    const response = await fetch(`${API_URL}/v0/markets/${marketId}`);

    if (!response.ok) {
        throw new Error(`Failed to fetch market details: ${response.status}`);
    }

    return await response.json();
};
