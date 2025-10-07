import React from 'react';
import SiteTabs from './SiteTabs';
import BuySharesLayout from '../layouts/trade/BuySharesLayout'
import SellSharesLayout from '../layouts/trade/SellSharesLayout'

const TradeTabs = ({ marketId, market, token, onTransactionSuccess }) => {
    const tabsData = [
        {
            label: 'Purchase Shares',
            content: <BuySharesLayout
                        marketId={marketId}
                        market={market}
                        token={token}
                        onTransactionSuccess={onTransactionSuccess}
                    />
        },
        {
            label: 'Sell Shares',
            content: <SellSharesLayout
                        marketId={marketId}
                        market={market}
                        token={token}
                        onTransactionSuccess={onTransactionSuccess}
                    />
        },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default TradeTabs;