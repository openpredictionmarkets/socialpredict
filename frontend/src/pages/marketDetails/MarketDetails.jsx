import { API_URL } from '../../config';
import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import TestMarketData from '../../tests/TestData'

function MarketDetails() {
    return (
        <div>
            <MarketDetailsTable
                data={TestMarketData.probabilityChanges}
                title={TestMarketData.market.questionTitle}
                market={TestMarketData.market}
                creator={TestMarketData.creator}
                probabilityChanges={TestMarketData.probabilityChanges}
                numUsers={TestMarketData.numUsers}
                totalVolume={TestMarketData.totalVolume}
            />
        </div>
    );
}

export default MarketDetails;