import React from 'react';
import SiteTabs from './SiteTabs';
import BuySharesLayout from '../layouts/trade/BuySharesLayout'

const TradeTabs = ({ marketId, token, onBetSuccess }) => {
    const tabsData = [
        {
            label: 'Purchase Shares',
            content: <BuySharesLayout
                        marketId={marketId}
                        token={token}
                        onBetSuccess={onBetSuccess}
                    />
        },
        { label: 'Sell Shares', content: <div>Selling Layout</div> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default TradeTabs;