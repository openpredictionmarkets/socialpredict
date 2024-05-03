import React, { useEffect, useState } from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { fetchMarketDataHook, calculateCurrentProbability } from './marketDetailsUtils'
import { useAuth } from '../../helpers/AuthContent';
import ResolveModalButton from '../../components/modals/resolution/ResolveModal'

const MarketDetails = () => {
    const { username } = useAuth();
    const [token, setToken] = useState(null);
    const details = fetchMarketDataHook();

    // check if username is the creator of this market
    const isCreator = username === details?.creator?.username;

    useEffect(() => {
        setToken(localStorage.getItem('token'));
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
            {isCreator && (
                <div className="flex-none ml-6" style={{ width: '10%' }}>
                    <ResolveModalButton marketId={details.market.id} token={token} className="text-xs px-2 py-1" />
                </div>
            )}
        </div>
    );
};

export default MarketDetails;