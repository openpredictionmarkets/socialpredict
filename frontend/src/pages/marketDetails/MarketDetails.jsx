import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { fetchMarketDataHook, calculateCurrentProbability } from './marketDetailsUtils'
import { useAuth } from '../../helpers/AuthContent';


const MarketDetails = () => {
    const { username } = useAuth();
    const details = fetchMarketDataHook();

    console.log("username: ", username)

    if (!details) {
        return <div>Loading...</div>;
    }

    const currentProbability = calculateCurrentProbability(details);

    // Check if the logged-in user is the creator
    const isCreator = username === details.creator.username;

    return (
        <div className="flex flex-col min-h-screen">
            <div className="flex-grow">
                <MarketDetailsTable
                    market={details.market}
                    creator={details.creator}
                    numUsers={details.numUsers}
                    totalVolume={details.totalVolume}
                    currentProbability={currentProbability}
                    probabilityChanges={details.probabilityChanges}
                />
            </div>
            {isCreator && (
                <div className="w-full bg-white p-4 shadow-md fixed inset-x-0 bottom-0">
                    <NeutralButton label="Resolve Market" />
                </div>
            )}
        </div>
    );
};

export default MarketDetails;