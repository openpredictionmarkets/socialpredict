import React, { useEffect, useState } from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { useFetchMarketData, calculateCurrentProbability } from './marketDetailsUtils'
import { useAuth } from '../../helpers/AuthContent';
import ResolveModalButton from '../../components/modals/resolution/ResolveModal'
import BetModalButton from '../../components/modals/bet/BetModal'
import ActivityTabs from '../../components/tabs/ActivityTabs';


const MarketDetails = () => {
    const { username } = useAuth();
    const [token, setToken] = useState(null);
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const { details, refetchData } = useFetchMarketData();

    // check if username is the creator of this market
    const isCreator = username === details?.creator?.username;
    // check if market is resolved
    const isResolved = details?.market?.isResolved === true;

    useEffect(() => {
        const fetchedToken = localStorage.getItem('token');
        setToken(fetchedToken);
        setIsLoggedIn(!!fetchedToken);
    }, []);

    if (!details) {
        return <div>Loading...</div>;
    }

    const currentProbability = calculateCurrentProbability(details);

    return (
        <div className="flex-col min-h-screen">
            <div className="flex-grow flex">
                <div className="flex-1">
                    <MarketDetailsTable
                        market={details.market}
                        creator={details.creator}
                        numUsers={details.numUsers}
                        totalVolume={details.totalVolume}
                        currentProbability={currentProbability}
                        probabilityChanges={details.probabilityChanges}
                    />
                </div>
            </div>
            <div className="flex items-center space-x-14" style={{ width: '110%' }}>
                {isCreator && !isResolved && (
                    <div className="flex-none ml-6 mr-6" style={{ width: '10%' }}>
                        <ResolveModalButton
                            marketId={details.market.id}
                            token={token}
                            disabled={!token}
                            className="text-xs px-2 py-1" />
                    </div>
                )}
                {!isResolved && isLoggedIn && (
                    <div className="flex-none ml-6 mr-6" style={{ width: '10%' }}>
                        <BetModalButton
                            marketId={details.market.id}
                            token={token}
                            disabled={!token}
                            onTransactionSuccess={refetchData}
                            className="text-xs px-2 py-1" />
                    </div>
                )}
            </div>
            <div className="max-w-4xl mx-auto mt-8">
                <ActivityTabs marketId={details.market.id} />
            </div>
        </div>
    );
};

export default MarketDetails;