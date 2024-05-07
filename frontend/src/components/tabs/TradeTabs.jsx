import React from 'react';
import SiteTabs from './SiteTabs';

const TradeTabs = ({ marketId }) => {
    const tabsData = [
        { label: 'Purchase Shares', content: <div>Buying Layout</div> },
        { label: 'Sell Shares', content: <div>Selling Layout</div> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default TradeTabs;