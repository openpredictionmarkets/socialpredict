import React from 'react';
import SiteTabs from './SiteTabs';
import BetsActivityLayout from '../layouts/activity/bets/BetsActivity';
import PositionsActivityLayout from '../layouts/activity/positions/PositionsActivity';

const ActivityTabs = ({ marketId }) => {
    const tabsData = [
        { label: 'Positions', content: <PositionsActivityLayout marketId={marketId} /> },
        { label: 'Bets', content: <BetsActivityLayout marketId={marketId} /> },
        { label: 'Comments', content: <div>Comments Go here...</div> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;
