import React, { useState, useEffect, useRef } from 'react';
import { API_URL } from '../../config';
import SiteTabs from './SiteTabs';
import BetsActivityLayout from '../layouts/activity/bets/BetsActivity';
import PositionsActivityLayout from '../layouts/activity/positions/PositionsActivity';
import LeaderboardActivity from '../layouts/activity/leaderboard/LeaderboardActivity';
import MarketComments from '../comments/MarketComments';

const ActivityTabs = ({ marketId, market, refreshTrigger, isLoggedIn, token }) => {
    const [bets, setBets] = useState(null); // null = not yet fetched
    const [betsLoading, setBetsLoading] = useState(false);
    const [betsError, setBetsError] = useState(null);
    const lastRefreshRef = useRef(refreshTrigger);

    const fetchBets = async () => {
        setBetsLoading(true);
        setBetsError(null);
        try {
            const response = await fetch(`${API_URL}/v0/markets/bets/${marketId}`);
            if (response.ok) {
                const data = await response.json();
                setBets(data.sort((a, b) => new Date(b.placedAt) - new Date(a.placedAt)));
            } else {
                setBetsError('Failed to load bets');
            }
        } catch {
            setBetsError('Failed to load bets');
        } finally {
            setBetsLoading(false);
        }
    };

    // Re-fetch when refreshTrigger changes (e.g. after a new bet is placed)
    useEffect(() => {
        if (bets !== null && refreshTrigger !== lastRefreshRef.current) {
            lastRefreshRef.current = refreshTrigger;
            fetchBets();
        }
    }, [refreshTrigger]);

    const handleBetsTabSelect = () => {
        // Fetch on first open, or if refreshTrigger advanced since last fetch
        if (bets === null || refreshTrigger !== lastRefreshRef.current) {
            lastRefreshRef.current = refreshTrigger;
            fetchBets();
        }
    };

    const tabsData = [
        {
            label: 'Positions',
            content: <PositionsActivityLayout marketId={marketId} market={market} refreshTrigger={refreshTrigger} />,
        },
        {
            label: 'Bets',
            content: <BetsActivityLayout bets={bets} loading={betsLoading} error={betsError} />,
            onSelect: handleBetsTabSelect,
        },
        {
            label: 'Leaderboard',
            content: <LeaderboardActivity marketId={marketId} market={market} />,
        },
        {
            label: 'Comments',
            content: <MarketComments marketId={marketId} isLoggedIn={isLoggedIn} token={token} />,
        },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;
