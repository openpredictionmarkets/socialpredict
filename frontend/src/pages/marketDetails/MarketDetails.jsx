import React from 'react';
import { useParams } from 'react-router-dom';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { fetchMarketDataHook } from './marketDetailsUtils'
import { useAuth } from '../../helpers/AuthContent';


const MarketDetails = () => {
    const { username } = useAuth();
    const details = fetchMarketDataHook();

    console.log("username: ", username)

    if (!details) {
        return <div>Loading...</div>;
    }

    // Check if the logged-in user is the creator
    const isCreator = username === details.creator.username;

    return (
        <div className="flex flex-col min-h-screen">
            <div className="flex-grow">
                <MarketDetailsTable
                    market={details.market}
                    creator={details.creator}
                    probabilityChanges={details.probabilityChanges}
                    numUsers={details.numUsers}
                    totalVolume={details.totalVolume}
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