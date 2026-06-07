import React, { useEffect, useState } from 'react';
import { apiRequest } from '../../../../api/httpClient';
import { useAuth } from '../../../../helpers/AuthContent';
import LoadingSpinner from '../../../loaders/LoadingSpinner';
import MarketLifecycleTable from '../MarketLifecycleTable';

const OwnedMarketsTabContent = ({ username }) => {
    const { isLoggedIn, token } = useAuth();
    const [markets, setMarkets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        let ignore = false;

        const fetchOwnedMarkets = async () => {
            setLoading(true);
            setError(null);
            try {
                const data = await apiRequest(`/v0/users/${username}/owned-markets?limit=50`, {
                    authenticated: true,
                    authToken: token,
                    fallbackMessage: 'Error fetching owned markets',
                });
                if (!ignore) {
                    setMarkets(data.markets || []);
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

    return (
        <MarketLifecycleTable
            markets={markets}
            emptyMessage="No owned markets found for this user."
            showCreator
            showSteward
        />
    );
};

export default OwnedMarketsTabContent;
