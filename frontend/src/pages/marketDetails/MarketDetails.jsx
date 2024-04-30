import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { fetchMarketDataHook, calculateCurrentProbability } from './marketDetailsUtils'
import { useAuth } from '../../helpers/AuthContent';


const MarketDetails = () => {
    const { username, isLoggedIn } = useAuth();
    const details = fetchMarketDataHook();
    // check if username is the creator of this market
    console.log("details.creator.username: ", details?.creator?.username)
    console.log("username: ", username)

    const isCreator = username === details?.creator?.username;
    console.log("isCreator: ", isCreator)

    if (!details) {
        return <div>Loading...</div>;
    }

    const currentProbability = calculateCurrentProbability(details);

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
            {isLoggedIn && isCreator && (
                <div className="w-full bg-white p-4 shadow-md fixed inset-x-0 bottom-0">

                </div>
            )}
        </div>
    );
};

export default MarketDetails;