import React from 'react';
import SiteTabs from './SiteTabs';
import BuySharesLayout from '../layouts/trade/BuySharesLayout'
import SellSharesLayout from '../layouts/trade/SellSharesLayout'

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