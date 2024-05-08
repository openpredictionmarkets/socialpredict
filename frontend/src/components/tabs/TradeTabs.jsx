import React from 'react';
import SiteTabs from './SiteTabs';
import BuySharesLayout from '../layouts/trade/BuySharesLayout'
import SellSharesLayout from '../layouts/trade/SellSharesLayout'

const TradeTabs = ({ marketId, token, onBetSuccess, onSaleSuccess }) => {
    const tabsData = [
        {
            label: 'Purchase Shares',
            content: <BuySharesLayout
                        marketId={marketId}
                        token={token}
                        onBetSuccess={onBetSuccess}
                    />
        },
        {
            label: 'Sell Shares',
            content: <SellSharesLayout
                        marketId={marketId}
                        token={token}
                        onSaleSuccess={onSaleSuccess}
                    />
        },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default TradeTabs;