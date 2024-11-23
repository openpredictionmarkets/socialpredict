import React from 'react';
import SiteTabs from './SiteTabs.jsx';
import BuySharesLayout from '../layouts/trade/BuySharesLayout.jsx'
import SellSharesLayout from '../layouts/trade/SellSharesLayout.jsx'

const TradeTabs = ({ marketId, token, onTransactionSuccess }) => {
    const tabsData = [
        {
            label: 'Purchase Shares',
            content: <BuySharesLayout
                marketId={marketId}
                token={token}
                onTransactionSuccess={onTransactionSuccess}
            />
        },
        {
            label: 'Sell Shares',
            content: <SellSharesLayout
                marketId={marketId}
                token={token}
                onTransactionSuccess={onTransactionSuccess}
            />
        },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default TradeTabs;