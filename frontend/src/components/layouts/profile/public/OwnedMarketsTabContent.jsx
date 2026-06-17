import React, { useEffect, useState } from 'react';
import { apiRequest } from '../../../../api/httpClient';
import { useAuth } from '../../../../helpers/AuthContent';
import LoadingSpinner from '../../../loaders/LoadingSpinner';
import MarketLifecycleTable, { groupLifecycleMarketRows } from '../MarketLifecycleTable';

const PAGE_SIZE = 20;
const FETCH_BATCH_SIZE = 100;
const paginationButtonClass = [
    'rounded',
    'border',
    'border-transparent',
    'bg-neutral-btn',
    'px-3',
    'py-1.5',
    'text-xs',
    'font-semibold',
    'text-white',
    'transition-colors',
    'duration-200',
    'hover:bg-neutral-btn-hover',
    'disabled:cursor-not-allowed',
    'disabled:bg-custom-gray-light',
    'disabled:text-gray-400',
    'disabled:opacity-60',
].join(' ');

const OwnedMarketsTabContent = ({ username }) => {
    const { isLoggedIn, token } = useAuth();
    const [markets, setMarkets] = useState([]);
    const [page, setPage] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        setPage(0);
    }, [username]);

    useEffect(() => {
        let ignore = false;

        const fetchOwnedMarkets = async () => {
            setLoading(true);
            setError(null);
            try {
                const rows = [];
                let offset = 0;
                let keepFetching = true;
                while (keepFetching) {
                    const params = new URLSearchParams({
                        limit: String(FETCH_BATCH_SIZE),
                        offset: String(offset),
                    });
                    const data = await apiRequest(`/v0/users/${username}/owned-markets?${params.toString()}`, {
                        authenticated: true,
                        authToken: token,
                        fallbackMessage: 'Error fetching owned markets',
                    });
                    const batch = data.markets || [];
                    rows.push(...batch);
                    keepFetching = batch.length === FETCH_BATCH_SIZE;
                    offset += FETCH_BATCH_SIZE;
                }
                if (!ignore) {
                    setMarkets(groupLifecycleMarketRows(rows));
                }
            } catch (err) {
                if (!ignore) {
                    setError(err.message);
                }
            } finally {
                if (!ignore) {
                    setLoading(false);
                }
            }
        };

        if (username && token) {
            fetchOwnedMarkets();
        } else {
            setMarkets([]);
            setError(null);
            setLoading(false);
        }

        return () => {
            ignore = true;
        };
    }, [username, token]);

    if (loading) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="flex items-center justify-center">
                    <LoadingSpinner />
                    <span className="ml-2 text-gray-300">Loading owned markets...</span>
                </div>
            </div>
        );
    }

    if (!isLoggedIn) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-center">
                    <div className="text-xl font-semibold text-blue-100">
                        Log in to see owned markets
                    </div>
                    <div className="text-sm text-gray-400 mt-2">
                        Created and stewarded markets are visible to logged-in players.
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-center text-red-400">
                    Error loading owned markets: {error}
                </div>
            </div>
        );
    }

    const start = page * PAGE_SIZE;
    const visibleMarkets = markets.slice(start, start + PAGE_SIZE);
    const hasNextPage = start + PAGE_SIZE < markets.length;

    return (
        <div className="space-y-3">
            <div className="flex flex-col gap-2 rounded-lg border border-gray-700 bg-gray-900/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="text-xs uppercase tracking-[0.16em] text-gray-400">
                    Showing owned markets page {page + 1}{visibleMarkets.length ? ` (${start + 1}-${start + visibleMarkets.length} of ${markets.length})` : ''}
                </div>
                <div className="flex gap-2">
                    <button
                        type="button"
                        onClick={() => setPage((current) => Math.max(0, current - 1))}
                        disabled={page === 0 || loading}
                        className={paginationButtonClass}
                    >
                        Previous
                    </button>
                    <button
                        type="button"
                        onClick={() => setPage((current) => current + 1)}
                        disabled={!hasNextPage || loading}
                        className={paginationButtonClass}
                    >
                        Next
                    </button>
                </div>
            </div>
            <MarketLifecycleTable
                markets={visibleMarkets}
                emptyMessage="No owned markets found for this user."
                showCreator
                showSteward
            />
        </div>
    );
};

export default OwnedMarketsTabContent;
