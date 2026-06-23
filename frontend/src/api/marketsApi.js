import { API_URL } from '../config';
import { apiRequest, authenticatedApiRequest } from './httpClient';

/**
 * Search markets by query and status
 * @param {string} query - Search query
 * @param {string} status - Market status filter ('all', 'active', 'closed', 'resolved')
 * @param {number} limit - Maximum number of results (optional, defaults to 20)
 * @returns {Promise<Object>} Search results object
 */
export const searchMarkets = async (query, status = 'all', limit = 20, options = {}) => {
    if (!query?.trim()) {
        throw new Error('Search query is required');
    }

    const params = new URLSearchParams({
        query: query.trim(),
        status: status,
        limit: limit.toString()
    });
    if (options.tagSlug) {
        params.set('tagSlug', options.tagSlug);
    }

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

export const createMarketGroup = async (marketGroupData) => {
    return authenticatedApiRequest('/v0/market-groups', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(marketGroupData),
        reasonMessages: {
            USER_NOT_APPROVED: 'User does not have approval to create markets in moderator mode.',
            AUTHORIZATION_DENIED: 'You are not allowed to create this market group.',
            INSUFFICIENT_BALANCE: 'You do not have enough credit to create this market group.',
            VALIDATION_FAILED: 'Check the market group fields and try again.',
            INVALID_REQUEST: 'Check the market group fields and try again.',
        },
        fallbackMessage: 'Market group creation failed. Please try again.',
    });
};

export const getMarketGroupDetails = async (groupId) => {
    if (!groupId) {
        throw new Error('Market group ID is required');
    }

    return apiRequest(`/v0/market-groups/${groupId}`, {
        reasonMessages: {
            RATE_LIMITED: 'Grouped market details are loading. Wait a moment and try again.',
        },
        fallbackMessage: 'Failed to fetch market group details.',
    });
};

export const proposeMarketGroupAnswerAddition = async ({ groupId, token, answerLabel }) => {
    const normalizedAnswerLabel = String(answerLabel || '').trim();
    if (!groupId) {
        throw new Error('Market group ID is required.');
    }
    if (!normalizedAnswerLabel) {
        throw new Error('Answer label is required.');
    }

    return authenticatedApiRequest(`/v0/market-groups/${groupId}/answers`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        authToken: token,
        body: JSON.stringify({ answerLabel: normalizedAnswerLabel }),
        reasonMessages: {
            AUTHORIZATION_DENIED: 'Only active moderators can propose grouped-market answers.',
            INSUFFICIENT_BALANCE: 'You do not have enough credit to add this answer.',
            INVALID_STATE: 'This grouped market is not accepting new answers.',
            VALIDATION_FAILED: 'Check the answer label and try again.',
            MARKET_NOT_FOUND: 'No market group was found for that ID.',
            RATE_LIMITED: 'Too many answer requests. Wait and try again.',
        },
        fallbackMessage: 'Answer addition request failed. Please try again.',
    });
};

export const listMarketGroupAnswerAdditionsForReview = async ({
    groupId = '',
    token,
    status = 'pending',
    limit = 50,
    offset = 0,
} = {}) => {
    const params = new URLSearchParams({
        status,
        limit: String(limit),
        offset: String(offset),
    });
    const path = groupId
        ? `/v0/market-groups/${groupId}/answer-additions?${params.toString()}`
        : `/v0/profile/market-group-answer-additions?${params.toString()}`;

    return authenticatedApiRequest(path, {
        authToken: token,
        reasonMessages: {
            AUTHORIZATION_DENIED: 'Only the grouped market steward can review these answer options.',
            RATE_LIMITED: 'Too many answer review requests. Wait and try again.',
        },
        fallbackMessage: 'Unable to load grouped answer options.',
    });
};

export const reviewMarketGroupAnswerAdditionForSteward = async ({
    additionId,
    token,
    status,
    reason = '',
    confirm = false,
}) => {
    const normalizedAdditionId = String(additionId || '').trim();
    const normalizedStatus = String(status || '').trim();
    if (!normalizedAdditionId || !normalizedStatus) {
        throw new Error('Answer addition ID and review status are required.');
    }

    return authenticatedApiRequest(`/v0/profile/market-group-answer-additions/${normalizedAdditionId}`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        authToken: token,
        body: JSON.stringify({
            status: normalizedStatus,
            reason: String(reason || '').trim(),
            confirm: Boolean(confirm),
        }),
        reasonMessages: {
            AUTHORIZATION_DENIED: 'Only the grouped market steward can review this answer option.',
            INSUFFICIENT_BALANCE: 'The proposing moderator no longer has enough credit to add this answer.',
            INVALID_STATE: 'This answer option has already been reviewed or can no longer be added.',
            RATE_LIMITED: 'Too many answer review requests. Wait and try again.',
        },
        fallbackMessage: 'Unable to review grouped answer option.',
    });
};

export const updateMarketGroupAnswerAdditionSettings = async ({
    groupId,
    token,
    autoApproveAnswerAdditions,
}) => {
    if (!groupId) {
        throw new Error('Market group ID is required.');
    }

    return authenticatedApiRequest(`/v0/market-groups/${groupId}/answer-addition-settings`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        authToken: token,
        body: JSON.stringify({
            autoApproveAnswerAdditions: Boolean(autoApproveAnswerAdditions),
        }),
        reasonMessages: {
            AUTHORIZATION_DENIED: 'Only the grouped market steward can change this setting.',
            INVALID_STATE: 'This grouped market can no longer accept answer option changes.',
            RATE_LIMITED: 'Too many answer setting requests. Wait and try again.',
        },
        fallbackMessage: 'Unable to update grouped answer option settings.',
    });
};

export const getMarketSummaryReadModel = async (marketId) => {
    if (!marketId) {
        throw new Error('Market ID is required');
    }

    return apiRequest(`/v0/read/markets/${marketId}/summary`, {
        fallbackMessage: 'Failed to fetch market summary.',
    });
};
