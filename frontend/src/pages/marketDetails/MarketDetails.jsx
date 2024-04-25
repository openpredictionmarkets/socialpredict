import { API_URL } from '../../config';
import React from 'react';
import MarketChart from '../../components/charts/MarketChart';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import TestMarketData from '../../tests/TestData'

function MarketDetails() {
    return (
        <div>
            <MarketDetailsTable
                data={TestMarketData.probabilityChanges}
                title={TestMarketData.market.questionTitle}
                className="shadow-md border border-custom-gray-light"
            />
        </div>
    );

}

export default MarketDetails;